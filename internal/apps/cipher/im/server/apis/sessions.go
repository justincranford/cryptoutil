// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for cipher-im service.
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

// Re-exported request/response types for session management operations.
type (
	// SessionIssueRequest is the request for creating a new session.
	SessionIssueRequest = cryptoutilTemplateAPIs.SessionIssueRequest
	// SessionIssueResponse is the response containing the issued session token.
	SessionIssueResponse = cryptoutilTemplateAPIs.SessionIssueResponse
	// SessionValidateRequest is the request for validating an existing session.
	SessionValidateRequest = cryptoutilTemplateAPIs.SessionValidateRequest
	// SessionValidateResponse is the response containing session validation results.
	SessionValidateResponse = cryptoutilTemplateAPIs.SessionValidateResponse
)
