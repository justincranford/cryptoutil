// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"context"
	json "encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const authzCoverageJWTToken = "header.payload.signature"

type authzIntrospectionTokenService struct {
	claims map[string]any
	err    error
}

func (*authzIntrospectionTokenService) IssueAccessToken(context.Context, map[string]any) (string, error) {
	return "", nil
}

func (*authzIntrospectionTokenService) IssueRefreshToken(context.Context) (string, error) {
	return "", nil
}

func (s *authzIntrospectionTokenService) ValidateAccessToken(context.Context, string) (map[string]any, error) {
	return s.claims, s.err
}

func (*authzIntrospectionTokenService) GetPublicKeys() []map[string]any {
	return nil
}

func TestHandleIntrospect_JWTClaimsAndValidationFallback(t *testing.T) {
	t.Parallel()

	issuedAt := float64(1_710_000_000)
	expiresAt := float64(1_710_003_600)
	clientID := "coverage-client"
	subject := "coverage-user"
	scope := "openid profile"
	issuer := "https://issuer.example.test"
	jti := "coverage-jti"
	audience := []string{"coverage-audience"}

	tests := []struct {
		name          string
		claims        map[string]any
		validationErr error
		storeRevoked  bool
		tokenTypeHint string
		wantActive    bool
		wantClaims    bool
		wantTokenHint bool
	}{
		{
			name: "valid jwt returns standard claims and token hint",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimClientID: clientID,
				cryptoutilSharedMagic.ClaimSub:      subject,
				cryptoutilSharedMagic.ClaimScope:    scope,
				cryptoutilSharedMagic.ClaimExp:      expiresAt,
				cryptoutilSharedMagic.ClaimIat:      issuedAt,
				cryptoutilSharedMagic.ClaimIss:      issuer,
				cryptoutilSharedMagic.ClaimAud:      audience,
				cryptoutilSharedMagic.ClaimJti:      jti,
			},
			tokenTypeHint: cryptoutilSharedMagic.TokenTypeAccessToken,
			wantActive:    true,
			wantClaims:    true,
			wantTokenHint: true,
		},
		{
			name:         "valid jwt with revoked database record is inactive",
			claims:       map[string]any{},
			storeRevoked: true,
		},
		{
			name:          "invalid jwt falls back to opaque token lookup",
			validationErr: errors.New("invalid signature"),
		},
		{
			name:       "valid jwt without optional claims remains active",
			claims:     map[string]any{},
			wantActive: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createIntrospectTestConfig(t)
			repoFactory := createIntrospectTestRepoFactory(t)

			if tc.storeRevoked {
				now := time.Now().UTC()
				testClient := createIntrospectRevokeTestClient(t, repoFactory)
				token := &cryptoutilIdentityDomain.Token{
					ID:         googleUuid.Must(googleUuid.NewV7()),
					TokenValue: authzCoverageJWTToken,
					TokenType:  cryptoutilIdentityDomain.TokenTypeAccess,
					ClientID:   testClient.ID,
					ExpiresAt:  now.Add(time.Hour),
					IssuedAt:   now,
					Revoked:    true,
				}
				require.NoError(t, repoFactory.TokenRepository().Create(context.Background(), token))
			}

			tokenService := &authzIntrospectionTokenService{claims: tc.claims, err: tc.validationErr}
			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenService)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			form := url.Values{cryptoutilSharedMagic.ParamToken: []string{authzCoverageJWTToken}}
			if tc.tokenTypeHint != "" {
				form.Set(cryptoutilSharedMagic.ParamTokenTypeHint, tc.tokenTypeHint)
			}

			request := httptest.NewRequest(testPost, testOauth2V1Introspect, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusOK, response.StatusCode)

			var body map[string]any
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			require.Equal(t, tc.wantActive, body["active"])
			require.Equal(t, tc.wantClaims, body[cryptoutilSharedMagic.ClaimClientID] != nil)
			require.Equal(t, tc.wantTokenHint, body[cryptoutilSharedMagic.ParamTokenTypeHint] != nil)

			if tc.wantClaims {
				require.Equal(t, clientID, body[cryptoutilSharedMagic.ClaimClientID])
				require.Equal(t, subject, body[cryptoutilSharedMagic.ClaimSub])
				require.Equal(t, scope, body[cryptoutilSharedMagic.ClaimScope])
				require.Equal(t, expiresAt, body[cryptoutilSharedMagic.ClaimExp])
				require.Equal(t, issuedAt, body[cryptoutilSharedMagic.ClaimIat])
				require.Equal(t, issuer, body[cryptoutilSharedMagic.ClaimIss])
				require.Equal(t, []any{audience[0]}, body[cryptoutilSharedMagic.ClaimAud])
				require.Equal(t, jti, body[cryptoutilSharedMagic.ClaimJti])
			}

			if tc.wantTokenHint {
				require.Equal(t, tc.tokenTypeHint, body[cryptoutilSharedMagic.ParamTokenTypeHint])
			}
		})
	}
}
