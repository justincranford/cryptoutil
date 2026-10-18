// Copyright (c) 2025-2026 Justin Cranford.
//

package apis

import (
	"context"
	json "encoding/json"
	"fmt"
	http "net/http"
	"net/http/httptest"
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

type fixedClientAuthenticator struct {
	client *cryptoutilIdentityDomain.Client
}

func (a *fixedClientAuthenticator) Method() string {
	return cryptoutilSharedMagic.ClientAuthMethodSecretBasic
}

func (a *fixedClientAuthenticator) Authenticate(_ context.Context, _, _ string) (*cryptoutilIdentityDomain.Client, error) {
	return a.client, nil
}

func TestClientSecretRotation_ClientLookupFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		dropClientTable bool
		wantStatus      int
		wantDescription string
	}{
		{
			name:            "unknown authenticated client",
			wantStatus:      fiber.StatusNotFound,
			wantDescription: "Client not found",
		},
		{
			name:            "client repository failure",
			dropClientTable: true,
			wantStatus:      fiber.StatusInternalServerError,
			wantDescription: valueFailedToRetrieveClient,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			databaseID := googleUuid.Must(googleUuid.NewV7())
			databaseConfig := &cryptoutilIdentityConfig.DatabaseConfig{
				Type: cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String()),
			}
			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(t.Context(), databaseConfig)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, repoFactory.Close())
			}()

			require.NoError(t, repoFactory.AutoMigrate(t.Context()))

			const clientID = "rotation-lookup-client"

			authenticatedClient := &cryptoutilIdentityDomain.Client{ClientID: clientID}
			service := NewService(&cryptoutilIdentityConfig.Config{
				Tokens: &cryptoutilIdentityConfig.TokenConfig{Issuer: testHTTPSAuthExampleCom},
			}, repoFactory, nil)
			service.clientAuth.RegisterAuthenticator(&fixedClientAuthenticator{client: authenticatedClient})

			if tc.dropClientTable {
				require.NoError(t, repoFactory.DB().Migrator().DropTable(&cryptoutilIdentityDomain.Client{}))
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			request := httptest.NewRequest(http.MethodPost, "/oauth2/v1/clients/"+clientID+"/rotate-secret", nil)
			request.SetBasicAuth(clientID, "test-secret")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			var result map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
			require.Equal(t, tc.wantStatus, response.StatusCode)
			require.Equal(t, tc.wantDescription, result["error_description"])
		})
	}
}

var _ cryptoutilIdentityClientAuth.ClientAuthenticator = (*fixedClientAuthenticator)(nil)
