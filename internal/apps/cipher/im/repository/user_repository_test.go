// Copyright (c) 2025 Justin Cranford
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewUserRepository(testDB)

	tests := []struct {
		name    string
		user    *cryptoutilAppsTemplateServiceServerRepository.User
		wantErr bool
	}{
		{
			name: "valid user creation",
			user: &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-user-" + testJWKGenService.GenerateUUIDv7().String(),
			},
			wantErr: false,
		},
		{
			name: "user with empty username",
			user: &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "",
			},
			wantErr: false, // Repository doesn't validate username (validation in handler)
		},
		{
			name: "user with long username",
			user: &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: string(make([]byte, 512)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create unique copy for this test to avoid shared mutations.
			testUser := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       tt.user.ID,
				Username: tt.user.Username,
			}

			err := repo.Create(ctx, testUser)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify retrieval works.
			retrieved, err := repo.FindByID(ctx, testUser.ID)
			require.NoError(t, err)
			require.Equal(t, testUser.ID, retrieved.ID)
			require.Equal(t, testUser.Username, retrieved.Username)

			// Cleanup.
			require.NoError(t, repo.Delete(ctx, testUser.ID))
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewUserRepository(testDB)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "found existing user",
			wantErr: false,
		},
		{
			name:    "nonexistent user",
			wantErr: true, // GORM returns gorm.ErrRecordNotFound
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantErr {
				// Test nonexistent user.
				nonexistentID := *testJWKGenService.GenerateUUIDv7()
				retrieved, err := repo.FindByID(ctx, nonexistentID)
				require.Error(t, err)
				require.Nil(t, retrieved)

				return
			}

			// Test found existing user.
			user := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-user-" + testJWKGenService.GenerateUUIDv7().String(),
			}
			require.NoError(t, repo.Create(ctx, user))

			defer func() { _ = repo.Delete(ctx, user.ID) }()

			retrieved, err := repo.FindByID(ctx, user.ID)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			require.Equal(t, user.ID, retrieved.ID)
			require.Equal(t, user.Username, retrieved.Username)
		})
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewUserRepository(testDB)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "found existing user by username",
			wantErr: false,
		},
		{
			name:    "nonexistent username",
			wantErr: true, // GORM returns gorm.ErrRecordNotFound
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantErr {
				// Test nonexistent username.
				nonexistentUsername := "nonexistent-user-" + testJWKGenService.GenerateUUIDv7().String()
				retrieved, err := repo.FindByUsername(ctx, nonexistentUsername)
				require.Error(t, err)
				require.Nil(t, retrieved)

				return
			}

			// Test found existing user.
			user := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-user-unique-" + testJWKGenService.GenerateUUIDv7().String(),
			}
			require.NoError(t, repo.Create(ctx, user))

			defer func() { _ = repo.Delete(ctx, user.ID) }()

			retrieved, err := repo.FindByUsername(ctx, user.Username)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			require.Equal(t, user.ID, retrieved.ID)
			require.Equal(t, user.Username, retrieved.Username)
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewUserRepository(testDB)

	// Create test user.
	user := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-user-original-" + testJWKGenService.GenerateUUIDv7().String(),
	}
	require.NoError(t, repo.Create(ctx, user))

	defer func() { _ = repo.Delete(ctx, user.ID) }()

	// Modify username.
	user.Username = "test-user-updated-" + testJWKGenService.GenerateUUIDv7().String()

	// Update user.
	err := repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify update persisted.
	retrieved, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.Username, retrieved.Username, "Updated username should be persisted")
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewUserRepository(testDB)

	// Create test user.
	user := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-user-" + testJWKGenService.GenerateUUIDv7().String(),
	}
	require.NoError(t, repo.Create(ctx, user))

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr bool
	}{
		{
			name:    "delete existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "delete nonexistent user (idempotent)",
			id:      *testJWKGenService.GenerateUUIDv7(),
			wantErr: false, // GORM doesn't error on 0 rows deleted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := repo.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify deletion.
			if tt.id == user.ID {
				_, err := repo.FindByID(ctx, tt.id)
				require.Error(t, err, "Should not find deleted user")
				require.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

func TestUserRepository_TransactionContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewUserRepository(testDB)

	// Test transaction rollback.
	tx := testDB.Begin()
	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	user := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-user-rollback-" + testJWKGenService.GenerateUUIDv7().String(),
	}

	// Create user within transaction.
	require.NoError(t, repo.Create(txCtx, user))

	// Rollback transaction.
	require.NoError(t, tx.Rollback().Error)

	// Verify user was NOT persisted (transaction rolled back).
	_, err := repo.FindByID(ctx, user.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Test transaction commit.
	tx = testDB.Begin()
	txCtx = cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	user2 := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-user-commit-" + testJWKGenService.GenerateUUIDv7().String(),
	}

	// Create user within transaction.
	require.NoError(t, repo.Create(txCtx, user2))

	// Commit transaction.
	require.NoError(t, tx.Commit().Error)

	defer func() { _ = repo.Delete(ctx, user2.ID) }()

	// Verify user WAS persisted (transaction committed).
	retrieved, err := repo.FindByID(ctx, user2.ID)
	require.NoError(t, err)
	require.Equal(t, user2.ID, retrieved.ID)
	require.Equal(t, user2.Username, retrieved.Username)
}
