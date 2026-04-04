// Copyright (c) 2025 Justin Cranford

// Package businesslogic Ã¢â‚¬â€ error path coverage tests for session manager.
// These tests inject errors via SessionManager struct fields, enabling t.Parallel().
package businesslogic

import (
	"context"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// =============================================================================
// initializeSessionJWK error path tests.
// =============================================================================

// TestInitializeSessionJWK_UnsupportedSessionAlgorithm covers the outer default case.
func TestInitializeSessionJWK_UnsupportedSessionAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	_, err := sm.initializeSessionJWK(context.Background(), true, "INVALID_ALGO")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported session algorithm")
}

// TestInitializeSessionJWK_GenerateJWKError covers the genErr != nil branch.
func TestInitializeSessionJWK_GenerateJWKError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	sm.generateRSAJWKFn = func(_ int) (joseJwk.Key, error) {
		return nil, fmt.Errorf("injected generate error")
	}

	_, err := sm.initializeSessionJWK(context.Background(), true, cryptoutilSharedMagic.SessionAlgorithmJWS)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate JWK")
}

// TestInitializeSessionJWK_MarshalJWKError covers the marshalErr != nil branch.
func TestInitializeSessionJWK_MarshalJWKError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	sm.jsonMarshalFn = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("injected marshal error")
	}

	_, err := sm.initializeSessionJWK(context.Background(), true, cryptoutilSharedMagic.SessionAlgorithmJWS)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to marshal JWK")
}

// TestInitializeSessionJWK_BarrierEncryptError covers the barrier encrypt error branch.
func TestInitializeSessionJWK_BarrierEncryptError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	sm.barrierEncryptFn = func(_ context.Context, _ *cryptoutilAppsFrameworkServiceServerBarrier.Service, _ []byte) ([]byte, error) {
		return nil, fmt.Errorf("injected barrier encrypt error")
	}

	_, err := sm.initializeSessionJWK(context.Background(), true, cryptoutilSharedMagic.SessionAlgorithmJWS)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to encrypt JWK")
}

// TestInitializeSessionJWK_StoreJWKDBError covers the DB create error branch.
func TestInitializeSessionJWK_StoreJWKDBError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	// Close DB to force error.
	sqlDB, err := sm.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	_, err = sm.initializeSessionJWK(context.Background(), true, cryptoutilSharedMagic.SessionAlgorithmJWS)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to")
}

// =============================================================================
// issueJWESession error path tests.
// =============================================================================

// TestIssueJWESession_LoadJWKError covers the JWK load error branch.
func TestIssueJWESession_LoadJWKError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.browserJWKID = &badID

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load session JWK")
}

// TestIssueJWESession_BarrierDecryptError covers the barrier decrypt error in issue.
func TestIssueJWESession_BarrierDecryptError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	sm.barrierDecryptFn = func(_ context.Context, _ *cryptoutilAppsFrameworkServiceServerBarrier.Service, _ []byte) ([]byte, error) {
		return nil, fmt.Errorf("injected barrier decrypt error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt session JWK")
}

// TestIssueJWESession_ParseJWKError covers the JWK parse error in issue.
func TestIssueJWESession_ParseJWKError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	sm.jwkParseKeyFn = func(_ []byte, _ ...joseJwk.ParseOption) (joseJwk.Key, error) {
		return nil, fmt.Errorf("injected parse error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWK")
}

// TestIssueJWESession_MarshalClaimsError covers the JSON marshal error in issue.
func TestIssueJWESession_MarshalClaimsError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	sm.jsonMarshalFn = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("injected marshal error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to marshal JWT claims")
}

// TestIssueJWESession_EncryptError covers the encrypt error in issue.
func TestIssueJWESession_EncryptError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	sm.encryptBytesFn = func(_ []joseJwk.Key, _ []byte) (*joseJwe.Message, []byte, error) {
		return nil, nil, fmt.Errorf("injected encrypt error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to encrypt JWT")
}

// TestIssueJWESession_HashError covers the hash error in issue.
func TestIssueJWESession_HashError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	sm.hashHighEntropyDeterministicFn = func(_ string) (string, error) {
		return "", fmt.Errorf("injected hash error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash jti")
}

// TestIssueJWESession_CreateSessionDBError covers the DB create error in issue.
func TestIssueJWESession_CreateSessionDBError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)

	// Drop session tables to force create error while JWK load still works.
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS browser_sessions").Error)
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS service_sessions").Error)

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
}

// =============================================================================
// validateJWESession error path tests.
// =============================================================================

// TestValidateJWESession_LoadJWKError covers the JWK load error in validate.
func TestValidateJWESession_LoadJWKError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.browserJWKID = &badID

	_, err := sm.ValidateBrowserSession(context.Background(), "some-token")
	require.Error(t, err)
}

// TestValidateJWESession_BarrierDecryptError covers the barrier decrypt error in validate.
func TestValidateJWESession_BarrierDecryptError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	sm.barrierDecryptFn = func(_ context.Context, _ *cryptoutilAppsFrameworkServiceServerBarrier.Service, _ []byte) ([]byte, error) {
		return nil, fmt.Errorf("injected barrier decrypt error")
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// TestValidateJWESession_DecryptTokenError covers the JWT decrypt error in validate.
func TestValidateJWESession_DecryptTokenError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	_, err := sm.ValidateBrowserSession(context.Background(), "not-a-valid-jwe-token")
	require.Error(t, err)
}

// TestValidateJWESession_DBQueryError covers DB errors during session lookup in validate.
func TestValidateJWESession_DBQueryError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Close DB to force query error.
	sqlDB, dbErr := sm.db.DB()
	require.NoError(t, dbErr)
	require.NoError(t, sqlDB.Close())

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// =============================================================================
// issueJWSSession error path tests.
// =============================================================================

// TestIssueJWSSession_LoadJWKError covers the JWK load error branch.
func TestIssueJWSSession_LoadJWKError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.browserJWKID = &badID

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load session JWK")
}

// TestIssueJWSSession_BarrierDecryptError covers the barrier decrypt error in issue.
func TestIssueJWSSession_BarrierDecryptError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.barrierDecryptFn = func(_ context.Context, _ *cryptoutilAppsFrameworkServiceServerBarrier.Service, _ []byte) ([]byte, error) {
		return nil, fmt.Errorf("injected barrier decrypt error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt JWK")
}

// TestIssueJWSSession_MarshalClaimsError covers the JSON marshal error.
func TestIssueJWSSession_MarshalClaimsError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.jsonMarshalFn = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("injected marshal error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to marshal JWT claims")
}

// TestIssueJWSSession_SignError covers the sign error in issue.
func TestIssueJWSSession_SignError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.signBytesFn = func(_ []joseJwk.Key, _ []byte) (*joseJws.Message, []byte, error) {
		return nil, nil, fmt.Errorf("injected sign error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to sign JWT")
}

// TestIssueJWSSession_HashError covers the hash error in issue.
func TestIssueJWSSession_HashError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.hashHighEntropyDeterministicFn = func(_ string) (string, error) {
		return "", fmt.Errorf("injected hash error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash jti")
}

// TestIssueJWSSession_CreateSessionDBError covers the DB create error in issue.
func TestIssueJWSSession_CreateSessionDBError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)

	// Drop session tables to force create error while JWK load still works.
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS browser_sessions").Error)
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS service_sessions").Error)

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to store session")
}

// =============================================================================
// validateJWSSession error path tests.
// =============================================================================

// TestValidateJWSSession_LoadJWKError covers the JWK load error in validate.
func TestValidateJWSSession_LoadJWKError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	badID := googleUuid.Must(googleUuid.NewV7())
	sm.browserJWKID = &badID

	_, err := sm.ValidateBrowserSession(context.Background(), "some-token")
	require.Error(t, err)
}

// TestValidateJWSSession_BarrierDecryptError covers the barrier decrypt error in validate.
func TestValidateJWSSession_BarrierDecryptError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	sm.barrierDecryptFn = func(_ context.Context, _ *cryptoutilAppsFrameworkServiceServerBarrier.Service, _ []byte) ([]byte, error) {
		return nil, fmt.Errorf("injected barrier decrypt error")
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// TestValidateJWSSession_VerifyError covers the JWT verify error in validate.
func TestValidateJWSSession_VerifyError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	_, err := sm.ValidateBrowserSession(context.Background(), "not-a-valid-jws-token")
	require.Error(t, err)
}

// TestValidateJWSSession_HashError covers the hash error in validate.
func TestValidateJWSSession_HashError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	sm.hashHighEntropyDeterministicFn = func(_ string) (string, error) {
		return "", fmt.Errorf("injected hash error")
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash jti")
}

// TestValidateJWSSession_DBQueryError covers DB errors during session lookup.
func TestValidateJWSSession_DBQueryError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	sqlDB, dbErr := sm.db.DB()
	require.NoError(t, dbErr)
	require.NoError(t, sqlDB.Close())

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}
