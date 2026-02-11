// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
)

// TestNewUUIDIssuer validates UUID issuer initialization.
func TestNewUUIDIssuer(t *testing.T) {
	t.Parallel()

	issuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	require.NotNil(t, issuer)
}

// TestIssueToken validates UUID token generation.
func TestIssueToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	issuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	token, err := issuer.IssueToken(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUID (v4 or v7).
	_, err = googleUuid.Parse(token)
	require.NoError(t, err)
}

// TestIssueTokenUniqueness validates that tokens are unique.
func TestIssueTokenUniqueness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	issuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	token1, err := issuer.IssueToken(ctx)
	require.NoError(t, err)

	token2, err := issuer.IssueToken(ctx)
	require.NoError(t, err)

	require.NotEqual(t, token1, token2)
}
