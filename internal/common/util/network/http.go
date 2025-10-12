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

// GetHealthzResponse returns the full response from /livez endpoint for content validation
func GetHealthzResponse(baseURL *string, rootCAsPool *x509.CertPool) ([]byte, int, error) {
	url := *baseURL + "/livez"
	return httpGetWithResponse(&url, 2*time.Second, rootCAsPool)
}

// GetReadyzResponse returns the full response from /readyz endpoint for content validation
func GetReadyzResponse(baseURL *string, rootCAsPool *x509.CertPool) ([]byte, int, error) {
	url := *baseURL + "/readyz"
	return httpGetWithResponse(&url, 2*time.Second, rootCAsPool)
}

// httpGetWithResponse returns the response body, status code, and any error
func httpGetWithResponse(url *string, timeout time.Duration, rootCAsPool *x509.CertPool) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{}

	if strings.HasPrefix(*url, "https://") {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{
			RootCAs:    rootCAsPool,
			MinVersion: tls.VersionTLS12,
		}}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("get %v failed: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp.StatusCode, nil
}
