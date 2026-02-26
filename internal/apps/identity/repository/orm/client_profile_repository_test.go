// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const updatedDescriptionClientProfile = "Updated description"

func TestClientProfileRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	profile := &cryptoutilIdentityDomain.ClientProfile{
		Name:               "standard_profile",
		Description:        "Standard OAuth client profile",
		RequiredScopes:     []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
		OptionalScopes:     []string{cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ScopePhone},
		ConsentScreenCount: 1,
		ConsentScreen1Text: "Grant access to your profile?",
		RequireClientMFA:   false,
		Enabled:            true,
	}

	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, profile.ID)

	retrieved, err := repo.GetByID(context.Background(), profile.ID)
	require.NoError(t, err)
	require.Equal(t, profile.Name, retrieved.Name)
	require.Len(t, retrieved.RequiredScopes, 2)
	require.Len(t, retrieved.OptionalScopes, 2)
	require.Equal(t, 1, retrieved.ConsentScreenCount)
}

func TestClientProfileRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	_, err := repo.GetByID(context.Background(), nonExistentID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrClientProfileNotFound)
}

func TestClientProfileRepository_GetByName(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "client_profile_found",
			setup: func() string {
				profile := &cryptoutilIdentityDomain.ClientProfile{
					Name:           "test_profile",
					RequiredScopes: []string{cryptoutilSharedMagic.ScopeOpenID},
					Enabled:        true,
				}
				err := repo.Create(context.Background(), profile)
				require.NoError(t, err)

				return profile.Name
			},
			wantErr: nil,
		},
		{
			name: "client_profile_not_found",
			setup: func() string {
				return "nonexistent_profile"
			},
			wantErr: cryptoutilIdentityAppErr.ErrClientProfileNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profileName := tc.setup()
			_, err := repo.GetByName(context.Background(), profileName)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientProfileRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	profile := &cryptoutilIdentityDomain.ClientProfile{
		Name:           "update_test_profile",
		RequiredScopes: []string{cryptoutilSharedMagic.ScopeOpenID},
		Enabled:        true,
	}
	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)

	profile.Description = updatedDescriptionClientProfile
	profile.RequireClientMFA = true
	profile.ClientMFAChain = []string{"mtls"}
	err = repo.Update(context.Background(), profile)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), profile.ID)
	require.NoError(t, err)
	require.Equal(t, updatedDescriptionClientProfile, retrieved.Description)
	require.True(t, retrieved.RequireClientMFA)
	require.Len(t, retrieved.ClientMFAChain, 1)
}

func TestClientProfileRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	profile := &cryptoutilIdentityDomain.ClientProfile{
		Name:           "delete_test_profile",
		RequiredScopes: []string{cryptoutilSharedMagic.ScopeOpenID},
		Enabled:        true,
	}
	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), profile.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), profile.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrClientProfileNotFound)
}

func TestClientProfileRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		profile := &cryptoutilIdentityDomain.ClientProfile{
			Name:           "list_test_profile_" + string(rune('a'+i)),
			RequiredScopes: []string{cryptoutilSharedMagic.ScopeOpenID},
			Enabled:        true,
		}
		err := repo.Create(context.Background(), profile)
		require.NoError(t, err)
	}

	profiles, err := repo.List(context.Background(), 0, 3)
	require.NoError(t, err)
	require.Len(t, profiles, 3)

	profiles, err = repo.List(context.Background(), 3, 3)
	require.NoError(t, err)
	require.Len(t, profiles, 2)
}

func TestClientProfileRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewClientProfileRepository(testDB.db)

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		profile := &cryptoutilIdentityDomain.ClientProfile{
			Name:           "count_test_profile_" + string(rune('a'+i)),
			RequiredScopes: []string{cryptoutilSharedMagic.ScopeOpenID},
			Enabled:        true,
		}
		err := repo.Create(context.Background(), profile)
		require.NoError(t, err)
	}

	count, err = repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), count)
}
