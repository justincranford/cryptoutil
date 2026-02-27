// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"fmt"
	"log/slog"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleDeviceAuthorization handles POST /device_authorization - RFC 8628 Section 3.1.
//
// Device requests authorization by providing client_id and optional scope.
// Server responds with device_code, user_code, verification_uri, and polling interval.
//
// Request parameters:
// - client_id (required): OAuth 2.0 client identifier.
// - scope (optional): Space-delimited list of requested scopes.
//
// Response fields:
// - device_code: Opaque device verification code for polling /token endpoint.
// - user_code: Human-readable code for user to enter on verification URI.
// - verification_uri: URI where user visits to authorize device.
// - verification_uri_complete: Optional URI with user_code pre-filled.
// - expires_in: Device code lifetime in seconds (default: 1800).
// - interval: Minimum polling interval in seconds (default: 5).
func (s *Service) handleDeviceAuthorization(c *fiber.Ctx) error {
	clientID := c.FormValue(cryptoutilSharedMagic.ParamClientID)
	scope := c.FormValue(cryptoutilSharedMagic.ParamScope)

	// Validate required parameters.
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "client_id is required",
		})
	}

	ctx := c.Context()

	// Validate client exists.
	clientRepo := s.repoFactory.ClientRepository()

	_, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		slog.ErrorContext(ctx, "Client not found for device authorization", cryptoutilSharedMagic.ClaimClientID, clientID, cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidClient,
			"error_description":               "Invalid client_id",
		})
	}

	// Generate device code (opaque token for polling).
	deviceCode, err := GenerateDeviceCode()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate device code", cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to generate device code",
		})
	}

	// Generate user code (human-readable code for verification).
	userCode, err := GenerateUserCode()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate user code", cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to generate user code",
		})
	}

	// Create device authorization record.
	authID := googleUuid.Must(googleUuid.NewV7())

	deviceAuth := &cryptoutilIdentityDomain.DeviceAuthorization{
		ID:         authID,
		ClientID:   clientID,
		DeviceCode: deviceCode,
		UserCode:   userCode,
		Scope:      scope,
		Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
		CreatedAt:  time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(cryptoutilSharedMagic.DefaultDeviceCodeLifetime),
	}

	deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()
	if err := deviceAuthRepo.Create(ctx, deviceAuth); err != nil {
		slog.ErrorContext(ctx, "Failed to store device authorization", cryptoutilSharedMagic.StringError, err, cryptoutilSharedMagic.ParamDeviceCode, deviceCode[:cryptoutilSharedMagic.IMMinPasswordLength]+"...")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to store device authorization",
		})
	}

	// Construct verification URIs.
	// Use s.config.Tokens.Issuer for the base URL (e.g., "https://authz.example.com").
	baseURL := s.config.Tokens.Issuer
	if baseURL == "" {
		// Fallback to hostname for development/testing.
		baseURL = fmt.Sprintf("https://%s", c.Hostname())
	}

	verificationURI := fmt.Sprintf("%s/device", baseURL)
	verificationURIComplete := fmt.Sprintf("%s?%s=%s", verificationURI, cryptoutilSharedMagic.ParamUserCode, userCode)

	slog.InfoContext(ctx, "Device authorization request created",
		"device_code_prefix", deviceCode[:cryptoutilSharedMagic.IMMinPasswordLength]+"...",
		cryptoutilSharedMagic.ParamUserCode, userCode,
		cryptoutilSharedMagic.ClaimClientID, clientID,
		cryptoutilSharedMagic.ClaimScope, scope,
		"expires_at", deviceAuth.ExpiresAt,
	)

	// Return RFC 8628 Section 3.2 response.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.ParamDeviceCode: deviceCode,
		cryptoutilSharedMagic.ParamUserCode:   userCode,
		"verification_uri":                    verificationURI,
		"verification_uri_complete":           verificationURIComplete,
		cryptoutilSharedMagic.ParamExpiresIn:  int(cryptoutilSharedMagic.DefaultDeviceCodeLifetime.Seconds()),
		"interval":                            int(cryptoutilSharedMagic.DefaultPollingInterval.Seconds()),
	})
}
