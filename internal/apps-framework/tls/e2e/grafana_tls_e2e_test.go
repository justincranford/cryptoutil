// Copyright (c) 2025-2026 Justin Cranford.
//
//

//go:build e2e

// Package e2e_test provides E2E tests for Grafana HTTPS + OTLP ingest mTLS connectivity.
// These tests use the same compose stack started by TestMain in otel_tls_e2e_test.go.
// Verification points:
//  1. Grafana HTTPS UI: server cert is Cat 2 (public-https-server-entity-grafana-otel-lgtm)
//  2. Grafana OTLP gRPC: Cat 9 infra client cert accepted; no-cert connection rejected
package e2e_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// waitForGrafanaHealth polls the Grafana HTTPS health endpoint until it returns 200.
func waitForGrafanaHealth(t *testing.T) {
	t.Helper()

	caPool := loadCACertPool(t, cryptoutilSharedMagic.GrafanaTLSE2ECACertPath)

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

	healthURL := fmt.Sprintf("https://127.0.0.1:%d/api/health", cryptoutilSharedMagic.GrafanaTLSE2EUIPort)
	deadline := time.Now().UTC().Add(cryptoutilSharedMagic.GrafanaTLSE2EHealthTimeout)

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

	t.Fatalf("Grafana did not become healthy within %s at %s", cryptoutilSharedMagic.GrafanaTLSE2EHealthTimeout, healthURL)
}

// TestGrafanaHTTPS_ServerCert verifies Grafana HTTPS UI TLS handshake succeeds
// and the server presents the expected Cat 2 cert CN.
func TestGrafanaHTTPS_ServerCert(t *testing.T) {
	t.Parallel()

	waitForGrafanaHealth(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.GrafanaTLSE2ECACertPath)

	addr := fmt.Sprintf("127.0.0.1:%d", cryptoutilSharedMagic.GrafanaTLSE2EUIPort)
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout},
		"tcp", addr,
		&tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    caPool,
		},
	)
	require.NoError(t, err, "TLS dial to Grafana HTTPS UI %s must succeed with Cat 1 CA", addr)

	defer func() { _ = conn.Close() }()

	certs := conn.ConnectionState().PeerCertificates
	require.NotEmpty(t, certs, "Grafana HTTPS must present a server certificate")

	cn := certs[0].Subject.CommonName
	assert.Equal(t, cryptoutilSharedMagic.GrafanaTLSE2EServerCertCN, cn,
		"Grafana HTTPS server cert CN must be Cat 2 identity")
}

// TestGrafanaHTTPS_APIHealth verifies Grafana /api/health returns 200 over HTTPS.
func TestGrafanaHTTPS_APIHealth(t *testing.T) {
	t.Parallel()

	waitForGrafanaHealth(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.GrafanaTLSE2ECACertPath)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    caPool,
		},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   cryptoutilSharedMagic.IMDefaultTimeout,
	}

	healthURL := fmt.Sprintf("https://127.0.0.1:%d/api/health", cryptoutilSharedMagic.GrafanaTLSE2EUIPort)

	resp, err := client.Get(healthURL) //nolint:noctx // Simple health check in E2E context.
	require.NoError(t, err, "GET %s must succeed", healthURL)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"Grafana /api/health must return 200; body: %s", body)
}

// TestGrafanaOTLP_GRPC_mTLS_Accepted verifies OTLP gRPC to Grafana succeeds with Cat 9 infra cert.
func TestGrafanaOTLP_GRPC_mTLS_Accepted(t *testing.T) {
	t.Parallel()

	waitForGrafanaHealth(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.GrafanaTLSE2ECACertPath)
	clientCert := loadClientCert(t, cryptoutilSharedMagic.GrafanaTLSE2EInfraCertPath, cryptoutilSharedMagic.GrafanaTLSE2EInfraKeyPath)

	addr := fmt.Sprintf("127.0.0.1:%d", cryptoutilSharedMagic.GrafanaTLSE2EOTLPGRPCPort)
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout},
		"tcp", addr,
		&tls.Config{
			MinVersion:   tls.VersionTLS13,
			RootCAs:      caPool,
			Certificates: []tls.Certificate{clientCert},
		},
	)
	require.NoError(t, err, "TLS dial to Grafana OTLP gRPC %s must succeed with Cat 9 infra cert", addr)

	defer func() { _ = conn.Close() }()

	certs := conn.ConnectionState().PeerCertificates
	require.NotEmpty(t, certs, "Grafana OTLP gRPC must present a server certificate")

	cn := certs[0].Subject.CommonName
	assert.Equal(t, cryptoutilSharedMagic.GrafanaTLSE2EServerCertCN, cn,
		"Grafana OTLP gRPC server cert CN must match Cat 2 identity")
}

// TestGrafanaOTLP_GRPC_mTLS_Rejected verifies OTLP gRPC to Grafana fails without client cert.
// Grafana's internal OTel receiver enforces mTLS via client_ca_file in the TLS override config.
func TestGrafanaOTLP_GRPC_mTLS_Rejected(t *testing.T) {
	t.Parallel()

	waitForGrafanaHealth(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.GrafanaTLSE2ECACertPath)

	addr := fmt.Sprintf("127.0.0.1:%d", cryptoutilSharedMagic.GrafanaTLSE2EOTLPGRPCPort)

	// No client certificate — Grafana must reject connection (mTLS enforced by Cat 8 client CA).
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout},
		"tcp", addr,
		&tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    caPool,
			// No Certificates field — client presents no cert.
		},
	)
	if err == nil {
		_ = conn.Close()

		t.Fatal("TLS dial to Grafana OTLP gRPC without client cert must fail (mTLS enforced)")
	}

	// Verify it's a TLS error, not a network error.
	assert.Contains(t, err.Error(), "tls",
		"error from no-cert Grafana OTLP gRPC dial must be a TLS handshake failure")
}
