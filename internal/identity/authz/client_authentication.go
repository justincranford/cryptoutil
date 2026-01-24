// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	fiber "github.com/gofiber/fiber/v2"

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
			clientSecret := parts[1]

			authenticator, ok := s.clientAuth.GetAuthenticator(cryptoutilIdentityMagic.ClientAuthMethodSecretBasic)
			if !ok {
				return nil, fiber.ErrUnauthorized
			}

			client, err := authenticator.Authenticate(c.Context(), clientID, clientSecret)
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

			// Extract client certificate from TLS connection.
			// Fiber stores TLS peer certificates in c.Context().TLS.PeerCertificates.
			// For mTLS auth, the client certificate chain must be provided via TLS handshake.
			tlsConn := c.Context().Conn()
			if tlsConn == nil {
				return nil, fmt.Errorf("mTLS required but no TLS connection found")
			}

			// Get the first certificate (client certificate) from the peer chain.
			// The TLS handshake has already validated the certificate chain using the configured ClientCAs.
			// Here we just need to extract it for application-level verification (subject, fingerprint).
			// Note: In production, Fiber/fasthttp must be configured with tls.Config.ClientAuth = tls.RequireAndVerifyClientCert
			// and tls.Config.ClientCAs set to the trusted CA pool.
			peerCerts := c.Context().TLSConnectionState().PeerCertificates
			if len(peerCerts) == 0 {
				return nil, fmt.Errorf("mTLS required but no client certificate provided")
			}

			// Encode peer certificates as PEM for authenticator.
			var pemChain strings.Builder
			for _, cert := range peerCerts {
				err := pem.Encode(&pemChain, &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: cert.Raw,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to encode certificate: %w", err)
				}
			}

			client, err := authenticator.Authenticate(c.Context(), clientID, pemChain.String())
			if err != nil {
				return nil, fmt.Errorf("mTLS auth failed: %w", err)
			}

			return client, nil
		}
	}

	return nil, fiber.ErrUnauthorized
}
