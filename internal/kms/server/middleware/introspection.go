// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// BatchIntrospector handles batch token introspection with deduplication.
type BatchIntrospector struct {
	config     IntrospectionConfig
	httpClient *http.Client

	// Cache for introspection results.
	cache     map[string]*introspectionCacheEntry
	cacheLock sync.RWMutex
}

// IntrospectionConfig configures the batch introspector.
type IntrospectionConfig struct {
	// IntrospectionURL is the token introspection endpoint.
	IntrospectionURL string

	// ClientID for authenticating introspection requests.
	ClientID string

	// ClientSecret for authenticating introspection requests.
	ClientSecret string

	// CacheTTL is how long to cache introspection results.
	CacheTTL time.Duration

	// MaxBatchSize is the maximum number of tokens per batch request.
	MaxBatchSize int

	// HTTPTimeout is the timeout for introspection requests.
	HTTPTimeout time.Duration
}

// introspectionCacheEntry stores cached introspection result.
type introspectionCacheEntry struct {
	active    bool
	expiresAt time.Time
}

// IntrospectionResult represents the result of token introspection.
type IntrospectionResult struct {
	// Active indicates whether the token is valid.
	Active bool `json:"active"`

	// TokenType is the type of token.
	TokenType string `json:"token_type,omitempty"`

	// Scope is the token's scope.
	Scope string `json:"scope,omitempty"`

	// ClientID is the client that requested the token.
	ClientID string `json:"client_id,omitempty"`

	// Username is the resource owner username.
	Username string `json:"username,omitempty"`

	// Subject is the token subject.
	Subject string `json:"sub,omitempty"`

	// Issuer is the token issuer.
	Issuer string `json:"iss,omitempty"`

	// Audience is the intended audience.
	Audience string `json:"aud,omitempty"`

	// ExpiresAt is the token expiration time (Unix timestamp).
	ExpiresAt int64 `json:"exp,omitempty"`

	// IssuedAt is when the token was issued (Unix timestamp).
	IssuedAt int64 `json:"iat,omitempty"`

	// NotBefore is when the token becomes valid (Unix timestamp).
	NotBefore int64 `json:"nbf,omitempty"`

	// JTI is the token identifier.
	JTI string `json:"jti,omitempty"`
}

// DefaultIntrospectionConfig returns a default configuration.
func DefaultIntrospectionConfig() IntrospectionConfig {
	return IntrospectionConfig{
		CacheTTL:     cryptoutilMagic.IntrospectionCacheTTL,
		MaxBatchSize: cryptoutilMagic.IntrospectionMaxBatchSize,
		HTTPTimeout:  cryptoutilMagic.IntrospectionHTTPTimeout,
	}
}

// NewBatchIntrospector creates a new batch introspector.
func NewBatchIntrospector(config IntrospectionConfig) (*BatchIntrospector, error) {
	if config.IntrospectionURL == "" {
		return nil, errors.New("introspection URL is required")
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = DefaultIntrospectionConfig().CacheTTL
	}

	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = DefaultIntrospectionConfig().MaxBatchSize
	}

	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = DefaultIntrospectionConfig().HTTPTimeout
	}

	return &BatchIntrospector{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		cache: make(map[string]*introspectionCacheEntry),
	}, nil
}

// Introspect checks a single token's validity.
func (b *BatchIntrospector) Introspect(ctx context.Context, token string) (*IntrospectionResult, error) {
	// Check cache first.
	if entry := b.getCached(token); entry != nil {
		return &IntrospectionResult{Active: entry.active}, nil
	}

	// Perform single introspection.
	result, err := b.performIntrospection(ctx, token)
	if err != nil {
		return nil, err
	}

	// Cache result.
	b.setCached(token, result.Active)

	return result, nil
}

// BatchIntrospect checks multiple tokens' validity with deduplication.
func (b *BatchIntrospector) BatchIntrospect(ctx context.Context, tokens []string) (map[string]*IntrospectionResult, error) {
	results := make(map[string]*IntrospectionResult)

	// Deduplicate tokens.
	uniqueTokens := b.deduplicateTokens(tokens)

	// Check cache for each token.
	uncachedTokens := make([]string, 0, len(uniqueTokens))

	for _, token := range uniqueTokens {
		if entry := b.getCached(token); entry != nil {
			results[token] = &IntrospectionResult{Active: entry.active}
		} else {
			uncachedTokens = append(uncachedTokens, token)
		}
	}

	// If all tokens are cached, return early.
	if len(uncachedTokens) == 0 {
		return results, nil
	}

	// Process uncached tokens in batches.
	for i := 0; i < len(uncachedTokens); i += b.config.MaxBatchSize {
		end := i + b.config.MaxBatchSize
		if end > len(uncachedTokens) {
			end = len(uncachedTokens)
		}

		batch := uncachedTokens[i:end]

		batchResults, err := b.processBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("batch introspection failed: %w", err)
		}

		for token, result := range batchResults {
			results[token] = result
			b.setCached(token, result.Active)
		}
	}

	return results, nil
}

// deduplicateTokens removes duplicate tokens from the list.
func (b *BatchIntrospector) deduplicateTokens(tokens []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0, len(tokens))

	for _, token := range tokens {
		if !seen[token] {
			seen[token] = true

			unique = append(unique, token)
		}
	}

	return unique
}

// processBatch introspects a batch of tokens.
// Note: Standard OAuth2 introspection only supports single tokens.
// This method processes tokens individually but could be extended for custom batch endpoints.
func (b *BatchIntrospector) processBatch(ctx context.Context, tokens []string) (map[string]*IntrospectionResult, error) {
	results := make(map[string]*IntrospectionResult)

	// Standard OAuth2 introspection doesn't support batch requests.
	// Process tokens individually.
	for _, token := range tokens {
		result, err := b.performIntrospection(ctx, token)
		if err != nil {
			// Continue on error but mark token as inactive.
			results[token] = &IntrospectionResult{Active: false}

			continue
		}

		results[token] = result
	}

	return results, nil
}

// performIntrospection makes an introspection request.
func (b *BatchIntrospector) performIntrospection(ctx context.Context, token string) (*IntrospectionResult, error) {
	reqBody := "token=" + token

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.config.IntrospectionURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add client authentication.
	if b.config.ClientID != "" && b.config.ClientSecret != "" {
		req.SetBasicAuth(b.config.ClientID, b.config.ClientSecret)
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection returned status %d", resp.StatusCode)
	}

	var result IntrospectionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	return &result, nil
}

// getCached retrieves a cached introspection result.
func (b *BatchIntrospector) getCached(token string) *introspectionCacheEntry {
	b.cacheLock.RLock()
	defer b.cacheLock.RUnlock()

	entry, exists := b.cache[token]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry
}

// setCached stores an introspection result in cache.
func (b *BatchIntrospector) setCached(token string, active bool) {
	b.cacheLock.Lock()
	defer b.cacheLock.Unlock()

	b.cache[token] = &introspectionCacheEntry{
		active:    active,
		expiresAt: time.Now().Add(b.config.CacheTTL),
	}
}

// ClearCache clears all cached introspection results.
func (b *BatchIntrospector) ClearCache() {
	b.cacheLock.Lock()
	defer b.cacheLock.Unlock()

	b.cache = make(map[string]*introspectionCacheEntry)
}

// CacheSize returns the current cache size.
func (b *BatchIntrospector) CacheSize() int {
	b.cacheLock.RLock()
	defer b.cacheLock.RUnlock()

	return len(b.cache)
}
