// Copyright (c) 2025-2026 Justin Cranford.
//

package apis_test

import (
	"bytes"
	crand "crypto/rand"
	json "encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Sequential: replaces crypto/rand.Reader package-global state.
func TestCodeGenerators_ReturnRandomReadErrors(t *testing.T) {
	config, repoFactory := createDeviceAuthTestDependencies(t)
	defer func() {
		require.NoError(t, repoFactory.Close())
	}()

	testClient := createTestClientForDevice(t.Context(), t, repoFactory)
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	service.RegisterRoutes(app)

	originalReader := crand.Reader

	crand.Reader = strings.NewReader("")
	defer func() {
		crand.Reader = originalReader
	}()

	tests := []struct {
		name     string
		generate func() (string, error)
	}{
		{name: "authorization code", generate: cryptoutilIdentityAuthz.GenerateAuthorizationCode},
		{name: "device code", generate: cryptoutilIdentityAuthz.GenerateDeviceCode},
		{name: "request URI", generate: cryptoutilIdentityAuthz.GenerateRequestURI},
		{name: "user code", generate: cryptoutilIdentityAuthz.GenerateUserCode},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, err := tc.generate()
			require.ErrorIs(t, err, io.EOF)
			require.Empty(t, code)
		})
	}

	deviceAuthorizationTests := []struct {
		name      string
		reader    io.Reader
		wantError string
	}{
		{name: "device code generation failure", reader: strings.NewReader(""), wantError: "Failed to generate device code"},
		{name: "user code generation failure", reader: bytes.NewReader(make([]byte, cryptoutilSharedMagic.DefaultDeviceCodeLength)), wantError: "Failed to generate user code"},
	}

	for _, tc := range deviceAuthorizationTests {
		t.Run(tc.name, func(t *testing.T) {
			crand.Reader = tc.reader
			formData := url.Values{
				cryptoutilSharedMagic.ParamClientID: []string{testClient.ClientID},
			}
			request := httptest.NewRequest(testPost, "/oauth2/v1/device_authorization", strings.NewReader(formData.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			var result map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
			require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
			require.Equal(t, tc.wantError, result["error_description"])
		})
	}

	t.Run("PAR request URI generation failure", func(t *testing.T) {
		crand.Reader = strings.NewReader("")
		formData := url.Values{
			cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
			cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
			cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.TestRedirectURI},
			cryptoutilSharedMagic.ParamCodeChallenge:       []string{cryptoutilSharedMagic.TEST_CODE_CHALLENGE},
			cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
		}
		request := httptest.NewRequest(testPost, "/oauth2/v1/par", strings.NewReader(formData.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		response, err := app.Test(request, -1)

		require.NoError(t, err)
		defer func() {
			require.NoError(t, response.Body.Close())
		}()

		var result map[string]string
		require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
		require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
		require.Equal(t, "Internal server error", result["error_description"])
	})
}
