// Copyright (c) 2025 Justin Cranford
//

package businesslogic

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSessionManagerService_ValidateServiceSession_NilContext(t *testing.T) {
	t.Parallel()

	svc := &SessionManagerService{
		sessionManager: setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
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
		sessionManager: setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
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
		sessionManager: setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
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
		sessionManager: setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
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
		sessionManager: setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
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

// TestNewSessionManagerService_HappyPath tests successful creation of SessionManagerService.
func TestNewSessionManagerService_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	telemetrySvc := setupTelemetryService(t)
	jwkGenSvc := setupJWKGenService(t, telemetrySvc)
	barrierSvc := setupBarrierService(t, db, telemetrySvc, jwkGenSvc)

	// Create config with valid session settings.
	config := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	svc, err := NewSessionManagerService(ctx, db, telemetrySvc, jwkGenSvc, barrierSvc, config)

	require.NoError(t, err)
	require.NotNil(t, svc)

	// Verify the service is functional by calling a simple method.
	// Use IssueBrowserSessionWithTenant which validates the session manager is initialized.
	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	token, err := svc.IssueBrowserSessionWithTenant(ctx, "test-user-123", tenantID, realmID)

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestNewSessionManagerService_InitializeError tests that NewSessionManagerService returns error when Initialize fails.
func TestNewSessionManagerService_InitializeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	telemetrySvc := setupTelemetryService(t)
	jwkGenSvc := setupJWKGenService(t, telemetrySvc)
	barrierSvc := setupBarrierService(t, db, telemetrySvc, jwkGenSvc)

	// Create config with valid session settings.
	config := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Close the underlying DB connection to make Initialize fail.
	sqlDB, err := db.DB()
	require.NoError(t, err)

	_ = sqlDB.Close()

	// NewSessionManagerService should fail because Initialize will fail on DB operations.
	svc, err := NewSessionManagerService(ctx, db, telemetrySvc, jwkGenSvc, barrierSvc, config)

	require.Error(t, err)
	require.Nil(t, svc)
	require.Contains(t, err.Error(), "failed to initialize session manager")
}
