// Copyright (c) 2025-2026 Justin Cranford.
//

package apis

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity-authz/clientauth"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAuthenticateClient_RegistryMissingAuthenticator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		authMethod    cryptoutilIdentityDomain.ClientAuthMethod
		basicAuth     bool
		includeSecret bool
	}{
		{
			name:       "basic authentication",
			authMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			basicAuth:  true,
		},
		{
			name:          "post authentication",
			authMethod:    cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			includeSecret: true,
		},
		{
			name:       "mtls authentication",
			authMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			databaseID := googleUuid.Must(googleUuid.NewV7())
			config := &cryptoutilIdentityConfig.Config{
				Database: &cryptoutilIdentityConfig.DatabaseConfig{
					Type: cryptoutilSharedMagic.TestDatabaseSQLite,
					DSN:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String()),
				},
				Tokens: &cryptoutilIdentityConfig.TokenConfig{
					Issuer: testHTTPSAuthExampleCom,
				},
			}

			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config.Database)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, repoFactory.Close())
			}()

			require.NoError(t, repoFactory.AutoMigrate(context.Background()))

			service := NewService(config, repoFactory, nil)
			service.clientAuth = &cryptoutilIdentityClientAuth.Registry{}

			form := url.Values{
				cryptoutilSharedMagic.ParamClientID: []string{"registry-error-client"},
			}
			if tc.includeSecret {
				form.Set(cryptoutilSharedMagic.ParamClientSecret, "secret")
			}

			if tc.authMethod == cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth {
				client := &cryptoutilIdentityDomain.Client{
					ID:                      googleUuid.Must(googleUuid.NewV7()),
					ClientID:                form.Get(cryptoutilSharedMagic.ParamClientID),
					Name:                    cryptoutilSharedMagic.TestClientName,
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: tc.authMethod,
				}
				require.NoError(t, repoFactory.ClientRepository().Create(context.Background(), client))
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/authenticate", func(c *fiber.Ctx) error {
				_, authErr := service.authenticateClient(c)
				if authErr != nil {
					return c.Status(fiber.StatusUnauthorized).SendString(authErr.Error())
				}

				return c.SendStatus(fiber.StatusOK)
			})

			request := httptest.NewRequest(http.MethodPost, "/authenticate", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			if tc.basicAuth {
				credentials := base64.StdEncoding.EncodeToString([]byte(form.Get(cryptoutilSharedMagic.ParamClientID) + ":secret"))
				request.Header.Set("Authorization", "Basic "+credentials)
			}

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			require.Equal(t, fiber.StatusUnauthorized, response.StatusCode)
			require.Equal(t, fiber.ErrUnauthorized.Message, string(body))
		})
	}
}
