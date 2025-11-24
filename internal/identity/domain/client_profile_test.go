package domain_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"cryptoutil/internal/identity/domain"
)

func TestClientProfile_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		clientProfile  *domain.ClientProfile
		expectIDChange bool
	}{
		{
			name: "generates ID when empty",
			clientProfile: &domain.ClientProfile{
				Name: "Test Profile",
			},
			expectIDChange: true,
		},
		{
			name: "preserves existing ID",
			clientProfile: &domain.ClientProfile{
				ID:   googleUuid.Must(googleUuid.NewV7()),
				Name: "Test Profile",
			},
			expectIDChange: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			originalID := tc.clientProfile.ID

			err := tc.clientProfile.BeforeCreate(nil)
			require.NoError(t, err)

			if tc.expectIDChange {
				require.NotEqual(t, googleUuid.Nil, tc.clientProfile.ID, "ID should be generated")
			} else {
				require.Equal(t, originalID, tc.clientProfile.ID, "ID should be preserved")
			}
		})
	}
}

func TestClientProfile_TableName(t *testing.T) {
	t.Parallel()

	clientProfile := domain.ClientProfile{}
	require.Equal(t, "client_profiles", clientProfile.TableName())
}
