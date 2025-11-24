package domain_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/domain"
)

func TestUser_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           *domain.User
		expectIDChange bool
	}{
		{
			name: "generates ID when empty",
			user: &domain.User{
				Sub: "test_user",
			},
			expectIDChange: true,
		},
		{
			name: "preserves existing ID",
			user: &domain.User{
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

	user := domain.User{}
	require.Equal(t, "users", user.TableName())
}
