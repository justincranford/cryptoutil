// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.AuthZ != nil {
		if err := c.AuthZ.Validate(); err != nil {
			return fmt.Errorf("authz config: %w", err)
		}
	}

	if c.IDP != nil {
		if err := c.IDP.Validate(); err != nil {
			return fmt.Errorf("idp config: %w", err)
		}
	}

	if c.RS != nil {
		if err := c.RS.Validate(); err != nil {
			return fmt.Errorf("rs config: %w", err)
		}
	}

	if c.Database != nil {
		if err := c.Database.Validate(); err != nil {
			return fmt.Errorf("database config: %w", err)
		}
	}

	if c.Tokens != nil {
		if err := c.Tokens.Validate(); err != nil {
			return fmt.Errorf("tokens config: %w", err)
		}
	}

	if c.Sessions != nil {
		if err := c.Sessions.Validate(); err != nil {
			return fmt.Errorf("sessions config: %w", err)
		}
	}

	if c.Security != nil {
		if err := c.Security.Validate(); err != nil {
			return fmt.Errorf("security config: %w", err)
		}
	}

	if c.Observability != nil {
		if err := c.Observability.Validate(); err != nil {
			return fmt.Errorf("observability config: %w", err)
		}
	}

	return nil
}

// Validate validates token configuration.
func (tc *TokenConfig) Validate() error {
	if tc.AccessTokenLifetime <= 0 {
		return fmt.Errorf("access token lifetime must be positive")
	}

	if tc.RefreshTokenLifetime <= 0 {
		return fmt.Errorf("refresh token lifetime must be positive")
	}

	if tc.IDTokenLifetime <= 0 {
		return fmt.Errorf("ID token lifetime must be positive")
	}

	if tc.AccessTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatJWS && tc.AccessTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatJWE && tc.AccessTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatUUID {
		return fmt.Errorf("access token format must be 'jws', 'jwe', or 'uuid'")
	}

	if tc.RefreshTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatUUID {
		return fmt.Errorf("refresh token format must be 'uuid'")
	}

	if tc.IDTokenFormat != cryptoutilSharedMagic.IdentityTokenFormatJWS {
		return fmt.Errorf("ID token format must be 'jws'")
	}

	if tc.Issuer == "" {
		return fmt.Errorf("token issuer is required")
	}

	if tc.SigningAlgorithm == "" {
		return fmt.Errorf("signing algorithm is required")
	}

	return nil
}

// Validate validates security configuration.
func (sc *SecurityConfig) Validate() error {
	if sc.PKCEChallengeMethod != cryptoutilSharedMagic.PKCEMethodS256 && sc.PKCEChallengeMethod != cryptoutilSharedMagic.PKCEMethodPlain {
		return fmt.Errorf("pKCE challenge method must be 'S256' or 'plain'")
	}

	if sc.RateLimitEnabled {
		if sc.RateLimitRequests <= 0 {
			return fmt.Errorf("rate limit requests must be positive")
		}

		if sc.RateLimitWindow <= 0 {
			return fmt.Errorf("rate limit window must be positive")
		}
	}

	return nil
}
