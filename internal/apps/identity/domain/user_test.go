// Copyright (c) 2025 Justin Cranford

package domain_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestUser_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           *cryptoutilIdentityDomain.User
		expectIDChange bool
	}{
		{
			name: "generates ID when empty",
			user: &cryptoutilIdentityDomain.User{
				Sub: "test_user",
			},
			expectIDChange: true,
		},
		{
			name: "preserves existing ID",
			user: &cryptoutilIdentityDomain.User{
				ID:  googleUuid.Must(googleUuid.NewV7()),
				Sub: "test_user",
			},
			expectIDChange: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			originalID := tc.user.ID

			err := tc.user.BeforeCreate(nil)
			require.NoError(t, err)

			if tc.expectIDChange {
				require.NotEqual(t, googleUuid.Nil, tc.user.ID, "ID should be generated")
			} else {
				require.Equal(t, originalID, tc.user.ID, "ID should be preserved")
			}
		})
	}
}

func TestUser_TableName(t *testing.T) {
	t.Parallel()

	user := cryptoutilIdentityDomain.User{}
	require.Equal(t, "users", user.TableName())
}
