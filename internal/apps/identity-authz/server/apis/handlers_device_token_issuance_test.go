// Copyright (c) 2025-2026 Justin Cranford.
//
//

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

const deviceTokenIssuanceErrorDescription = "Failed to issue token"

func TestHandleDeviceCodeGrant_TokenIssuance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		userIDValid            bool
		refreshTokenLifetime   int
		tokenServiceConfigured bool
		accessTokenErr         error
		refreshTokenErr        error
		wantStatus             int
		wantDescription        string
		wantAccessCalls        int
		wantRefreshCalls       int
		wantRefreshToken       bool
	}{
		{
			name:                 "token service is missing",
			userIDValid:          true,
			refreshTokenLifetime: 600,
			wantStatus:           fiber.StatusInternalServerError,
			wantDescription:      "Token service not configured",
		},
		{
			name:                   "authorized device has no user",
			refreshTokenLifetime:   600,
			tokenServiceConfigured: true,
			wantStatus:             fiber.StatusInternalServerError,
			wantDescription:        "Device authorization missing user ID",
		},
		{
			name:                   "access token issuance fails",
			userIDValid:            true,
			refreshTokenLifetime:   600,
			tokenServiceConfigured: true,
			accessTokenErr:         errors.New("access token failure"),
			wantStatus:             fiber.StatusInternalServerError,
			wantDescription:        deviceTokenIssuanceErrorDescription,
			wantAccessCalls:        1,
		},
		{
			name:                   "refresh token issuance fails",
			userIDValid:            true,
			refreshTokenLifetime:   600,
			tokenServiceConfigured: true,
			refreshTokenErr:        errors.New("refresh token failure"),
			wantStatus:             fiber.StatusInternalServerError,
			wantDescription:        deviceTokenIssuanceErrorDescription,
			wantAccessCalls:        1,
			wantRefreshCalls:       1,
		},
		{
			name:                   "successful token response includes refresh token",
			userIDValid:            true,
			refreshTokenLifetime:   600,
			tokenServiceConfigured: true,
			wantStatus:             fiber.StatusOK,
			wantAccessCalls:        1,
			wantRefreshCalls:       1,
			wantRefreshToken:       true,
		},
		{
			name:                   "successful token response without refresh token",
			userIDValid:            true,
			tokenServiceConfigured: true,
			wantStatus:             fiber.StatusOK,
			wantAccessCalls:        1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			config, repoFactory := createIntegrationTestDependencies(t)

			clientID := "device-token-client-" + googleUuid.NewString()
			client := &cryptoutilIdentityDomain.Client{
				ID:                   googleUuid.Must(googleUuid.NewV7()),
				ClientID:             clientID,
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
				RefreshTokenLifetime: tc.refreshTokenLifetime,
			}
			require.NoError(t, repoFactory.ClientRepository().Create(ctx, client))

			if tc.refreshTokenLifetime == 0 {
				require.NoError(t, repoFactory.DB().
					Model(&cryptoutilIdentityDomain.Client{}).
					Where("id = ?", client.ID).
					UpdateColumn("refresh_token_lifetime", 0).Error)
				storedClient, err := repoFactory.ClientRepository().GetByClientID(ctx, clientID)
				require.NoError(t, err)
				require.Zero(t, storedClient.RefreshTokenLifetime)
			}

			deviceCode := "device-code-" + googleUuid.NewString()

			deviceAuth := &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:         googleUuid.Must(googleUuid.NewV7()),
				ClientID:   clientID,
				DeviceCode: deviceCode,
				Status:     cryptoutilIdentityDomain.DeviceAuthStatusAuthorized,
				CreatedAt:  time.Now().UTC(),
				ExpiresAt:  time.Now().UTC().Add(time.Hour),
			}
			if tc.userIDValid {
				deviceAuth.UserID = cryptoutilIdentityDomain.NullableUUID{
					UUID:  googleUuid.Must(googleUuid.NewV7()),
					Valid: true,
				}
			}

			require.NoError(t, repoFactory.DeviceAuthorizationRepository().Create(ctx, deviceAuth))

			var (
				tokenService     cryptoutilIdentityAuthz.TokenService
				tokenServiceStub *authzCodeTokenServiceStub
			)
			if tc.tokenServiceConfigured {
				tokenServiceStub = &authzCodeTokenServiceStub{
					accessTokenErr:  tc.accessTokenErr,
					refreshTokenErr: tc.refreshTokenErr,
				}
				tokenService = tokenServiceStub
			}

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenService)
			app := fiber.New()
			service.RegisterRoutes(app)

			form := url.Values{
				cryptoutilSharedMagic.ParamGrantType:  []string{cryptoutilSharedMagic.GrantTypeDeviceCode},
				cryptoutilSharedMagic.ParamClientID:   []string{clientID},
				cryptoutilSharedMagic.ParamDeviceCode: []string{deviceCode},
			}
			req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.wantStatus, resp.StatusCode)

			var body map[string]any
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

			if tc.wantDescription != "" {
				require.Equal(t, tc.wantDescription, body["error_description"])
			} else {
				require.Equal(t, testAccessToken, body[cryptoutilSharedMagic.TokenTypeAccessToken])
				require.Equal(t, cryptoutilSharedMagic.AuthorizationBearer, body[cryptoutilSharedMagic.ParamTokenType])
				_, hasRefreshToken := body[cryptoutilSharedMagic.GrantTypeRefreshToken]
				require.Equal(t, tc.wantRefreshToken, hasRefreshToken)
			}

			if tokenServiceStub != nil {
				require.Equal(t, tc.wantAccessCalls, tokenServiceStub.accessTokenCalls)
				require.Equal(t, tc.wantRefreshCalls, tokenServiceStub.refreshTokenCalls)
			}

			storedDeviceAuth, err := repoFactory.DeviceAuthorizationRepository().GetByDeviceCode(ctx, deviceCode)
			require.NoError(t, err)
			require.Equal(t, cryptoutilIdentityDomain.DeviceAuthStatusUsed, storedDeviceAuth.Status)
		})
	}
}
