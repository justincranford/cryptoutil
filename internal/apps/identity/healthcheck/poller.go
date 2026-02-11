// Copyright (c) 2025 Justin Cranford

// Package healthcheck provides health check polling functionality with exponential backoff.
package healthcheck

import (
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"time"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// Response represents a health check JSON response.
type Response struct {
	Status   string            `json:"status"`
	Database string            `json:"database,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

const (
	defaultInitialInterval = 1 * time.Second
	defaultMaxInterval     = 30 * time.Second
)

// Poller polls service health endpoints with exponential backoff retry.
type Poller struct {
	client          *http.Client
	maxRetries      int
	initialInterval time.Duration
	maxInterval     time.Duration
}

// NewPoller creates a new health check poller.
// The skipTLSVerify parameter should only be true in development/testing environments.
func NewPoller(timeout time.Duration, maxRetries int, skipTLSVerify bool) *Poller {
	// Use internal/infra/tls/ for consistent TLS configuration across the project.
	tlsConfig, err := cryptoutilSharedCryptoTls.NewClientConfig(&cryptoutilSharedCryptoTls.ClientConfigOptions{
		SkipVerify: skipTLSVerify, // Only true in dev/test per Session 4 Q4
	})
	if err != nil {
		// Fallback to nil transport if TLS config fails (should not happen).
		return &Poller{
			client: &http.Client{
				Timeout: timeout,
			},
			maxRetries:      maxRetries,
			initialInterval: defaultInitialInterval,
			maxInterval:     defaultMaxInterval,
		}
	}

	return &Poller{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig.TLSConfig,
			},
		},
		maxRetries:      maxRetries,
		initialInterval: defaultInitialInterval,
		maxInterval:     defaultMaxInterval,
	}
}

// Poll polls a health endpoint until it returns healthy or max retries reached.
func (p *Poller) Poll(ctx context.Context, url string) (*Response, error) {
	interval := p.initialInterval

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("polling canceled: %w", ctx.Err())
		default:
		}

		resp, err := p.check(ctx, url)
		if err == nil && resp.Status == "healthy" {
			return resp, nil
		}

		// Wait with exponential backoff before retry
		if attempt < p.maxRetries-1 {
			time.Sleep(interval)

			interval = interval * 2
			if interval > p.maxInterval {
				interval = p.maxInterval
			}
		}
	}

	return nil, fmt.Errorf("health check failed after %d attempts", p.maxRetries)
}

// Check performs a single health check request.
func (p *Poller) check(ctx context.Context, url string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpResp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		_ = httpResp.Body.Close() //nolint:errcheck // HTTP response body close - error not critical in health check context
	}()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &resp, nil
}
