// Copyright (c) 2025 Justin Cranford

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// SendEmailOTPRequest represents a request to send an email OTP.
type SendEmailOTPRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
}

// SendEmailOTPResponse represents a response after sending an email OTP.
type SendEmailOTPResponse struct {
	Sent bool `json:"sent"`
}

// VerifyEmailOTPRequest represents a request to verify an email OTP.
type VerifyEmailOTPRequest struct {
	Code string `json:"code" validate:"required"`
}

// VerifyEmailOTPResponse represents a response after verifying an email OTP.
type VerifyEmailOTPResponse struct {
	Verified bool `json:"verified"`
}

// handleSendEmailOTP handles POST /oidc/v1/mfa/email-otp/send.
func (s *Service) handleSendEmailOTP(c *fiber.Ctx) error {
	var req SendEmailOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "invalid request body",
		})
	}

	// Parse user_id.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "invalid user_id format",
		})
	}

	// Verify user exists.
	ctx := c.Context()

	user, err := s.repoFactory.UserRepository().GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":             "user_not_found",
			"error_description": fmt.Sprintf("user not found: %v", err),
		})
	}

	// Send OTP using email OTP service.
	emailOTPService := s.emailOTPService
	if err := emailOTPService.SendOTP(ctx, user.ID, req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             "server_error",
			"error_description": fmt.Sprintf("failed to send OTP: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(SendEmailOTPResponse{Sent: true})
}

// handleVerifyEmailOTP handles POST /oidc/v1/mfa/email-otp/verify.
func (s *Service) handleVerifyEmailOTP(c *fiber.Ctx) error {
	var req VerifyEmailOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "invalid request body",
		})
	}

	// Get user_id from X-User-ID header.
	userIDStr := c.Get("X-User-ID")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "missing X-User-ID header",
		})
	}

	userID, err := googleUuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "invalid_request",
			"error_description": "invalid X-User-ID format",
		})
	}

	// Verify OTP using email OTP service.
	ctx := c.Context()

	emailOTPService := s.emailOTPService
	if err := emailOTPService.VerifyOTP(ctx, userID, req.Code); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "invalid_otp",
			"error_description": fmt.Sprintf("OTP verification failed: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(VerifyEmailOTPResponse{Verified: true})
}
