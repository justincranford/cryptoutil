// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"log/slog"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// handleToken handles POST /token - OAuth 2.1 token endpoint.
func (s *Service) handleDeviceCodeGrant(c *fiber.Ctx) error {
	// Extract device_code and client_id parameters.
	deviceCode := c.FormValue(cryptoutilIdentityMagic.ParamDeviceCode)
	clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)

	// Validate required parameters.
	if deviceCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "device_code is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	ctx := c.Context()

	// Retrieve device authorization from repository.
	deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		slog.ErrorContext(ctx, "Device authorization not found", "device_code", deviceCode, "error", err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "Invalid or expired device_code",
		})
	}

	// Validate client_id matches.
	if deviceAuth.ClientID != clientID {
		slog.ErrorContext(ctx, "Client ID mismatch for device authorization", "expected", deviceAuth.ClientID, "actual", clientID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "client_id does not match device_code",
		})
	}

	// Check if device code is expired (RFC 8628 Section 3.5).
	if deviceAuth.IsExpired() {
		slog.InfoContext(ctx, "Device code expired", "device_code", deviceCode, "expires_at", deviceAuth.ExpiresAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorExpiredToken,
			"error_description": "device_code has expired",
		})
	}

	// Check if device code has already been used (RFC 8628 Section 3.5).
	if deviceAuth.IsUsed() {
		slog.WarnContext(ctx, "Device code already used", "device_code", deviceCode, "used_at", deviceAuth.UsedAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidGrant,
			"error_description": "device_code has already been used",
		})
	}

	// Enforce polling rate limiting (RFC 8628 Section 3.5).
	if deviceAuth.LastPolledAt != nil {
		minPollTime := deviceAuth.LastPolledAt.Add(cryptoutilIdentityMagic.DefaultPollingInterval)
		if time.Now().UTC().Before(minPollTime) {
			slog.DebugContext(ctx, "Polling too fast", "device_code", deviceCode, "last_polled_at", deviceAuth.LastPolledAt)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorSlowDown,
				"error_description": "Polling too fast, slow down",
			})
		}
	}

	// Update last polled timestamp.
	now := time.Now().UTC()
	deviceAuth.LastPolledAt = &now

	if err := deviceAuthRepo.Update(ctx, deviceAuth); err != nil {
		slog.ErrorContext(ctx, "Failed to update device authorization polling timestamp", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update polling timestamp",
		})
	}

	// Check authorization status (RFC 8628 Section 3.5).
	switch {
	case deviceAuth.IsPending():
		// User has not yet authorized - client should continue polling.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAuthorizationPending,
			"error_description": "Authorization pending, user has not yet authorized",
		})
	case deviceAuth.IsDenied():
		// User explicitly denied authorization.
		slog.InfoContext(ctx, "Device authorization denied by user", "device_code", deviceCode)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorAccessDenied,
			"error_description": "User denied authorization",
		})
	case deviceAuth.IsAuthorized():
		// User has authorized - issue tokens.
		return s.issueDeviceCodeTokens(c, deviceAuth)
	default:
		// Unknown status - should never happen.
		slog.ErrorContext(ctx, "Invalid device authorization status", "status", deviceAuth.Status)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Invalid device authorization status",
		})
	}
}

// issueDeviceCodeTokens issues access and refresh tokens for authorized device code.
func (s *Service) issueDeviceCodeTokens(c *fiber.Ctx, deviceAuth *cryptoutilIdentityDomain.DeviceAuthorization) error {
	ctx := c.Context()

	// Mark device code as used.
	now := time.Now().UTC()
	deviceAuth.Status = cryptoutilIdentityDomain.DeviceAuthStatusUsed
	deviceAuth.UsedAt = &now

	deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()
	if err := deviceAuthRepo.Update(ctx, deviceAuth); err != nil {
		slog.ErrorContext(ctx, "Failed to mark device code as used", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update device authorization status",
		})
	}

	// Validate UserID is set (device must be authorized by a user).
	if !deviceAuth.UserID.Valid {
		slog.ErrorContext(ctx, "Device authorization missing user ID", "device_code", deviceAuth.DeviceCode)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Device authorization missing user ID",
		})
	}

	// Retrieve client for token generation.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, deviceAuth.ClientID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve client for device code token", "client_id", deviceAuth.ClientID, "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to retrieve client",
		})
	}

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Token service not configured",
		})
	}

	// Generate access token with standard OAuth 2.1 claims.
	accessTokenClaims := map[string]any{
		"sub":       deviceAuth.UserID.UUID.String(),
		"client_id": deviceAuth.ClientID,
		"scope":     deviceAuth.Scope,
		"exp":       time.Now().UTC().Add(time.Duration(client.AccessTokenLifetime) * time.Second).Unix(),
		"iat":       time.Now().UTC().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed for device code", "error", err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": appErr.Message,
		})
	}

	// Generate refresh token (optional, based on client configuration).
	var refreshToken string

	if client.RefreshTokenLifetime > 0 {
		refreshToken, err = s.tokenSvc.IssueRefreshToken(ctx)
		if err != nil {
			appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

			slog.ErrorContext(ctx, "Refresh token issuance failed for device code", "error", err)

			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				"error":             cryptoutilIdentityMagic.ErrorServerError,
				"error_description": appErr.Message,
			})
		}
	}

	slog.InfoContext(ctx, "Device code token exchange successful",
		"client_id", deviceAuth.ClientID,
		"user_id", deviceAuth.UserID.UUID.String(),
		"scope", deviceAuth.Scope,
	)

	// Build token response (RFC 8628 Section 3.5).
	tokenResponse := fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   client.AccessTokenLifetime,
		"scope":        deviceAuth.Scope,
	}

	if refreshToken != "" {
		tokenResponse["refresh_token"] = refreshToken
	}

	return c.Status(fiber.StatusOK).JSON(tokenResponse)
}
