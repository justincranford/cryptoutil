// Copyright (c) 2025-2026 Justin Cranford.
package test_help_tls

import (
	"context"
	"errors"
	"io"
	http "net/http"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type readinessRetryBody struct {
	io.Reader
	closed *int
	err    error
}

func (body readinessRetryBody) Close() error {
	*body.closed++

	return body.err
}

func TestDoReadinessRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		statuses     []int
		transportErr error
		closeErr     error
		wantCode     int
		wantClosed   int
		wantErr      bool
	}{
		{name: cryptoutilSharedMagic.DockerServiceHealthHealthy, statuses: []int{http.StatusOK}, wantCode: http.StatusOK},
		{name: "transient fail closed", statuses: []int{http.StatusServiceUnavailable, http.StatusServiceUnavailable, http.StatusOK}, wantCode: http.StatusOK, wantClosed: 2},
		{name: "unexpected status remains visible", statuses: []int{http.StatusUnauthorized}, wantCode: http.StatusUnauthorized},
		{name: "redirect remains visible and is never followed", statuses: []int{http.StatusFound}, wantCode: http.StatusFound},
		{name: "transport failure remains visible", transportErr: errors.New("transport failed"), wantErr: true},
		{name: "body close failure remains visible", statuses: []int{http.StatusServiceUnavailable}, closeErr: errors.New("body close failed"), wantClosed: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			calls, closed := 0, 0
			client := &http.Client{Transport: readinessRoundTripper(func(request *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodGet, request.Method)
				require.Equal(t, cryptoutilSharedMagic.HostnameLocalhost, request.URL.Hostname(), "redirect must never leave the allowed loopback host")

				calls++

				if tt.transportErr != nil {
					return nil, tt.transportErr
				}

				return &http.Response{
					StatusCode: tt.statuses[calls-1],
					Header:     http.Header{"Location": []string{"https://remote.invalid/"}},
					Body:       readinessRetryBody{Reader: strings.NewReader("{}"), closed: &closed, err: tt.closeErr},
				}, nil
			})}

			ctx, cancel := context.WithTimeout(t.Context(), cryptoutilSharedMagic.TestTimeout)
			defer cancel()

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, testHealthEndpointURL, nil)
			require.NoError(t, err)
			response, err := DoReadinessRequest(client, request)
			require.Equal(t, tt.wantErr, err != nil)

			if tt.transportErr != nil {
				require.Equal(t, 1, calls)
			} else {
				require.Len(t, tt.statuses, calls)
			}

			require.Equal(t, tt.wantClosed, closed, "retried bodies close, successful body remains caller-owned")

			if tt.wantErr {
				require.Nil(t, response)
			} else {
				require.Equal(t, tt.wantCode, response.StatusCode)
				require.NoError(t, response.Body.Close())
			}
		})
	}
}

func TestDoReadinessRequestRejectsInvalidInputsAndHonorsDeadline(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: readinessRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	})}
	request, err := http.NewRequest(http.MethodGet, testHealthEndpointURL, nil)
	require.NoError(t, err)
	err = readinessErrorRequest(t, nil, request)
	require.Error(t, err)
	err = readinessErrorRequest(t, client, nil)
	require.Error(t, err)
	err = readinessErrorRequest(t, client, request)
	require.ErrorContains(t, err, "bounded deadline")

	post := request.Clone(t.Context())
	post.Method = http.MethodPost
	err = readinessErrorRequest(t, client, post)
	require.ErrorContains(t, err, "GET")

	for _, endpoint := range []string{"http://localhost/readyz", "https://remote.invalid/readyz", "https://user@localhost/readyz"} {
		unsafe, parseErr := http.NewRequest(http.MethodGet, endpoint, nil)
		require.NoError(t, parseErr)
		require.Error(t, readinessErrorRequest(t, client, unsafe))
	}

	noURL := request.Clone(t.Context())
	noURL.URL = nil
	require.Error(t, readinessErrorRequest(t, client, noURL))

	cancelled, cancel := context.WithCancel(t.Context())
	cancel()

	err = readinessErrorRequest(t, client, request.WithContext(cancelled))
	require.ErrorIs(t, err, context.Canceled)

	bounded, cancelBounded := context.WithTimeout(t.Context(), time.Millisecond)
	defer cancelBounded()

	err = readinessErrorRequest(t, client, request.WithContext(bounded))
	require.Error(t, err, "persistent unready response must fail within the caller's deadline")
}

func readinessErrorRequest(t *testing.T, client *http.Client, request *http.Request) error {
	t.Helper()

	response, err := DoReadinessRequest(client, request)
	if response != nil {
		require.NoError(t, response.Body.Close())
	}

	require.Nil(t, response)

	return err
}
