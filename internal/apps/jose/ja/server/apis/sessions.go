// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
// Session handling is delegated to the template package.
package apis

import (
	cryptoutilTemplateAPIs "cryptoutil/internal/apps/template/service/server/apis"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
)

// SessionHandler is an alias to the template's SessionHandler.
type SessionHandler = cryptoutilTemplateAPIs.SessionHandler

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(sessionManager *cryptoutilTemplateBusinessLogic.SessionManagerService) *SessionHandler {
	return cryptoutilTemplateAPIs.NewSessionHandler(sessionManager)
}

type (
	// SessionIssueRequest represents a session issue request.
	SessionIssueRequest = cryptoutilTemplateAPIs.SessionIssueRequest
	// SessionIssueResponse represents a session issue response.
	SessionIssueResponse = cryptoutilTemplateAPIs.SessionIssueResponse
	// SessionValidateRequest represents a session validate request.
	SessionValidateRequest = cryptoutilTemplateAPIs.SessionValidateRequest
	// SessionValidateResponse represents a session validate response.
	SessionValidateResponse = cryptoutilTemplateAPIs.SessionValidateResponse
)
