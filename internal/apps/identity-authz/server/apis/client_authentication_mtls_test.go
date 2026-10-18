// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAuthenticateClient_MTLSWithoutPeerCertificate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		authMethod cryptoutilIdentityDomain.ClientAuthMethod
	}{
		{
			name:       "CA-validated client authentication",
			authMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
		},
		{
			name:       "self-signed client authentication",
			authMethod: cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createClientAuthFlowTestConfig(t)
			repoFactory := createClientAuthFlowTestRepoFactory(t, config)
			client := createClientAuthFlowTestClient(t, repoFactory, tc.authMethod)

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			app := fiber.New()
			service.RegisterRoutes(app)

			form := url.Values{
				cryptoutilSharedMagic.ParamGrantType: []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
				cryptoutilSharedMagic.ParamClientID:  []string{client.ClientID},
			}
			request := httptest.NewRequest("POST", testOauth2V1Token, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
		})
	}
}
