package businesslogic

import (
	"context"
	json "encoding/json"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestSessionManager_IssueBrowserSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Load JWK and decrypt
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

	// Decrypt JWE
	claimsBytes, decryptErr := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	require.NoError(t, decryptErr)

	var claims map[string]any

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
}

func TestSessionManager_ValidateBrowserSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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

func TestSessionManager_ValidateBrowserSession_JWE_InvalidToken(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Validate should fail with invalid token
	session, validateErr := sm.ValidateBrowserSession(ctx, "invalid.jwe.token")
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Invalid session token")
}

func TestSessionManager_ValidateBrowserSession_JWE_ExpiredJWT(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())
	jti := googleUuid.Must(googleUuid.NewV7())
	now := time.Now().UTC()
	exp := now.Add(-1 * time.Hour) // Expired 1 hour ago

	// Create expired JWT claims
	claims := map[string]any{
		"jti":       jti.String(),
		"iat":       now.Add(-2 * time.Hour).Unix(),
		"exp":       exp.Unix(),
		"sub":       userID,
		"tenant_id": tenantID.String(),
		"realm_id":  realmID.String(),
	}

	claimsBytes, _ := json.Marshal(claims)

	// Load JWK to encrypt
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

	// Encrypt claims
	_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, encryptErr)

	// Validate should fail due to expiration
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jweBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Session expired")
}

func TestSessionManager_ValidateBrowserSession_JWE_RevokedSession(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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

	jwk, parseErr := joseJwk.ParseKey([]byte(browserJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	claimsBytes, decryptErr := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	require.NoError(t, decryptErr)

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

func TestSessionManager_IssueServiceSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWE)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Load JWK and decrypt
	var serviceJWK struct {
		EncryptedJWK string
	}

	findErr := sm.db.Table("service_session_jwks").
		Where("id = ?", sm.serviceJWKID).
		Select("encrypted_jwk").
		First(&serviceJWK).Error
	require.NoError(t, findErr)

	jwk, parseErr := joseJwk.ParseKey([]byte(serviceJWK.EncryptedJWK))
	require.NoError(t, parseErr)

	claimsBytes, decryptErr := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	require.NoError(t, decryptErr)

	var claims map[string]any

	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	require.NoError(t, unmarshalErr)

	require.Equal(t, clientID, claims["sub"])
	require.Equal(t, tenantID.String(), claims["tenant_id"])
	require.Equal(t, realmID.String(), claims["realm_id"])
}

func TestSessionManager_ValidateServiceSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWE)
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

// TestSessionManager_ValidateBrowserSession_JWE_MissingExpClaim tests validation when exp claim is missing.
func TestSessionManager_ValidateBrowserSession_JWE_MissingExpClaim(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to encrypt custom token.
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
	_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, encryptErr)

	// Validate should fail due to missing exp claim.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jweBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Missing or invalid exp claim")
}

// TestSessionManager_ValidateBrowserSession_JWE_MissingJtiClaim tests validation when jti claim is missing.
func TestSessionManager_ValidateBrowserSession_JWE_MissingJtiClaim(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to encrypt custom token.
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
	_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, encryptErr)

	// Validate should fail due to missing jti claim.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jweBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Missing or invalid jti claim")
}

// TestSessionManager_ValidateBrowserSession_JWE_InvalidJtiFormat tests validation when jti is not a valid UUID.
func TestSessionManager_ValidateBrowserSession_JWE_InvalidJtiFormat(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to encrypt custom token.
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
	_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, encryptErr)

	// Validate should fail due to invalid jti format.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jweBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Invalid jti")
}

// TestSessionManager_ValidateBrowserSession_JWE_InvalidExpType tests validation when exp claim is not a number.
func TestSessionManager_ValidateBrowserSession_JWE_InvalidExpType(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Load JWK from database to encrypt custom token.
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
	_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, encryptErr)

	// Validate should fail due to invalid exp type.
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jweBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Missing or invalid exp claim")
}
