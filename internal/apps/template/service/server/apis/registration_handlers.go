// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"fmt"
	http "net/http"
	"regexp"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// emailRegex is a simple email validation pattern.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// registrationHandlersHashSecretPBKDF2Fn allows overriding hash function for testing.
var registrationHandlersHashSecretPBKDF2Fn = cryptoutilSharedCryptoHash.HashSecretPBKDF2

// RegistrationHandlers handles tenant registration endpoints.
type RegistrationHandlers struct {
	registrationService *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService
}

// NewRegistrationHandlers creates registration API handlers.
func NewRegistrationHandlers(
	registrationService *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService,
) *RegistrationHandlers {
	return &RegistrationHandlers{
		registrationService: registrationService,
	}
}

// RegisterUserRequest is the request body for user registration.
type RegisterUserRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=50"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	TenantName   string `json:"tenant_name" validate:"required,min=3,max=100"`
	CreateTenant bool   `json:"create_tenant"`
}

// RegisterUserResponse is the response for user registration.
type RegisterUserResponse struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id,omitempty"`
	Message  string `json:"message"`
}

// HandleRegisterUser handles POST /browser/api/v1/auth/register.
func (h *RegistrationHandlers) HandleRegisterUser(c *fiber.Ctx) error {
	var req RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request fields.
	if err := validateRegistrationRequest(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Hash password using FIPS-approved PBKDF2-HMAC-SHA256.
	passwordHash, err := registrationHandlersHashSecretPBKDF2Fn(req.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process password",
		})
	}

	// Generate UUIDv7 for new user (time-ordered).
	userID := googleUuid.Must(googleUuid.NewV7())

	// Call registration service with validated and hashed data.
	tenant, err := h.registrationService.RegisterUserWithTenant(
		c.Context(),
		userID,
		req.Username,
		req.Email,
		passwordHash,
		req.TenantName,
		req.CreateTenant,
	)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Registration failed",
		})
	}

	response := RegisterUserResponse{
		UserID:  userID.String(),
		Message: "User registered successfully",
	}

	if tenant != nil {
		response.TenantID = tenant.ID.String()
		response.Message = "User registered and tenant created"
	}

	return c.Status(http.StatusCreated).JSON(response)
}

// validateRegistrationRequest validates the registration request fields.
func validateRegistrationRequest(req *RegisterUserRequest) error {
	// Trim whitespace.
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.TenantName = strings.TrimSpace(req.TenantName)

	// Validate username length.
	if len(req.Username) < cryptoutilSharedMagic.CipherMinUsernameLength {
		return fmt.Errorf("username must be at least %d characters", cryptoutilSharedMagic.CipherMinUsernameLength)
	}

	if len(req.Username) > cryptoutilSharedMagic.CipherMaxUsernameLength {
		return fmt.Errorf("username must be at most %d characters", cryptoutilSharedMagic.CipherMaxUsernameLength)
	}

	// Validate email format.
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password length.
	if len(req.Password) < cryptoutilSharedMagic.CipherMinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", cryptoutilSharedMagic.CipherMinPasswordLength)
	}

	// Validate tenant name length.
	if len(req.TenantName) < cryptoutilSharedMagic.CipherMinUsernameLength {
		return fmt.Errorf("tenant name must be at least %d characters", cryptoutilSharedMagic.CipherMinUsernameLength)
	}

	if len(req.TenantName) > cryptoutilSharedMagic.CipherMaxTenantNameLength {
		return fmt.Errorf("tenant name must be at most %d characters", cryptoutilSharedMagic.CipherMaxTenantNameLength)
	}

	return nil
}

// JoinRequestSummary is a summary of a join request.
type JoinRequestSummary struct {
	ID          string  `json:"id"`
	UserID      *string `json:"user_id,omitempty"`
	ClientID    *string `json:"client_id,omitempty"`
	TenantID    string  `json:"tenant_id"`
	Status      string  `json:"status"`
	RequestedAt string  `json:"requested_at"`
	ProcessedAt *string `json:"processed_at,omitempty"`
	ProcessedBy *string `json:"processed_by,omitempty"`
}

// HandleListJoinRequests handles GET /browser/api/v1/admin/join-requests.
func (h *RegistrationHandlers) HandleListJoinRequests(c *fiber.Ctx) error {
	// Extract tenant ID from authenticated user's context (set by middleware)
	tenantIDVal := c.Locals("tenant_id")
	if tenantIDVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing tenant_id in session context",
		})
	}

	tenantID, ok := tenantIDVal.(googleUuid.UUID)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid tenant_id type in context",
		})
	}

	// TODO: Verify user has admin role

	requests, err := h.registrationService.ListJoinRequests(c.Context(), tenantID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list join requests",
		})
	}

	summaries := make([]JoinRequestSummary, len(requests))

	for i, req := range requests {
		summary := JoinRequestSummary{
			ID:          req.ID.String(),
			TenantID:    req.TenantID.String(),
			Status:      req.Status,
			RequestedAt: req.RequestedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if req.UserID != nil {
			userIDStr := req.UserID.String()
			summary.UserID = &userIDStr
		}

		if req.ClientID != nil {
			clientIDStr := req.ClientID.String()
			summary.ClientID = &clientIDStr
		}

		if req.ProcessedAt != nil {
			processedAtStr := req.ProcessedAt.Format("2006-01-02T15:04:05Z07:00")
			summary.ProcessedAt = &processedAtStr
		}

		if req.ProcessedBy != nil {
			processedByStr := req.ProcessedBy.String()
			summary.ProcessedBy = &processedByStr
		}

		summaries[i] = summary
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"requests": summaries,
	})
}

// ProcessJoinRequestRequest is the request body for processing join requests (approve/reject).
type ProcessJoinRequestRequest struct {
	Approved bool `json:"approved"`
}

// HandleProcessJoinRequest handles PUT /admin/api/v1/join-requests/:id.
// Processes a join request by approving or rejecting it.
func (h *RegistrationHandlers) HandleProcessJoinRequest(c *fiber.Ctx) error {
	requestID, err := googleUuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request ID",
		})
	}

	var req ProcessJoinRequestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Extract admin user ID from authenticated user's context (set by middleware)
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing user_id in session context",
		})
	}

	adminUserID, ok := userIDVal.(googleUuid.UUID)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user_id type in context",
		})
	}

	err = h.registrationService.AuthorizeJoinRequest(
		c.Context(),
		requestID,
		adminUserID,
		req.Approved,
	)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	message := "Join request rejected"
	if req.Approved {
		message = "Join request approved"
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": message,
	})
}
