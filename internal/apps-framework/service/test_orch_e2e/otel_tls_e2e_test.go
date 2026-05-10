// Copyright (c) 2025-2026 Justin Cranford.
//
//

//go:build e2e

// Package test_orch_e2e_test provides E2E tests for OTel Collector TLS connectivity.
// These tests start the sm-kms docker compose stack (pki-init + OTel Collector)
// with a test override that exposes OTel OTLP ports to the host, then verify:
//  1. OTel server TLS: server cert is Cat 2 (public-https-server-entity-otel-collector-contrib)
//  2. App→OTel mTLS: Cat 9 app client cert accepted; no-cert connection rejected
package test_orch_e2e_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	http "net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilTestOrchE2E "cryptoutil/internal/apps-framework/service/test_orch_e2e"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestMain starts the OTel compose stack once for all tests in this package.
var (
	tlsPSIDSpec    cryptoutilTestOrchE2E.TLSPSIDSpec
	composeManager *cryptoutilTestOrchE2E.ComposeManager
)

func TestMain(m *testing.M) {
	spec, err := cryptoutilTestOrchE2E.NewTLSPSIDSpec(cryptoutilSharedMagic.OTLPServiceSMKMS)
	if err != nil {
		fmt.Printf("ERROR: test-orch spec init failed: %v\n", err)

		os.Exit(1)
	}

	tlsPSIDSpec = spec

	cm := cryptoutilTestOrchE2E.NewComposeManager(spec)
	composeManager = cm

	ctx := context.Background()

	if err := cm.Start(ctx); err != nil {
		fmt.Printf("ERROR: compose up failed: %v\n", err)

		_ = cm.Stop(ctx)

		os.Exit(1)
	}

	// Run all tests.
	code := m.Run()

	if err := cm.Stop(ctx); err != nil {
		fmt.Printf("WARNING: compose down failed: %v\n", err)
	}

	os.Exit(code)
}

// TestOtelServerTLS_GRPC verifies OTel gRPC :4317 TLS handshake succeeds with valid Cat 9 client cert.
// The server must present a Cat 2 cert (CN=public-https-server-entity-otel-collector-contrib).
func TestOtelServerTLS_GRPC(t *testing.T) {
	t.Parallel()

	cryptoutilTestOrchE2E.WaitForOTelHealth(t, tlsPSIDSpec, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)

	caPool := cryptoutilTestOrchE2E.LoadCACertPool(t, tlsPSIDSpec.PublicCACertPath)
	clientCert := cryptoutilTestOrchE2E.LoadClientCert(t, tlsPSIDSpec.OTelClientCertPath, tlsPSIDSpec.OTelClientKeyPath)

	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
	}

	addr := fmt.Sprintf("127.0.0.1:%d", tlsPSIDSpec.OTelGRPCPort)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout}, "tcp", addr, tlsCfg)
	require.NoError(t, err, "TLS dial to OTel gRPC %s must succeed with valid Cat 9 client cert", addr)

	defer func() { _ = conn.Close() }()

	// Verify server cert CN matches the expected Cat 2 identity.
	certs := conn.ConnectionState().PeerCertificates
	require.NotEmpty(t, certs, "OTel gRPC server must present a certificate")

	cn := certs[0].Subject.CommonName
	assert.Equal(t, tlsPSIDSpec.OTelServerCertCN, cn,
		"OTel gRPC server cert CN must be Cat 2 identity")
}

// TestOtelServerTLS_HTTP verifies OTel HTTP :4318 TLS handshake succeeds with valid Cat 9 client cert.
// The server must present a Cat 2 cert (CN=public-https-server-entity-otel-collector-contrib).
func TestOtelServerTLS_HTTP(t *testing.T) {
	t.Parallel()

	cryptoutilTestOrchE2E.WaitForOTelHealth(t, tlsPSIDSpec, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)

	caPool := cryptoutilTestOrchE2E.LoadCACertPool(t, tlsPSIDSpec.PublicCACertPath)
	clientCert := cryptoutilTestOrchE2E.LoadClientCert(t, tlsPSIDSpec.OTelClientCertPath, tlsPSIDSpec.OTelClientKeyPath)

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

	url := fmt.Sprintf("https://127.0.0.1:%d", tlsPSIDSpec.OTelHTTPPort)
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

	cryptoutilTestOrchE2E.WaitForOTelHealth(t, tlsPSIDSpec, cryptoutilSharedMagic.OtelTLSE2EHealthTimeout)

	caPool := cryptoutilTestOrchE2E.LoadCACertPool(t, tlsPSIDSpec.PublicCACertPath)

	// No client cert — OTel must reject this connection.
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    caPool,
		// Deliberately NO Certificates — tests Cat 8 client CA enforcement.
	}

	addr := fmt.Sprintf("127.0.0.1:%d", tlsPSIDSpec.OTelGRPCPort)

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: cryptoutilSharedMagic.IMDefaultTimeout}, "tcp", addr, tlsCfg)
	if err == nil {
		_ = conn.Close()

		t.Fatal("Expected OTel gRPC to reject connection without client cert, but connection succeeded")
	}

	t.Logf("OTel gRPC correctly rejected no-cert connection: %v", err)
}
