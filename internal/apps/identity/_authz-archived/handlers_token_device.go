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
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleToken handles POST /token - OAuth 2.1 token endpoint.
func (s *Service) handleDeviceCodeGrant(c *fiber.Ctx) error {
	// Extract device_code and client_id parameters.
	deviceCode := c.FormValue(cryptoutilSharedMagic.ParamDeviceCode)
	clientID := c.FormValue(cryptoutilSharedMagic.ParamClientID)

	// Validate required parameters.
	if deviceCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "device_code is required",
		})
	}

	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "client_id is required",
		})
	}

	ctx := c.Context()

	// Retrieve device authorization from repository.
	deviceAuthRepo := s.repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		slog.ErrorContext(ctx, "Device authorization not found", cryptoutilSharedMagic.ParamDeviceCode, deviceCode, cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description":               "Invalid or expired device_code",
		})
	}

	// Validate client_id matches.
	if deviceAuth.ClientID != clientID {
		slog.ErrorContext(ctx, "Client ID mismatch for device authorization", "expected", deviceAuth.ClientID, "actual", clientID)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description":               "client_id does not match device_code",
		})
	}

	// Check if device code is expired (RFC 8628 Section 3.5).
	if deviceAuth.IsExpired() {
		slog.InfoContext(ctx, "Device code expired", cryptoutilSharedMagic.ParamDeviceCode, deviceCode, "expires_at", deviceAuth.ExpiresAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorExpiredToken,
			"error_description":               "device_code has expired",
		})
	}

	// Check if device code has already been used (RFC 8628 Section 3.5).
	if deviceAuth.IsUsed() {
		slog.WarnContext(ctx, "Device code already used", cryptoutilSharedMagic.ParamDeviceCode, deviceCode, "used_at", deviceAuth.UsedAt)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidGrant,
			"error_description":               "device_code has already been used",
		})
	}

	// Enforce polling rate limiting (RFC 8628 Section 3.5).
	if deviceAuth.LastPolledAt != nil {
		minPollTime := deviceAuth.LastPolledAt.Add(cryptoutilSharedMagic.DefaultPollingInterval)
		if time.Now().UTC().Before(minPollTime) {
			slog.DebugContext(ctx, "Polling too fast", cryptoutilSharedMagic.ParamDeviceCode, deviceCode, "last_polled_at", deviceAuth.LastPolledAt)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorSlowDown,
				"error_description":               "Polling too fast, slow down",
			})
		}
	}

	// Update last polled timestamp.
	now := time.Now().UTC()
	deviceAuth.LastPolledAt = &now

	if err := deviceAuthRepo.Update(ctx, deviceAuth); err != nil {
		slog.ErrorContext(ctx, "Failed to update device authorization polling timestamp", cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to update polling timestamp",
		})
	}

	// Check authorization status (RFC 8628 Section 3.5).
	switch {
	case deviceAuth.IsPending():
		// User has not yet authorized - client should continue polling.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorAuthorizationPending,
			"error_description":               "Authorization pending, user has not yet authorized",
		})
	case deviceAuth.IsDenied():
		// User explicitly denied authorization.
		slog.InfoContext(ctx, "Device authorization denied by user", cryptoutilSharedMagic.ParamDeviceCode, deviceCode)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorAccessDenied,
			"error_description":               "User denied authorization",
		})
	case deviceAuth.IsAuthorized():
		// User has authorized - issue tokens.
		return s.issueDeviceCodeTokens(c, deviceAuth)
	default:
		// Unknown status - should never happen.
		slog.ErrorContext(ctx, "Invalid device authorization status", cryptoutilSharedMagic.StringStatus, deviceAuth.Status)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Invalid device authorization status",
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
		slog.ErrorContext(ctx, "Failed to mark device code as used", cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to update device authorization status",
		})
	}

	// Validate UserID is set (device must be authorized by a user).
	if !deviceAuth.UserID.Valid {
		slog.ErrorContext(ctx, "Device authorization missing user ID", cryptoutilSharedMagic.ParamDeviceCode, deviceAuth.DeviceCode)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Device authorization missing user ID",
		})
	}

	// Retrieve client for token generation.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, deviceAuth.ClientID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve client for device code token", cryptoutilSharedMagic.ClaimClientID, deviceAuth.ClientID, cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to retrieve client",
		})
	}

	// Ensure token service is configured.
	if s.tokenSvc == nil {
		slog.ErrorContext(ctx, "Token service not configured")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Token service not configured",
		})
	}

	// Generate access token with standard OAuth 2.1 claims.
	accessTokenClaims := map[string]any{
		cryptoutilSharedMagic.ClaimSub:      deviceAuth.UserID.UUID.String(),
		cryptoutilSharedMagic.ClaimClientID: deviceAuth.ClientID,
		cryptoutilSharedMagic.ClaimScope:    deviceAuth.Scope,
		cryptoutilSharedMagic.ClaimExp:      time.Now().UTC().Add(time.Duration(client.AccessTokenLifetime) * time.Second).Unix(),
		cryptoutilSharedMagic.ClaimIat:      time.Now().UTC().Unix(),
	}

	accessToken, err := s.tokenSvc.IssueAccessToken(ctx, accessTokenClaims)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

		slog.ErrorContext(ctx, "Access token issuance failed for device code", cryptoutilSharedMagic.StringError, err)

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               appErr.Message,
		})
	}

	// Generate refresh token (optional, based on client configuration).
	var refreshToken string

	if client.RefreshTokenLifetime > 0 {
		refreshToken, err = s.tokenSvc.IssueRefreshToken(ctx)
		if err != nil {
			appErr := cryptoutilIdentityAppErr.ErrTokenIssuanceFailed

			slog.ErrorContext(ctx, "Refresh token issuance failed for device code", cryptoutilSharedMagic.StringError, err)

			return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
				"error_description":               appErr.Message,
			})
		}
	}

	slog.InfoContext(ctx, "Device code token exchange successful",
		cryptoutilSharedMagic.ClaimClientID, deviceAuth.ClientID,
		"user_id", deviceAuth.UserID.UUID.String(),
		cryptoutilSharedMagic.ClaimScope, deviceAuth.Scope,
	)

	// Build token response (RFC 8628 Section 3.5).
	tokenResponse := fiber.Map{
		cryptoutilSharedMagic.TokenTypeAccessToken: accessToken,
		cryptoutilSharedMagic.ParamTokenType:       cryptoutilSharedMagic.AuthorizationBearer,
		cryptoutilSharedMagic.ParamExpiresIn:       client.AccessTokenLifetime,
		cryptoutilSharedMagic.ClaimScope:           deviceAuth.Scope,
	}

	if refreshToken != "" {
		tokenResponse[cryptoutilSharedMagic.GrantTypeRefreshToken] = refreshToken
	}

	return c.Status(fiber.StatusOK).JSON(tokenResponse)
}
