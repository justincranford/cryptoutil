// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
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

type (
	// SessionIssueRequest represents a session issue request.
	SessionIssueRequest = cryptoutilAppsFrameworkServiceServerApis.SessionIssueRequest
	// SessionIssueResponse represents a session issue response.
	SessionIssueResponse = cryptoutilAppsFrameworkServiceServerApis.SessionIssueResponse
	// SessionValidateRequest represents a session validate request.
	SessionValidateRequest = cryptoutilAppsFrameworkServiceServerApis.SessionValidateRequest
	// SessionValidateResponse represents a session validate response.
	SessionValidateResponse = cryptoutilAppsFrameworkServiceServerApis.SessionValidateResponse
)
