// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	http "net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "modernc.org/sqlite"             // CGO-free SQLite driver

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func imServiceLivez(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageLivez)

		return 0
	}

	// Parse flags.
	url := defaultLivezURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultLivezURL { // Only set if not already set
				baseURL := args[i+1]

				livezPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath
				if !strings.HasSuffix(baseURL, livezPath) {
					url = baseURL + livezPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call livez endpoint.
	statusCode, body, err := httpGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Liveness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		_, _ = fmt.Fprintf(stdout, "✅ Service is alive (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "❌ Service is not alive (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// imServiceReadyz implements the readyz subcommand.
// CLI wrapper calling the admin readiness check API.
func imServiceReadyz(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageReadyz)

		return 0
	}

	// Parse flags.
	url := defaultReadyzURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultReadyzURL { // Only set if not already set
				baseURL := args[i+1]

				readyzPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath
				if !strings.HasSuffix(baseURL, readyzPath) {
					url = baseURL + readyzPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call readyz endpoint.
	statusCode, body, err := httpGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Readiness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		_, _ = fmt.Fprintf(stdout, "✅ Service is ready (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "❌ Service is not ready (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// imServiceShutdown implements the shutdown subcommand.
// CLI wrapper calling the admin graceful shutdown API.
func imServiceShutdown(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, IMUsageShutdown)

		return 0
	}

	// Parse flags.
	url := defaultShutdownURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultShutdownURL { // Only set if not already set
				baseURL := args[i+1]

				shutdownPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath
				if !strings.HasSuffix(baseURL, shutdownPath) {
					url = baseURL + shutdownPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call shutdown endpoint.
	statusCode, body, err := httpPost(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Shutdown request failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK || statusCode == http.StatusAccepted {
		_, _ = fmt.Fprintf(stdout, "✅ Shutdown initiated (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "❌ Shutdown request failed (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// loadCACertPool loads a CA certificate from file and returns an x509.CertPool.
func loadCACertPool(cacertPath string) (*x509.CertPool, error) {
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

		if block.Type == "CERTIFICATE" {
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

// httpGet performs an HTTP GET request with optional CA certificate validation.
// Used by health check CLI wrappers to call API endpoints.
func httpGet(url, cacertPath string) (int, string, error) {
	// Load CA certificate pool if specified.
	caCertPool, err := loadCACertPool(cacertPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Create HTTP client with proper TLS configuration.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				RootCAs:            caCertPool,        // Use CA cert pool if provided, nil = system defaults
				InsecureSkipVerify: caCertPool == nil, // Skip verification if no CA cert provided (backward compatibility)
			},
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

// httpPost performs an HTTP POST request with optional CA certificate validation.
// Used by shutdown CLI wrapper to call admin API endpoint.
func httpPost(url, cacertPath string) (int, string, error) {
	// Load CA certificate pool if specified.
	caCertPool, err := loadCACertPool(cacertPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Create HTTP client with proper TLS configuration.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				RootCAs:            caCertPool,        // Use CA cert pool if provided, nil = system defaults
				InsecureSkipVerify: caCertPool == nil, // Skip verification if no CA cert provided (backward compatibility)
			},
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
