// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
)

func TestUUIDIssuer_ValidateToken(t *testing.T) {
	t.Parallel()

	issuer := NewUUIDIssuer()
	ctx := context.Background()

	// Generate test UUIDs dynamically to avoid hardcoded values.
	validLowercaseUUID := googleUuid.Must(googleUuid.NewV7()).String()
	validUppercaseUUID := googleUuid.Must(googleUuid.NewV7()).String()

	tests := []struct {
		name        string
		token       string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid v4 UUID",
			token:       googleUuid.New().String(),
			expectError: false,
		},
		{
			name:        "valid v7 UUID",
			token:       googleUuid.Must(googleUuid.NewV7()).String(),
			expectError: false,
		},
		{
			name:        "valid lowercase UUID",
			token:       validLowercaseUUID,
			expectError: false,
		},
		{
			name:        "valid uppercase UUID",
			token:       validUppercaseUUID,
			expectError: false,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			errorType:   cryptoutilIdentityAppErr.ErrInvalidToken,
		},
		{
			name:        "invalid format - too short",
			token:       "550e8400-e29b-41d4",
			expectError: true,
			errorType:   cryptoutilIdentityAppErr.ErrInvalidToken,
		},
		{
			name:        "invalid format - no dashes",
			token:       "550e8400e29b41d4a716446655440000",
			expectError: false, // UUIDs without dashes are valid.
		},
		{
			name:        "invalid format - wrong characters",
			token:       "gggggggg-gggg-gggg-gggg-gggggggggggg",
			expectError: true,
			errorType:   cryptoutilIdentityAppErr.ErrInvalidToken,
		},
		{
			name:        "invalid format - random string",
			token:       "not-a-valid-uuid-at-all",
			expectError: true,
			errorType:   cryptoutilIdentityAppErr.ErrInvalidToken,
		},
		{
			name:        "nil UUID string",
			token:       googleUuid.Nil.String(),
			expectError: false, // Nil UUID is valid format.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := issuer.ValidateToken(ctx, tc.token)

			if tc.expectError {
				testify.Error(t, err)
				testify.ErrorIs(t, err, tc.errorType)
			} else {
				testify.NoError(t, err)
			}
		})
	}
}

func TestUUIDIssuer_IssueAndValidate(t *testing.T) {
	t.Parallel()

	issuer := NewUUIDIssuer()
	ctx := context.Background()

	// Issue a token.
	token, err := issuer.IssueToken(ctx)
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	// Validate the issued token.
	err = issuer.ValidateToken(ctx, token)
	testify.NoError(t, err)
}
