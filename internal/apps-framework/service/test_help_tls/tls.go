// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_tls provides TLS material creation, certificate/client construction,
// and secure/insecure test client helpers for integration and E2E test suites.
// It handles test TLS certificate generation, mTLS client setup, and client configuration.
//
// Consumed by:
//   - test_orch_e2e: TLS material for E2E tests
//   - test_orch_integration: TLS clients and certificates
//   - TLS test suites: certificate validation and mTLS testing
package test_help_tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	http "net/http"
	"testing"

	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type generateAutoTLSSettingsFn func([]string, []string, int) (*cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings, error)

// NewTestTLSSettings generates ephemeral TLS settings for tests.
//
// The returned settings include a server certificate chain and private key suitable
// for local HTTPS listeners using localhost/loopback SANs.
func NewTestTLSSettings(t *testing.T) *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings {
	t.Helper()

	return mustTestTLSSettings(cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings)
}

// NewTestTLSSettingsForTestMain returns TLS settings without requiring *testing.T.
func NewTestTLSSettingsForTestMain() *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings {
	return mustTestTLSSettings(cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings)
}

func mustTestTLSSettings(generatorFn generateAutoTLSSettingsFn) *cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings {
	tlsSettings, err := newTestTLSSettingsWithGenerator(generatorFn)
	if err != nil {
		panic(fmt.Sprintf("test_help_tls: generate auto TLS settings: %v", err))
	}

	return tlsSettings
}

func newTestTLSSettingsWithGenerator(generatorFn generateAutoTLSSettingsFn) (*cryptoutilAppsFrameworkServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
	tlsSettings, err := generatorFn(
		[]string{cryptoutilSharedMagic.DefaultOTLPHostnameDefault},
		[]string{cryptoutilSharedMagic.IPv4Loopback, cryptoutilSharedMagic.IPv6Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("generate auto TLS settings: %w", err)
	}

	return tlsSettings, nil
}

// NewInsecureHTTPSClient returns an HTTPS client configured for test-only insecure TLS.
func NewInsecureHTTPSClient(t *testing.T) *http.Client {
	t.Helper()

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Test helper intentionally bypasses cert validation.
		},
		DisableKeepAlives: true,
	}

	return &http.Client{Transport: transport}
}

// NewMTLSClient returns an HTTPS client configured with a client certificate and
// optional CA trust pool for server verification.
func NewMTLSClient(t *testing.T, certPath, keyPath string, caPool *x509.CertPool) *http.Client {
	t.Helper()

	client, err := newMTLSClient(certPath, keyPath, caPool)
	if err != nil {
		panic(fmt.Sprintf("test_help_tls: create mTLS client: %v", err))
	}

	return client
}

func newMTLSClient(certPath, keyPath string, caPool *x509.CertPool) (*http.Client, error) {
	if certPath == "" {
		return nil, fmt.Errorf("certPath must be non-empty")
	}

	if keyPath == "" {
		return nil, fmt.Errorf("keyPath must be non-empty")
	}

	clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("load client certificate/key pair: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      caPool,
			MinVersion:   tls.VersionTLS13,
		},
		DisableKeepAlives: true,
	}

	return &http.Client{Transport: transport}, nil
}
