// Copyright (c) 2025 Justin Cranford

package healthcheck

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Response represents a health check JSON response.
type Response struct {
	Status   string            `json:"status"`
	Database string            `json:"database,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

// Poller polls service health endpoints with exponential backoff retry.
type Poller struct {
	client          *http.Client
	maxRetries      int
	initialInterval time.Duration
	maxInterval     time.Duration
}

// NewPoller creates a new health check poller.
func NewPoller(timeout time.Duration, maxRetries int) *Poller {
	return &Poller{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Development/testing only
				},
			},
		},
		maxRetries:      maxRetries,
		initialInterval: 1 * time.Second,
		maxInterval:     30 * time.Second,
	}
}

// Poll polls a health endpoint until it returns healthy or max retries reached.
func (p *Poller) Poll(ctx context.Context, url string) (*Response, error) {
	interval := p.initialInterval

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
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
	defer httpResp.Body.Close()

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
