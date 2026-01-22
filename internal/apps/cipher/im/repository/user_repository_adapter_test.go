// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestUserRepositoryAdapter_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := NewUserRepository(testDB)
	adapter := NewUserRepositoryAdapter(userRepo)

	tests := []struct {
		name    string
		user    *cryptoutilRepository.User
		wantErr bool
	}{
		{
			name: "valid user creation",
			user: &cryptoutilRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "adapter-test-" + testJWKGenService.GenerateUUIDv7().String(),
			},
			wantErr: false,
		},
		{
			name: "user with minimal fields",
			user: &cryptoutilRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "minimal-" + testJWKGenService.GenerateUUIDv7().String(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := adapter.Create(ctx, tt.user)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify user was created via adapter.
			found, err := adapter.FindByID(ctx, tt.user.ID)
			require.NoError(t, err)
			require.Equal(t, tt.user.ID, found.GetID())
			require.Equal(t, tt.user.Username, found.GetUsername())

			// Cleanup.
			require.NoError(t, userRepo.Delete(ctx, tt.user.ID))
		})
	}
}

func TestUserRepositoryAdapter_FindByUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := NewUserRepository(testDB)
	adapter := NewUserRepositoryAdapter(userRepo)

	tests := []struct {
		name        string
		setupUser   bool
		username    string
		wantErr     bool
		expectFound bool
	}{
		{
			name:        "existing user",
			setupUser:   true,
			username:    "", // Will be set to created user's username.
			wantErr:     false,
			expectFound: true,
		},
		{
			name:        "non-existent user",
			setupUser:   false,
			username:    "nonexistent-" + testJWKGenService.GenerateUUIDv7().String(),
			wantErr:     true,
			expectFound: false,
		},
		{
			name:        "empty username",
			setupUser:   false,
			username:    "",
			wantErr:     true,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var user *cryptoutilRepository.User

			if tt.setupUser {
				// Create test user in subtest context.
				user = &cryptoutilRepository.User{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					Username: "find-username-" + testJWKGenService.GenerateUUIDv7().String(),
				}
				require.NoError(t, userRepo.Create(ctx, user))

				defer func() { _ = userRepo.Delete(ctx, user.ID) }()

				// Use actual username for this test.
				tt.username = user.Username
			}

			found, err := adapter.FindByUsername(ctx, tt.username)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, found)

				return
			}

			require.NoError(t, err)

			if tt.expectFound {
				require.NotNil(t, found)
				require.Equal(t, user.ID, found.GetID())
				require.Equal(t, user.Username, found.GetUsername())
			}
		})
	}
}

func TestUserRepositoryAdapter_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userRepo := NewUserRepository(testDB)
	adapter := NewUserRepositoryAdapter(userRepo)

	tests := []struct {
		name        string
		setupUser   bool
		userID      googleUuid.UUID
		wantErr     bool
		expectFound bool
	}{
		{
			name:        "existing user by ID",
			setupUser:   true,
			userID:      googleUuid.UUID{}, // Will be set to created user's ID.
			wantErr:     false,
			expectFound: true,
		},
		{
			name:        "non-existent user ID",
			setupUser:   false,
			userID:      *testJWKGenService.GenerateUUIDv7(),
			wantErr:     true,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var user *cryptoutilRepository.User

			if tt.setupUser {
				// Create test user in subtest context.
				user = &cryptoutilRepository.User{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					Username: "find-id-" + testJWKGenService.GenerateUUIDv7().String(),
				}
				require.NoError(t, userRepo.Create(ctx, user))

				defer func() { _ = userRepo.Delete(ctx, user.ID) }()

				// Use actual user ID for this test.
				tt.userID = user.ID
			}

			found, err := adapter.FindByID(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, found)

				return
			}

			require.NoError(t, err)

			if tt.expectFound {
				require.NotNil(t, found)
				require.Equal(t, user.ID, found.GetID())
				require.Equal(t, user.Username, found.GetUsername())
			}
		})
	}
}

func TestUserRepositoryAdapter_AdapterConformance(t *testing.T) {
	t.Parallel()

	userRepo := NewUserRepository(testDB)
	adapter := NewUserRepositoryAdapter(userRepo)

	// Verify adapter implements interface.
	require.NotNil(t, adapter)
	require.NotNil(t, adapter.repo)
}
