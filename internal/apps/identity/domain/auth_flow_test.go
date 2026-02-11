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

func TestAuthFlow_BeforeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFlow   func() *cryptoutilIdentityDomain.AuthFlow
		validateID  bool
		expectedErr bool
	}{
		{
			name: "generates ID when empty",
			setupFlow: func() *cryptoutilIdentityDomain.AuthFlow {
				return &cryptoutilIdentityDomain.AuthFlow{}
			},
			validateID:  true,
			expectedErr: false,
		},
		{
			name: "preserves existing ID",
			setupFlow: func() *cryptoutilIdentityDomain.AuthFlow {
				existingID := googleUuid.Must(googleUuid.NewV7())

				return &cryptoutilIdentityDomain.AuthFlow{ID: existingID}
			},
			validateID:  false,
			expectedErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flow := tc.setupFlow()
			originalID := flow.ID

			err := flow.BeforeCreate(nil)

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tc.validateID {
					require.NotEqual(t, googleUuid.Nil, flow.ID, "ID should be generated")
				} else {
					require.Equal(t, originalID, flow.ID, "ID should be preserved")
				}
			}
		})
	}
}

func TestAuthFlow_TableName(t *testing.T) {
	t.Parallel()

	flow := cryptoutilIdentityDomain.AuthFlow{}
	require.Equal(t, "auth_flows", flow.TableName())
}
