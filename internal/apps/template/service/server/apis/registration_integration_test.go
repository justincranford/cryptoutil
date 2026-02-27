// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration

package apis

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"os"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	// Use modernc SQLite driver (CGO-free).
	_ "modernc.org/sqlite"
)

const testIntegrationPassword = "SecurePass123!"

var (
	testDB                   *gorm.DB
	testRegistrationSvc      *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService
	testSessionManager       *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	testRegistrationApp      *fiber.App
	testJoinRequestMgmtApp   *fiber.App
	testTenantID             googleUuid.UUID
	testUserID               googleUuid.UUID
	testMockSessionValidator *mockSessionValidatorIntegration
)

// mockSessionValidatorIntegration is a mock SessionValidator for integration tests.
// It bypasses actual session validation and returns predefined tenant/user IDs.
type mockSessionValidatorIntegration struct {
	tenantID googleUuid.UUID
	realmID  googleUuid.UUID
	userID   string
}

func (m *mockSessionValidatorIntegration) ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error) {
	// Return a mock session with predefined tenant_id and user_id.
	// Note: BrowserSession embeds Session (which has TenantID/RealmID as UUID)
	// and adds UserID as *string.
	return &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
		Session: cryptoutilAppsTemplateServiceServerRepository.Session{
			TenantID: m.tenantID,
			RealmID:  m.realmID,
		},
		UserID: &m.userID,
	}, nil
}

func (m *mockSessionValidatorIntegration) ValidateServiceSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error) {
	// Return a mock session with predefined tenant_id.
	// Note: ServiceSession embeds Session (which has TenantID/RealmID as UUID)
	// and adds ClientID as *string.
	return &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
		Session: cryptoutilAppsTemplateServiceServerRepository.Session{
			TenantID: m.tenantID,
			RealmID:  m.realmID,
		},
		ClientID: &m.userID, // Use userID as clientID for simplicity
	}, nil
}

// addAuthHeader adds a mock Bearer token to the request for testing.
// The mockSessionValidatorIntegration will accept any token and return a valid session.
func addAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer test-mock-token")
}

func TestMain(m *testing.M) {
	// Use SQLite with modernc driver (CGO-free).
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:        cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}, &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to open SQLite: %v", err))
	}

	testDB = db

	// Get underlying SQL DB for configuration.
	sqlDB, err := testDB.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get SQL DB: %v", err))
	}

	// Configure SQLite for concurrent operations.
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		panic(fmt.Sprintf("failed to enable WAL: %v", err))
	}

	if _, err := sqlDB.Exec("PRAGMA busy_timeout = 30000;"); err != nil {
		panic(fmt.Sprintf("failed to set busy timeout: %v", err))
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetConnMaxLifetime(0)

	// Create barrier tables (before GORM migrations).
	barrierSchema := `
	CREATE TABLE IF NOT EXISTS barrier_root_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS barrier_intermediate_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (kek_uuid) REFERENCES barrier_root_keys(uuid)
	);

	CREATE TABLE IF NOT EXISTS barrier_content_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (kek_uuid) REFERENCES barrier_intermediate_keys(uuid)
	);
	`

	if _, err := sqlDB.Exec(barrierSchema); err != nil {
		panic(fmt.Sprintf("failed to create barrier tables: %v", err))
	}

	// Run migrations including session tables.
	if err := testDB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerRepository.TenantRealm{},
		&cryptoutilAppsTemplateServiceServerRepository.User{},
		&cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate: %v", err))
	}

	// Create repositories.
	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	// Create service.
	testRegistrationSvc = cryptoutilAppsTemplateServiceServerBusinesslogic.NewTenantRegistrationService(
		testDB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)

	// Create minimal session manager for integration tests.
	// For full-featured tests, see server_builder.go which sets up all dependencies.
	// Here we create minimal infrastructure for session testing.
	ctx := context.Background()

	// Create telemetry service (minimal - no OTLP export).
	testConfig := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, testConfig.ToTelemetrySettings())
	if err != nil {
		panic(fmt.Sprintf("failed to create telemetry service: %v", err))
	}
	defer telemetryService.Shutdown()

	// Create JWK generation service.
	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		panic(fmt.Sprintf("failed to create JWK generation service: %v", err))
	}
	defer jwkGenService.Shutdown()

	// Create barrier repository and service for session encryption.
	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(testDB)
	if err != nil {
		panic(fmt.Sprintf("failed to create barrier repository: %v", err))
	}

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic(fmt.Sprintf("failed to generate unseal JWK: %v", err))
	}

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	if err != nil {
		panic(fmt.Sprintf("failed to create unseal service: %v", err))
	}
	defer unsealService.Shutdown()

	// Create barrier service.
	barrierService, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(
		ctx,
		telemetryService,
		jwkGenService,
		barrierRepo,
		unsealService,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create barrier service: %v", err))
	}

	// Create session manager.
	testSessionManager, err = cryptoutilAppsTemplateServiceServerBusinesslogic.NewSessionManagerService(
		ctx,
		testDB,
		telemetryService,
		jwkGenService,
		barrierService,
		testConfig,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create session manager: %v", err))
	}

	// Create Fiber apps for testing.
	testRegistrationApp = fiber.New()

	// Create testJoinRequestMgmtApp with custom error handler for apperr.Error types.
	// This ensures SessionMiddleware's 401 errors are correctly converted to HTTP 401 responses.
	testJoinRequestMgmtApp = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *cryptoutilSharedApperr.Error
			if errors.As(err, &appErr) {
				return c.Status(int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode)).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: appErr.Summary,
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: err.Error(),
			})
		},
	})

	// Create test authentication middleware that sets tenant_id and user_id.
	// In production, this would be set by JWT/session validation middleware.
	testTenantID = googleUuid.New()
	testUserID = googleUuid.New()
	testRealmID := googleUuid.New() // Create realm ID for mock validator
	authMiddleware := func(c *fiber.Ctx) error {
		c.Locals("tenant_id", testTenantID)
		c.Locals("user_id", testUserID)

		return c.Next()
	}
	testJoinRequestMgmtApp.Use(authMiddleware)

	// Create mock session validator for testing that bypasses real session validation.
	// This allows tests to use simple Authorization headers without actual session tokens.
	testMockSessionValidator = &mockSessionValidatorIntegration{
		tenantID: testTenantID,
		realmID:  testRealmID,
		userID:   testUserID.String(), // Convert UUID to string
	}

	// Register routes.
	RegisterRegistrationRoutes(testRegistrationApp, testRegistrationSvc, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	RegisterJoinRequestManagementRoutes(testJoinRequestMgmtApp, testRegistrationSvc, testMockSessionValidator)

	// Run tests.
	exitCode := m.Run()

	// Cleanup.
	os.Exit(exitCode)
}

func TestIntegration_RegisterUser_CreateTenant(t *testing.T) {
	t.Parallel()

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])
	password := testIntegrationPassword
	tenantName := fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	reqBody := RegisterUserRequest{
		Username:     username,
		Password:     password,
		Email:        fmt.Sprintf("%s@example.com", username),
		TenantName:   tenantName,
		CreateTenant: true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testRegistrationApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "tenant_id")
	require.Contains(t, result, "user_id")
	require.Contains(t, result, "message")
}

func TestIntegration_RegisterUser_JoinExistingTenant(t *testing.T) {
	t.Skip("Join existing tenant flow not yet implemented in service")
	t.Parallel()

	ctx := context.Background()

	// Create tenant first.
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create realm.
	realm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		RealmID:  googleUuid.New(),
		Type:     cryptoutilSharedMagic.AuthMethodUsernamePassword,
		Active:   true,
		Source:   "db",
	}
	require.NoError(t, testDB.Create(realm).Error)

	// Register user to join existing tenant.
	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])
	password := testIntegrationPassword

	reqBody := RegisterUserRequest{
		Username:     username,
		Password:     password,
		Email:        fmt.Sprintf("%s@example.com", username),
		CreateTenant: false,
		TenantName:   tenant.Name,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testRegistrationApp.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Read response body for debugging.
	bodyBytes, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)

	if resp.StatusCode != http.StatusOK {
		t.Logf("Response status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.Unmarshal(bodyBytes, &result))
	require.Contains(t, result, "message")
	require.Contains(t, result["message"], "pending")

	// Verify join request created.
	var joinRequests []cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).Where("tenant_id = ?", tenant.ID).Find(&joinRequests).Error)
	require.GreaterOrEqual(t, len(joinRequests), 1)
	require.Equal(t, "pending", joinRequests[0].Status)
}

func TestIntegration_RateLimiting_ExceedsLimit(t *testing.T) {
	t.Parallel()

	// Create separate Fiber app with low rate limit for testing.
	app := fiber.New()

	// Create custom rate limiter with 3 requests/min and burst 3 (so exactly 3 requests allowed).
	rateLimiter := NewRateLimiter(3, 3)

	// Rate limit middleware.
	rateLimitMiddleware := func(c *fiber.Ctx) error {
		ipAddress := c.IP()
		if !rateLimiter.Allow(ipAddress) {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: "Rate limit exceeded. Please try again later.",
			})
		}

		return c.Next()
	}

	// Create handlers and register with custom middleware.
	handlers := NewRegistrationHandlers(testRegistrationSvc)
	app.Post("/browser/api/v1/auth/register", rateLimitMiddleware, handlers.HandleRegisterUser)

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength])

	// Make 4 requests quickly (should exceed limit).
	for i := 0; i < 4; i++ {
		reqBody := RegisterUserRequest{
			Username:     fmt.Sprintf("%s_%d", username, i),
			Password:     testIntegrationPassword,
			Email:        fmt.Sprintf("%s_%d@example.com", username, i),
			TenantName:   fmt.Sprintf("tenant_%d", i),
			CreateTenant: true,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.100") // Same IP for rate limiting

		resp, err := app.Test(req, -1)
		require.NoError(t, err)

		_ = resp.Body.Close()

		if i < 3 {
			require.Equal(t, http.StatusCreated, resp.StatusCode, "Request %d should succeed", i+1)
		} else {
			require.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Request %d should be rate limited", i+1)
		}
	}
}
