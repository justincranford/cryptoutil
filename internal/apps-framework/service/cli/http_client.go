// Copyright (c) 2025-2026 Justin Cranford.
//

package cli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	http "net/http"
	"os"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// LoadCACertPool loads a CA certificate from file and returns an x509.CertPool.
// Returns nil if cacertPath is empty (uses system defaults).
func LoadCACertPool(cacertPath string) (*x509.CertPool, error) {
	if cacertPath == "" {
		return nil, nil //nolint:nilnil // Valid pattern: no CA cert specified means use system defaults
	}

	// Read CA certificate file.
	caCertPEM, err := os.ReadFile(cacertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	// Create certificate pool.
	caCertPool := x509.NewCertPool()

	// Parse and add certificates to pool.
	for {
		block, rest := pem.Decode(caCertPEM)
		if block == nil {
			break
		}

		if block.Type == cryptoutilSharedMagic.StringPEMTypeCertificate {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
			}

			caCertPool.AddCert(cert)
		}

		caCertPEM = rest
	}

	if len(caCertPool.Subjects()) == 0 { //nolint:staticcheck // Subjects() is safe for manually created CertPools
		return nil, fmt.Errorf("no CA certificates found in file")
	}

	return caCertPool, nil
}

// LoadClientCert loads a client certificate and key for mTLS authentication.
// Returns nil if both certPath and keyPath are empty (no mTLS client cert).
// Returns an error if only one of certPath/keyPath is provided.
func LoadClientCert(certPath, keyPath string) (*tls.Certificate, error) {
	switch {
	case certPath == "" && keyPath == "":
		return nil, nil //nolint:nilnil // Valid: no mTLS client cert requested
	case certPath == "" || keyPath == "":
		return nil, fmt.Errorf("both --cert and --key must be provided together (got cert=%q key=%q)", certPath, keyPath)
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
	}

	return &cert, nil
}

// buildTLSConfig constructs a TLS config with optional CA pool and optional client cert.
func buildTLSConfig(caCertPool *x509.CertPool, clientCert *tls.Certificate) *tls.Config {
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		RootCAs:            caCertPool,
		InsecureSkipVerify: caCertPool == nil, //nolint:gosec // Skip verification if no CA cert provided (backward compatibility)
	}

	if clientCert != nil {
		cfg.Certificates = []tls.Certificate{*clientCert}
	}

	return cfg
}

// HTTPGet performs an HTTP GET request with optional CA certificate validation and optional
// mTLS client certificate. Used by health check CLI wrappers to call API endpoints.
func HTTPGet(url, cacertPath, certPath, keyPath string) (int, string, error) {
	caCertPool, err := LoadCACertPool(cacertPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	clientCert, err := LoadClientCert(certPath, keyPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load client certificate: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: buildTLSConfig(caCertPool, clientCert),
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP GET failed: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

// HTTPPost performs an HTTP POST request with optional CA certificate validation and optional
// mTLS client certificate. Used by shutdown CLI wrapper to call admin API endpoint.
func HTTPPost(url, cacertPath, certPath, keyPath string) (int, string, error) {
	caCertPool, err := LoadCACertPool(cacertPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	clientCert, err := LoadClientCert(certPath, keyPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load client certificate: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: buildTLSConfig(caCertPool, clientCert),
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP POST failed: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, string(body), nil
}
