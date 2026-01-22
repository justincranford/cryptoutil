// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration

package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	// Use modernc SQLite driver (CGO-free).
	_ "modernc.org/sqlite"
)

var (
	testDB                 *gorm.DB
	testRegistrationSvc    *cryptoutilTemplateBusinessLogic.TenantRegistrationService
	testSessionManager     *cryptoutilTemplateBusinessLogic.SessionManagerService
	testRegistrationApp    *fiber.App
	testJoinRequestMgmtApp *fiber.App
	testTenantID           googleUuid.UUID
	testUserID             googleUuid.UUID
)

func TestMain(m *testing.M) {
	// Use SQLite with modernc driver (CGO-free).
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
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

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
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
		&cryptoutilTemplateRepository.Tenant{},
		&cryptoutilTemplateRepository.TenantRealm{},
		&cryptoutilTemplateRepository.User{},
		&cryptoutilTemplateDomain.TenantJoinRequest{},
		&cryptoutilTemplateRepository.BrowserSession{},
		&cryptoutilTemplateRepository.ServiceSession{},
		&cryptoutilTemplateRepository.BrowserSessionJWK{},
		&cryptoutilTemplateRepository.ServiceSessionJWK{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate: %v", err))
	}

	// Create repositories.
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(testDB)

	// Create service.
	testRegistrationSvc = cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(
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
	testConfig := cryptoutilConfig.NewTestConfig("127.0.0.1", 0, true)
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, testConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to create telemetry service: %v", err))
	}
	defer telemetryService.Shutdown()
	
	// Create JWK generation service.
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		panic(fmt.Sprintf("failed to create JWK generation service: %v", err))
	}
	defer jwkGenService.Shutdown()
	
	// Create barrier repository and service for session encryption.
	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(testDB)
	if err != nil {
		panic(fmt.Sprintf("failed to create barrier repository: %v", err))
	}
	
	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	if err != nil {
		panic(fmt.Sprintf("failed to generate unseal JWK: %v", err))
	}
	
	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	if err != nil {
		panic(fmt.Sprintf("failed to create unseal service: %v", err))
	}
	defer unsealService.Shutdown()
	
	// Create barrier service.
	barrierService, err := cryptoutilTemplateBarrier.NewBarrierService(
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
	testSessionManager, err = cryptoutilTemplateBusinessLogic.NewSessionManagerService(
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
	testJoinRequestMgmtApp = fiber.New()

	// Add test authentication middleware that sets tenant_id and user_id.
	// In production, this would be set by JWT/session validation middleware.
	testTenantID = googleUuid.New()
	testUserID = googleUuid.New()
	authMiddleware := func(c *fiber.Ctx) error {
		c.Locals("tenant_id", testTenantID)
		c.Locals("user_id", testUserID)
		return c.Next()
	}
	testJoinRequestMgmtApp.Use(authMiddleware)

	// Register routes.
	RegisterRegistrationRoutes(testRegistrationApp, testRegistrationSvc, 10)
	RegisterJoinRequestManagementRoutes(testJoinRequestMgmtApp, testRegistrationSvc)

	// Run tests.
	exitCode := m.Run()

	// Cleanup.
	os.Exit(exitCode)
}

func TestIntegration_RegisterUser_CreateTenant(t *testing.T) {
	t.Parallel()

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:8])
	password := "SecurePass123!"
	tenantName := fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8])

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
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
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
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create realm.
	realm := &cryptoutilTemplateRepository.TenantRealm{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		RealmID:  googleUuid.New(),
		Type:     "username_password",
		Active:   true,
		Source:   "db",
	}
	require.NoError(t, testDB.Create(realm).Error)

	// Register user to join existing tenant.
	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:8])
	password := "SecurePass123!"

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
	defer resp.Body.Close()

	// Read response body for debugging.
	bodyBytes, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	if resp.StatusCode != http.StatusOK {
		t.Logf("Response status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(bodyBytes, &result))
	require.Contains(t, result, "message")
	require.Contains(t, result["message"], "pending")

	// Verify join request created.
	var joinRequests []cryptoutilTemplateDomain.TenantJoinRequest
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
				"error": "Rate limit exceeded. Please try again later.",
			})
		}

		return c.Next()
	}

	// Create handlers and register with custom middleware.
	handlers := NewRegistrationHandlers(testRegistrationSvc)
	app.Post("/browser/api/v1/auth/register", rateLimitMiddleware, handlers.HandleRegisterUser)

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:8])

	// Make 4 requests quickly (should exceed limit).
	for i := 0; i < 4; i++ {
		reqBody := RegisterUserRequest{
			Username:     fmt.Sprintf("%s_%d", username, i),
			Password:     "SecurePass123!",
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
		resp.Body.Close()

		if i < 3 {
			require.Equal(t, http.StatusCreated, resp.StatusCode, "Request %d should succeed", i+1)
		} else {
			require.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Request %d should be rate limited", i+1)
		}
	}
}

func TestIntegration_ListJoinRequests(t *testing.T) {
	t.Skip("Join request management requires join flow to be implemented first")
	t.Parallel()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join requests.
	userID1 := googleUuid.New()
	userID2 := googleUuid.New()
	jr1 := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID1,
		Status:   "pending",
	}
	jr2 := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID2,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr1).Error)
	require.NoError(t, testDB.Create(jr2).Error)

	// List join requests.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)
	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "requests")

	requests := result["requests"].([]interface{})
	require.GreaterOrEqual(t, len(requests), 2, "Should have at least 2 join requests")
}

func TestIntegration_ProcessJoinRequest_Approve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request.
	userID := googleUuid.New()
	jr := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr).Error)

	// Approve join request.
	reqBody := ProcessJoinRequestRequest{
		Approved: true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/admin/api/v1/join-requests/%s", jr.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify status updated.
	var updated cryptoutilTemplateDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).First(&updated, "id = ?", jr.ID).Error)
	require.Equal(t, "approved", updated.Status)
}

func TestIntegration_ProcessJoinRequest_Reject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request.
	userID := googleUuid.New()
	jr := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr).Error)

	// Reject join request.
	reqBody := ProcessJoinRequestRequest{
		Approved: false,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/admin/api/v1/join-requests/%s", jr.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify status updated.
	var updated cryptoutilTemplateDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).First(&updated, "id = ?", jr.ID).Error)
	require.Equal(t, "rejected", updated.Status)
}

func TestIntegration_DuplicateUsername_SameTenant(t *testing.T) {
	t.Skip("Join existing tenant flow not yet implemented in service")
	t.Parallel()

	ctx := context.Background()

	// Create tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	username := fmt.Sprintf("user_%s", googleUuid.NewString()[:8])

	// Create first join request.
	userID1 := googleUuid.New()
	jr1 := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: tenant.ID,
		UserID:   &userID1,
		Status:   "pending",
	}
	require.NoError(t, testDB.Create(jr1).Error)

	// Try to create second join request with same username.
	reqBody := RegisterUserRequest{
		Username:     username,
		Password:     "SecurePass123!",
		Email:        fmt.Sprintf("%s@example.com", username),
		CreateTenant: false,
		TenantName:   tenant.Name,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testRegistrationApp.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should still succeed (duplicate check happens during approval).
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify two join requests exist.
	var joinRequests []cryptoutilTemplateDomain.TenantJoinRequest
	require.NoError(t, testDB.WithContext(ctx).Where("tenant_id = ?", tenant.ID).Find(&joinRequests).Error)
	require.GreaterOrEqual(t, len(joinRequests), 2, "Should have at least 2 join requests (duplicate checking deferred to approval)")
}

// TestIntegration_PostgreSQL tests with real PostgreSQL container (slow, only run with -tags=integration).
// NOTE: Disabled on Windows due to testcontainers "rootless Docker" error. Run on Linux/Mac instead.
func TestIntegration_PostgreSQL(t *testing.T) {
	t.Skip("PostgreSQL container test disabled on Windows - rootless Docker not supported")

	ctx := context.Background()

	// Start PostgreSQL container.
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewString())),
		postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewString())),
		postgres.WithPassword("password"),
	)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := gorm.Open(postgresDriver.New(postgresDriver.Config{DSN: connStr}), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations.
	require.NoError(t, db.AutoMigrate(
		&cryptoutilTemplateRepository.Tenant{},
		&cryptoutilTemplateRepository.TenantRealm{},
		&cryptoutilTemplateRepository.User{},
		&cryptoutilTemplateDomain.TenantJoinRequest{},
	))

	// Create repositories and service.
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(db)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(db)
	_ = cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(db, tenantRepo, userRepo, joinRequestRepo)

	// Create tenant via service.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   googleUuid.New(),
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, db.Create(tenant).Error)

	// Verify tenant exists.
	var retrieved cryptoutilTemplateRepository.Tenant
	require.NoError(t, db.First(&retrieved, "id = ?", tenant.ID).Error)
	require.Equal(t, tenant.Name, retrieved.Name)

	t.Logf("PostgreSQL integration test passed with tenant: %s", tenant.Name)
}

// TestIntegration_RegisterUser_InvalidJSON tests HandleRegisterUser with malformed JSON.
func TestIntegration_RegisterUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/browser/api/v1/auth/register", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testRegistrationApp.Test(req, -1)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "error")
	require.Contains(t, result["error"], "Invalid request body")
}

// TestIntegration_ListJoinRequests_NoRequests tests list when no requests exist.
// Note: This test creates its own isolated Fiber app to ensure no state pollution.
func TestIntegration_ListJoinRequests_NoRequests(t *testing.T) {
	t.Parallel()

	// Create empty database for this test only
	emptyDB, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	require.NoError(t, err)

	// Run migrations on empty DB
	require.NoError(t, emptyDB.AutoMigrate(&cryptoutilTemplateDomain.TenantJoinRequest{}))

	// Create repositories with empty DB
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(emptyDB)
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(emptyDB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(emptyDB)

	// Create service with empty DB
	svc := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(emptyDB, tenantRepo, userRepo, joinRequestRepo)

	// Create isolated Fiber app with auth middleware
	app := fiber.New()
	testTenantID := googleUuid.New()
	authMiddleware := func(c *fiber.Ctx) error {
		c.Locals("tenant_id", testTenantID)
		return c.Next()
	}
	app.Use(authMiddleware)
	RegisterJoinRequestManagementRoutes(app, svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "requests")

	requests, ok := result["requests"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 0, len(requests))
}

// TestIntegration_ListJoinRequests_WithData tests list with existing requests.
func TestIntegration_ListJoinRequests_WithData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create tenant using the same tenant ID from auth middleware.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   testTenantID,
		Name: fmt.Sprintf("tenant_%s", googleUuid.NewString()[:8]),
	}
	require.NoError(t, testDB.Create(tenant).Error)

	// Create join request with unique ID.
	userID := googleUuid.New()
	joinReq := &cryptoutilTemplateDomain.TenantJoinRequest{
		ID:       googleUuid.New(),
		TenantID: testTenantID,
		UserID:   &userID,
		Status:   "pending",
	}
	require.NoError(t, testDB.WithContext(ctx).Create(joinReq).Error)

	// Verify join request was created in database.
	var dbCount int64
	require.NoError(t, testDB.Model(&cryptoutilTemplateDomain.TenantJoinRequest{}).Count(&dbCount).Error)
	t.Logf("Total join requests in DB: %d", dbCount)

	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/join-requests", nil)

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	t.Logf("Response body: %s", string(bodyBytes))

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(bodyBytes, &result))
	require.Contains(t, result, "requests")

	requests, ok := result["requests"].([]interface{})
	require.True(t, ok)
	t.Logf("Returned requests count: %d", len(requests))
	// Note: There might be multiple requests from parallel tests, so just check >= 1
	require.GreaterOrEqual(t, len(requests), 1)

	// Find our specific request in the list
	foundOurRequest := false
	for _, reqItem := range requests {
		reqMap, ok := reqItem.(map[string]interface{})
		if !ok {
			continue
		}
		if reqMap["id"] == joinReq.ID.String() {
			foundOurRequest = true
			require.Equal(t, "pending", reqMap["status"])
			require.Equal(t, tenant.ID.String(), reqMap["tenant_id"])
			break
		}
	}
	require.True(t, foundOurRequest, "Our join request should be in the list")
}

// TestIntegration_ProcessJoinRequest_InvalidID tests handling of invalid request IDs.
func TestIntegration_ProcessJoinRequest_InvalidID(t *testing.T) {
	t.Parallel()

	reqBody := `{"approved": true}`
	req := httptest.NewRequest(http.MethodPut, "/admin/api/v1/join-requests/invalid-uuid", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "error")
	require.Equal(t, "Invalid request ID", result["error"])
}

// TestIntegration_ProcessJoinRequest_InvalidJSON tests handling of malformed JSON in request body.
func TestIntegration_ProcessJoinRequest_InvalidJSON(t *testing.T) {
	t.Parallel()

	validID := googleUuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/admin/api/v1/join-requests/"+validID, strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testJoinRequestMgmtApp.Test(req, -1)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Contains(t, result, "error")
	require.Equal(t, "Invalid request body", result["error"])
}

// TestRegistrationRoutes_MethodNotAllowed tests unsupported HTTP methods return 405.
func TestRegistrationRoutes_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET on register endpoint", http.MethodGet, "/browser/api/v1/auth/register"},
		{"DELETE on register endpoint", http.MethodDelete, "/browser/api/v1/auth/register"},
		{"PATCH on register endpoint", http.MethodPatch, "/browser/api/v1/auth/register"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			resp, err := testRegistrationApp.Test(req, -1)
			require.NoError(t, err)
			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})
	}
}
