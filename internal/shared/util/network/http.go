// Copyright (c) 2025 Justin Cranford
//
//

package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// HTTPGetLivez performs a GET /livez request to the private health endpoint.
func HTTPGetLivez(ctx context.Context, baseURL, adminContextPath string, timeout time.Duration, rootCAsPool *x509.CertPool, insecureSkipVerify bool) (int, http.Header, []byte, error) {
	fullPath := adminContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodGet, baseURL+fullPath, timeout, true, rootCAsPool, insecureSkipVerify)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get %s: %w", fullPath, err)
	}

	return statusCode, headers, body, nil
}

// HTTPGetReadyz performs a GET /readyz request to the private readiness endpoint.
func HTTPGetReadyz(ctx context.Context, baseURL, adminContextPath string, timeout time.Duration, rootCAsPool *x509.CertPool, insecureSkipVerify bool) (int, http.Header, []byte, error) {
	fullPath := adminContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodGet, baseURL+fullPath, timeout, true, rootCAsPool, insecureSkipVerify)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get %s: %w", fullPath, err)
	}

	return statusCode, headers, body, nil
}

// HTTPPostShutdown performs a POST /shutdown request to the private shutdown endpoint.
func HTTPPostShutdown(ctx context.Context, baseURL, adminContextPath string, timeout time.Duration, rootCAsPool *x509.CertPool, insecureSkipVerify bool) (int, http.Header, []byte, error) {
	fullPath := adminContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath

	statusCode, headers, body, err := HTTPResponse(ctx, http.MethodPost, baseURL+fullPath, timeout, true, rootCAsPool, insecureSkipVerify)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to post %s: %w", fullPath, err)
	}

	return statusCode, headers, body, nil
}

// HTTPResponse performs an HTTP request and returns the response details.
// It supports custom TLS configuration, timeout control, and redirect handling.
//
// Parameters:
//   - ctx: context for cancellation, deadlines, and tracing
//   - method: HTTP method (GET, POST, etc.)
//   - url: target URL
//   - timeout: request timeout (0 = no timeout)
//   - followRedirects: whether to follow HTTP redirects
//   - rootCAsPool: custom root CA pool for TLS verification (nil = system defaults)
//   - insecureSkipVerify: skip TLS certificate verification (for development)
//
// Returns the status code, headers, response body, and any error encountered.
func HTTPResponse(ctx context.Context, method, url string, timeout time.Duration, followRedirects bool, rootCAsPool *x509.CertPool, insecureSkipVerify bool) (int, http.Header, []byte, error) {
	if timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create %s request: %w", method, err)
	}

	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	if !followRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		}
	}

	if strings.HasPrefix(url, "https://") {
		transport := &http.Transport{}
		if rootCAsPool == nil && !insecureSkipVerify {
			transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		} else {
			tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
			if rootCAsPool != nil {
				tlsConfig.RootCAs = rootCAsPool
			}

			if insecureSkipVerify {
				tlsConfig.InsecureSkipVerify = true
			}

			transport.TLSClientConfig = tlsConfig
		}

		client.Transport = transport
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to make %s request: %w", method, err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, resp.Header, body, nil
}
