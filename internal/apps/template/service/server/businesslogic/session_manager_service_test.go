// Copyright (c) 2025 Justin Cranford
//

package businesslogic

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// setupTelemetryService creates a TelemetryService for testing.
func setupTelemetryService(t *testing.T) *cryptoutilTelemetry.TelemetryService {
	t.Helper()

	ctx := context.Background()
	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)

	t.Cleanup(func() {
		telemetrySvc.Shutdown()
	})

	return telemetrySvc
}

// setupJWKGenService creates a JWKGenService for testing.
func setupJWKGenService(t *testing.T, telemetrySvc *cryptoutilTelemetry.TelemetryService) *cryptoutilJose.JWKGenService {
	t.Helper()

	ctx := context.Background()
	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)

	t.Cleanup(func() {
		jwkGenSvc.Shutdown()
	})

	return jwkGenSvc
}

// setupBarrierService creates a BarrierService for testing.
func setupBarrierService(t *testing.T, db *gorm.DB, telemetrySvc *cryptoutilTelemetry.TelemetryService, jwkGenSvc *cryptoutilJose.JWKGenService) *cryptoutilTemplateBarrier.BarrierService {
	t.Helper()

	ctx := context.Background()

	// Create barrier tables.
	sqlDB, err := db.DB()
	require.NoError(t, err)

	schema := `
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
		updated_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS barrier_content_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`

	_, err = sqlDB.ExecContext(ctx, schema)
	require.NoError(t, err)

	// Create unseal JWK.
	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	t.Cleanup(func() {
		unsealService.Shutdown()
	})

	// Create barrier repository.
	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)

	t.Cleanup(func() {
		barrierRepo.Shutdown()
	})

	// Create barrier service.
	barrierSvc, err := cryptoutilTemplateBarrier.NewBarrierService(ctx, telemetrySvc, jwkGenSvc, barrierRepo, unsealService)
	require.NoError(t, err)

	t.Cleanup(func() {
		barrierSvc.Shutdown()
	})

	return barrierSvc
}

// TestNewSessionManagerService_NilContext tests that NewSessionManagerService returns an error when context is nil.
func TestNewSessionManagerService_NilContext(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	svc, err := NewSessionManagerService(nil, nil, nil, nil, nil, nil)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewSessionManagerService_NilDB tests that NewSessionManagerService returns an error when db is nil.
func TestNewSessionManagerService_NilDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc, err := NewSessionManagerService(ctx, nil, nil, nil, nil, nil)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "database cannot be nil")
}

// TestNewSessionManagerService_NilTelemetry tests that NewSessionManagerService returns an error when telemetryService is nil.
func TestNewSessionManagerService_NilTelemetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)

	svc, err := NewSessionManagerService(ctx, db, nil, nil, nil, nil)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "telemetry service cannot be nil")
}

// Note: Testing NilJWKGenService, NilBarrierService, NilConfig requires creating real
// TelemetryService, JWKGenService, and BarrierService instances, which are complex to
// instantiate. Those validation paths are tested implicitly via the SessionManager tests
// or integration tests.

// --- Method Validation Tests ---
// These tests verify the input validation in the SessionManagerService wrapper methods.
// They bypass the constructor by directly creating a SessionManagerService with a valid SessionManager.

// TestSessionManagerService_IssueBrowserSessionWithTenant_NilContext tests context validation.
func TestSessionManagerService_IssueBrowserSessionWithTenant_NilContext(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	token, err := svc.IssueBrowserSessionWithTenant(nil, "user123", googleUuid.New(), googleUuid.New())

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestSessionManagerService_IssueBrowserSessionWithTenant_EmptyUserID tests userID validation.
func TestSessionManagerService_IssueBrowserSessionWithTenant_EmptyUserID(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	token, err := svc.IssueBrowserSessionWithTenant(ctx, "", googleUuid.New(), googleUuid.New())

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "user ID cannot be empty")
}

// TestSessionManagerService_IssueBrowserSessionWithTenant_Success tests successful session issuance.
func TestSessionManagerService_IssueBrowserSessionWithTenant_Success(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	userID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	token, err := svc.IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID)

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestSessionManagerService_ValidateBrowserSession_NilContext tests context validation.
// TestSessionManagerService_ValidateBrowserSession_NilContext tests context validation.
func TestSessionManagerService_ValidateBrowserSession_NilContext(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	session, err := svc.ValidateBrowserSession(nil, "some-token")

	require.Error(t, err)
	require.Nil(t, session)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestSessionManagerService_ValidateBrowserSession_EmptyToken tests token validation.
func TestSessionManagerService_ValidateBrowserSession_EmptyToken(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	session, err := svc.ValidateBrowserSession(ctx, "")

	require.Error(t, err)
	require.Nil(t, session)
	require.Contains(t, err.Error(), "token cannot be empty")
}

// TestSessionManagerService_ValidateBrowserSession_Success tests successful session validation.
func TestSessionManagerService_ValidateBrowserSession_Success(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	userID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// First issue a session.
	token, err := svc.IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Then validate it.
	session, err := svc.ValidateBrowserSession(ctx, token)

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
}

// TestSessionManagerService_IssueServiceSessionWithTenant_NilContext tests context validation.
func TestSessionManagerService_IssueServiceSessionWithTenant_NilContext(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	token, err := svc.IssueServiceSessionWithTenant(nil, "client123", googleUuid.New(), googleUuid.New())

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestSessionManagerService_IssueServiceSessionWithTenant_EmptyClientID tests clientID validation.
func TestSessionManagerService_IssueServiceSessionWithTenant_EmptyClientID(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	token, err := svc.IssueServiceSessionWithTenant(ctx, "", googleUuid.New(), googleUuid.New())

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "client ID cannot be empty")
}

// TestSessionManagerService_IssueServiceSessionWithTenant_Success tests successful session issuance.
func TestSessionManagerService_IssueServiceSessionWithTenant_Success(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	clientID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	token, err := svc.IssueServiceSessionWithTenant(ctx, clientID, tenantID, realmID)

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestSessionManagerService_ValidateServiceSession_NilContext tests context validation.
func TestSessionManagerService_ValidateServiceSession_NilContext(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	session, err := svc.ValidateServiceSession(nil, "some-token")

	require.Error(t, err)
	require.Nil(t, session)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestSessionManagerService_ValidateServiceSession_EmptyToken tests token validation.
func TestSessionManagerService_ValidateServiceSession_EmptyToken(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	session, err := svc.ValidateServiceSession(ctx, "")

	require.Error(t, err)
	require.Nil(t, session)
	require.Contains(t, err.Error(), "token cannot be empty")
}

// TestSessionManagerService_ValidateServiceSession_Success tests successful session validation.
func TestSessionManagerService_ValidateServiceSession_Success(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()
	clientID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// First issue a session.
	token, err := svc.IssueServiceSessionWithTenant(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Then validate it.
	session, err := svc.ValidateServiceSession(ctx, token)

	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
}

// TestSessionManagerService_CleanupExpiredSessions_NilContext tests context validation.
func TestSessionManagerService_CleanupExpiredSessions_NilContext(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	//nolint:staticcheck // SA1012: Intentionally passing nil context to test validation.
	err := svc.CleanupExpiredSessions(nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestSessionManagerService_CleanupExpiredSessions_Success tests successful cleanup.
func TestSessionManagerService_CleanupExpiredSessions_Success(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE),
	}

	ctx := context.Background()

	// First create some sessions.
	userID := googleUuid.New().String()
	clientID := googleUuid.New().String()
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	_, err := svc.IssueBrowserSessionWithTenant(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	_, err = svc.IssueServiceSessionWithTenant(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)

	// Cleanup should succeed (no expired sessions yet, but function should complete).
	err = svc.CleanupExpiredSessions(ctx)

	require.NoError(t, err)
}

// TestNewSessionManagerService_NilJWKGenService tests that NewSessionManagerService returns an error when jwkGenService is nil.
func TestNewSessionManagerService_NilJWKGenService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	telemetrySvc := setupTelemetryService(t)

	svc, err := NewSessionManagerService(ctx, db, telemetrySvc, nil, nil, nil)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewSessionManagerService_NilBarrierService tests that NewSessionManagerService returns an error when barrierService is nil.
func TestNewSessionManagerService_NilBarrierService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	telemetrySvc := setupTelemetryService(t)
	jwkGenSvc := setupJWKGenService(t, telemetrySvc)

	svc, err := NewSessionManagerService(ctx, db, telemetrySvc, jwkGenSvc, nil, nil)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "barrier service cannot be nil")
}

// TestNewSessionManagerService_NilConfig tests that NewSessionManagerService returns an error when config is nil.
func TestNewSessionManagerService_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	telemetrySvc := setupTelemetryService(t)
	jwkGenSvc := setupJWKGenService(t, telemetrySvc)
	barrierSvc := setupBarrierService(t, db, telemetrySvc, jwkGenSvc)

	svc, err := NewSessionManagerService(ctx, db, telemetrySvc, jwkGenSvc, barrierSvc, nil)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "config cannot be nil")
}
