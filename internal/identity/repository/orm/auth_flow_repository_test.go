// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestAuthFlowRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	flow := &cryptoutilIdentityDomain.AuthFlow{
		Name:                "authorization_code_flow",
		Description:         "Standard authorization code flow with PKCE",
		FlowType:            cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
		RequirePKCE:         true,
		PKCEChallengeMethod: "S256",
		AllowedScopes:       []string{"openid", "profile", "email"},
		RequireConsent:      true,
		ConsentScreenCount:  1,
		RememberConsent:     false,
		RequireState:        true,
		Enabled:             true,
	}

	err := repo.Create(context.Background(), flow)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, flow.ID)

	retrieved, err := repo.GetByID(context.Background(), flow.ID)
	require.NoError(t, err)
	require.Equal(t, flow.Name, retrieved.Name)
	require.Equal(t, flow.FlowType, retrieved.FlowType)
	require.Equal(t, flow.RequirePKCE, retrieved.RequirePKCE)
	require.Len(t, retrieved.AllowedScopes, 3)
}

func TestAuthFlowRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	_, err := repo.GetByID(context.Background(), nonExistentID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthFlowNotFound)
}

func TestAuthFlowRepository_GetByName(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "auth_flow_found",
			setup: func() string {
				flow := &cryptoutilIdentityDomain.AuthFlow{
					Name:        "test_flow",
					FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
					RequirePKCE: true,
					Enabled:     true,
				}
				_ = repo.Create(context.Background(), flow)
				return flow.Name
			},
			wantErr: nil,
		},
		{
			name: "auth_flow_not_found",
			setup: func() string {
				return "nonexistent_flow"
			},
			wantErr: cryptoutilIdentityAppErr.ErrAuthFlowNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flowName := tc.setup()
			_, err := repo.GetByName(context.Background(), flowName)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthFlowRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	flow := &cryptoutilIdentityDomain.AuthFlow{
		Name:        "update_test_flow",
		FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
		RequirePKCE: true,
		Enabled:     true,
	}
	err := repo.Create(context.Background(), flow)
	require.NoError(t, err)

	flow.Description = "Updated description"
	flow.RequireConsent = true
	err = repo.Update(context.Background(), flow)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), flow.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated description", retrieved.Description)
	require.True(t, retrieved.RequireConsent)
}

func TestAuthFlowRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	flow := &cryptoutilIdentityDomain.AuthFlow{
		Name:        "delete_test_flow",
		FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
		RequirePKCE: true,
		Enabled:     true,
	}
	err := repo.Create(context.Background(), flow)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), flow.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), flow.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthFlowNotFound)
}

func TestAuthFlowRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	for i := 0; i < 5; i++ {
		flow := &cryptoutilIdentityDomain.AuthFlow{
			Name:        "list_test_flow_" + string(rune('a'+i)),
			FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
			RequirePKCE: true,
			Enabled:     true,
		}
		err := repo.Create(context.Background(), flow)
		require.NoError(t, err)
	}

	flows, err := repo.List(context.Background(), 0, 3)
	require.NoError(t, err)
	require.Len(t, flows, 3)

	flows, err = repo.List(context.Background(), 3, 3)
	require.NoError(t, err)
	require.Len(t, flows, 2)
}

func TestAuthFlowRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthFlowRepository(testDB.db)

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := 0; i < 5; i++ {
		flow := &cryptoutilIdentityDomain.AuthFlow{
			Name:        "count_test_flow_" + string(rune('a'+i)),
			FlowType:    cryptoutilIdentityDomain.AuthFlowTypeAuthorizationCode,
			RequirePKCE: true,
			Enabled:     true,
		}
		err := repo.Create(context.Background(), flow)
		require.NoError(t, err)
	}

	count, err = repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(5), count)
}
