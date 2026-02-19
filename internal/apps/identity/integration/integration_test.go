// Copyright (c) 2025 Justin Cranford
//
//

//go:build integration

// Package integration provides integration tests for the identity server.
// These tests require real network listeners and use hardcoded ports,
// so they must be run separately from unit tests using:
//
//	go test -tags=integration ./internal/identity/integration/...
//
// Do NOT run these with `go test ./...` as they will conflict with parallel tests.
package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	cryptoutilIdentityPKCE "cryptoutil/internal/apps/identity/authz/pkce"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/apps/identity/server"
	cryptoutilDigests "cryptoutil/internal/shared/crypto/digests"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

// mockKeyGenerator implements KeyGenerator for integration tests.
type mockKeyGenerator struct{}

func (m *mockKeyGenerator) GenerateSigningKey(ctx context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	// Use test key bytes (mock RSA private key for testing).
	keyBytes := []byte("mock-rsa-private-key-bytes-for-testing")

	return &cryptoutilIdentityIssuer.SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           keyBytes,
		Algorithm:     algorithm,
		CreatedAt:     time.Now().UTC(),
		Active:        false,
		ValidForVerif: false,
	}, nil
}

func (m *mockKeyGenerator) GenerateEncryptionKey(ctx context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	// Use test key bytes (32-byte AES-256 key for testing).
	keyBytes := []byte("01234567890123456789012345678901")

	return &cryptoutilIdentityIssuer.EncryptionKey{
		KeyID:        googleUuid.NewString(),
		Key:          keyBytes,
		CreatedAt:    time.Now().UTC(),
		Active:       false,
		ValidForDecr: false,
	}, nil
}

const (
	testTimeout       = 30 * time.Second
	serverStartDelay  = 500 * time.Millisecond
	httpClientTimeout = 5 * time.Second
	shutdownTimeout   = 3 * time.Second
	testAuthZPort     = 18080
	testIDPPort       = 18081
	testRSPort        = 18082
	testAuthZBaseURL  = "http://127.0.0.1:18080"
	testIDPBaseURL    = "http://127.0.0.1:18081"
	testRSBaseURL     = "http://127.0.0.1:18082"
	testRedirectURI   = "http://127.0.0.1:18083/callback"
	testClientID      = "test-client"
	testClientSecret  = "test-secret" // pragma: allowlist secret
	testScope         = "read:resource write:resource"
	testAdminScope    = "admin"
	testUserID        = "test-user"
	testUsername      = "testuser"
	testPassword      = "testpass123" // pragma: allowlist secret
	authorizationCode = "test-auth-code"
	bearerTokenPrefix = "Bearer "
)

// testMutex ensures integration tests run sequentially to avoid port conflicts.
// These tests use hardcoded ports (18080, 18081, 18082) and cannot run in parallel.
var testMutex sync.Mutex //nolint:gochecknoglobals // Required for test synchronization

// testServers holds all three identity servers for integration testing.
//
// Validates requirements:
// - R11-01: All integration tests passing
// - R11-02: Code coverage meets target (≥90%).
type testServers struct {
	authzServer *cryptoutilIdentityServer.AuthZServer
	idpServer   *cryptoutilIdentityServer.IDPServer
	rsServer    *cryptoutilIdentityServer.RSServer
	httpClient  *http.Client
}

// setupTestServers creates and starts all three identity servers.
func setupTestServers(t *testing.T) (*testServers, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)

	// Generate unique DB name per test using UUID to ensure test isolation.
	testUUID := googleUuid.New()
	testDBName := fmt.Sprintf("file:%s.db?mode=memory&cache=shared", testUUID.String())

	authzConfig := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "test-authz",
			BindAddress: "127.0.0.1",
			Port:        testAuthZPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  testDBName,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: "jws",
			Issuer:            testAuthZBaseURL,
			SigningAlgorithm:  "RS256",
		},
	}

	idpConfig := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "test-idp",
			BindAddress: "127.0.0.1",
			Port:        testIDPPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  testDBName,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: "jws",
			Issuer:            testAuthZBaseURL,
			SigningAlgorithm:  "RS256",
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			SessionLifetime: 3600 * time.Second,
			IdleTimeout:     1800 * time.Second,
			CookieName:      "session_id",
			CookiePath:      "/",
			CookieSecure:    false,
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
		},
	}

	rsConfig := &cryptoutilIdentityConfig.Config{
		RS: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "test-rs",
			BindAddress: "127.0.0.1",
			Port:        testRSPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  testDBName,
		},
	}

	// Initialize repository factory (shared across all servers).
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, authzConfig.Database)
	testify.NoError(t, err, "Failed to initialize repository factory")

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	testify.NoError(t, err, "Failed to run database migrations")

	// Create logger for tests.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in test output.
	}))

	// TEMPORARY: Use legacy JWS issuer without key rotation for integration tests.
	// TODO: Implement proper key rotation testing infrastructure.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	testify.NoError(t, err, "Failed to generate RSA key")

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		authzConfig.Tokens.Issuer,
		privateKey,
		authzConfig.Tokens.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	testify.NoError(t, err, "Failed to create JWS issuer")

	// Create key rotation manager for JWE encryption only.
	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockKeyGenerator{},
		nil,
	)
	testify.NoError(t, err, "Failed to create key rotation manager")

	// Initialize encryption keys by triggering first rotation.
	err = keyRotationMgr.RotateEncryptionKey(ctx)
	testify.NoError(t, err, "Failed to rotate initial encryption key")

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	testify.NoError(t, err, "Failed to create JWE issuer")

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, authzConfig.Tokens)

	// Create all three servers.
	authzServer := cryptoutilIdentityServer.NewAuthZServer(authzConfig, repoFactory, tokenSvc)
	idpServer := cryptoutilIdentityServer.NewIDPServer(idpConfig, repoFactory, tokenSvc)
	rsServer, err := cryptoutilIdentityServer.NewRSServer(ctx, rsConfig, logger, tokenSvc)
	testify.NoError(t, err, "Failed to create RS server")

	// Create HTTP client for testing.
	httpClient := &http.Client{
		Timeout: httpClientTimeout,
	}

	servers := &testServers{
		authzServer: authzServer,
		idpServer:   idpServer,
		rsServer:    rsServer,
		httpClient:  httpClient,
	}

	// Start all servers in background with synchronized error handling.
	// Use buffered channel to prevent goroutine leaks if server starts before we read.
	errChan := make(chan error, 3)

	go func() {
		if err := authzServer.Start(ctx); err != nil {
			t.Logf("AuthZ server error: %v", err)

			errChan <- err
		}
	}()

	go func() {
		if err := idpServer.Start(ctx); err != nil {
			t.Logf("IdP server error: %v", err)

			errChan <- err
		}
	}()

	go func() {
		if err := rsServer.Start(ctx); err != nil {
			t.Logf("RS server error: %v", err)

			errChan <- err
		}
	}()

	// Wait for servers to start, checking for startup errors.
	time.Sleep(serverStartDelay)

	// Check if any server failed to start.
	select {
	case err := <-errChan:
		testify.FailNowf(t, "Server failed to start: %v", err.Error())
	default:
		// All servers started successfully.
	}

	// Seed test data: create test client.
	seedTestData(t, ctx, repoFactory)

	return servers, cancel
}

// shutdownTestServers gracefully shuts down all test servers.
func shutdownTestServers(t *testing.T, servers *testServers) {
	t.Helper()

	// Use longer timeout to allow servers to clean up properly.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown servers in parallel to speed up test cleanup.
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		if err := servers.authzServer.Stop(shutdownCtx); err != nil {
			// Don't fail tests on shutdown errors - database may already be closed.
			if !strings.Contains(err.Error(), "database is closed") {
				t.Logf("AuthZ server shutdown error: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		if err := servers.idpServer.Stop(shutdownCtx); err != nil {
			// Don't fail tests on shutdown errors - database may already be closed.
			if !strings.Contains(err.Error(), "database is closed") {
				t.Logf("IdP server shutdown error: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		if err := servers.rsServer.Stop(shutdownCtx); err != nil {
			// Don't fail tests on shutdown errors - database may already be closed.
			if !strings.Contains(err.Error(), "database is closed") {
				t.Logf("RS server shutdown error: %v", err)
			}
		}
	}()

	// Wait for all servers to finish shutting down.
	wg.Wait()

	// CRITICAL: Add delay after shutdown to allow OS to release ports.
	// Without this, next test may fail with "bind: address already in use".
	// Increased from 100ms to 500ms to handle TCP TIME_WAIT state properly.
	time.Sleep(500 * time.Millisecond)
}

// seedTestData seeds the database with test client.
func seedTestData(t *testing.T, ctx context.Context, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	// Use GORM AutoMigrate now that gorm.Model duplication is fixed.
	db := repoFactory.DB()

	// Create test user for authentication.
	userRepo := repoFactory.UserRepository()
	testUserUUID := googleUuid.Must(googleUuid.NewV7())
	now := time.Now().UTC()

	// Generate password hash using the same crypto package used by authentication
	passwordHash, err := cryptoutilDigests.HashLowEntropyNonDeterministic(testPassword)
	testify.NoError(t, err, "Failed to hash test password")

	testUser := &cryptoutilIdentityDomain.User{
		ID:                testUserUUID,
		Sub:               testUserID,
		PreferredUsername: testUsername,
		Email:             "testuser@example.com",
		PasswordHash:      passwordHash,
		Enabled:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = userRepo.Create(ctx, testUser)
	testify.NoError(t, err, "Failed to create test user")

	// Create test client using repository.
	clientRepo := repoFactory.ClientRepository()
	testClientUUID := googleUuid.Must(googleUuid.NewV7())

	// Hash the client secret using PBKDF2-HMAC-SHA256 (same as production).
	// Integration test uses same hashing as production to validate authentication flow.
	hashedClientSecret, err := cryptoutilDigests.HashLowEntropyNonDeterministic(testClientSecret)
	testify.NoError(t, err, "Failed to hash client secret")

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      testClientUUID,
		ClientID:                testClientID,
		ClientSecret:            hashedClientSecret, // Store hashed secret
		ClientType:              "confidential",
		Name:                    "Test Client",
		Description:             "Test client for integration tests",
		RedirectURIs:            []string{testRedirectURI},
		AllowedGrantTypes:       []string{"authorization_code", "client_credentials", "refresh_token"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"read:resource", "write:resource", "delete:resource", "admin"},
		TokenEndpointAuthMethod: "client_secret_post",
		RequirePKCE:             boolPtr(true),
		PKCEChallengeMethod:     "S256",
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 boolPtr(true),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	err = clientRepo.Create(ctx, testClient)
	testify.NoError(t, err, "Failed to create test client")

	// Verify client was created.
	var clientCount int64

	err = db.Table("clients").Count(&clientCount).Error
	testify.NoError(t, err, "Failed to count clients")
	testify.Equal(t, int64(1), clientCount, "Should have exactly 1 client")
}

// TestHealthCheckEndpoints verifies all servers respond to health checks.
//
// Validates requirements:
// - R03-01: Integration tests use real SQLite database
// - R03-02: Integration tests start all three servers
// - R03-04: Integration tests validate cross-server interactions.
func TestHealthCheckEndpoints(t *testing.T) {
	t.Parallel()
	testMutex.Lock()

	servers, cancel := setupTestServers(t)

	defer func() {
		shutdownTestServers(t, servers)
		cancel()
		testMutex.Unlock()
	}()

	tests := []struct {
		name     string
		endpoint string
	}{
		{"AuthZ Health Check", testAuthZBaseURL + "/health"},
		{"IdP Health Check", testIDPBaseURL + "/health"},
		{"RS Health Check", testRSBaseURL + "/api/v1/public/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.endpoint, nil)
			testify.NoError(t, err, "Failed to create request")

			resp, err := servers.httpClient.Do(req)
			testify.NoError(t, err, "Health check request failed")

			defer func() {
				defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test code cleanup
			}()

			testify.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200 OK")
		})
	}
}

// TestOAuth2AuthorizationCodeFlow tests the complete OAuth 2.0 authorization code flow.
//
// Validates requirements:
// - R01-01: /oauth2/v1/authorize stores authorization request and redirects to login
// - R01-02: User login associates real user ID with authorization request
// - R01-03: Consent approval generates authorization code with user context
// - R01-04: Token exchange returns access token with real user ID in sub claim
// - R01-05: Authorization code single-use enforcement
// - R01-06: Integration test validates end-to-end OAuth 2.1 flow.
