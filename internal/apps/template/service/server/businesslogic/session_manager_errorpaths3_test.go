// Copyright (c) 2025 Justin Cranford

// Package businesslogic — error path tests for OPAQUE, unsupported algorithms, cleanup, and service-side.
package businesslogic

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testInvalidAlgorithm = "INVALID_ALGO"

// =============================================================================
// OPAQUE session error path tests.
// =============================================================================

// TestIssueOPAQUESession_HashError covers the hash error in OPAQUE issue.
func TestIssueOPAQUESession_HashError(t *testing.T) {
	t.Parallel()

	orig := hashHighEntropyDeterministicFn

	defer func() { hashHighEntropyDeterministicFn = orig }()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	hashHighEntropyDeterministicFn = func(_ string) (string, error) {
		return "", fmt.Errorf("injected hash error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash session token")
}

// TestIssueOPAQUESession_CreateDBError covers the DB create error in OPAQUE issue.
func TestIssueOPAQUESession_CreateDBError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	sqlDB, err := sm.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	_, err = sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to store session")
}

// TestValidateOPAQUESession_HashError covers the hash error in OPAQUE validate.
func TestValidateOPAQUESession_HashError(t *testing.T) {
	t.Parallel()

	orig := hashHighEntropyDeterministicFn

	defer func() { hashHighEntropyDeterministicFn = orig }()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	hashHighEntropyDeterministicFn = func(_ string) (string, error) {
		return "", fmt.Errorf("injected hash error")
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash session token")
}

// TestValidateOPAQUESession_DBQueryError covers DB errors during session lookup.
func TestValidateOPAQUESession_DBQueryError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	sqlDB, dbErr := sm.db.DB()
	require.NoError(t, dbErr)
	require.NoError(t, sqlDB.Close())

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// =============================================================================
// Unsupported algorithm tests.
// =============================================================================

// TestValidateBrowserSessionJWE_UnsupportedAlgorithm covers the default case.
func TestValidateBrowserSessionJWE_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	sm.browserAlgorithm = testInvalidAlgorithm

	_, err := sm.ValidateBrowserSession(context.Background(), cryptoutilSharedMagic.ParamToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported browser session algorithm")
}

// TestValidateServiceSessionJWE_UnsupportedAlgorithm covers the default case.
func TestValidateServiceSessionJWE_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	sm.serviceAlgorithm = testInvalidAlgorithm

	_, err := sm.ValidateServiceSession(context.Background(), cryptoutilSharedMagic.ParamToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported service session algorithm")
}

// TestGenerateJWSKey_UnsupportedAlgorithm covers the default case.
func TestGenerateJWSKey_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	_, err := sm.generateJWSKey("INVALID_JWS_ALG")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWS algorithm")
}

// TestGenerateJWEKey_UnsupportedAlgorithm covers the default case.
func TestGenerateJWEKey_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	_, err := sm.generateJWEKey("INVALID_JWE_ALG")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWE algorithm")
}

// TestGenerateSessionJWK_UnsupportedAlgorithm covers the default case.
func TestGenerateSessionJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	_, err := sm.generateSessionJWK(true, "INVALID_ALGO")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported session algorithm")
}

// =============================================================================
// CleanupExpiredSessions error path tests.
// =============================================================================

// TestCleanupExpiredSessions_ServiceCleanupError covers the service session cleanup error.
func TestCleanupExpiredSessions_ServiceCleanupError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	// Drop only the service_sessions table to trigger error on service cleanup
	// while browser cleanup succeeds.
	err := sm.db.Exec("DROP TABLE IF EXISTS service_sessions").Error
	require.NoError(t, err)

	err = sm.CleanupExpiredSessions(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to cleanup service sessions")
}

// =============================================================================
// Service session error path tests (isBrowser=false paths).
// =============================================================================

// TestIssueServiceSession_JWE_LoadError covers service-side JWE issue with bad JWK ID.
func TestIssueServiceSession_JWE_LoadError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.serviceJWKID = &badID

	_, err := sm.IssueServiceSession(context.Background(), "client1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load session JWK")
}

// TestIssueServiceSession_JWS_LoadError covers service-side JWS issue with bad JWK ID.
func TestIssueServiceSession_JWS_LoadError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.serviceJWKID = &badID

	_, err := sm.IssueServiceSession(context.Background(), "client1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load session JWK")
}

// TestValidateServiceSession_JWE_LoadError covers service-side JWE validate with bad JWK ID.
func TestValidateServiceSession_JWE_LoadError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.serviceJWKID = &badID

	_, err := sm.ValidateServiceSession(context.Background(), "some-token")
	require.Error(t, err)
}

// TestValidateServiceSession_JWS_LoadError covers service-side JWS validate with bad JWK ID.
func TestValidateServiceSession_JWS_LoadError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.serviceJWKID = &badID

	_, err := sm.ValidateServiceSession(context.Background(), "some-token")
	require.Error(t, err)
}
