// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleClientSecretRotation handles POST /oauth2/v1/clients/{id}/rotate-secret.
// Rotates the client secret, invalidating the old secret and generating a new one.
func (s *Service) handleClientSecretRotation(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse client_id from URL parameter (string, not UUID).
	clientIDParam := c.Params("id")

	// Authenticate the requesting client (must be the client itself or an admin).
	authenticatedClient, err := s.authenticateClient(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidClient,
			"error_description":               "Client authentication failed",
		})
	}

	// Verify the authenticated client is the same as the client being rotated.
	// In production, you might also allow admin clients to rotate any client's secret.
	if authenticatedClient.ClientID != clientIDParam {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorAccessDenied,
			"error_description":               "Client can only rotate its own secret",
		})
	}

	// Retrieve the client from database.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientIDParam)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrClientNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
				"error_description":               "Client not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to retrieve client",
		})
	}

	// Generate new client secret (32 bytes = 256 bits of entropy).
	secretBytes := make([]byte, cryptoutilSharedMagic.ClientSecretLength)
	if _, err := crand.Read(secretBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to generate new secret",
		})
	}

	newSecretPlaintext := base64.URLEncoding.EncodeToString(secretBytes)

	// Hash the new secret using PBKDF2-HMAC-SHA256 (FIPS-approved).
	hashedSecret, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(newSecretPlaintext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to hash new secret",
		})
	}

	// Rotate secret in repository (archives old secret in history table, updates client).
	rotatedBy := authenticatedClient.ClientID
	reason := "Client-initiated rotation"

	err = clientRepo.RotateSecret(ctx, client.ID, hashedSecret, rotatedBy, reason)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to rotate secret",
		})
	}

	// Return the new plaintext secret (this is the ONLY time it will be available).
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.ClaimClientID:     client.ClientID,
		cryptoutilSharedMagic.ParamClientSecret: newSecretPlaintext,
		"rotated_at":                            time.Now().UTC(),
		"message":                               "Client secret rotated successfully. Store this secret securely - it will not be shown again.",
	})
}
