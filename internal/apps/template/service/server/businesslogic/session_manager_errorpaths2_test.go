// Copyright (c) 2025 Justin Cranford

// Package businesslogic — additional error path coverage tests (part 2).
// Covers JWK parse errors, verify/decrypt result errors, and DB query/update errors.
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

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// =============================================================================
// JWK parse error tests for issue and validate paths.
// =============================================================================

// TestIssueJWSSession_ParseJWKError covers the JWK parse error in JWS issue.
func TestIssueJWSSession_ParseJWKError(t *testing.T) {
	orig := jwkParseKeyFn

	defer func() { jwkParseKeyFn = orig }()

	sm := setupJWSSessionManager(t)

	jwkParseKeyFn = func(_ []byte, _ ...joseJwk.ParseOption) (joseJwk.Key, error) {
		return nil, fmt.Errorf("injected parse error")
	}

	_, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWK")
}

// TestValidateJWESession_ParseJWKError covers the JWK parse error in JWE validate.
func TestValidateJWESession_ParseJWKError(t *testing.T) {
	orig := jwkParseKeyFn

	defer func() { jwkParseKeyFn = orig }()

	sm := setupJWESessionManager(t)

	// Issue a valid token first.
	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Now inject parse failure for validate path.
	jwkParseKeyFn = func(_ []byte, _ ...joseJwk.ParseOption) (joseJwk.Key, error) {
		return nil, fmt.Errorf("injected parse error")
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// TestValidateJWSSession_ParseJWKError covers the JWK parse error in JWS validate.
func TestValidateJWSSession_ParseJWKError(t *testing.T) {
	orig := jwkParseKeyFn

	defer func() { jwkParseKeyFn = orig }()

	sm := setupJWSSessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	jwkParseKeyFn = func(_ []byte, _ ...joseJwk.ParseOption) (joseJwk.Key, error) {
		return nil, fmt.Errorf("injected parse error")
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// =============================================================================
// Verify/decrypt result error tests (non-JSON payloads trigger unmarshal errors).
// =============================================================================

// TestValidateJWESession_UnmarshalError covers the claims unmarshal error in JWE validate.
func TestValidateJWESession_UnmarshalError(t *testing.T) {
	orig := decryptBytesFn

	defer func() { decryptBytesFn = orig }()

	sm := setupJWESessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Override decryptBytesFn to return non-JSON bytes, triggering unmarshal error.
	decryptBytesFn = func(_ []joseJwk.Key, _ []byte) ([]byte, error) {
		return []byte("not-valid-json"), nil
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// TestValidateJWSSession_UnmarshalError covers the claims unmarshal error in JWS validate.
func TestValidateJWSSession_UnmarshalError(t *testing.T) {
	orig := verifyBytesFn

	defer func() { verifyBytesFn = orig }()

	sm := setupJWSSessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Override verifyBytesFn to return non-JSON bytes, triggering unmarshal error.
	verifyBytesFn = func(_ []joseJwk.Key, _ []byte) ([]byte, error) {
		return []byte("not-valid-json"), nil
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// =============================================================================
// JWS validate PublicKey error test.
// =============================================================================

// TestValidateJWSSession_PublicKeyError covers the PublicKey extraction error.
func TestValidateJWSSession_PublicKeyError(t *testing.T) {
	orig := jwkParseKeyFn

	defer func() { jwkParseKeyFn = orig }()

	sm := setupJWSSessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Override jwkParseKeyFn to return a symmetric (HMAC) key.
	// Symmetric keys do not support PublicKey(), which triggers the error.
	jwkParseKeyFn = func(_ []byte, _ ...joseJwk.ParseOption) (joseJwk.Key, error) {
		return cryptoutilSharedCryptoJose.GenerateHMACJWK(cryptoutilSharedMagic.HMACKeySize256)
	}

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// =============================================================================
// DB query/update errors in validate paths.
// =============================================================================

// TestValidateJWESession_SessionNotFound covers the session not found error (non-DB error).
func TestValidateJWESession_SessionNotFound(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Delete all sessions to trigger "not found" error.
	require.NoError(t, sm.db.Exec("DELETE FROM browser_sessions").Error)

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// TestValidateJWSSession_SessionNotFound covers session not found after hash lookup.
func TestValidateJWSSession_SessionNotFound(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Delete all sessions.
	require.NoError(t, sm.db.Exec("DELETE FROM browser_sessions").Error)

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
}

// TestValidateJWESession_QuerySessionDBError covers non-RecordNotFound DB error in JWE validate.
func TestValidateJWESession_QuerySessionDBError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Drop session table to trigger generic DB error (not ErrRecordNotFound).
	require.NoError(t, sm.db.Exec("DROP TABLE browser_sessions").Error)

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to query session")
}

// TestValidateJWSSession_QuerySessionDBError covers non-RecordNotFound DB error in JWS validate.
func TestValidateJWSSession_QuerySessionDBError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Drop session table to trigger generic DB error.
	require.NoError(t, sm.db.Exec("DROP TABLE browser_sessions").Error)

	_, err = sm.ValidateBrowserSession(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to query session")
}

// =============================================================================
// Validate update activity error tests.
// Uses GORM callback override to force update errors while queries still work.
// =============================================================================

// TestValidateJWESession_UpdateActivityError covers the update activity warning path.
func TestValidateJWESession_UpdateActivityError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Rename the column to force update error while query still works.
	require.NoError(t, sm.db.Exec("ALTER TABLE browser_sessions RENAME COLUMN last_activity TO last_activity_old").Error)

	session, err := sm.ValidateBrowserSession(context.Background(), token)
	// The validate should succeed (update error is just logged, not returned).
	require.NoError(t, err)
	require.NotNil(t, session)
}

// TestValidateJWSSession_UpdateActivityError covers the update activity warning path.
func TestValidateJWSSession_UpdateActivityError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Rename the column to force update error while query still works.
	require.NoError(t, sm.db.Exec("ALTER TABLE browser_sessions RENAME COLUMN last_activity TO last_activity_old").Error)

	session, err := sm.ValidateBrowserSession(context.Background(), token)
	require.NoError(t, err)
	require.NotNil(t, session)
}

// TestValidateOPAQUESession_UpdateActivityError covers the update activity warning path.
func TestValidateOPAQUESession_UpdateActivityError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	token, err := sm.IssueBrowserSession(context.Background(), "user1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.NoError(t, err)

	// Rename the column to force update error while query still works.
	require.NoError(t, sm.db.Exec("ALTER TABLE browser_sessions RENAME COLUMN last_activity TO last_activity_old").Error)

	session, err := sm.ValidateBrowserSession(context.Background(), token)
	require.NoError(t, err)
	require.NotNil(t, session)
}

// =============================================================================
// Service-side issue/validate create error tests.
// =============================================================================

// TestIssueServiceSession_OPAQUE_CreateError covers service OPAQUE create error.
func TestIssueServiceSession_OPAQUE_CreateError(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	// Drop service session table to force create error.
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS service_sessions").Error)

	_, err := sm.IssueServiceSession(context.Background(), "client1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to store session")
}

// TestIssueServiceSession_JWE_CreateError covers service JWE create error.
func TestIssueServiceSession_JWE_CreateError(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)

	// Drop service session table to force create error.
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS service_sessions").Error)

	_, err := sm.IssueServiceSession(context.Background(), "client1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
}

// TestIssueServiceSession_JWS_CreateError covers service JWS create error.
func TestIssueServiceSession_JWS_CreateError(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)

	// Drop service session table to force create error.
	require.NoError(t, sm.db.Exec("DROP TABLE IF EXISTS service_sessions").Error)

	_, err := sm.IssueServiceSession(context.Background(), "client1", googleUuid.Must(googleUuid.NewV7()), googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
}

// =============================================================================
// Suppress unused import warnings.
// =============================================================================

var (
	_ = (*joseJwe.Message)(nil)
	_ = (*joseJws.Message)(nil)
)

// =============================================================================
// Keygen error path tests (injectable keygen vars in session_manager_session.go).
// =============================================================================

// TestGenerateJWSKey_RSAKeygenError covers RSA keygen failure in generateJWSKey.
func TestGenerateJWSKey_RSAKeygenError(t *testing.T) {
	orig := generateRSAKeyPairSessionFn

	defer func() { generateRSAKeyPairSessionFn = orig }()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	generateRSAKeyPairSessionFn = func(_ int) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return nil, fmt.Errorf("injected RSA keygen error")
	}

	_, err := sm.generateJWSKey(cryptoutilSharedMagic.SessionJWSAlgorithmRS256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate RSA key pair")
}

// TestGenerateJWEKey_AESKeygenError covers AES keygen failure in generateJWEKey.
func TestGenerateJWEKey_AESKeygenError(t *testing.T) {
	orig := generateAESKeySessionFn

	defer func() { generateAESKeySessionFn = orig }()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	generateAESKeySessionFn = func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
		return nil, fmt.Errorf("injected AES keygen error")
	}

	_, err := sm.generateJWEKey(cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate AES key")
}

// =============================================================================
// Service-side store error test (initializeSessionJWK with isBrowser=false).
// =============================================================================

// TestInitializeSessionJWK_ServiceStoreError covers the service-side JWK store error path.
func TestInitializeSessionJWK_ServiceStoreError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		ServiceSessionJWSAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
	}

	sm := NewSessionManager(db, nil, config)

	// Add trigger to prevent INSERT on service_session_jwks (forces store error).
	require.NoError(t, db.Exec(`
		CREATE TRIGGER prevent_service_jwk_insert BEFORE INSERT ON service_session_jwks
		BEGIN
			SELECT RAISE(ABORT, 'forced insert error');
		END;
	`).Error)

	_, err := sm.initializeSessionJWK(context.Background(), false, cryptoutilSharedMagic.SessionAlgorithmJWS)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to store JWK")
}
