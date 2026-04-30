// Copyright (c) 2025-2026 Justin Cranford.
//
//

//go:build e2e

// Package e2e_test provides E2E tests for the full telemetry and public HTTPS pipeline.
// These tests use the same compose stack started by TestMain in otel_tls_e2e_test.go,
// which now starts all sm-kms variants (sqlite-1, sqlite-2, postgresql-1, postgresql-2)
// alongside OTel Collector and Grafana LGTM.
//
// Verification points (Phase 11):
//  1. App→OTel→Grafana mTLS pipeline: Grafana HTTPS reachable; Grafana OTLP gRPC mTLS accepted/rejected
//  2. App public HTTPS: each variant serves Cat 3 cert; Cat 4 mTLS enforcement active
//  3. Health endpoints reachable over mTLS for all 4 sm-kms variants
package e2e_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	http "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// waitForAppsHealthy waits for all 4 sm-kms app variants to become reachable over HTTPS.
// Uses the Cat 1 CA pool (server cert verification) and Cat 5 service-user cert (client auth).
// Returns when all 4 variants respond to the health endpoint, or fatals on timeout.
func waitForAppsHealthy(t *testing.T) {
	t.Helper()

	type appVariant struct {
		name     string
		port     int
		certPath string
		keyPath  string
	}

	variants := []appVariant{
		{
			name:     cryptoutilSharedMagic.AppSMKMSSQLite1Container,
			port:     cryptoutilSharedMagic.AppSMKMSSQLite1PublicPort,
			certPath: cryptoutilSharedMagic.AppSMKMSSQLite1ClientCertPath,
			keyPath:  cryptoutilSharedMagic.AppSMKMSSQLite1ClientKeyPath,
		},
		{
			name:     cryptoutilSharedMagic.AppSMKMSSQLite2Container,
			port:     cryptoutilSharedMagic.AppSMKMSSQLite2PublicPort,
			certPath: cryptoutilSharedMagic.AppSMKMSSQLite2ClientCertPath,
			keyPath:  cryptoutilSharedMagic.AppSMKMSSQLite2ClientKeyPath,
		},
		{
			name:     cryptoutilSharedMagic.AppSMKMSPostgres1Container,
			port:     cryptoutilSharedMagic.AppSMKMSPostgres1PublicPort,
			certPath: cryptoutilSharedMagic.AppSMKMSPostgresClientCertPath,
			keyPath:  cryptoutilSharedMagic.AppSMKMSPostgresClientKeyPath,
		},
		{
			name:     cryptoutilSharedMagic.AppSMKMSPostgres2Container,
			port:     cryptoutilSharedMagic.AppSMKMSPostgres2PublicPort,
			certPath: cryptoutilSharedMagic.AppSMKMSPostgresClientCertPath,
			keyPath:  cryptoutilSharedMagic.AppSMKMSPostgresClientKeyPath,
		},
	}

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath)

	for _, v := range variants {
		clientCert := loadClientCert(t, v.certPath, v.keyPath)

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

		healthURL := fmt.Sprintf("https://127.0.0.1:%d%s", v.port, cryptoutilSharedMagic.IdentityE2EHealthEndpoint)
		deadline := time.Now().UTC().Add(cryptoutilSharedMagic.FullPipelineTLSE2ETimeout)

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
			require.Failf(t, "app did not become healthy", "app %q did not become healthy within %s at %s", v.name, cryptoutilSharedMagic.FullPipelineTLSE2ETimeout, healthURL)
		}
	}
}

// TestFullPipeline_GrafanaHealth verifies Grafana HTTPS UI returns 200 from /api/health.
// This confirms the OTel→Grafana mTLS pipeline is active (Grafana runs OTel internally).
func TestFullPipeline_GrafanaHealth(t *testing.T) {
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
	require.NoError(t, err, "GET %s must succeed over HTTPS with Cat 1 CA", healthURL)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"Grafana /api/health must return 200; body: %s", body)
}

// TestFullPipeline_GrafanaOTLP_ServerCert verifies Grafana OTLP gRPC presents Cat 2 server cert.
// This verifies the full pipeline: app→OTel→Grafana uses the same Cat 1/2 cert hierarchy.
func TestFullPipeline_GrafanaOTLP_ServerCert(t *testing.T) {
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
	require.Equal(t, cryptoutilSharedMagic.GrafanaTLSE2EServerCertCN, cn,
		"Grafana OTLP gRPC server cert CN must be Cat 2 identity")
}

// TestFullPipeline_GrafanaOTLP_MTLSRejected verifies Grafana OTLP gRPC rejects connections without client cert.
// This confirms Cat 8 CA enforcement is active for the OTel→Grafana mTLS ingest pipeline.
func TestFullPipeline_GrafanaOTLP_MTLSRejected(t *testing.T) {
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

		t.Fatal("TLS dial to Grafana OTLP gRPC without client cert must fail (Cat 8 mTLS enforced)")
	}

	// Verify it's a TLS error, not a network error.
	require.Contains(t, err.Error(), "tls",
		"error from no-cert Grafana OTLP gRPC dial must be a TLS handshake failure")
}

// appVariantTestCase is a table row for parametric app variant tests.
type appVariantTestCase struct {
	name           string
	port           int
	serverCertCN   string
	clientCertPath string
	clientKeyPath  string
}

// allAppVariants returns the test cases for all 4 sm-kms app variants.
func allAppVariants() []appVariantTestCase {
	return []appVariantTestCase{
		{
			name:           cryptoutilSharedMagic.AppSMKMSSQLite1Container,
			port:           cryptoutilSharedMagic.AppSMKMSSQLite1PublicPort,
			serverCertCN:   cryptoutilSharedMagic.AppSMKMSSQLite1ServerCertCN,
			clientCertPath: cryptoutilSharedMagic.AppSMKMSSQLite1ClientCertPath,
			clientKeyPath:  cryptoutilSharedMagic.AppSMKMSSQLite1ClientKeyPath,
		},
		{
			name:           cryptoutilSharedMagic.AppSMKMSSQLite2Container,
			port:           cryptoutilSharedMagic.AppSMKMSSQLite2PublicPort,
			serverCertCN:   cryptoutilSharedMagic.AppSMKMSSQLite2ServerCertCN,
			clientCertPath: cryptoutilSharedMagic.AppSMKMSSQLite2ClientCertPath,
			clientKeyPath:  cryptoutilSharedMagic.AppSMKMSSQLite2ClientKeyPath,
		},
		{
			name:           cryptoutilSharedMagic.AppSMKMSPostgres1Container,
			port:           cryptoutilSharedMagic.AppSMKMSPostgres1PublicPort,
			serverCertCN:   cryptoutilSharedMagic.AppSMKMSPostgres1ServerCertCN,
			clientCertPath: cryptoutilSharedMagic.AppSMKMSPostgresClientCertPath,
			clientKeyPath:  cryptoutilSharedMagic.AppSMKMSPostgresClientKeyPath,
		},
		{
			name:           cryptoutilSharedMagic.AppSMKMSPostgres2Container,
			port:           cryptoutilSharedMagic.AppSMKMSPostgres2PublicPort,
			serverCertCN:   cryptoutilSharedMagic.AppSMKMSPostgres2ServerCertCN,
			clientCertPath: cryptoutilSharedMagic.AppSMKMSPostgresClientCertPath,
			clientKeyPath:  cryptoutilSharedMagic.AppSMKMSPostgresClientKeyPath,
		},
	}
}

// TestFullPipeline_AppPublicHTTPS_ServerCert verifies each sm-kms variant presents the correct Cat 3
// server certificate CN on its public HTTPS port.
// Cat 3 = public-https-server-entity-{PS-ID}-{variant}, signed by Cat 1 issuing CA.
func TestFullPipeline_AppPublicHTTPS_ServerCert(t *testing.T) {
	t.Parallel()

	waitForAppsHealthy(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath) // Cat 1 CA

	for _, tc := range allAppVariants() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clientCert := loadClientCert(t, tc.clientCertPath, tc.clientKeyPath)

			addr := fmt.Sprintf("127.0.0.1:%d", tc.port)
			conn, err := tls.DialWithDialer(
				&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout},
				"tcp", addr,
				&tls.Config{
					MinVersion:   tls.VersionTLS13,
					RootCAs:      caPool,
					Certificates: []tls.Certificate{clientCert},
				},
			)
			require.NoError(t, err, "TLS dial to app %q at %s must succeed with valid Cat 5 client cert", tc.name, addr)

			defer func() { _ = conn.Close() }()

			certs := conn.ConnectionState().PeerCertificates
			require.NotEmpty(t, certs, "app %q must present a server certificate", tc.name)

			cn := certs[0].Subject.CommonName
			require.Equal(t, tc.serverCertCN, cn,
				"app %q server cert CN must be Cat 3 identity", tc.name)
		})
	}
}

// TestFullPipeline_AppPublicHTTPS_HealthEndpoint verifies each sm-kms variant serves the health
// endpoint over HTTPS with mTLS. The Cat 5 service-user cert authenticates the test client.
func TestFullPipeline_AppPublicHTTPS_HealthEndpoint(t *testing.T) {
	t.Parallel()

	waitForAppsHealthy(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath) // Cat 1 CA

	for _, tc := range allAppVariants() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clientCert := loadClientCert(t, tc.clientCertPath, tc.clientKeyPath)

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
				Timeout:   cryptoutilSharedMagic.IMDefaultTimeout,
			}

			healthURL := fmt.Sprintf("https://127.0.0.1:%d%s", tc.port, cryptoutilSharedMagic.IdentityE2EHealthEndpoint)

			resp, err := client.Get(healthURL) //nolint:noctx // Simple E2E health check.
			require.NoError(t, err, "GET %s for app %q must succeed over HTTPS mTLS", healthURL, tc.name)

			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, resp.StatusCode,
				"app %q health endpoint must return 200; body: %s", tc.name, body)
		})
	}
}

// TestFullPipeline_AppPublicHTTPS_MTLSRejected verifies each sm-kms variant enforces Cat 4 mTLS.
// Connections without a client certificate must be rejected (RequireAndVerifyClientCert).
func TestFullPipeline_AppPublicHTTPS_MTLSRejected(t *testing.T) {
	t.Parallel()

	waitForAppsHealthy(t)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath) // Cat 1 CA

	for _, tc := range allAppVariants() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			addr := fmt.Sprintf("127.0.0.1:%d", tc.port)

			// No client certificate — app must reject connection (Cat 4 mTLS enforced).
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

				require.Failf(t, "TLS dial must fail for mTLS enforcement", "TLS dial to app %q at %s without client cert must fail (Cat 4 mTLS enforced)", tc.name, addr)
			}

			// Verify it's a TLS error, not a network error.
			require.Contains(t, err.Error(), "tls",
				"error from no-cert dial to app %q must be a TLS handshake failure", tc.name)
		})
	}
}
