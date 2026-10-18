// Copyright (c) 2025-2026 Justin Cranford.

package test_help_tls

import (
	"context"
	"errors"
	"io"
	http "net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testHealthEndpointURL = "https://localhost/health"

type readinessRoundTripper func(*http.Request) (*http.Response, error)

func (roundTripper readinessRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTripper(request)
}

type readinessResponseBody struct {
	io.Reader
	closeErr error
}

func (body readinessResponseBody) Close() error {
	return body.closeErr
}

func TestRequestHTTPSHealthEndpoint_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		client      *http.Client
		endpointURL string
		wantErr     string
	}{
		{
			name:        "invalid endpoint",
			client:      &http.Client{},
			endpointURL: "https://invalid\x00endpoint",
			wantErr:     "create HTTPS health request",
		},
		{
			name: "request failure",
			client: &http.Client{Transport: readinessRoundTripper(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("injected request failure")
			})},
			endpointURL: testHealthEndpointURL,
			wantErr:     "send HTTPS health request",
		},
		{
			name: "response body close failure",
			client: &http.Client{Transport: readinessRoundTripper(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readinessResponseBody{Reader: strings.NewReader(""), closeErr: errors.New("injected close failure")},
				}, nil
			})},
			endpointURL: testHealthEndpointURL,
			wantErr:     "close HTTPS health response",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := requestHTTPSHealthEndpoint(context.Background(), testCase.client, testCase.endpointURL)
			require.ErrorContains(t, err, testCase.wantErr)
		})
	}
}
