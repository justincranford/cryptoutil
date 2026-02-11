// Copyright (c) 2025 Justin Cranford
//

// Package apis provides HTTP API handlers for jose-ja service.
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

type (
	// SessionIssueRequest represents a session issue request.
	SessionIssueRequest = cryptoutilAppsTemplateServiceServerApis.SessionIssueRequest
	// SessionIssueResponse represents a session issue response.
	SessionIssueResponse = cryptoutilAppsTemplateServiceServerApis.SessionIssueResponse
	// SessionValidateRequest represents a session validate request.
	SessionValidateRequest = cryptoutilAppsTemplateServiceServerApis.SessionValidateRequest
	// SessionValidateResponse represents a session validate response.
	SessionValidateResponse = cryptoutilAppsTemplateServiceServerApis.SessionValidateResponse
)
