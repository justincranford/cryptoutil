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
	t.Parallel()

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
	require.Contains(t, claims, cryptoutilSharedMagic.ClaimJti)
	require.Contains(t, claims, cryptoutilSharedMagic.ClaimIat)
	require.Contains(t, claims, cryptoutilSharedMagic.ClaimExp)
	require.Contains(t, claims, cryptoutilSharedMagic.ClaimSub)
	require.Contains(t, claims, "tenant_id")
	require.Contains(t, claims, "realm_id")
	require.Equal(t, userID, claims[cryptoutilSharedMagic.ClaimSub])
	require.Equal(t, tenantID.String(), claims["tenant_id"])
	require.Equal(t, realmID.String(), claims["realm_id"])
}

func TestSessionManager_ValidateBrowserSession_JWE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupToken       func(t *testing.T, sm *SessionManager, ctx context.Context) string
		wantErr          bool
		wantErrorMessage string
	}{
		{
			name: "success - valid JWE token",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				userID := googleUuid.Must(googleUuid.NewV7()).String()
				tenantID := googleUuid.Must(googleUuid.NewV7())
				realmID := googleUuid.Must(googleUuid.NewV7())
				token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name: "invalid token - malformed JWE",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				return "invalid.jwe.token"
			},
			wantErr:          true,
			wantErrorMessage: "Invalid session token",
		},
		{
			name: "expired JWT",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				// Load JWK from database to encrypt custom token
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

				// Create expired JWT
				now := time.Now().UTC()
				exp := now.Add(-1 * time.Hour) // Expired 1 hour ago
				jti := googleUuid.Must(googleUuid.NewV7())

				claims := map[string]any{
					cryptoutilSharedMagic.ClaimJti: jti.String(),
					cryptoutilSharedMagic.ClaimIat: now.Add(-2 * time.Hour).Unix(),
					cryptoutilSharedMagic.ClaimExp: exp.Unix(),
					cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
					"tenant_id":                    googleUuid.Must(googleUuid.NewV7()).String(),
					"realm_id":                     googleUuid.Must(googleUuid.NewV7()).String(),
				}

				claimsBytes, _ := json.Marshal(claims)
				_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
				require.NoError(t, encryptErr)

				return string(jweBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Session expired",
		},
		{
			name: "revoked session",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				userID := googleUuid.Must(googleUuid.NewV7()).String()
				tenantID := googleUuid.Must(googleUuid.NewV7())
				realmID := googleUuid.Must(googleUuid.NewV7())

				// Issue session
				token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
				require.NoError(t, err)

				// Load JWK and decrypt to get jti
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

				jtiStr, ok := claims[cryptoutilSharedMagic.ClaimJti].(string)
				require.True(t, ok, "jti claim should be string")

				// Delete session from database (simulate revocation)
				jti, parseJTIErr := googleUuid.Parse(jtiStr)
				require.NoError(t, parseJTIErr)

				deleteErr := sm.db.Where("id = ?", jti).Delete(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).Error
				require.NoError(t, deleteErr)

				return token
			},
			wantErr:          true,
			wantErrorMessage: "Session revoked or not found",
		},
		{
			name: "missing exp claim",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				// Load JWK from database to encrypt custom token
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

				// Create JWT without exp claim
				now := time.Now().UTC()
				jti := googleUuid.Must(googleUuid.NewV7())

				claims := map[string]any{
					cryptoutilSharedMagic.ClaimJti: jti.String(),
					cryptoutilSharedMagic.ClaimIat: now.Unix(),
					// No exp claim - intentionally missing
					cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
					"tenant_id":                    googleUuid.Must(googleUuid.NewV7()).String(),
					"realm_id":                     googleUuid.Must(googleUuid.NewV7()).String(),
				}

				claimsBytes, _ := json.Marshal(claims)
				_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
				require.NoError(t, encryptErr)

				return string(jweBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Missing or invalid exp claim",
		},
		{
			name: "missing jti claim",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				// Load JWK from database to encrypt custom token
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

				// Create JWT without jti claim
				now := time.Now().UTC()
				exp := now.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)

				claims := map[string]any{
					// No jti claim - intentionally missing
					cryptoutilSharedMagic.ClaimIat: now.Unix(),
					cryptoutilSharedMagic.ClaimExp: exp.Unix(),
					cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
					"tenant_id":                    googleUuid.Must(googleUuid.NewV7()).String(),
					"realm_id":                     googleUuid.Must(googleUuid.NewV7()).String(),
				}

				claimsBytes, _ := json.Marshal(claims)
				_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
				require.NoError(t, encryptErr)

				return string(jweBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Missing or invalid jti claim",
		},
		{
			name: "invalid jti format - not a UUID",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				// Load JWK from database to encrypt custom token
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

				// Create JWT with invalid jti format
				now := time.Now().UTC()
				exp := now.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)

				claims := map[string]any{
					cryptoutilSharedMagic.ClaimJti: "not-a-valid-uuid",
					cryptoutilSharedMagic.ClaimIat: now.Unix(),
					cryptoutilSharedMagic.ClaimExp: exp.Unix(),
					cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
					"tenant_id":                    googleUuid.Must(googleUuid.NewV7()).String(),
					"realm_id":                     googleUuid.Must(googleUuid.NewV7()).String(),
				}

				claimsBytes, _ := json.Marshal(claims)
				_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
				require.NoError(t, encryptErr)

				return string(jweBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Invalid jti",
		},
		{
			name: "invalid exp type - not a number",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				// Load JWK from database to encrypt custom token
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

				// Create JWT with non-numeric exp claim
				now := time.Now().UTC()
				jti := googleUuid.Must(googleUuid.NewV7())

				claims := map[string]any{
					cryptoutilSharedMagic.ClaimJti: jti.String(),
					cryptoutilSharedMagic.ClaimIat: now.Unix(),
					cryptoutilSharedMagic.ClaimExp: "not-a-number", // Invalid type
					cryptoutilSharedMagic.ClaimSub: googleUuid.Must(googleUuid.NewV7()).String(),
					"tenant_id":                    googleUuid.Must(googleUuid.NewV7()).String(),
					"realm_id":                     googleUuid.Must(googleUuid.NewV7()).String(),
				}

				claimsBytes, _ := json.Marshal(claims)
				_, jweBytes, encryptErr := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jwk}, claimsBytes)
				require.NoError(t, encryptErr)

				return string(jweBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Missing or invalid exp claim",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
			ctx := context.Background()

			token := tc.setupToken(t, sm, ctx)

			session, validateErr := sm.ValidateBrowserSession(ctx, token)

			if tc.wantErr {
				require.Error(t, validateErr)
				require.Nil(t, session)

				if tc.wantErrorMessage != "" {
					require.Contains(t, validateErr.Error(), tc.wantErrorMessage)
				}
			} else {
				require.NoError(t, validateErr)
				require.NotNil(t, session)
			}
		})
	}
}

func TestSessionManager_IssueServiceSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWE)
	t.Parallel()

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

	require.Equal(t, clientID, claims[cryptoutilSharedMagic.ClaimSub])
	require.Equal(t, tenantID.String(), claims["tenant_id"])
	require.Equal(t, realmID.String(), claims["realm_id"])
}

func TestSessionManager_ValidateServiceSession_JWE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWE)
	t.Parallel()

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
