// Copyright (c) 2025-2026 Justin Cranford.

package client

import (
	"context"
	json "encoding/json"
	"errors"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testItemsPath = "/items"

func TestClient_Ping(t *testing.T) {
	t.Parallel()

	body := clientTestJSONBody(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, cryptoutilSharedMagic.DefaultServiceHealthEndpoint, r.URL.Path)

		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	response, err := New(server.URL + "/").Ping(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ok", response[cryptoutilSharedMagic.StringStatus])
}

func TestClient_DoJSON(t *testing.T) {
	t.Parallel()

	body := clientTestJSONBody(t)
	tests := []struct {
		name        string
		status      int
		body        string
		input       any
		response    any
		path        string
		closeServer bool
		wantErr     string
		wantRequest bool
	}{
		{
			name:        "success with JSON input and output",
			status:      http.StatusOK,
			body:        body,
			input:       map[string]string{cryptoutilSharedMagic.ClaimName: "item"},
			response:    &map[string]any{},
			path:        testItemsPath,
			wantRequest: true,
		},
		{name: "success without response decoding", status: http.StatusNoContent, path: testItemsPath},
		{
			name:     "HTTP status error",
			status:   http.StatusBadRequest,
			body:     body,
			path:     testItemsPath,
			response: &map[string]any{},
			wantErr:  "unexpected status code",
		},
		{
			name:     "invalid JSON response",
			status:   http.StatusOK,
			body:     "{",
			path:     testItemsPath,
			response: &map[string]any{},
			wantErr:  "decode response",
		},
		{name: "invalid request URL", status: http.StatusOK, path: "/bad\npath", wantErr: "build request"},
		{
			name:    "unmarshalable request",
			status:  http.StatusOK,
			path:    testItemsPath,
			input:   make(chan int),
			wantErr: "marshal request",
		},
		{
			name:        "transport error",
			status:      http.StatusOK,
			path:        testItemsPath,
			closeServer: true,
			wantErr:     "do request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requestSeen := make(chan struct{}, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestSeen <- struct{}{}

				if tc.wantRequest {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "application/json", r.Header.Get("Content-Type"))

					var requestBody map[string]string
					require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))
					require.Equal(t, "item", requestBody[cryptoutilSharedMagic.ClaimName])
				}

				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			if tc.closeServer {
				server.Close()
			}

			client := New(server.URL)

			err := client.doJSON(context.Background(), http.MethodPost, tc.path, tc.input, tc.response)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			if tc.wantRequest {
				require.Len(t, requestSeen, 1)

				decoded, ok := tc.response.(*map[string]any)
				require.True(t, ok)
				require.Equal(t, "ok", (*decoded)[cryptoutilSharedMagic.StringStatus])
			}
		})
	}
}

func TestClient_DoJSON_CanceledContext(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := New(server.URL).doJSON(ctx, http.MethodGet, testItemsPath, nil, nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, context.Canceled) || strings.Contains(err.Error(), context.Canceled.Error()))
}

func clientTestJSONBody(t *testing.T) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{cryptoutilSharedMagic.StringStatus: "ok"})
	require.NoError(t, err)

	return string(body)
}
