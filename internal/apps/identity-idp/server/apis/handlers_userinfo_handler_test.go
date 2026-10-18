// Copyright (c) 2025-2026 Justin Cranford.

package apis

import (
	"context"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	userInfoIssueJWTToken          = "jwt"
	userInfoIssueJWTWithoutSubject = "jwt-without-subject"
)

func TestService_HandleUserInfoPaths(t *testing.T) {
	t.Parallel()

	uuidTokenService, signedJWTTokenService, unsignedJWTTokenService := newUserInfoHandlerTokenServices(t)

	tests := []struct {
		name           string
		tokenService   *cryptoutilIdentityIssuer.TokenService
		authorization  string
		tokenKind      string
		accept         string
		createUser     bool
		tokenExpiresAt time.Time
		wantStatus     int
		wantContent    string
		wantJWT        bool
	}{
		{name: "missing authorization header", tokenService: uuidTokenService, wantStatus: http.StatusUnauthorized},
		{name: "invalid authorization scheme", tokenService: uuidTokenService, authorization: testBasicAuthorizationHeader, wantStatus: http.StatusUnauthorized},
		{name: testInvalidUUIDTokenCase, tokenService: uuidTokenService, authorization: cryptoutilSharedMagic.AuthorizationBearerPrefix + testInvalidUUIDValue, wantStatus: http.StatusUnauthorized},
		{name: "UUID token missing from database", tokenService: uuidTokenService, tokenKind: cryptoutilSharedMagic.TokenFormatUUID, wantStatus: http.StatusUnauthorized},
		{name: "expired UUID token", tokenService: uuidTokenService, tokenKind: cryptoutilSharedMagic.TokenFormatUUID, tokenExpiresAt: time.Now().UTC().Add(-time.Hour), wantStatus: http.StatusUnauthorized},
		{name: "user missing for UUID token", tokenService: uuidTokenService, tokenKind: cryptoutilSharedMagic.TokenFormatUUID, tokenExpiresAt: time.Now().UTC().Add(time.Hour), wantStatus: http.StatusNotFound},
		{name: "JSON response from UUID token", tokenService: uuidTokenService, tokenKind: cryptoutilSharedMagic.TokenFormatUUID, tokenExpiresAt: time.Now().UTC().Add(time.Hour), createUser: true, wantStatus: http.StatusOK, wantContent: testJSONContentType},
		{name: "JWT response from JWT access token", tokenService: signedJWTTokenService, tokenKind: userInfoIssueJWTToken, accept: MIMEApplicationJWT, createUser: true, wantStatus: http.StatusOK, wantContent: MIMEApplicationJWT, wantJWT: true},
		{name: "JWT signing failure falls back to JSON", tokenService: unsignedJWTTokenService, tokenKind: cryptoutilSharedMagic.TokenFormatUUID, accept: MIMEApplicationJWT, tokenExpiresAt: time.Now().UTC().Add(time.Hour), createUser: true, wantStatus: http.StatusOK, wantContent: testJSONContentType},
		{name: "JWT token missing subject", tokenService: signedJWTTokenService, tokenKind: userInfoIssueJWTWithoutSubject, wantStatus: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, &cryptoutilIdentityConfig.DatabaseConfig{
				Type: cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
			})
			require.NoError(t, err)
			require.NoError(t, repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.User{}, &cryptoutilIdentityDomain.Token{}))

			service := &Service{repoFactory: repoFactory, tokenSvc: tc.tokenService}
			app := fiber.New()
			app.Get("/", service.handleUserInfo)
			t.Cleanup(func() {
				require.NoError(t, repoFactory.Close())
			})

			tokenValue := ""

			switch tc.tokenKind {
			case cryptoutilSharedMagic.TokenFormatUUID:
				tokenValue, err = tc.tokenService.IssueAccessToken(ctx, nil)
				require.NoError(t, err)
			case userInfoIssueJWTToken:
				userID := googleUuid.Must(googleUuid.NewV7())
				tokenValue, err = tc.tokenService.IssueAccessToken(ctx, map[string]any{
					cryptoutilSharedMagic.ClaimAud:   testUserInfoClientID,
					cryptoutilSharedMagic.ClaimSub:   userID.String(),
					cryptoutilSharedMagic.ClaimScope: "profile email",
				})
				require.NoError(t, err)

				tc.authorization = cryptoutilSharedMagic.AuthorizationBearerPrefix + tokenValue

				if tc.createUser {
					require.NoError(t, repoFactory.UserRepository().Create(ctx, &cryptoutilIdentityDomain.User{
						ID:    userID,
						Sub:   userID.String(),
						Name:  testUserInfoDisplayName,
						Email: testUserInfoEmail,
					}))
				}
			case userInfoIssueJWTWithoutSubject:
				tokenValue, err = tc.tokenService.IssueAccessToken(ctx, map[string]any{
					cryptoutilSharedMagic.ClaimAud: testUserInfoClientID,
				})
				require.NoError(t, err)
			}

			if tc.tokenExpiresAt.After(time.Time{}) {
				userID := googleUuid.Must(googleUuid.NewV7())
				if tc.createUser {
					user := &cryptoutilIdentityDomain.User{
						ID:    userID,
						Sub:   userID.String(),
						Name:  testUserInfoDisplayName,
						Email: testUserInfoEmail,
					}
					require.NoError(t, repoFactory.UserRepository().Create(ctx, user))
				}

				clientID := googleUuid.Must(googleUuid.NewV7())
				require.NoError(t, repoFactory.TokenRepository().Create(ctx, &cryptoutilIdentityDomain.Token{
					TokenValue:  tokenValue,
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    clientID,
					UserID:      cryptoutilIdentityDomain.NewNullableUUID(&userID),
					ExpiresAt:   tc.tokenExpiresAt,
				}))
			}

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			if tc.tokenKind == cryptoutilSharedMagic.TokenFormatUUID {
				tc.authorization = cryptoutilSharedMagic.AuthorizationBearerPrefix + tokenValue
			}

			request.Header.Set("Authorization", tc.authorization)

			if tc.accept != "" {
				request.Header.Set("Accept", tc.accept)
			}

			response, err := app.Test(request, -1)
			require.NoError(t, err)

			defer func() { _ = response.Body.Close() }() //nolint:errcheck // Test cleanup.

			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.wantContent != "" {
				require.Contains(t, response.Header.Get(fiber.HeaderContentType), tc.wantContent)
			}

			if tc.wantJWT {
				responseBody, readErr := io.ReadAll(response.Body)
				require.NoError(t, readErr)
				require.Equal(t, cryptoutilSharedMagic.JWSPartCount-1, strings.Count(string(responseBody), "."))
				claims, validateErr := tc.tokenService.ValidateIDToken(ctx, string(responseBody))
				require.NoError(t, validateErr)
				require.NotEmpty(t, claims[cryptoutilSharedMagic.ClaimSub])
			}
		})
	}
}

func newUserInfoHandlerTokenServices(t *testing.T) (
	*cryptoutilIdentityIssuer.TokenService,
	*cryptoutilIdentityIssuer.TokenService,
	*cryptoutilIdentityIssuer.TokenService,
) {
	t.Helper()

	ctx := context.Background()
	keyRotationManager, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, keyRotationManager.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm))

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://userinfo.example",
		keyRotationManager,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		time.Hour,
		time.Hour,
	)
	require.NoError(t, err)

	unsignedKeyRotationManager, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err)
	unsignedJWSIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		"https://userinfo.example",
		unsignedKeyRotationManager,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		time.Hour,
		time.Hour,
	)
	require.NoError(t, err)

	return cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, cryptoutilIdentityIssuer.NewUUIDIssuer(), &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: cryptoutilSharedMagic.IdentityTokenFormatUUID,
		}),
		cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, nil, &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: cryptoutilSharedMagic.IdentityTokenFormatJWS,
		}),
		cryptoutilIdentityIssuer.NewTokenService(unsignedJWSIssuer, nil, cryptoutilIdentityIssuer.NewUUIDIssuer(), &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: cryptoutilSharedMagic.IdentityTokenFormatUUID,
		})
}
