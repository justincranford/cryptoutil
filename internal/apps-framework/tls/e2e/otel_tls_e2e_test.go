// Copyright (c) 2025-2026 Justin Cranford.
//
//

//go:build e2e

// Package e2e provides E2E tests for OTel Collector TLS connectivity.
// These tests start the sm-kms docker compose stack (pki-init + OTel Collector)
// with a test override that exposes OTel OTLP ports to the host, then verify:
//  1. OTel server TLS: server cert is Cat 2 (public-https-server-entity-otel-collector-contrib)
//  2. App→OTel mTLS: Cat 9 app client cert accepted; no-cert connection rejected
package e2e_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	http "net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// otelComposeManager provides minimal compose lifecycle for OTel TLS tests.
// It runs TWO compose files (main + test port-expose override) so Go tests
// can directly dial OTel gRPC/HTTP from the host.
type otelComposeManager struct {
	mainFile     string
	overrideFile string
}

func newOtelComposeManager(main, override string) *otelComposeManager {
	return &otelComposeManager{mainFile: main, overrideFile: override}
}

func (m *otelComposeManager) start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", m.mainFile,
		"-f", m.overrideFile,
		"up", "-d", "--build",
		cryptoutilSharedMagic.PSIDPKIInit,
		cryptoutilSharedMagic.OtelTLSE2EContainer,
		cryptoutilSharedMagic.GrafanaTLSE2EContainer,
		cryptoutilSharedMagic.AppSMKMSSQLite1Container,
		cryptoutilSharedMagic.AppSMKMSSQLite2Container,
		cryptoutilSharedMagic.AppSMKMSPostgres1Container,
		cryptoutilSharedMagic.AppSMKMSPostgres2Container,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compose up failed: %w", err)
	}

	return nil
}

func (m *otelComposeManager) stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", m.mainFile,
		"-f", m.overrideFile,
		"down", "-v",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compose down failed: %w", err)
	}

	return nil
}

// loadCACertPool reads a PEM CA cert file and returns an x509.CertPool.
func loadCACertPool(t *testing.T, caPath string) *x509.CertPool {
	t.Helper()

	caPEM, err := os.ReadFile(caPath) //nolint:gosec // CA cert path from trusted test config.
	require.NoError(t, err, "read CA cert %q", caPath)

	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(caPEM), "parse CA cert from %q", caPath)

	return pool
}

// loadClientCert reads a PEM cert+key pair for mTLS.
func loadClientCert(t *testing.T, certPath, keyPath string) tls.Certificate {
	t.Helper()

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	require.NoError(t, err, "load client cert %q / %q", certPath, keyPath)

	return cert
}

// waitForOtelHealth polls the OTel health endpoint until ready or timeout.
func waitForOtelHealth(t *testing.T, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().UTC().Add(timeout)

	client := &http.Client{
		Timeout: cryptoutilSharedMagic.OtelCollectorHealthCheckTimeout,
	}

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/", cryptoutilSharedMagic.OtelTLSE2EHealthPort)

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

// TestMain starts the OTel compose stack once for all tests in this package.
var composeManager *otelComposeManager

func TestMain(m *testing.M) {
	cm := newOtelComposeManager(cryptoutilSharedMagic.OtelTLSE2EComposeFile, cryptoutilSharedMagic.OtelTLSE2EComposeOverrideFile)
	composeManager = cm

	ctx := context.Background()

	if err := cm.start(ctx); err != nil {
		fmt.Printf("ERROR: compose up failed: %v\n", err)

		_ = cm.stop(ctx)

		os.Exit(1)
	}

	// Wait for OTel health endpoint to be ready.
	waitFn := func(t *testing.T) {
		t.Helper()
		waitForOtelHealth(t, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)
	}

	_ = waitFn // used in sub-tests below

	// Run all tests.
	code := m.Run()

	if err := cm.stop(ctx); err != nil {
		fmt.Printf("WARNING: compose down failed: %v\n", err)
	}

	os.Exit(code)
}

// TestOtelServerTLS_GRPC verifies OTel gRPC :4317 TLS handshake succeeds with valid Cat 9 client cert.
// The server must present a Cat 2 cert (CN=public-https-server-entity-otel-collector-contrib).
func TestOtelServerTLS_GRPC(t *testing.T) {
	t.Parallel()

	waitForOtelHealth(t, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath)
	clientCert := loadClientCert(t, cryptoutilSharedMagic.OtelTLSE2EClientCertPath, cryptoutilSharedMagic.OtelTLSE2EClientKeyPath)

	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
	}

	addr := fmt.Sprintf("127.0.0.1:%d", cryptoutilSharedMagic.OtelTLSE2EGRPCPort)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout}, "tcp", addr, tlsCfg)
	require.NoError(t, err, "TLS dial to OTel gRPC %s must succeed with valid Cat 9 client cert", addr)

	defer func() { _ = conn.Close() }()

	// Verify server cert CN matches the expected Cat 2 identity.
	certs := conn.ConnectionState().PeerCertificates
	require.NotEmpty(t, certs, "OTel gRPC server must present a certificate")

	cn := certs[0].Subject.CommonName
	assert.Equal(t, cryptoutilSharedMagic.OtelTLSE2EOtelServerCertCN, cn,
		"OTel gRPC server cert CN must be Cat 2 identity")
}

// TestOtelServerTLS_HTTP verifies OTel HTTP :4318 TLS handshake succeeds with valid Cat 9 client cert.
// The server must present a Cat 2 cert (CN=public-https-server-entity-otel-collector-contrib).
func TestOtelServerTLS_HTTP(t *testing.T) {
	t.Parallel()

	waitForOtelHealth(t, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath)
	clientCert := loadClientCert(t, cryptoutilSharedMagic.OtelTLSE2EClientCertPath, cryptoutilSharedMagic.OtelTLSE2EClientKeyPath)

	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
	}

	transport := &http.Transport{
		TLSClientConfig:   tlsCfg,
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   cryptoutilSharedMagic.IMDefaultTimeout,
	}

	url := fmt.Sprintf("https://127.0.0.1:%d", cryptoutilSharedMagic.OtelTLSE2EHTTPPort)
	resp, err := client.Get(url) //nolint:noctx // Direct TLS verification, no context needed.
	// OTel HTTP receiver responds with 405 on GET (expects OTLP POST) — that's OK.
	// We only care that the TLS handshake succeeded (no certificate errors).
	if err == nil {
		defer func() { _ = resp.Body.Close() }()
		// Any non-TLS HTTP response confirms TLS handshake success.
		t.Logf("OTel HTTP TLS handshake succeeded: status=%d", resp.StatusCode)
	} else {
		// err may be HTTP-level (method not allowed) but TLS succeeded — check it's not TLS error.
		require.NotContains(t, err.Error(), "certificate", "TLS handshake to OTel HTTP must succeed")
		require.NotContains(t, err.Error(), "tls", "TLS handshake to OTel HTTP must succeed")
	}
}

// TestOtelMTLS_Rejection verifies that connections WITHOUT a client cert are rejected by OTel.
// The OTel Collector is configured with client_ca_file (Cat 8 CA), so no-cert connections must fail.
func TestOtelMTLS_Rejection(t *testing.T) {
	t.Parallel()

	waitForOtelHealth(t, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)

	caPool := loadCACertPool(t, cryptoutilSharedMagic.OtelTLSE2ECACertPath)

	// No client cert — OTel must reject this connection.
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    caPool,
		// Deliberately NO Certificates — tests Cat 8 client CA enforcement.
	}

	addr := fmt.Sprintf("127.0.0.1:%d", cryptoutilSharedMagic.OtelTLSE2EGRPCPort)

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout}, "tcp", addr, tlsCfg)
	if err == nil {
		_ = conn.Close()

		t.Fatal("Expected OTel gRPC to reject connection without client cert, but connection succeeded")
	}

	t.Logf("OTel gRPC correctly rejected no-cert connection: %v", err)
}
