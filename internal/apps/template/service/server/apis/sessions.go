// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

// Constants for repeated strings.
const (
	errInvalidRequestBody = "Invalid request body format"
	sessionTypeBrowser    = "browser"
)

// SessionManager defines the interface for session management operations.
// This interface enables testing by allowing mock implementations.
type SessionManager interface {
	IssueBrowserSessionWithTenant(ctx context.Context, userID string, tenantID, realmID googleUuid.UUID) (string, error)
	IssueServiceSessionWithTenant(ctx context.Context, clientID string, tenantID, realmID googleUuid.UUID) (string, error)
	ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error)
	ValidateServiceSession(ctx context.Context, token string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error)
}

// SessionHandler handles session management endpoints.
type SessionHandler struct {
	sessionManager SessionManager
}

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(sessionManager SessionManager) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
	}
}

// Ensure *SessionManagerService implements SessionManager at compile time.
var _ SessionManager = (*cryptoutilTemplateBusinessLogic.SessionManagerService)(nil)

// SessionIssueRequest represents the request body for issuing a session.
type SessionIssueRequest struct {
	UserID      string `json:"user_id" validate:"required,min=1,max=255"`
	TenantID    string `json:"tenant_id" validate:"required,uuid"`
	RealmID     string `json:"realm_id" validate:"required,uuid"`
	SessionType string `json:"session_type" validate:"required,oneof=browser service"`
}

// SessionIssueResponse represents the response body for session issuance.
type SessionIssueResponse struct {
	Token string `json:"token"`
}

// SessionValidateRequest represents the request body for session validation.
type SessionValidateRequest struct {
	Token       string `json:"token" validate:"required,min=1"`
	SessionType string `json:"session_type" validate:"required,oneof=browser service"`
}

// SessionValidateResponse represents the response body for session validation.
type SessionValidateResponse struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RealmID  string `json:"realm_id"`
	Valid    bool   `json:"valid"`
}

// IssueSession creates a new session token.
func (h *SessionHandler) IssueSession(c *fiber.Ctx) error {
	var req SessionIssueRequest
	if err := c.BodyParser(&req); err != nil {
		summary := errInvalidRequestBody

		return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, err)
	}

	// Parse tenant and realm IDs.
	tenantID, err := googleUuid.Parse(req.TenantID)
	if err != nil {
		summary := "Invalid tenant_id format"

		return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, err)
	}

	realmID, err := googleUuid.Parse(req.RealmID)
	if err != nil {
		summary := "Invalid realm_id format"

		return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, err)
	}

	// Create context from request context.
	ctx := context.Background()

	// Issue session based on type.
	var token string
	if req.SessionType == sessionTypeBrowser {
		token, err = h.sessionManager.IssueBrowserSessionWithTenant(ctx, req.UserID, tenantID, realmID)
	} else {
		token, err = h.sessionManager.IssueServiceSessionWithTenant(ctx, req.UserID, tenantID, realmID)
	}

	if err != nil {
		summary := "Failed to issue session token"

		return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, err)
	}

	// Format response.
	resp := SessionIssueResponse{
		Token: token,
	}

	if err := c.JSON(resp); err != nil {
		return fmt.Errorf("failed to encode JSON response: %w", err)
	}

	return nil
}

// ValidateSession validates an existing session token.
func (h *SessionHandler) ValidateSession(c *fiber.Ctx) error {
	var req SessionValidateRequest
	if err := c.BodyParser(&req); err != nil {
		summary := errInvalidRequestBody

		return cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, err)
	}

	// Create context from request context.
	ctx := context.Background()

	// Validate session based on type.
	var (
		userID, tenantID, realmID string
		valid                     bool
	)

	if req.SessionType == sessionTypeBrowser {
		browserSession, err := h.sessionManager.ValidateBrowserSession(ctx, req.Token)
		if err != nil {
			valid = false
		} else {
			if browserSession.UserID != nil {
				userID = *browserSession.UserID
			}

			tenantID = browserSession.TenantID.String()
			realmID = browserSession.RealmID.String()
			valid = true
		}
	} else {
		serviceSession, err := h.sessionManager.ValidateServiceSession(ctx, req.Token)
		if err != nil {
			valid = false
		} else {
			if serviceSession.ClientID != nil {
				userID = *serviceSession.ClientID
			}

			tenantID = serviceSession.TenantID.String()
			realmID = serviceSession.RealmID.String()
			valid = true
		}
	}

	// Format response.
	resp := SessionValidateResponse{
		UserID:   userID,
		TenantID: tenantID,
		RealmID:  realmID,
		Valid:    valid,
	}

	if err := c.JSON(resp); err != nil {
		return fmt.Errorf("failed to encode JSON response: %w", err)
	}

	return nil
}
