package client

import (
	"context"
	"fmt"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testClientStatusOK        = "ok"
	testClientJSONEndpoint    = "/endpoint"
	testClientReadErrorPrefix = "read error: "
	testClientDecodeError     = "decode response"
	testClientRequestError    = "do request"
)

func TestClientNewAndPing(t *testing.T) {
	t.Parallel()

	requests := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.Path + " " + r.Header.Get("Content-Type")

		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(server.Close)

	client := New(server.URL + "///")
	require.Equal(t, server.URL, client.baseURL)
	require.Positive(t, client.httpClient.Timeout)
	result, err := client.Ping(context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string]any{cryptoutilSharedMagic.StringStatus: testClientStatusOK}, result)
	require.Equal(t, http.MethodGet+" /service/api/v1/health application/json", <-requests)
}

func TestClientDoJSONResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		response   string
		decode     bool
		wantErr    string
		wantResult map[string]any
	}{
		{name: testClientDecodeError, status: http.StatusOK, response: `{"ok":true}`, decode: true, wantResult: map[string]any{"ok": true}},
		{name: "discard response", status: http.StatusNoContent},
		{name: "reject server error", status: http.StatusBadRequest, response: `{"error":"bad"}`, wantErr: "unexpected status code: 400"},
		{name: "reject invalid JSON", status: http.StatusOK, response: `{`, decode: true, wantErr: testClientDecodeError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requestBody := make(chan string, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					requestBody <- testClientReadErrorPrefix + err.Error()
				} else {
					requestBody <- string(body)
				}

				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.response)
			}))
			t.Cleanup(server.Close)

			var (
				result map[string]any
				output any
			)
			if tc.decode {
				output = &result
			}

			err := New(server.URL).doJSON(context.Background(), http.MethodPost, testClientJSONEndpoint, map[string]string{"key": "value"}, output)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}

			require.Equal(t, tc.wantResult, result)
			require.JSONEq(t, `{"key":"value"}`, <-requestBody)
		})
	}
}

func TestClientDoJSONRequestErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   string
		input    any
		closed   bool
		canceled bool
		wantErr  string
	}{
		{name: "marshal failure", method: http.MethodPost, input: make(chan int), wantErr: "marshal request"},
		{name: "invalid method", method: "bad\nmethod", wantErr: "build request"},
		{name: "closed server", method: http.MethodGet, closed: true, wantErr: testClientRequestError},
		{name: "canceled context", method: http.MethodGet, canceled: true, wantErr: testClientRequestError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.NotFoundHandler())
			if tc.closed {
				server.Close()
			} else {
				t.Cleanup(server.Close)
			}

			ctx := context.Background()

			if tc.canceled {
				var cancel context.CancelFunc

				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := New(server.URL).doJSON(ctx, tc.method, testClientJSONEndpoint, tc.input, nil)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestClientDoJSONNilInput(t *testing.T) {
	t.Parallel()

	requestBody := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			requestBody <- testClientReadErrorPrefix + err.Error()

			return
		}

		requestBody <- string(body)

		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	err := New(server.URL).doJSON(context.Background(), http.MethodPost, testClientJSONEndpoint, nil, nil)
	require.NoError(t, err)
	require.Empty(t, <-requestBody)
}
