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
)

// HTTPGetHealthzResponse returns the full response from /livez endpoint for content validation.
func HTTPGetHealthzResponse(baseURL *string, rootCAsPool *x509.CertPool) ([]byte, int, error) {
	url := *baseURL + "/livez"
	body, _, statusCode, err := HTTPResponse(context.Background(), http.MethodGet, url, 2*time.Second, true, rootCAsPool)
	return body, statusCode, err
}

// HTTPGetReadyzResponse returns the full response from /readyz endpoint for content validation.
func HTTPGetReadyzResponse(baseURL *string, rootCAsPool *x509.CertPool) ([]byte, int, error) {
	url := *baseURL + "/readyz"
	body, _, statusCode, err := HTTPResponse(context.Background(), http.MethodGet, url, 2*time.Second, true, rootCAsPool)
	return body, statusCode, err
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
//
// Returns the response body, headers, status code, and any error encountered.
func HTTPResponse(ctx context.Context, method, url string, timeout time.Duration, followRedirects bool, rootCAsPool *x509.CertPool) ([]byte, http.Header, int, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create %s request: %w", method, err)
	}
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		}
	}
	if strings.HasPrefix(url, "https://") {
		transport := &http.Transport{}
		if rootCAsPool == nil {
			transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		} else {
			transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: rootCAsPool}
		}
		client.Transport = transport
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to make %s request: %w", method, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp.Header, resp.StatusCode, nil
}
