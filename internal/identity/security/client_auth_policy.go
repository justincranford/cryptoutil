// Copyright (c) 2025 Justin Cranford
//
//

// Package security provides authentication and authorization security policies.
package security

import (
	"context"
	"fmt"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// ClientAuthPolicy defines policies for client authentication methods.
type ClientAuthPolicy struct {
	// AllowedMethods defines which authentication methods are permitted for this policy.
	AllowedMethods []cryptoutilIdentityDomain.ClientAuthMethod

	// RequireMTLS indicates whether mTLS is required for this policy.
	RequireMTLS bool

	// RequireJWTSignature indicates whether JWT-based auth must use specific algorithms.
	RequireJWTSignature bool

	// AllowedJWTAlgorithms defines permitted JWT signing algorithms.
	AllowedJWTAlgorithms []string

	// RequireCertificateValidation indicates whether certificate validation is enforced.
	RequireCertificateValidation bool

	// AllowSelfSignedCertificates indicates whether self-signed certificates are permitted.
	AllowSelfSignedCertificates bool

	// MaxCertificateAge defines maximum certificate age in days (0 = no limit).
	MaxCertificateAge int
}

// DefaultClientAuthPolicy returns the default authentication policy.
func DefaultClientAuthPolicy() *ClientAuthPolicy {
	return &ClientAuthPolicy{
		AllowedMethods: []cryptoutilIdentityDomain.ClientAuthMethod{
			cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			cryptoutilIdentityDomain.ClientAuthMethodSecretJWT,
			cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT,
			cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
		},
		RequireMTLS:                  false,
		RequireJWTSignature:          true,
		AllowedJWTAlgorithms:         []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"},
		RequireCertificateValidation: true,
		AllowSelfSignedCertificates:  false,
		MaxCertificateAge:            cryptoutilIdentityMagic.DefaultCertificateMaxAgeDays,
	}
}

// StrictClientAuthPolicy returns a strict authentication policy for high-security environments.
func StrictClientAuthPolicy() *ClientAuthPolicy {
	return &ClientAuthPolicy{
		AllowedMethods: []cryptoutilIdentityDomain.ClientAuthMethod{
			cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT,
			cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
		},
		RequireMTLS:                  true,
		RequireJWTSignature:          true,
		AllowedJWTAlgorithms:         []string{"RS256", "ES256"},
		RequireCertificateValidation: true,
		AllowSelfSignedCertificates:  false,
		MaxCertificateAge:            cryptoutilIdentityMagic.StrictCertificateMaxAgeDays,
	}
}

// PublicClientAuthPolicy returns a policy for public clients (SPAs, mobile apps).
func PublicClientAuthPolicy() *ClientAuthPolicy {
	return &ClientAuthPolicy{
		AllowedMethods: []cryptoutilIdentityDomain.ClientAuthMethod{
			cryptoutilIdentityDomain.ClientAuthMethodNone,
		},
		RequireMTLS:                  false,
		RequireJWTSignature:          false,
		AllowedJWTAlgorithms:         []string{},
		RequireCertificateValidation: false,
		AllowSelfSignedCertificates:  false,
		MaxCertificateAge:            0,
	}
}

// DevelopmentClientAuthPolicy returns a permissive policy for development environments.
func DevelopmentClientAuthPolicy() *ClientAuthPolicy {
	return &ClientAuthPolicy{
		AllowedMethods: []cryptoutilIdentityDomain.ClientAuthMethod{
			cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			cryptoutilIdentityDomain.ClientAuthMethodSecretJWT,
			cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT,
			cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
			cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
			cryptoutilIdentityDomain.ClientAuthMethodNone,
		},
		RequireMTLS:                  false,
		RequireJWTSignature:          false,
		AllowedJWTAlgorithms:         []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "HS256"},
		RequireCertificateValidation: false,
		AllowSelfSignedCertificates:  true,
		MaxCertificateAge:            0,
	}
}

// ClientAuthPolicyManager manages authentication policies for different client profiles.
type ClientAuthPolicyManager struct {
	policies map[string]*ClientAuthPolicy
}

// NewClientAuthPolicyManager creates a new policy manager with default policies.
func NewClientAuthPolicyManager() *ClientAuthPolicyManager {
	return &ClientAuthPolicyManager{
		policies: map[string]*ClientAuthPolicy{
			"default":     DefaultClientAuthPolicy(),
			"strict":      StrictClientAuthPolicy(),
			"public":      PublicClientAuthPolicy(),
			"development": DevelopmentClientAuthPolicy(),
		},
	}
}

// GetPolicy retrieves the policy for a given profile name.
func (m *ClientAuthPolicyManager) GetPolicy(profileName string) (*ClientAuthPolicy, error) {
	policy, ok := m.policies[profileName]
	if !ok {
		return nil, fmt.Errorf("policy profile not found: %s", profileName)
	}

	return policy, nil
}

// RegisterPolicy registers a custom policy.
func (m *ClientAuthPolicyManager) RegisterPolicy(profileName string, policy *ClientAuthPolicy) {
	m.policies[profileName] = policy
}

// ValidateClientAuthMethod validates if the given auth method is allowed by the policy.
func (p *ClientAuthPolicy) ValidateClientAuthMethod(_ context.Context, method cryptoutilIdentityDomain.ClientAuthMethod) error {
	for _, allowed := range p.AllowedMethods {
		if allowed == method {
			return nil
		}
	}

	return fmt.Errorf("authentication method %s not allowed by policy", method)
}

// ValidateJWTAlgorithm validates if the given JWT algorithm is allowed by the policy.
func (p *ClientAuthPolicy) ValidateJWTAlgorithm(_ context.Context, algorithm string) error {
	if !p.RequireJWTSignature {
		return nil
	}

	for _, allowed := range p.AllowedJWTAlgorithms {
		if allowed == algorithm {
			return nil
		}
	}

	return fmt.Errorf("JWT algorithm %s not allowed by policy", algorithm)
}
