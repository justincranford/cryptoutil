// Copyright (c) 2025-2026 Justin Cranford.
//

package e2e_infra

import (
	"context"
	"errors"
	"io"
	http "net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testHealthURL          = "http://health.test/health"
	testServiceName        = "test-svc"
	testHealthTimeoutError = "health check timeout after"
)

type testHealthRoundTripper func(*http.Request) (*http.Response, error)

func (roundTripper testHealthRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTripper(request)
}

func newTestHealthHTTPClient(statusCode int, responseErr error, calls *atomic.Int32) *http.Client {
	return &http.Client{Transport: testHealthRoundTripper(func(_ *http.Request) (*http.Response, error) {
		calls.Add(1)

		if responseErr != nil {
			return nil, responseErr
		}

		return &http.Response{
			StatusCode: statusCode,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})}
}

func TestWaitForHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		url              string
		timeout          time.Duration
		statusCode       int
		responseErr      error
		pollInterval     time.Duration
		wantErr          string
		wantRoundTrip    bool
		wantNoRoundTrips bool
	}{
		{
			name:       "timeout without a poll",
			url:        testHealthURL,
			timeout:    cryptoutilSharedMagic.TestUnitPollInterval / 2,
			statusCode: http.StatusServiceUnavailable,
			wantErr:    testHealthTimeoutError,
		},
		{
			name:         "negative poll interval is rejected",
			timeout:      cryptoutilSharedMagic.TestUnitShortTimeout,
			pollInterval: -cryptoutilSharedMagic.TestUnitPollInterval,
			wantErr:      "health poll interval must be positive",
		},
		{
			name:          "connection errors retry until timeout",
			url:           testHealthURL,
			timeout:       cryptoutilSharedMagic.TestUnitShortTimeout,
			responseErr:   errors.New("injected connection error"),
			wantErr:       testHealthTimeoutError,
			wantRoundTrip: true,
		},
		{
			name:          "non-success status retries until timeout",
			url:           testHealthURL,
			timeout:       cryptoutilSharedMagic.TestUnitMediumTimeout,
			statusCode:    http.StatusServiceUnavailable,
			wantErr:       testHealthTimeoutError,
			wantRoundTrip: true,
		},
		{
			name:          "success status returns nil",
			url:           testHealthURL,
			timeout:       cryptoutilSharedMagic.TestUnitMediumTimeout,
			statusCode:    http.StatusOK,
			wantRoundTrip: true,
		},
		{
			name:             "invalid URL retries request creation until timeout",
			url:              "http://invalid\x00host/health",
			timeout:          cryptoutilSharedMagic.TestUnitShortTimeout,
			statusCode:       http.StatusOK,
			wantErr:          testHealthTimeoutError,
			wantNoRoundTrips: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32

			interval := tc.pollInterval
			if interval == 0 {
				interval = cryptoutilSharedMagic.TestUnitPollInterval
			}

			cm := &ComposeManager{
				HTTPClient:         newTestHealthHTTPClient(tc.statusCode, tc.responseErr, &calls),
				healthPollInterval: interval,
			}

			err := cm.WaitForHealth(tc.url, tc.timeout)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)

				if tc.wantRoundTrip || tc.wantNoRoundTrips {
					require.NotContains(t, err.Error(), "(0 attempts)")
				}
			}

			if tc.wantRoundTrip {
				require.Positive(t, calls.Load())
			}

			if tc.wantNoRoundTrips {
				require.Zero(t, calls.Load())
			}
		})
	}
}

func TestWaitForMultipleServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceNames []string
		statusCode   int
		responseErr  error
		timeout      time.Duration
		wantErr      string
		wantRequests bool
	}{
		{
			name:    "empty services succeed",
			timeout: cryptoutilSharedMagic.TestUnitShortTimeout,
		},
		{
			name:         "unhealthy service reports its name",
			serviceNames: []string{testServiceName},
			responseErr:  errors.New("injected connection error"),
			timeout:      cryptoutilSharedMagic.TestUnitShortTimeout,
			wantErr:      testServiceName,
			wantRequests: true,
		},
		{
			name:         "all healthy services succeed",
			serviceNames: []string{"svc-a", "svc-b"},
			statusCode:   http.StatusOK,
			timeout:      cryptoutilSharedMagic.TestUnitHealthCheckTimeout,
			wantRequests: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32

			cm := &ComposeManager{
				HTTPClient:         newTestHealthHTTPClient(tc.statusCode, tc.responseErr, &calls),
				healthPollInterval: cryptoutilSharedMagic.TestUnitPollInterval,
			}

			services := make(map[string]string, len(tc.serviceNames))
			for _, name := range tc.serviceNames {
				services[name] = testHealthURL
			}

			err := cm.WaitForMultipleServices(services, tc.timeout)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}

			if tc.wantRequests {
				require.Positive(t, calls.Load())
			}
		})
	}
}

func TestWaitForServicesHealthy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		psOutput    string
		psOutputErr error
		wantErr     string
	}{
		{
			name:        "docker command error",
			psOutputErr: errors.New("injected docker error"),
			wantErr:     "failed to get docker compose ps output",
		},
		{
			name:     "invalid JSON output",
			psOutput: "not valid json",
			wantErr:  "failed to parse docker compose ps output",
		},
		{
			name:     "service missing from output",
			psOutput: `{"Name":"compose-other-svc-1","State":"running","Health":""}`,
			wantErr:  "unhealthy services",
		},
		{
			name:     "running service without healthcheck is healthy",
			psOutput: `{"Name":"compose-test-svc-1","State":"running"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cm := NewComposeManager("test-compose.yml")
			cm.psOutputFn = func(_ context.Context) ([]byte, error) {
				if tc.psOutputErr != nil {
					return nil, tc.psOutputErr
				}

				return []byte(tc.psOutput), nil
			}

			err := cm.WaitForServicesHealthy(context.Background(), []ServiceAndJob{{Service: testServiceName}})
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestDefaultTestmainFactoryDeps_ClosureBodies(t *testing.T) {
	t.Parallel()

	deps := defaultTestmainFactoryDeps()
	ctx := context.Background()

	cm := deps.newComposeManagerFn("nonexistent-compose.yml")
	require.NotNil(t, cm)

	insecureClient := deps.newInsecureClientFn()
	require.NotNil(t, insecureClient)

	require.Panics(t, func() {
		_ = deps.newSecureClientFn("nonexistent-ca.pem")
	})

	err := deps.startFn(ctx, cm)
	require.Error(t, err, "startFn must fail with nonexistent compose file")

	err = deps.waitForServicesFn(cm, map[string]string{}, cryptoutilSharedMagic.TestTLSClientRetryWait)
	require.NoError(t, err, "waitForServicesFn with empty services must return nil")

	err = deps.stopFn(ctx, cm)
	require.Error(t, err, "stopFn must fail with nonexistent compose file")
}

func TestSetupE2ETestMain_SetupFails(t *testing.T) {
	t.Parallel()

	cfg := E2ETestConfig{
		ComposeFile:    "nonexistent-compose.yml",
		HealthChecks:   map[string]string{},
		HealthTimeout:  cryptoutilSharedMagic.TestTLSClientRetryWait,
		ServiceLogName: testTest,
	}

	code := SetupE2ETestMain(nil, cfg, func(*E2ETestEnv) {})
	require.Equal(t, 1, code)
}
