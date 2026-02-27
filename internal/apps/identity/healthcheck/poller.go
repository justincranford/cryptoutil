// Copyright (c) 2025 Justin Cranford

// Package healthcheck provides health check polling functionality with exponential backoff.
package healthcheck

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"time"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

// newClientConfigFn is an injectable function for TLS client configuration, enabling testing of the fallback path.
var newClientConfigFn = cryptoutilSharedCryptoTls.NewClientConfig

// Response represents a health check JSON response.
type Response struct {
	Status   string            `json:"status"`
	Database string            `json:"database,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

const (
	defaultInitialInterval = 1 * time.Second
)

// Poller polls service health endpoints using poll.Until for retry.
type Poller struct {
	client   *http.Client
	timeout  time.Duration
	interval time.Duration
}

// NewPoller creates a new health check poller.
// The skipTLSVerify parameter should only be true in development/testing environments.
func NewPoller(timeout time.Duration, maxRetries int, skipTLSVerify bool) *Poller {
	// Use internal/infra/tls/ for consistent TLS configuration across the project.
	tlsConfig, err := newClientConfigFn(&cryptoutilSharedCryptoTls.ClientConfigOptions{
		SkipVerify: skipTLSVerify, // Only true in dev/test per Session 4 Q4
	})
	if err != nil {
		// Fallback to nil transport if TLS config fails (should not happen).
		return &Poller{
			client: &http.Client{
				Timeout: timeout,
			},
			timeout:  time.Duration(maxRetries) * defaultInitialInterval,
			interval: defaultInitialInterval,
		}
	}

	return &Poller{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig.TLSConfig,
			},
		},
		timeout:  time.Duration(maxRetries) * defaultInitialInterval,
		interval: defaultInitialInterval,
	}
}

// Poll polls a health endpoint until it returns healthy or the timeout elapses.
func (p *Poller) Poll(ctx context.Context, url string) (*Response, error) {
	var lastResp *Response

	err := cryptoutilSharedUtilPoll.Until(ctx, p.timeout, p.interval, func(ctx context.Context) (bool, error) {
		resp, checkErr := p.check(ctx, url)
		if checkErr == nil && resp.Status == cryptoutilSharedMagic.DockerServiceHealthHealthy {
			lastResp = resp

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return lastResp, nil
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
