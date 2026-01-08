// Copyright (c) 2025 Justin Cranford
//

package apis

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/apps/cipher/im/server/businesslogic"
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
)

// Session type constants.
const (
	SessionTypeBrowser = "browser"
	SessionTypeService = "service"
)

// Error message constants.
const (
	errMsgInvalidRequestBody = "Invalid request body format"
)

// SessionHandler handles session management endpoints for cipher-im.
type SessionHandler struct {
	sessionManager *businesslogic.SessionManagerService
}

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(sessionManager *businesslogic.SessionManagerService) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
	}
}

// SessionIssueRequest represents the request body for issuing a session.
type SessionIssueRequest struct {
	UserID      string `json:"user_id" validate:"required,min=1,max=255"`
	Realm       string `json:"realm" validate:"required,min=1,max=255"`
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
	UserID string `json:"user_id"`
	Realm  string `json:"realm"`
	Valid  bool   `json:"valid"`
}

// IssueSession creates a new session token.
func (h *SessionHandler) IssueSession(c *fiber.Ctx) error {
	var req SessionIssueRequest

	if err := c.BodyParser(&req); err != nil {
		summary := errMsgInvalidRequestBody

		return cryptoutilAppErr.NewHTTP400BadRequest(&summary, err)
	}

	// Create context from request context.
	ctx := context.Background()

	// Issue session based on type.
	var token string

	var err error

	if req.SessionType == SessionTypeBrowser {
		token, err = h.sessionManager.IssueBrowserSession(ctx, req.UserID, req.Realm)
	} else {
		token, err = h.sessionManager.IssueServiceSession(ctx, req.UserID, req.Realm)
	}

	if err != nil {
		summary := "Failed to issue session token"

		return cryptoutilAppErr.NewHTTP500InternalServerError(&summary, err)
	}

	// Format response.
	resp := SessionIssueResponse{
		Token: token,
	}

	if jsonErr := c.JSON(resp); jsonErr != nil {
		return fmt.Errorf("failed to write JSON response: %w", jsonErr)
	}

	return nil
}

// ValidateSession validates an existing session token.
func (h *SessionHandler) ValidateSession(c *fiber.Ctx) error {
	var req SessionValidateRequest

	if err := c.BodyParser(&req); err != nil {
		summary := errMsgInvalidRequestBody

		return cryptoutilAppErr.NewHTTP400BadRequest(&summary, err)
	}

	// Create context from request context.
	ctx := context.Background()

	// Validate session based on type.
	var userID, realm string

	var valid bool

	if req.SessionType == SessionTypeBrowser {
		browserSession, err := h.sessionManager.ValidateBrowserSession(ctx, req.Token)
		if err != nil {
			valid = false
		} else {
			if browserSession.UserID != nil {
				userID = *browserSession.UserID
			}

			if browserSession.Realm != nil {
				realm = *browserSession.Realm
			}

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

			if serviceSession.Realm != nil {
				realm = *serviceSession.Realm
			}

			valid = true
		}
	}

	// Format response.
	resp := SessionValidateResponse{
		UserID: userID,
		Realm:  realm,
		Valid:  valid,
	}

	if jsonErr := c.JSON(resp); jsonErr != nil {
		return fmt.Errorf("failed to write JSON response: %w", jsonErr)
	}

	return nil
}