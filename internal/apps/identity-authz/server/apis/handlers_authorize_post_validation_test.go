// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	"io"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const authzPostUnknownClientID = "unknown-authz-client"

func TestHandleAuthorizePOST_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		clientID     string
		redirectURI  string
		responseType string
		wantStatus   int
	}{
		{
			name:         "missing redirect URI",
			clientID:     testTestClient,
			responseType: cryptoutilSharedMagic.ResponseTypeCode,
			wantStatus:   fiber.StatusBadRequest,
		},
		{
			name:        "missing response type",
			clientID:    testTestClient,
			redirectURI: cryptoutilSharedMagic.TestRedirectURI,
			wantStatus:  fiber.StatusBadRequest,
		},
		{
			name:         "unsupported response type",
			clientID:     testTestClient,
			redirectURI:  cryptoutilSharedMagic.TestRedirectURI,
			responseType: cryptoutilSharedMagic.ParamToken,
			wantStatus:   fiber.StatusBadRequest,
		},
		{
			name:         "unknown client",
			clientID:     authzPostUnknownClientID,
			redirectURI:  cryptoutilSharedMagic.TestRedirectURI,
			responseType: cryptoutilSharedMagic.ResponseTypeCode,
			wantStatus:   http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoFactory := createAuthorizeTestRepoFactory(t)
			service := cryptoutilIdentityAuthz.NewService(createAuthorizeTestConfig(t), repoFactory, nil)
			require.NotNil(t, service)

			app := fiber.New()
			service.RegisterRoutes(app)

			formData := url.Values{
				cryptoutilSharedMagic.ParamClientID:            {tc.clientID},
				cryptoutilSharedMagic.ParamRedirectURI:         {tc.redirectURI},
				cryptoutilSharedMagic.ParamResponseType:        {tc.responseType},
				cryptoutilSharedMagic.ParamScope:               {testOpenidProfile},
				cryptoutilSharedMagic.ParamState:               {testTestState},
				cryptoutilSharedMagic.ParamCodeChallenge:       {testTestChallenge},
				cryptoutilSharedMagic.ParamCodeChallengeMethod: {cryptoutilSharedMagic.PKCEMethodS256},
			}
			request := httptest.NewRequest(http.MethodPost, testOauth2V1Authorize, strings.NewReader(formData.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, tc.wantStatus, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)
		})
	}
}
