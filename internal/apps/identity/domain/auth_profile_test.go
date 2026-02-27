// Copyright (c) 2025 Justin Cranford
//
//

package domain_test

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestAuthProfile_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupProf   func() *cryptoutilIdentityDomain.AuthProfile
		validateID  bool
		expectedErr bool
	}{
		{
			name: "generates ID when empty",
			setupProf: func() *cryptoutilIdentityDomain.AuthProfile {
				return &cryptoutilIdentityDomain.AuthProfile{}
			},
			validateID:  true,
			expectedErr: false,
		},
		{
			name: "preserves existing ID",
			setupProf: func() *cryptoutilIdentityDomain.AuthProfile {
				existingID := googleUuid.Must(googleUuid.NewV7())

				return &cryptoutilIdentityDomain.AuthProfile{ID: existingID}
			},
			validateID:  false,
			expectedErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile := tc.setupProf()
			originalID := profile.ID

			err := profile.BeforeCreate(nil)

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tc.validateID {
					require.NotEqual(t, googleUuid.Nil, profile.ID, "ID should be generated")
				} else {
					require.Equal(t, originalID, profile.ID, "ID should be preserved")
				}
			}
		})
	}
}

func TestAuthProfile_TableName(t *testing.T) {
	t.Parallel()

	profile := cryptoutilIdentityDomain.AuthProfile{}
	require.Equal(t, "auth_profiles", profile.TableName())
}
