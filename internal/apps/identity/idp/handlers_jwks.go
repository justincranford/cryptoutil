// Copyright (c) 2025 Justin Cranford

package idp

import (
	json "encoding/json"
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// handleJWKS serves the JSON Web Key Set endpoint.
// GET /.well-known/jwks.json
//
// Returns the public keys used for signing tokens, allowing clients
// to verify JWT signatures.
func (s *Service) handleJWKS(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get key repository from factory.
	keyRepo := s.repoFactory.KeyRepository()

	// Get all active public signing keys.
	// Note: If no keys exist yet, we return an empty JWKS (valid per spec).
	keys, err := keyRepo.FindByUsage(ctx, cryptoutilSharedMagic.KeyUsageSigning, true)
	if err != nil {
		// For errors like "no keys found" or no Key table, return empty JWKS.
		// This is valid behavior before keys are created.
		keys = nil
	}

	// Create JWK Set.
	jwkSet := joseJwk.NewSet()

	// Add each public key to the set.
	for _, key := range keys {
		if key.PublicKey == "" {
			continue // Skip symmetric keys (no public key).
		}

		// Parse public key JWK.
		publicJWK, parseErr := joseJwk.ParseKey([]byte(key.PublicKey))
		if parseErr != nil {
			continue // Skip invalid keys.
		}

		// Add to set.
		if addErr := jwkSet.AddKey(publicJWK); addErr != nil {
			continue // Skip on add failure.
		}
	}

	// Marshal JWKS to JSON.
	jwksBytes, err := json.Marshal(jwkSet)
	if err != nil {
		jsonErr := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": "Failed to serialize JWKS",
		})

		return fmt.Errorf("failed to send JWKS serialization error: %w", jsonErr)
	}

	// Set cache headers for JWKS.
	c.Set("Content-Type", "application/json")
	c.Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour.

	if sendErr := c.Send(jwksBytes); sendErr != nil {
		return fmt.Errorf("failed to send JWKS response: %w", sendErr)
	}

	return nil
}
