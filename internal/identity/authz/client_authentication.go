// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// authenticateClient authenticates a client using the configured authentication method.
func (s *Service) authenticateClient(c *fiber.Ctx) (*cryptoutilIdentityDomain.Client, error) {
	// Try HTTP Basic authentication first.
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Basic ") {
		credential := strings.TrimPrefix(authHeader, "Basic ")

		// Decode to get client_id:client_secret.
		decoded, err := base64.StdEncoding.DecodeString(credential)
		if err != nil {
			return nil, fmt.Errorf("failed to decode basic auth: %w", err)
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) == 2 {
			clientID := parts[0]

			authenticator, ok := s.clientAuth.GetAuthenticator(cryptoutilIdentityMagic.ClientAuthMethodSecretBasic)
			if !ok {
				return nil, fiber.ErrUnauthorized
			}

			client, err := authenticator.Authenticate(c.Context(), clientID, credential)
			if err != nil {
				return nil, fmt.Errorf("basic auth failed: %w", err)
			}

			return client, nil
		}
	}

	// Try POST body authentication.
	clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)
	clientSecret := c.FormValue(cryptoutilIdentityMagic.ParamClientSecret)

	if clientID != "" && clientSecret != "" {
		authenticator, ok := s.clientAuth.GetAuthenticator(cryptoutilIdentityMagic.ClientAuthMethodSecretPost)
		if !ok {
			return nil, fiber.ErrUnauthorized
		}

		client, err := authenticator.Authenticate(c.Context(), clientID, clientSecret)
		if err != nil {
			return nil, fmt.Errorf("post auth failed: %w", err)
		}

		return client, nil
	}

	// Try mTLS client authentication.
	// For mTLS, we need the client_id to determine which auth method to use
	if clientID != "" {
		// Get client from database to determine auth method
		client, err := s.repoFactory.ClientRepository().GetByClientID(c.Context(), clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get client: %w", err)
		}

		// Check if this client uses mTLS authentication
		if client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth ||
			client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth {
			authenticator, ok := s.clientAuth.GetAuthenticator(string(client.TokenEndpointAuthMethod))
			if !ok {
				return nil, fiber.ErrUnauthorized
			}

			// For mTLS, the credential is the client certificate chain
			// In a real implementation, this would be extracted from the TLS connection
			// For now, we'll expect it to be passed as a form parameter
			clientCert := c.FormValue("client_certificate")
			if clientCert == "" {
				return nil, fiber.ErrUnauthorized
			}

			client, err := authenticator.Authenticate(c.Context(), clientID, clientCert)
			if err != nil {
				return nil, fmt.Errorf("mTLS auth failed: %w", err)
			}

			return client, nil
		}
	}

	return nil, fiber.ErrUnauthorized
}
