// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for sm-im service.
// Session handling is delegated to the template package.
package apis

import (
	cryptoutilAppsFrameworkServiceServerApis "cryptoutil/internal/apps/framework/service/server/apis"
	cryptoutilAppsFrameworkServiceServerBusinesslogic "cryptoutil/internal/apps/framework/service/server/businesslogic"
)

// SessionHandler is an alias to the template's SessionHandler.
type SessionHandler = cryptoutilAppsFrameworkServiceServerApis.SessionHandler

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(sessionManager *cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService) *SessionHandler {
	return cryptoutilAppsFrameworkServiceServerApis.NewSessionHandler(sessionManager)
}

// Re-exported request/response types for session management operations.
type (
	// SessionIssueRequest is the request for creating a new session.
	SessionIssueRequest = cryptoutilAppsFrameworkServiceServerApis.SessionIssueRequest
	// SessionIssueResponse is the response containing the issued session token.
	SessionIssueResponse = cryptoutilAppsFrameworkServiceServerApis.SessionIssueResponse
	// SessionValidateRequest is the request for validating an existing session.
	SessionValidateRequest = cryptoutilAppsFrameworkServiceServerApis.SessionValidateRequest
	// SessionValidateResponse is the response containing session validation results.
	SessionValidateResponse = cryptoutilAppsFrameworkServiceServerApis.SessionValidateResponse
)
