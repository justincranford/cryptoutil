package businesslogic

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cryptoutilRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilJOSE "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestSessionManager_IssueBrowserSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmJWE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "test-realm"

	token, err := sm.IssueBrowserSession(ctx, userID, realm)
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
	claimsBytes, decryptErr := cryptoutilJOSE.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	require.NoError(t, decryptErr)

	var claims map[string]interface{}
	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	require.NoError(t, unmarshalErr)

	// Validate JWT claims
	require.Contains(t, claims, "jti")
	require.Contains(t, claims, "iat")
	require.Contains(t, claims, "exp")
	require.Contains(t, claims, "sub")
	require.Contains(t, claims, "realm")
	require.Equal(t, userID, claims["sub"])
	require.Equal(t, realm, claims["realm"])
}

func TestSessionManager_ValidateBrowserSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmJWE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "test-realm"

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, realm)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
	require.NotNil(t, session.Realm)
	require.Equal(t, realm, *session.Realm)
}

func TestSessionManager_ValidateBrowserSession_JWE_InvalidToken(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmJWE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Validate should fail with invalid token
	session, validateErr := sm.ValidateBrowserSession(ctx, "invalid.jwe.token")
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Invalid session token")
}

func TestSessionManager_ValidateBrowserSession_JWE_ExpiredJWT(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmJWE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "test-realm"
	jti := googleUuid.Must(googleUuid.NewV7())
	now := time.Now()
	exp := now.Add(-1 * time.Hour) // Expired 1 hour ago

	// Create expired JWT claims
	claims := map[string]interface{}{
		"jti":   jti.String(),
		"iat":   now.Add(-2 * time.Hour).Unix(),
		"exp":   exp.Unix(),
		"sub":   userID,
		"realm": realm,
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
	_, jweBytes, encryptErr := cryptoutilJOSE.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
	require.NoError(t, encryptErr)

	// Validate should fail due to expiration
	session, validateErr := sm.ValidateBrowserSession(ctx, string(jweBytes))
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Session expired")
}

func TestSessionManager_ValidateBrowserSession_JWE_RevokedSession(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmJWE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "test-realm"

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, realm)
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

	claimsBytes, _ := cryptoutilJOSE.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	var claims map[string]interface{}
	_ = json.Unmarshal(claimsBytes, &claims)
	jtiStr := claims["jti"].(string)

	// Delete session from database (simulate revocation)
	jti, _ := googleUuid.Parse(jtiStr)
	deleteErr := sm.db.Where("id = ?", jti).Delete(&cryptoutilRepository.BrowserSession{}).Error
	require.NoError(t, deleteErr)

	// Validate should fail (session revoked)
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr)
	require.Nil(t, session)
	require.Contains(t, validateErr.Error(), "Session revoked or not found")
}

func TestSessionManager_IssueServiceSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmJWE)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "service-realm"

	token, err := sm.IssueServiceSession(ctx, clientID, realm)
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

	claimsBytes, decryptErr := cryptoutilJOSE.DecryptBytes([]joseJwk.Key{jwk}, []byte(token))
	require.NoError(t, decryptErr)

	var claims map[string]interface{}
	unmarshalErr := json.Unmarshal(claimsBytes, &claims)
	require.NoError(t, unmarshalErr)

	require.Equal(t, clientID, claims["sub"])
	require.Equal(t, realm, claims["realm"])
}

func TestSessionManager_ValidateServiceSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmJWE)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "service-realm"

	// Issue session
	token, err := sm.IssueServiceSession(ctx, clientID, realm)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateServiceSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
	require.NotNil(t, session.Realm)
	require.Equal(t, realm, *session.Realm)
}
