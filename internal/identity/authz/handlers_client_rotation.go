// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilCrypto "cryptoutil/internal/crypto"
	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// handleClientSecretRotation handles POST /oauth2/v1/clients/{id}/rotate-secret.
// Rotates the client secret, invalidating the old secret and generating a new one.
func (s *Service) handleClientSecretRotation(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse client ID from URL parameter.
	idParam := c.Params("id")

	clientID, err := googleUuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": fmt.Sprintf("Invalid client ID format: %v", err),
		})
	}

	// Authenticate the requesting client (must be the client itself or an admin).
	// For this implementation, we'll use the Authorization header to authenticate.
	authenticatedClient, err := s.authenticateClient(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": "Client authentication failed",
		})
	}

	// Verify the authenticated client is the same as the client being rotated.
	// In production, you might also allow admin clients to rotate any client's secret.
	if authenticatedClient.ID != clientID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "Client can only rotate its own secret",
		})
	}

	// Retrieve the client from database.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByID(ctx, clientID)
	if err != nil {
		if errors.Is(err, cryptoutilIdentityAppErr.ErrClientNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
				"error_description": "Client not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to retrieve client",
		})
	}

	// Generate new client secret (32 bytes = 256 bits of entropy).
	secretBytes := make([]byte, cryptoutilIdentityMagic.ClientSecretLength)
	if _, err := crand.Read(secretBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to generate new secret",
		})
	}

	newSecretPlaintext := base64.URLEncoding.EncodeToString(secretBytes)

	// Hash the new secret using PBKDF2-HMAC-SHA256 (FIPS-approved).
	hashedSecret, err := cryptoutilCrypto.HashSecret(newSecretPlaintext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to hash new secret",
		})
	}

	// Update client with new secret.
	// Store old secret hash in rotation history (future enhancement: ClientSecretHistory table).
	oldSecretHash := client.ClientSecret
	client.ClientSecret = hashedSecret
	client.UpdatedAt = time.Now()

	err = clientRepo.Update(ctx, client)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to rotate secret",
		})
	}

	// Note: In production, log rotation for audit trail with old/new hash prefixes.
	_ = oldSecretHash // Suppress unused variable warning

	// Return the new plaintext secret (this is the ONLY time it will be available).
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"client_id":     client.ClientID,
		"client_secret": newSecretPlaintext,
		"rotated_at":    client.UpdatedAt,
		"message":       "Client secret rotated successfully. Store this secret securely - it will not be shown again.",
	})
}
