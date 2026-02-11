// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
)

// TestMIMEApplicationJWT validates the MIME type constant is defined correctly.
func TestMIMEApplicationJWT(t *testing.T) {
	t.Parallel()

	require.Equal(t, "application/jwt", cryptoutilIdentityIdp.MIMEApplicationJWT)
}

// TestAddScopeBasedClaimsHelper validates the helper function signature exists.
// The actual functionality is tested via integration tests in handlers_oidc_e2e_test.go.
func TestUserInfoJWTResponseConstants(t *testing.T) {
	t.Parallel()

	// Validate Accept header contains check works.
	acceptHeader := "application/jwt"
	require.True(t, strings.Contains(acceptHeader, cryptoutilIdentityIdp.MIMEApplicationJWT))

	acceptHeader = "application/json"
	require.False(t, strings.Contains(acceptHeader, cryptoutilIdentityIdp.MIMEApplicationJWT))
}
