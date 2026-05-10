// Copyright (c) 2025-2026 Justin Cranford.
//

//go:build e2e

package test_orch_e2e

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// LoadCACertPool reads a PEM CA cert file and returns an x509.CertPool.
func LoadCACertPool(t *testing.T, caPath string) *x509.CertPool {
	t.Helper()

	caPEM, err := os.ReadFile(caPath) //nolint:gosec // CA cert path from trusted test config.
	require.NoError(t, err, "read CA cert %q", caPath)

	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(caPEM), "parse CA cert from %q", caPath)

	return pool
}

// LoadClientCert reads a PEM cert+key pair for mTLS.
func LoadClientCert(t *testing.T, certPath, keyPath string) tls.Certificate {
	t.Helper()

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	require.NoError(t, err, "load client cert %q / %q", certPath, keyPath)

	return cert
}

// WaitForOTelHealth polls the OTel Collector health endpoint until ready or timeout.
func WaitForOTelHealth(t *testing.T, spec TLSPSIDSpec, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().UTC().Add(timeout)

	client := &http.Client{
		Timeout: cryptoutilSharedMagic.OtelCollectorHealthCheckTimeout,
	}

	healthURL := spec.OTelHealthURL()

	for time.Now().UTC().Before(deadline) {
		resp, err := client.Get(healthURL) //nolint:noctx // Simple health poll, no context needed.
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(cryptoutilSharedMagic.KMSE2EHealthPollInterval)
	}

	t.Fatalf("OTel Collector did not become healthy within %s at %s", timeout, healthURL)
}

// WaitForGrafanaHealth polls the Grafana HTTPS health endpoint until it returns 200.
func WaitForGrafanaHealth(t *testing.T, spec TLSPSIDSpec, timeout time.Duration) {
	t.Helper()

	caPool := LoadCACertPool(t, spec.PublicCACertPath)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    caPool,
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cryptoutilSharedMagic.OtelCollectorHealthCheckTimeout,
	}

	healthURL := fmt.Sprintf("https://127.0.0.1:%d/api/health", spec.GrafanaUIPort)
	deadline := time.Now().UTC().Add(timeout)

	for time.Now().UTC().Before(deadline) {
		resp, err := client.Get(healthURL) //nolint:noctx // Simple health poll, no context needed.
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(cryptoutilSharedMagic.KMSE2EHealthPollInterval)
	}

	t.Fatalf("Grafana did not become healthy within %s at %s", timeout, healthURL)
}

// WaitForAppsHealthy waits for all app variants in the spec to become reachable over HTTPS.
func WaitForAppsHealthy(t *testing.T, spec TLSPSIDSpec, timeout time.Duration) {
	t.Helper()

	caPool := LoadCACertPool(t, spec.PublicCACertPath)

	for _, v := range spec.AppVariants {
		clientCert := LoadClientCert(t, v.ClientCertPath, v.ClientKeyPath)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS13,
				RootCAs:      caPool,
				Certificates: []tls.Certificate{clientCert},
			},
			DisableKeepAlives: true,
		}

		client := &http.Client{
			Transport: transport,
			Timeout:   cryptoutilSharedMagic.OtelCollectorHealthCheckTimeout,
		}

		healthURL := fmt.Sprintf("https://127.0.0.1:%d%s", v.PublicPort, spec.AppHealthEndpoint)
		deadline := time.Now().UTC().Add(timeout)

		for time.Now().UTC().Before(deadline) {
			resp, err := client.Get(healthURL) //nolint:noctx // Simple health poll, no context needed.
			if err == nil {
				_ = resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					break
				}
			}

			time.Sleep(cryptoutilSharedMagic.KMSE2EHealthPollInterval)
		}

		if time.Now().UTC().After(deadline) {
			require.Failf(t, "app did not become healthy",
				"app %q did not become healthy within %s at %s", v.Name, timeout, healthURL)
		}
	}
}

// WaitForStack waits for OTel, Grafana, and all app variants to become healthy.
func WaitForStack(ctx context.Context, t *testing.T, spec TLSPSIDSpec) { //nolint:revive // ctx passed for future use.
	t.Helper()
	WaitForOTelHealth(t, spec, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)
	WaitForGrafanaHealth(t, spec, cryptoutilSharedMagic.GrafanaTLSE2EHealthTimeout)
	WaitForAppsHealthy(t, spec, cryptoutilSharedMagic.FullPipelineTLSE2ETimeout)
}
