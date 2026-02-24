// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for sm-im service.
// Session handling is delegated to the template package.
package apis

import (
	cryptoutilAppsTemplateServiceServerApis "cryptoutil/internal/apps/template/service/server/apis"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
)

// SessionHandler is an alias to the template's SessionHandler.
type SessionHandler = cryptoutilAppsTemplateServiceServerApis.SessionHandler

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(sessionManager *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService) *SessionHandler {
	return cryptoutilAppsTemplateServiceServerApis.NewSessionHandler(sessionManager)
}

// Re-exported request/response types for session management operations.
type (
	// SessionIssueRequest is the request for creating a new session.
	SessionIssueRequest = cryptoutilAppsTemplateServiceServerApis.SessionIssueRequest
	// SessionIssueResponse is the response containing the issued session token.
	SessionIssueResponse = cryptoutilAppsTemplateServiceServerApis.SessionIssueResponse
	// SessionValidateRequest is the request for validating an existing session.
	SessionValidateRequest = cryptoutilAppsTemplateServiceServerApis.SessionValidateRequest
	// SessionValidateResponse is the response containing session validation results.
	SessionValidateResponse = cryptoutilAppsTemplateServiceServerApis.SessionValidateResponse
)
