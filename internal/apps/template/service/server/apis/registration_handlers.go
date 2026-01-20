// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package apis

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
)

// RegistrationHandlers handles tenant registration endpoints.
type RegistrationHandlers struct {
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService
}

// NewRegistrationHandlers creates registration API handlers.
func NewRegistrationHandlers(
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService,
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

	// TODO: Validate request fields
	// TODO: Hash password
	// TODO: Create user in database
	// TODO: Call registration service

	userID := googleUuid.New()

	tenant, err := h.registrationService.RegisterUserWithTenant(
		c.Context(),
		userID,
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
	// TODO: Extract tenant ID from authenticated user's context
	// TODO: Verify user has admin role
	tenantID := googleUuid.New() // Placeholder

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

	// TODO: Extract admin user ID from authenticated user's context
	adminUserID := googleUuid.New() // Placeholder

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
