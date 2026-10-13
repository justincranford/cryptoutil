// Copyright (c) 2025-2026 Justin Cranford.

package test_help_tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	http "net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type readinessEndpoint struct {
	url    string
	rootCA *x509.CertPool
}

type readinessProbeResult struct {
	statusCode int
	err        error
}

func TestWaitForHTTPSHealthEndpoint_Table(t *testing.T) {
	t.Parallel()

	rootCA := x509.NewCertPool()
	healthURL := cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + strconv.Itoa(cryptoutilSharedMagic.TestAdminPort) + cryptoutilSharedMagic.DefaultAdminLivezEndpoint
	testTimeout := cryptoutilSharedMagic.TestServerWaitTickerInterval / 2
	tests := []struct {
		name          string
		endpoints     []readinessEndpoint
		probeResults  []readinessProbeResult
		timeout       time.Duration
		wantProbeCall int
		wantError     string
	}{
		{
			name: "waits for endpoint and trust pool",
			endpoints: []readinessEndpoint{
				{},
				{url: healthURL},
				{url: healthURL, rootCA: rootCA},
			},
			probeResults:  []readinessProbeResult{{statusCode: http.StatusOK}},
			timeout:       cryptoutilSharedMagic.DefaultHealthCheckTimeout,
			wantProbeCall: 1,
		},
		{
			name:      "retries transient request failures and unhealthy status",
			endpoints: []readinessEndpoint{{url: healthURL, rootCA: rootCA}},
			probeResults: []readinessProbeResult{
				{err: errors.New("injected request failure")},
				{statusCode: http.StatusServiceUnavailable},
				{statusCode: http.StatusOK},
			},
			timeout:       cryptoutilSharedMagic.DefaultHealthCheckTimeout,
			wantProbeCall: 3,
		},
		{
			name:          "timeout includes last unhealthy status",
			endpoints:     []readinessEndpoint{{url: healthURL, rootCA: rootCA}},
			probeResults:  []readinessProbeResult{{statusCode: http.StatusServiceUnavailable}},
			timeout:       testTimeout,
			wantProbeCall: 1,
			wantError:     "HTTP 503",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			endpointIndex := 0
			endpointFn := func() (string, *x509.CertPool) {
				index := endpointIndex
				if index >= len(tc.endpoints) {
					index = len(tc.endpoints) - 1
				}

				endpointIndex++

				endpoint := tc.endpoints[index]

				return endpoint.url, endpoint.rootCA
			}

			probeIndex := 0
			probeCalls := 0
			probe := func(_ context.Context, client *http.Client, endpointURL string) (int, error) {
				probeCalls++
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				require.Equal(t, uint16(tls.VersionTLS13), transport.TLSClientConfig.MinVersion)
				require.Same(t, rootCA, transport.TLSClientConfig.RootCAs)
				require.True(t, transport.DisableKeepAlives)
				require.Equal(t, healthURL, endpointURL)

				index := probeIndex
				if index >= len(tc.probeResults) {
					index = len(tc.probeResults) - 1
				}

				probeIndex++
				result := tc.probeResults[index]

				return result.statusCode, result.err
			}

			err := waitForHTTPSHealthEndpoint(context.Background(), endpointFn, tc.timeout, cryptoutilSharedMagic.TestServerWaitTickerInterval, probe)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantProbeCall, probeCalls)
		})
	}
}

func TestWaitForHTTPSHealthEndpoint_InvalidInputs(t *testing.T) {
	t.Parallel()

	rootCA := x509.NewCertPool()
	endpointFn := func() (string, *x509.CertPool) { return "https://localhost", rootCA }
	probe := func(context.Context, *http.Client, string) (int, error) { return http.StatusOK, nil }
	tests := []struct {
		name       string
		ctx        context.Context
		endpointFn httpsHealthEndpointFn
		timeout    time.Duration
		interval   time.Duration
		probe      httpsHealthProbeFn
		wantError  string
	}{
		{name: "nil context", endpointFn: endpointFn, timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout, interval: cryptoutilSharedMagic.TestServerWaitTickerInterval, probe: probe, wantError: "context cannot be nil"},
		{name: "nil endpoint source", ctx: context.Background(), timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout, interval: cryptoutilSharedMagic.TestServerWaitTickerInterval, probe: probe, wantError: "endpoint source cannot be nil"},
		{name: "nil probe", ctx: context.Background(), endpointFn: endpointFn, timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout, interval: cryptoutilSharedMagic.TestServerWaitTickerInterval, wantError: "health probe cannot be nil"},
		{name: "non-positive timeout", ctx: context.Background(), endpointFn: endpointFn, interval: cryptoutilSharedMagic.TestServerWaitTickerInterval, probe: probe, wantError: "timeout must be positive"},
		{name: "non-positive interval", ctx: context.Background(), endpointFn: endpointFn, timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout, probe: probe, wantError: "interval must be positive"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := waitForHTTPSHealthEndpoint(tc.ctx, tc.endpointFn, tc.timeout, tc.interval, tc.probe)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestWaitForHTTPSHealthEndpoint_TLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		trustServerCert bool
		wantTLSFailure  bool
	}{
		{name: "trusted TLS and healthy response", trustServerCert: true},
		{name: "untrusted TLS certificate is rejected", wantTLSFailure: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			rootCAs := x509.NewCertPool()
			if tc.trustServerCert {
				rootCAs.AddCert(server.Certificate())
			}

			if tc.wantTLSFailure {
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							MinVersion: tls.VersionTLS13,
							RootCAs:    rootCAs,
						},
						DisableKeepAlives: true,
					},
					Timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout,
				}
				statusCode, err := requestHTTPSHealthEndpoint(context.Background(), client, server.URL)
				require.Zero(t, statusCode)
				require.ErrorContains(t, err, cryptoutilSharedMagic.TLSName)

				return
			}

			endpointFn := func() (string, *x509.CertPool) { return server.URL, rootCAs }
			err := WaitForHTTPSHealthEndpoint(context.Background(), endpointFn, cryptoutilSharedMagic.DefaultHealthCheckTimeout)
			require.NoError(t, err)
		})
	}
}
