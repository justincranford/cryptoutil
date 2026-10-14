// Copyright (c) 2025-2026 Justin Cranford.
package domain_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestClientProfile_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		clientProfile  *cryptoutilIdentityDomain.ClientProfile
		expectIDChange bool
	}{
		{
			name: testGeneratesIDWhenEmpty,
			clientProfile: &cryptoutilIdentityDomain.ClientProfile{
				Name: testTestProfile,
			},
			expectIDChange: true,
		},
		{
			name: testPreservesExistingID,
			clientProfile: &cryptoutilIdentityDomain.ClientProfile{
				ID:   googleUuid.Must(googleUuid.NewV7()),
				Name: testTestProfile,
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

	clientProfile := cryptoutilIdentityDomain.ClientProfile{}
	require.Equal(t, "client_profiles", clientProfile.TableName())
}
