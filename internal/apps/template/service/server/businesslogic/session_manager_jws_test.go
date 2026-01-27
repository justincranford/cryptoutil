// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	json "encoding/json"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSessionManager_IssueBrowserSession_JWS_RS256_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Parse token as JWT
	var claims map[string]any

	// Load JWK from database to verify signature
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	privateJWK, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	// Extract public key from private JWK for verification
	publicJWK, publicKeyErr := privateJWK.PublicKey()
	require.NoError(t, publicKeyErr)

	// Verify JWS signature
	claimsBytes, verifyErr := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{publicJWK}, []byte(token))
	require.NoError(t, verifyErr)

	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	require.NoError(t, unmarshalErr)

	// Validate JWT claims
	require.Contains(t, claims, "jti")
	require.Contains(t, claims, "iat")
	require.Contains(t, claims, "exp")
	require.Contains(t, claims, "sub")
	require.Contains(t, claims, "tenant_id")
	require.Contains(t, claims, "realm_id")

	require.Equal(t, userID, claims["sub"])
	require.Equal(t, tenantID.String(), claims["tenant_id"])
	require.Equal(t, realmID.String(), claims["realm_id"])

	// Verify expiration is in future.
	expFloat, ok := claims["exp"].(float64)
	require.True(t, ok, "exp claim should be float64")

	exp := time.Unix(int64(expFloat), 0)
	require.True(t, time.Now().UTC().Before(exp), "Expiration should be in future")
}

func TestSessionManager_ValidateBrowserSession_JWS_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
}

func TestSessionManager_ValidateBrowserSession_JWS_InvalidSignature(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Create invalid JWT (malformed token)
	invalidToken := "invalid.jwt.token"

	session, err := sm.ValidateBrowserSession(ctx, invalidToken)
	require.Error(t, err)
	require.Nil(t, session)
}

func TestSessionManager_ValidateBrowserSession_JWS_ExpiredJWT(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Load JWK from database
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	jwk, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	// Create expired JWT manually
	now := time.Now().UTC()
	exp := now.Add(-1 * time.Hour) // Already expired
	jti := googleUuid.Must(googleUuid.NewV7())

	claims := map[string]any{
		"jti":       jti.String(),
		"iat":       now.Add(-2 * time.Hour).Unix(),
		"exp":       exp.Unix(),
		"sub":       userID,
		"tenant_id": tenantID.String(),
		"realm_id":  realmID.String(),
	}

	claimsBytes, _ := json.Marshal(claims)
	_, jwsBytes, signErr := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, signErr)

	// Validate should fail due to expiration
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jwsBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "JWT expired")
}

func TestSessionManager_ValidateBrowserSession_JWS_RevokedSession(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Parse token to extract jti
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	privateJWK, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	publicJWK, publicKeyErr := privateJWK.PublicKey()
	require.NoError(t, publicKeyErr)

	claimsBytes, verifyErr := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{publicJWK}, []byte(token))
	require.NoError(t, verifyErr)

	var claims map[string]any

	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	require.NoError(t, unmarshalErr)

	jtiStr, ok := claims["jti"].(string)
	require.True(t, ok, "jti claim should be string")

	// Delete session from database (simulate revocation).
	jti, parseJTIErr := googleUuid.Parse(jtiStr)
	require.NoError(t, parseJTIErr)

	deleteErr := sm.db.Where("id = ?", jti).Delete(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).Error
	require.NoError(t, deleteErr)

	// Validate should fail (session revoked)
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Session revoked or not found")
}

func TestSessionManager_IssueServiceSession_JWS_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWS)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Load JWK and verify signature
	var serviceJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("service_session_jwks").
		Where("id = ?", sm.serviceJWKID).
		Select("encrypted_jwk").
		First(&serviceJWK).Error
	require.NoError(t, findErr)

	privateJWK, parseErr := joseJwk.ParseKey([]byte(serviceJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	publicJWK, publicKeyErr := privateJWK.PublicKey()
	require.NoError(t, publicKeyErr)

	claimsBytes, verifyErr := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{publicJWK}, []byte(token))
	require.NoError(t, verifyErr)

	var claims map[string]any

	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	require.NoError(t, unmarshalErr)

	require.Equal(t, clientID, claims["sub"])
	require.Equal(t, tenantID.String(), claims["tenant_id"])
	require.Equal(t, realmID.String(), claims["realm_id"])
}

func TestSessionManager_ValidateServiceSession_JWS_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWS)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateServiceSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
}

// TestSessionManager_ValidateBrowserSession_JWS_MissingExpClaim tests validation when exp claim is missing.
func TestSessionManager_ValidateBrowserSession_JWS_MissingExpClaim(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to sign custom token.
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	jwk, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	// Create JWT without exp claim.
	now := time.Now().UTC()
	jti := googleUuid.Must(googleUuid.NewV7())

	claims := map[string]any{
		"jti": jti.String(),
		"iat": now.Unix(),
		// No exp claim - intentionally missing
		"sub":       googleUuid.Must(googleUuid.NewV7()).String(),
		"tenant_id": googleUuid.Must(googleUuid.NewV7()).String(),
		"realm_id":  googleUuid.Must(googleUuid.NewV7()).String(),
	}

	claimsBytes, _ := json.Marshal(claims)
	_, jwsBytes, signErr := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, signErr)

	// Validate should fail due to missing exp claim.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jwsBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Missing or invalid exp claim")
}

// TestSessionManager_ValidateBrowserSession_JWS_MissingJtiClaim tests validation when jti claim is missing.
func TestSessionManager_ValidateBrowserSession_JWS_MissingJtiClaim(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to sign custom token.
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	jwk, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	// Create JWT without jti claim.
	now := time.Now().UTC()
	exp := now.Add(24 * time.Hour)

	claims := map[string]any{
		// No jti claim - intentionally missing
		"iat":       now.Unix(),
		"exp":       exp.Unix(),
		"sub":       googleUuid.Must(googleUuid.NewV7()).String(),
		"tenant_id": googleUuid.Must(googleUuid.NewV7()).String(),
		"realm_id":  googleUuid.Must(googleUuid.NewV7()).String(),
	}

	claimsBytes, _ := json.Marshal(claims)
	_, jwsBytes, signErr := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, signErr)

	// Validate should fail due to missing jti claim.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jwsBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Missing or invalid jti claim")
}

// TestSessionManager_ValidateBrowserSession_JWS_InvalidJtiFormat tests validation when jti is not a valid UUID.
func TestSessionManager_ValidateBrowserSession_JWS_InvalidJtiFormat(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to sign custom token.
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	jwk, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	// Create JWT with invalid jti format.
	now := time.Now().UTC()
	exp := now.Add(24 * time.Hour)

	claims := map[string]any{
		"jti":       "not-a-valid-uuid",
		"iat":       now.Unix(),
		"exp":       exp.Unix(),
		"sub":       googleUuid.Must(googleUuid.NewV7()).String(),
		"tenant_id": googleUuid.Must(googleUuid.NewV7()).String(),
		"realm_id":  googleUuid.Must(googleUuid.NewV7()).String(),
	}

	claimsBytes, _ := json.Marshal(claims)
	_, jwsBytes, signErr := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, signErr)

	// Validate should fail due to invalid jti format.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jwsBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Invalid jti format")
}

// TestSessionManager_ValidateBrowserSession_JWS_InvalidExpType tests validation when exp claim is not a number.
func TestSessionManager_ValidateBrowserSession_JWS_InvalidExpType(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to sign custom token.
	var browserJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("browser_session_jwks").
		Where("id = ?", sm.browserJWKID).
		Select("encrypted_jwk").
		First(&browserJWK).Error
	require.NoError(t, findErr)

	jwk, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	// Create JWT with non-numeric exp claim.
	now := time.Now().UTC()
	jti := googleUuid.Must(googleUuid.NewV7())

	claims := map[string]any{
		"jti":       jti.String(),
		"iat":       now.Unix(),
		"exp":       "not-a-number", // Invalid type
		"sub":       googleUuid.Must(googleUuid.NewV7()).String(),
		"tenant_id": googleUuid.Must(googleUuid.NewV7()).String(),
		"realm_id":  googleUuid.Must(googleUuid.NewV7()).String(),
	}

	claimsBytes, _ := json.Marshal(claims)
	_, jwsBytes, signErr := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, signErr)

	// Validate should fail due to invalid exp type.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jwsBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Missing or invalid exp claim")
}
