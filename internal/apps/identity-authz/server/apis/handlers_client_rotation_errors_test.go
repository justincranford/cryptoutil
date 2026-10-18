// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	json "encoding/json"
	"fmt"
	http "net/http"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestClientSecretRotation_PersistenceFailure(t *testing.T) {
	t.Parallel()

	repoFactory, app, client, originalSecret := newClientRotationErrorTestFixture(t)
	defer func() {
		require.NoError(t, repoFactory.Close())
	}()

	require.NoError(t, repoFactory.DB().Migrator().DropTable(&cryptoutilIdentityDomain.ClientSecretHistory{}))

	response := rotateTestClientSecret(t, app, client.ClientID, originalSecret)
	defer func() {
		require.NoError(t, response.Body.Close())
	}()

	var result map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
	require.Equal(t, "Failed to rotate secret", result["error_description"])
}

func newClientRotationErrorTestFixture(t *testing.T) (*cryptoutilIdentityRepository.RepositoryFactory, *fiber.App, *cryptoutilIdentityDomain.Client, string) {
	t.Helper()

	databaseID := googleUuid.Must(googleUuid.NewV7())
	databaseConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String()),
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(t.Context(), databaseConfig)
	require.NoError(t, err)
	require.NoError(t, repoFactory.AutoMigrate(t.Context()))

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout,
	}
	service := cryptoutilIdentityAuthz.NewService(&cryptoutilIdentityConfig.Config{Tokens: tokenConfig}, repoFactory, nil)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterRoutes(app)

	originalSecret := "rotation-test-secret-" + googleUuid.Must(googleUuid.NewV7()).String()
	hashedSecret, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(originalSecret)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:                "rotation-test-client-" + googleUuid.Must(googleUuid.NewV7()).String(),
		ClientSecret:            hashedSecret,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Client Rotation Error Test",
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}
	require.NoError(t, repoFactory.ClientRepository().Create(t.Context(), client))

	return repoFactory, app, client, originalSecret
}

func rotateTestClientSecret(t *testing.T, app *fiber.App, clientID, secret string) *http.Response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth2/v1/clients/"+clientID+"/rotate-secret", nil)
	require.NoError(t, err)

	request.Host = cryptoutilSharedMagic.HostnameLocalhost
	request.SetBasicAuth(clientID, secret)

	response, err := app.Test(request, -1)
	require.NoError(t, err)

	return response
}
