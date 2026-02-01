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
	t.Parallel()

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

func TestSessionManager_ValidateBrowserSession_JWS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupToken       func(t *testing.T, sm *SessionManager, ctx context.Context) string
		wantErr          bool
		wantErrorMessage string
	}{
		{
			name: "success - valid JWT token",
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
			name: "invalid signature - malformed JWT",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
				return "invalid.jwt.token"
			},
			wantErr:          true,
			wantErrorMessage: "",
		},
		{
			name: "expired JWT",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
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
					"sub":       googleUuid.Must(googleUuid.NewV7()).String(),
					"tenant_id": googleUuid.Must(googleUuid.NewV7()).String(),
					"realm_id":  googleUuid.Must(googleUuid.NewV7()).String(),
				}

				claimsBytes, _ := json.Marshal(claims)
				_, jwsBytes, signErr := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwk}, claimsBytes)
				require.NoError(t, signErr)

				return string(jwsBytes)
			},
			wantErr:          true,
			wantErrorMessage: "JWT expired",
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

				// Create JWT without exp claim
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

				return string(jwsBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Missing or invalid exp claim",
		},
		{
			name: "missing jti claim",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
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

				// Create JWT without jti claim
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

				return string(jwsBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Missing or invalid jti claim",
		},
		{
			name: "invalid jti format - not a UUID",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
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

				// Create JWT with invalid jti format
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

				return string(jwsBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Invalid jti format",
		},
		{
			name: "invalid exp type - not a number",
			setupToken: func(t *testing.T, sm *SessionManager, ctx context.Context) string {
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

				// Create JWT with non-numeric exp claim
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

				return string(jwsBytes)
			},
			wantErr:          true,
			wantErrorMessage: "Missing or invalid exp claim",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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

func TestSessionManager_IssueServiceSession_JWS_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmJWS)
	t.Parallel()

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
