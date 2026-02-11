// Copyright (c) 2025 Justin Cranford
//
//

// Package realm provides identity provider federation for KMS.
package realm

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	http "net/http"
	"strings"
	"sync"
	"time"
)

// FederatedProvider represents a federated identity provider.
type FederatedProvider struct {
	// ID is the unique provider identifier.
	ID string `yaml:"id" json:"id"`

	// Name is the human-readable provider name.
	Name string `yaml:"name" json:"name"`

	// Type is the provider type (oidc only).
	Type FederationProviderType `yaml:"type" json:"type"`

	// IssuerURL is the OIDC issuer URL.
	IssuerURL string `yaml:"issuer_url" json:"issuer_url"`

	// JWKSURL is the JWKS endpoint URL.
	JWKSURL string `yaml:"jwks_url,omitempty" json:"jwks_url,omitempty"`

	// ClientID is the OIDC client ID.
	ClientID string `yaml:"client_id" json:"client_id"`

	// ClientSecret is the OIDC client secret (optional for public clients).
	ClientSecret string `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`

	// TenantMappings maps provider claims to tenant IDs.
	TenantMappings []TenantMapping `yaml:"tenant_mappings,omitempty" json:"tenant_mappings,omitempty"`

	// AllowedAudiences is the list of allowed audience values.
	AllowedAudiences []string `yaml:"allowed_audiences,omitempty" json:"allowed_audiences,omitempty"`

	// Enabled indicates if the provider is active.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// TrustSelfSigned allows self-signed certificates (for development).
	TrustSelfSigned bool `yaml:"trust_self_signed,omitempty" json:"trust_self_signed,omitempty"`
}

// FederationProviderType defines the type of federated provider.
type FederationProviderType string

const (
	// FederationTypeOIDC is an OpenID Connect provider.
	FederationTypeOIDC FederationProviderType = "oidc"
)

// TenantMapping maps provider claims to tenant IDs.
type TenantMapping struct {
	// ClaimName is the claim to match.
	ClaimName string `yaml:"claim_name" json:"claim_name"`

	// ClaimValue is the value to match.
	ClaimValue string `yaml:"claim_value" json:"claim_value"`

	// TenantID is the mapped tenant ID.
	TenantID string `yaml:"tenant_id" json:"tenant_id"`

	// Priority determines matching order (lower = higher priority).
	Priority int `yaml:"priority" json:"priority"`
}

// OIDCDiscoveryDocument represents OIDC discovery metadata.
type OIDCDiscoveryDocument struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint,omitempty"`
	JWKSURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported,omitempty"`
	ClaimsSupported       []string `json:"claims_supported,omitempty"`
}

// FederationManager manages federated identity providers.
type FederationManager struct {
	providers      map[string]*FederatedProvider
	discoveryCache map[string]*discoveryEntry
	httpClient     *http.Client
	mu             sync.RWMutex
}

// discoveryEntry holds cached discovery documents.
type discoveryEntry struct {
	document  *OIDCDiscoveryDocument
	expiresAt time.Time
	lastFetch time.Time
}

// FederationManagerConfig configures the federation manager.
type FederationManagerConfig struct {
	// HTTPTimeout is the HTTP client timeout.
	HTTPTimeout time.Duration `yaml:"http_timeout" json:"http_timeout"`

	// DiscoveryCacheTTL is the cache TTL for discovery documents.
	DiscoveryCacheTTL time.Duration `yaml:"discovery_cache_ttl" json:"discovery_cache_ttl"`
}

// Default federation configuration values.
const (
	defaultHTTPTimeout       = 30 * time.Second
	defaultDiscoveryCacheTTL = 1 * time.Hour
)

// NewFederationManager creates a new federation manager.
func NewFederationManager(config *FederationManagerConfig) *FederationManager {
	if config == nil {
		config = &FederationManagerConfig{
			HTTPTimeout:       defaultHTTPTimeout,
			DiscoveryCacheTTL: defaultDiscoveryCacheTTL,
		}
	}

	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = defaultHTTPTimeout
	}

	return &FederationManager{
		providers:      make(map[string]*FederatedProvider),
		discoveryCache: make(map[string]*discoveryEntry),
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

// RegisterProvider registers a federated identity provider.
func (m *FederationManager) RegisterProvider(provider *FederatedProvider) error {
	if provider == nil {
		return errors.New("provider cannot be nil")
	}

	if provider.ID == "" {
		return errors.New("provider ID is required")
	}

	if provider.IssuerURL == "" {
		return errors.New("issuer URL is required")
	}

	if provider.Type != FederationTypeOIDC {
		return fmt.Errorf("unsupported provider type: %s (only OIDC supported)", provider.Type)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[provider.ID]; exists {
		return fmt.Errorf("provider %s already exists", provider.ID)
	}

	m.providers[provider.ID] = provider

	return nil
}

// GetProvider retrieves a provider by ID.
func (m *FederationManager) GetProvider(providerID string) (*FederatedProvider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[providerID]

	return provider, ok
}

// GetProviderByIssuer retrieves a provider by issuer URL.
func (m *FederationManager) GetProviderByIssuer(issuerURL string) (*FederatedProvider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Normalize issuer URL.
	issuerURL = strings.TrimSuffix(issuerURL, "/")

	for _, provider := range m.providers {
		normalizedIssuer := strings.TrimSuffix(provider.IssuerURL, "/")
		if normalizedIssuer == issuerURL {
			return provider, true
		}
	}

	return nil, false
}

// ListProviders returns all registered providers.
func (m *FederationManager) ListProviders() []FederatedProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]FederatedProvider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, *provider)
	}

	return providers
}

// UnregisterProvider removes a provider.
func (m *FederationManager) UnregisterProvider(providerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[providerID]; !exists {
		return fmt.Errorf("provider %s not found", providerID)
	}

	delete(m.providers, providerID)

	// Clear discovery cache for this provider.
	for key := range m.discoveryCache {
		if strings.HasPrefix(key, providerID+":") {
			delete(m.discoveryCache, key)
		}
	}

	return nil
}

// DiscoverOIDC fetches OIDC discovery document for a provider.
func (m *FederationManager) DiscoverOIDC(ctx context.Context, providerID string) (*OIDCDiscoveryDocument, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %s not found", providerID)
	}

	if provider.Type != FederationTypeOIDC {
		return nil, fmt.Errorf("provider %s is not OIDC type", providerID)
	}

	// Check cache.
	cacheKey := providerID + ":discovery"

	m.mu.RLock()
	entry, cached := m.discoveryCache[cacheKey]
	m.mu.RUnlock()

	if cached && time.Now().UTC().Before(entry.expiresAt) {
		return entry.document, nil
	}

	// Fetch discovery document.
	discoveryURL := strings.TrimSuffix(provider.IssuerURL, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read discovery response: %w", err)
	}

	var doc OIDCDiscoveryDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse discovery document: %w", err)
	}

	// Cache the result.
	m.mu.Lock()
	m.discoveryCache[cacheKey] = &discoveryEntry{
		document:  &doc,
		expiresAt: time.Now().UTC().Add(defaultDiscoveryCacheTTL),
		lastFetch: time.Now().UTC(),
	}
	m.mu.Unlock()

	return &doc, nil
}

// MapTenantFromClaims maps JWT claims to a tenant ID using provider mappings.
func (m *FederationManager) MapTenantFromClaims(providerID string, claims map[string]any) (string, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerID]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("provider %s not found", providerID)
	}

	if len(provider.TenantMappings) == 0 {
		return "", errors.New("no tenant mappings configured")
	}

	// Sort mappings by priority.
	mappings := make([]TenantMapping, len(provider.TenantMappings))
	copy(mappings, provider.TenantMappings)

	// Simple priority sort.
	for i := range mappings {
		for j := i + 1; j < len(mappings); j++ {
			if mappings[j].Priority < mappings[i].Priority {
				mappings[i], mappings[j] = mappings[j], mappings[i]
			}
		}
	}

	// Find matching mapping.
	for _, mapping := range mappings {
		claimValue, ok := claims[mapping.ClaimName]
		if !ok {
			continue
		}

		// Convert to string for comparison.
		var strValue string

		switch v := claimValue.(type) {
		case string:
			strValue = v
		case []any:
			// Check if any element matches.
			for _, elem := range v {
				if s, ok := elem.(string); ok && s == mapping.ClaimValue {
					return mapping.TenantID, nil
				}
			}

			continue
		default:
			strValue = fmt.Sprintf("%v", v)
		}

		if strValue == mapping.ClaimValue {
			return mapping.TenantID, nil
		}
	}

	return "", errors.New("no matching tenant mapping found")
}

// ValidateAudience validates token audience against provider configuration.
func (m *FederationManager) ValidateAudience(providerID string, audience []string) error {
	m.mu.RLock()
	provider, ok := m.providers[providerID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("provider %s not found", providerID)
	}

	if len(provider.AllowedAudiences) == 0 {
		// No audience restriction.
		return nil
	}

	// Check if any audience matches.
	for _, aud := range audience {
		for _, allowed := range provider.AllowedAudiences {
			if aud == allowed {
				return nil
			}
		}
	}

	return errors.New("token audience not allowed")
}

// GetJWKSURL returns the JWKS URL for a provider.
func (m *FederationManager) GetJWKSURL(ctx context.Context, providerID string) (string, error) {
	m.mu.RLock()
	provider, ok := m.providers[providerID]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("provider %s not found", providerID)
	}

	// Use configured JWKS URL if available.
	if provider.JWKSURL != "" {
		return provider.JWKSURL, nil
	}

	// Otherwise, discover from OIDC metadata.
	doc, err := m.DiscoverOIDC(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("failed to discover JWKS URL: %w", err)
	}

	return doc.JWKSURI, nil
}
