// Copyright (c) 2025 Justin Cranford
//
//

// Package businesslogic — additional coverage tests (part 2).
package businesslogic

import (
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// setupJWESessionManager creates a SessionManager initialized with JWE for both sessions.
func setupJWESessionManager(t *testing.T) *SessionManager {
	t.Helper()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWE),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWEAlgorithm: cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM,
		ServiceSessionJWEAlgorithm: cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	return sm
}

// encryptCustomJWEClaims creates a JWE-encrypted token using the SM's browser JWK with custom claims.
func encryptCustomJWEClaims(t *testing.T, sm *SessionManager, claimsJSON []byte) string {
	t.Helper()

	ctx := context.Background()

	var browserJWK cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK

	err := sm.db.WithContext(ctx).Where("id = ?", *sm.browserJWKID).First(&browserJWK).Error
	require.NoError(t, err)

	jwkBytes := []byte(browserJWK.EncryptedJWK)
	jwk, err := joseJwk.ParseKey(jwkBytes)
	require.NoError(t, err)

	_, encryptedBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsJSON)
	require.NoError(t, err)

	return string(encryptedBytes)
}

// TestValidateBrowserSession_JWE_NoExpClaim covers the missing exp claim branch
// in validateJWESession.
func TestValidateBrowserSession_JWE_NoExpClaim(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	claimsJSON := []byte(`{"jti":"` + googleUuid.Must(googleUuid.NewV7()).String() + `","sub":"user123"}`)
	token := encryptCustomJWEClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, err)
}

// TestValidateBrowserSession_JWE_NoJTIClaim covers the missing jti claim branch
// in validateJWESession.
func TestValidateBrowserSession_JWE_NoJTIClaim(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	futureExp := time.Now().UTC().Add(24 * time.Hour).Unix()
	claimsJSON := []byte(fmt.Sprintf(`{"exp":%d,"sub":"user123"}`, futureExp))
	token := encryptCustomJWEClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, err)
}

// TestValidateBrowserSession_JWE_InvalidJTI covers the invalid jti UUID branch
// in validateJWESession.
func TestValidateBrowserSession_JWE_InvalidJTI(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	futureExp := time.Now().UTC().Add(24 * time.Hour).Unix()
	claimsJSON := []byte(fmt.Sprintf(`{"exp":%d,"jti":"not-a-valid-uuid"}`, futureExp))
	token := encryptCustomJWEClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, err)
}

// TestCleanupExpiredSessions_BrowserSessionsError covers the error return in
// CleanupExpiredSessions when browser sessions delete fails.
func TestCleanupExpiredSessions_BrowserSessionsError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   time.Hour,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	// Close the underlying DB to force errors.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	ctx := context.Background()
	cleanupErr := sm.CleanupExpiredSessions(ctx)
	require.Error(t, cleanupErr)
	require.Contains(t, cleanupErr.Error(), "failed to cleanup browser sessions")
}

// TestValidateServiceSession_JWS_NoExpClaim covers the missing exp claim branch
// in validateJWSSession for service sessions.
func TestValidateServiceSession_JWS_NoExpClaim(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)

	ctx := context.Background()

	var serviceJWK cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK

	err := sm.db.WithContext(ctx).Where("id = ?", *sm.serviceJWKID).First(&serviceJWK).Error
	require.NoError(t, err)

	privateJWK, err := joseJwk.ParseKey([]byte(serviceJWK.EncryptedJWK))
	require.NoError(t, err)

	claimsJSON := []byte(`{"sub":"client123"}`)
	_, signedBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{privateJWK}, claimsJSON)
	require.NoError(t, err)

	_, valErr := sm.ValidateServiceSession(ctx, string(signedBytes))
	require.Error(t, valErr)
}

// encryptCustomJWEServiceClaims creates a JWE-encrypted token using the SM's service JWK.
func encryptCustomJWEServiceClaims(t *testing.T, sm *SessionManager, claimsJSON []byte) string {
	t.Helper()

	ctx := context.Background()

	var serviceJWK cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK

	err := sm.db.WithContext(ctx).Where("id = ?", *sm.serviceJWKID).First(&serviceJWK).Error
	require.NoError(t, err)

	jwkBytes := []byte(serviceJWK.EncryptedJWK)
	jwk, err := joseJwk.ParseKey(jwkBytes)
	require.NoError(t, err)

	_, encryptedBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsJSON)
	require.NoError(t, err)

	return string(encryptedBytes)
}

// TestValidateServiceSession_JWE_NoExpClaim covers the missing exp claim branch
// in validateJWESession for service sessions.
func TestValidateServiceSession_JWE_NoExpClaim(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	claimsJSON := []byte(`{"jti":"` + googleUuid.Must(googleUuid.NewV7()).String() + `","sub":"client123"}`)
	token := encryptCustomJWEServiceClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err := sm.ValidateServiceSession(ctx, token)
	require.Error(t, err)
}

// TestValidateServiceSession_JWE_NoJTIClaim covers the missing jti branch
// in validateJWESession for service sessions.
func TestValidateServiceSession_JWE_NoJTIClaim(t *testing.T) {
	t.Parallel()

	sm := setupJWESessionManager(t)
	futureExp := time.Now().UTC().Add(24 * time.Hour).Unix()
	claimsJSON := []byte(fmt.Sprintf(`{"exp":%d,"sub":"client123"}`, futureExp))
	token := encryptCustomJWEServiceClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err := sm.ValidateServiceSession(ctx, token)
	require.Error(t, err)
}

// TestStartCleanupTask_LogsError covers the error logging branch in
// StartCleanupTask when CleanupExpiredSessions returns an error.
func TestStartCleanupTask_LogsError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   50 * time.Millisecond, // Very short for testing.
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	// Close the underlying DB to force cleanup errors.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	// Start cleanup task; errors will be logged (not returned).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sm.StartCleanupTask(ctx)

	// Wait for at least one cleanup cycle to trigger the error logging branch.
	time.Sleep(200 * time.Millisecond)

	cancel()
	time.Sleep(50 * time.Millisecond)
}
