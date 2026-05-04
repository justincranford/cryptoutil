// Copyright (c) 2025-2026 Justin Cranford.
package client

import (
	"context"
	json "encoding/json"
	"fmt"
	http "net/http"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Client is a typed HTTP client for identity-idp APIs.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a identity-idp client.
func New(baseURL string) *Client {
	requestTimeout := time.Duration(cryptoutilSharedMagic.DefaultMaxIdleConns) * time.Second

	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: &http.Client{Timeout: requestTimeout}}
}

// Ping executes a health request against the service path.
func (c *Client) Ping(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodGet, cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/health", nil, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any) error {
	var bodyReader *strings.Reader
	if in == nil {
		bodyReader = strings.NewReader("")
	} else {
		payload, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		bodyReader = strings.NewReader(string(payload))
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
