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

const (
	nonexistentProfile            = "nonexistent_profile"
	updatedDescriptionAuthProfile = "Updated description"
)

func TestAuthProfileRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	profile := &cryptoutilIdentityDomain.AuthProfile{
		Name:        "username_password_profile",
		Description: "Standard username/password authentication",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		RequireMFA:  true,
		MFAChain:    []string{cryptoutilSharedMagic.MFATypeTOTP, cryptoutilSharedMagic.AMRSMS},
		Enabled:     true,
	}

	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, profile.ID)

	retrieved, err := repo.GetByID(context.Background(), profile.ID)
	require.NoError(t, err)
	require.Equal(t, profile.Name, retrieved.Name)
	require.Equal(t, profile.ProfileType, retrieved.ProfileType)
	require.True(t, retrieved.RequireMFA)
	require.Len(t, retrieved.MFAChain, 2)
}

func TestAuthProfileRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	_, err := repo.GetByID(context.Background(), nonExistentID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthProfileNotFound)
}

func TestAuthProfileRepository_GetByName(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "auth_profile_found",
			setup: func() string {
				profile := &cryptoutilIdentityDomain.AuthProfile{
					Name:        "test_profile",
					ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
					RequireMFA:  false,
					Enabled:     true,
				}
				err := repo.Create(context.Background(), profile)
				require.NoError(t, err)

				return profile.Name
			},
			wantErr: nil,
		},
		{
			name: "auth_profile_not_found",
			setup: func() string {
				return nonexistentProfile
			},
			wantErr: cryptoutilIdentityAppErr.ErrAuthProfileNotFound,
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

func TestAuthProfileRepository_Update(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	profile := &cryptoutilIdentityDomain.AuthProfile{
		Name:        "update_test_profile",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		RequireMFA:  false,
		Enabled:     true,
	}
	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)

	profile.Description = updatedDescriptionAuthProfile
	profile.RequireMFA = true
	profile.MFAChain = []string{cryptoutilSharedMagic.MFATypeTOTP}
	err = repo.Update(context.Background(), profile)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), profile.ID)
	require.NoError(t, err)
	require.Equal(t, updatedDescriptionAuthProfile, retrieved.Description)
	require.True(t, retrieved.RequireMFA)
	require.Len(t, retrieved.MFAChain, 1)
}

func TestAuthProfileRepository_Delete(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	profile := &cryptoutilIdentityDomain.AuthProfile{
		Name:        "delete_test_profile",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		RequireMFA:  false,
		Enabled:     true,
	}
	err := repo.Create(context.Background(), profile)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), profile.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), profile.ID)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrAuthProfileNotFound)
}

func TestAuthProfileRepository_List(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		profile := &cryptoutilIdentityDomain.AuthProfile{
			Name:        "list_test_profile_" + string(rune('a'+i)),
			ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
			RequireMFA:  false,
			Enabled:     true,
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

func TestAuthProfileRepository_Count(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	repo := NewAuthProfileRepository(testDB.db)

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		profile := &cryptoutilIdentityDomain.AuthProfile{
			Name:        "count_test_profile_" + string(rune('a'+i)),
			ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
			RequireMFA:  false,
			Enabled:     true,
		}
		err := repo.Create(context.Background(), profile)
		require.NoError(t, err)
	}

	count, err = repo.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), count)
}
