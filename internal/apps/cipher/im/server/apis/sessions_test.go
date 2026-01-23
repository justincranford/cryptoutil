// Copyright (c) 2025 Justin Cranford

package apis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewSessionHandler tests the NewSessionHandler constructor.
func TestNewSessionHandler(t *testing.T) {
	// NewSessionHandler should not panic with nil - it just passes through to template handler.
	sessionHandler := NewSessionHandler(nil)
	require.NotNil(t, sessionHandler)
}

// TestSessionHandlerTypeAlias verifies that the type alias works correctly.
func TestSessionHandlerTypeAlias(t *testing.T) {
	// Verify that SessionHandler type alias resolves correctly.
	var handler *SessionHandler
	handler = nil
	require.Nil(t, handler)
}

// TestSessionRequestResponseTypes verifies the re-exported types work.
func TestSessionRequestResponseTypes(t *testing.T) {
	// Test that the type aliases compile and work.
	var issueReq SessionIssueRequest
	issueReq.UserID = "testuser"
	issueReq.TenantID = "00000000-0000-0000-0000-000000000001"
	issueReq.RealmID = "00000000-0000-0000-0000-000000000002"
	issueReq.SessionType = "browser"
	require.Equal(t, "testuser", issueReq.UserID)

	var issueResp SessionIssueResponse
	issueResp.Token = "test-token"
	require.Equal(t, "test-token", issueResp.Token)

	var validateReq SessionValidateRequest
	validateReq.Token = "test-token"
	validateReq.SessionType = "browser"
	require.Equal(t, "test-token", validateReq.Token)

	var validateResp SessionValidateResponse
	validateResp.Valid = true
	require.True(t, validateResp.Valid)
}


