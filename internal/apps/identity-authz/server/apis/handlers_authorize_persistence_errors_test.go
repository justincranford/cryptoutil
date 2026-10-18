// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	"context"
	json "encoding/json"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestHandleAuthorize_RequestPersistenceFailure(t *testing.T) {
	t.Parallel()

	databaseID := googleUuid.Must(googleUuid.NewV7())
	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, databaseID.String()),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: testHTTPSLocalhost8080,
		},
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config.Database)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, repoFactory.Close())
	}()

	require.NoError(t, repoFactory.AutoMigrate(context.Background()))

	client := createTestClient(context.Background(), t, repoFactory)
	require.NoError(t, repoFactory.DB().Migrator().DropTable(&cryptoutilIdentityDomain.AuthorizationRequest{}))

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterRoutes(app)

	form := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{client.ClientID},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.TestRedirectURI},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamScope:               []string{cryptoutilSharedMagic.ScopeOpenID},
		cryptoutilSharedMagic.ParamState:               []string{testTestState},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{cryptoutilSharedMagic.TEST_CODE_CHALLENGE},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	invalidMethodForm := make(url.Values, len(form))
	for key, values := range form {
		invalidMethodForm[key] = values
	}

	invalidMethodForm.Set(cryptoutilSharedMagic.ParamCodeChallengeMethod, cryptoutilSharedMagic.PKCEMethodPlain)

	tests := []struct {
		name             string
		method           string
		path             string
		body             string
		wantStatus       int
		wantErrorDescrip string
	}{
		{
			name:             "GET persistence failure",
			method:           http.MethodGet,
			path:             testOauth2V1Authorize + "?" + form.Encode(),
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorDescrip: testFailedToStoreAuthorizationRequest,
		},
		{
			name:             "POST persistence failure",
			method:           http.MethodPost,
			path:             testOauth2V1Authorize,
			body:             form.Encode(),
			wantStatus:       fiber.StatusInternalServerError,
			wantErrorDescrip: testFailedToStoreAuthorizationRequest,
		},
		{
			name:       "POST rejects unsupported PKCE method",
			method:     http.MethodPost,
			path:       testOauth2V1Authorize,
			body:       invalidMethodForm.Encode(),
			wantStatus: fiber.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			var result map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.wantErrorDescrip != "" {
				require.Equal(t, tc.wantErrorDescrip, result["error_description"])
			}
		})
	}
}
