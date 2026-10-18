// Copyright (c) 2025-2026 Justin Cranford.

package test_help_api

import (
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAssertJSONEncodingFailure(t *testing.T) {
	t.Parallel()

	AssertJSONEncodingFailure(
		t,
		"/response",
		httptest.NewRequest(http.MethodGet, "/response", nil),
		func(ctx *fiber.Ctx) error {
			return ctx.JSON(map[string]string{cryptoutilSharedMagic.StringStatus: "ok"})
		},
	)
}

func TestAssertJSONEncodingFailureWith_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		appTest    func(*http.Request, ...int) (*http.Response, error)
		wantFatal  []string
		wantErrors []string
	}{
		{
			name: "request failure",
			appTest: func(*http.Request, ...int) (*http.Response, error) {
				return nil, fmt.Errorf("injected request failure")
			},
			wantFatal: []string{"request failed: injected request failure"},
		},
		{
			name: "response body close failure",
			appTest: func(*http.Request, ...int) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       &errCloseReader{reader: strings.NewReader("body")},
				}, nil
			},
			wantErrors: []string{"failed to close response body: close failure"},
		},
		{
			name: "unexpected status",
			appTest: func(*http.Request, ...int) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("body")),
				}, nil
			},
			wantErrors: []string{"unexpected status: got 200, want 500"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				fatalMessages []string
				errorMessages []string
			)

			assertJSONEncodingFailureWith(
				httptest.NewRequest(http.MethodGet, "/response", nil),
				tc.appTest,
				func(format string, args ...any) {
					fatalMessages = append(fatalMessages, fmt.Sprintf(format, args...))
				},
				func(format string, args ...any) {
					errorMessages = append(errorMessages, fmt.Sprintf(format, args...))
				},
			)

			require.Equal(t, tc.wantFatal, fatalMessages)
			require.Equal(t, tc.wantErrors, errorMessages)
		})
	}
}

type errCloseReader struct {
	reader io.Reader
}

func (e *errCloseReader) Read(p []byte) (int, error) {
	return e.reader.Read(p)
}

func (e *errCloseReader) Close() error {
	return fmt.Errorf("close failure")
}

func TestNewHealthClientAndEndpoints_Table(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()
	handler.HandleFunc(cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+cryptoutilSharedMagic.PrivateAdminLivezRequestPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler.HandleFunc(cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler.HandleFunc(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler.HandleFunc(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath+"/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tlsServer := httptest.NewTLSServer(handler)
	t.Cleanup(tlsServer.Close)

	client := NewHealthClient(tlsServer.URL, tlsServer.URL)
	require.NotNil(t, client)
	require.NotNil(t, client.client)

	tests := []struct {
		name string
		call func(*HealthClient) (*http.Response, error)
	}{
		{name: "livez", call: (*HealthClient).Livez},
		{name: "readyz", call: (*HealthClient).Readyz},
		{name: "servicehealth", call: (*HealthClient).ServiceHealth},
		{name: "browserhealth", call: (*HealthClient).BrowserHealth},
		{name: "publichealth alias", call: (*HealthClient).PublicHealth},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := tc.call(client)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.NoError(t, client.DrainAndClose(resp))
		})
	}
}

func TestHealthClientErrorPaths_Table(t *testing.T) {
	t.Parallel()

	client := NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

	tests := []struct {
		name string
		call func(*HealthClient) (*http.Response, error)
		want string
	}{
		{name: "livez error", call: (*HealthClient).Livez, want: "livez request failed"},
		{name: "readyz error", call: (*HealthClient).Readyz, want: "readyz request failed"},
		{name: "servicehealth error", call: (*HealthClient).ServiceHealth, want: "servicehealth request failed"},
		{name: "browserhealth error", call: (*HealthClient).BrowserHealth, want: "browserhealth request failed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := tc.call(client)
			if resp != nil {
				require.NoError(t, client.DrainAndClose(resp))
			}

			require.Nil(t, resp)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestDrainAndClose_Table(t *testing.T) {
	t.Parallel()

	client := NewHealthClient("https://127.0.0.1:1", "https://127.0.0.1:1")

	tests := []struct {
		name string
		resp *http.Response
		want string
	}{
		{name: "nil response", resp: nil},
		{name: "nil body", resp: &http.Response{Body: nil}},
		{name: "close success", resp: &http.Response{Body: io.NopCloser(strings.NewReader("ok"))}},
		{name: "close error", resp: &http.Response{Body: &errCloseReader{reader: strings.NewReader("ok")}}, want: "close response body"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := client.DrainAndClose(tc.resp)
			if tc.want == "" {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			require.ErrorContains(t, err, tc.want)
		})
	}
}
