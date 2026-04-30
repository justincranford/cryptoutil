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

// Client is a typed HTTP client for jose-ja APIs.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a jose-ja client.
func New(baseURL string) *Client {
	requestTimeout := time.Duration(cryptoutilSharedMagic.DefaultMaxIdleConns) * time.Second

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// GetJWKS fetches the JSON Web Key Set payload.
func (c *Client) GetJWKS(ctx context.Context) (map[string]any, error) {
	var out map[string]any

	err := c.doJSON(ctx, http.MethodGet, "/service/api/v1/jwks", nil, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// CreateJWK creates a new JWK with the given request payload.
func (c *Client) CreateJWK(ctx context.Context, request map[string]any) error {
	return c.doJSON(ctx, http.MethodPost, "/service/api/v1/jwks", request, nil)
}

// RotateJWK rotates active key material.
func (c *Client) RotateJWK(ctx context.Context, request map[string]any) error {
	return c.doJSON(ctx, http.MethodPost, "/service/api/v1/jwks/rotate", request, nil)
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
