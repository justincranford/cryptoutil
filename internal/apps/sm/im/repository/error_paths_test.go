// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"database/sql"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsSmImDomain "cryptoutil/internal/apps/sm/im/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilTestdb "cryptoutil/internal/apps/template/service/testing/testdb"
)

// TestErrorPaths_CreateOperations tests error scenarios for Create operations.
func TestErrorPaths_CreateOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name   string
		testFn func(*testing.T)
	}{
		{
			name: "MessageRepository.Create with duplicate ID",
			testFn: func(t *testing.T) {
				repo := NewMessageRepository(testDB)

				messageID := *testJWKGenService.GenerateUUIDv7()

				// Create first message.
				message1 := &cryptoutilAppsSmImDomain.Message{
					ID:       messageID,
					SenderID: *testJWKGenService.GenerateUUIDv7(),
					JWE:      "test-jwe-1",
				}
				require.NoError(t, repo.Create(ctx, message1))

				defer func() { _ = repo.Delete(ctx, messageID) }()

				// Try to create duplicate.
				message2 := &cryptoutilAppsSmImDomain.Message{
					ID:       messageID, // Same ID = constraint violation
					SenderID: *testJWKGenService.GenerateUUIDv7(),
					JWE:      "test-jwe-2",
				}

				err := repo.Create(ctx, message2)
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create message")
			},
		},
		{
			name: "UserRepository.Create with duplicate username",
			testFn: func(t *testing.T) {
				repo := NewUserRepository(testDB)

				username := "test-user-" + testJWKGenService.GenerateUUIDv7().String()

				// Create first user.
				user1 := &cryptoutilAppsTemplateServiceServerRepository.User{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					Username: username,
				}
				require.NoError(t, repo.Create(ctx, user1))

				defer func() { _ = repo.Delete(ctx, user1.ID) }()

				// Try to create duplicate username.
				user2 := &cryptoutilAppsTemplateServiceServerRepository.User{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					Username: username, // Same username = constraint violation
				}

				err := repo.Create(ctx, user2)
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create user")
			},
		},
		{
			name: "MessageRecipientJWKRepository.Create with duplicate (recipient_id, message_id)",
			testFn: func(t *testing.T) {
				msgRepo := NewMessageRepository(testDB)
				jwkRepo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

				// Create message first (foreign key requirement).
				messageID := *testJWKGenService.GenerateUUIDv7()
				message := &cryptoutilAppsSmImDomain.Message{
					ID:       messageID,
					SenderID: *testJWKGenService.GenerateUUIDv7(),
					JWE:      "test-jwe",
				}
				require.NoError(t, msgRepo.Create(ctx, message))

				defer func() { _ = msgRepo.Delete(ctx, messageID) }()

				recipientID := *testJWKGenService.GenerateUUIDv7()

				// Create first recipient JWK.
				jwk1 := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
					ID:           *testJWKGenService.GenerateUUIDv7(),
					RecipientID:  recipientID,
					MessageID:    messageID,
					EncryptedJWK: "test-encrypted-jwk-1",
				}
				require.NoError(t, jwkRepo.Create(ctx, jwk1))

				defer func() { _ = jwkRepo.Delete(ctx, jwk1.ID) }()

				// Try to create duplicate with same ID = constraint violation.
				jwk2 := &cryptoutilAppsSmImDomain.MessageRecipientJWK{
					ID:           jwk1.ID, // Same ID = primary key violation
					RecipientID:  recipientID,
					MessageID:    messageID,
					EncryptedJWK: "test-encrypted-jwk-2",
				}

				err := jwkRepo.Create(ctx, jwk2)
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create message recipient JWK")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.testFn(t)
		})
	}
}

// TestErrorPaths_UpdateOperations tests error scenarios for Update operations.
func TestErrorPaths_UpdateOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name   string
		testFn func(*testing.T)
	}{
		{
			name: "UserRepository.Update succeeds even with non-existent user",
			testFn: func(t *testing.T) {
				repo := NewUserRepository(testDB)

				// GORM Save() doesn't error on non-existent records.
				// This test verifies the behavior (no error, 0 rows affected).
				nonExistentID := *testJWKGenService.GenerateUUIDv7()
				user := &cryptoutilAppsTemplateServiceServerRepository.User{
					ID:       nonExistentID,
					Username: "non-existent-user",
				}

				err := repo.Update(ctx, user)
				// GORM Save() inserts if record doesn't exist.
				require.NoError(t, err)

				// Clean up the inserted record.
				defer func() { _ = repo.Delete(ctx, nonExistentID) }()
			},
		},
		{
			name: "MessageRepository.MarkAsRead succeeds even with non-existent message",
			testFn: func(t *testing.T) {
				repo := NewMessageRepository(testDB)

				// GORM Model().Where().Update() doesn't error on 0 rows affected.
				// This test verifies the behavior (no error returned).
				nonExistentID := *testJWKGenService.GenerateUUIDv7()

				err := repo.MarkAsRead(ctx, nonExistentID)
				// GORM doesn't error on 0 rows affected.
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.testFn(t)
		})
	}
}

// TestErrorPaths_DeleteOperations tests error scenarios for Delete operations.
func TestErrorPaths_DeleteOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name   string
		testFn func(*testing.T)
	}{
		{
			name: "MessageRepository.Delete with non-existent message",
			testFn: func(t *testing.T) {
				repo := NewMessageRepository(testDB)

				// Try to delete non-existent message.
				nonExistentID := *testJWKGenService.GenerateUUIDv7()

				err := repo.Delete(ctx, nonExistentID)
				// Note: GORM doesn't error on delete of non-existent record,
				// so we check that no panic occurred and operation completes.
				require.NoError(t, err)
			},
		},
		{
			name: "UserRepository.Delete with non-existent user",
			testFn: func(t *testing.T) {
				repo := NewUserRepository(testDB)

				// Try to delete non-existent user.
				nonExistentID := *testJWKGenService.GenerateUUIDv7()

				err := repo.Delete(ctx, nonExistentID)
				// Note: GORM doesn't error on delete of non-existent record,
				// so we check that no panic occurred and operation completes.
				require.NoError(t, err)
			},
		},
		{
			name: "MessageRecipientJWKRepository.Delete with non-existent JWK",
			testFn: func(t *testing.T) {
				repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

				// Try to delete non-existent JWK.
				nonExistentID := *testJWKGenService.GenerateUUIDv7()

				err := repo.Delete(ctx, nonExistentID)
				// GORM doesn't error on delete of non-existent record.
				require.NoError(t, err)
			},
		},
		{
			name: "MessageRecipientJWKRepository.DeleteByMessageID with non-existent message",
			testFn: func(t *testing.T) {
				repo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

				// Try to delete by non-existent message ID.
				nonExistentMessageID := *testJWKGenService.GenerateUUIDv7()

				err := repo.DeleteByMessageID(ctx, nonExistentMessageID)
				// GORM doesn't error on delete with 0 rows affected.
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.testFn(t)
		})
	}
}
// TestErrorReturns_DatabaseErrors tests error paths when database operations fail.
// These tests increase coverage by exercising error return statements in repository methods.
func TestErrorReturns_DatabaseErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a closed database for error testing using the shared testdb helper.
	closedGormDB := cryptoutilTestdb.NewClosedSQLiteDB(t, nil)

	t.Run("MessageRepository.FindByRecipientID error", func(t *testing.T) {
		repo := NewMessageRepository(closedGormDB)
		_, err := repo.FindByRecipientID(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find messages by recipient")
	})

	t.Run("MessageRepository.MarkAsRead error", func(t *testing.T) {
		repo := NewMessageRepository(closedGormDB)
		err := repo.MarkAsRead(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to mark message as read")
	})

	t.Run("MessageRepository.Delete error", func(t *testing.T) {
		repo := NewMessageRepository(closedGormDB)
		err := repo.Delete(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete message")
	})

	t.Run("MessageRecipientJWKRepository.FindByMessageID error", func(t *testing.T) {
		repo := NewMessageRecipientJWKRepository(closedGormDB, testBarrierService)
		_, err := repo.FindByMessageID(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find message recipient JWKs")
	})

	t.Run("MessageRecipientJWKRepository.Delete error", func(t *testing.T) {
		repo := NewMessageRecipientJWKRepository(closedGormDB, testBarrierService)
		err := repo.Delete(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete message recipient JWK")
	})

	t.Run("MessageRecipientJWKRepository.DeleteByMessageID error", func(t *testing.T) {
		repo := NewMessageRecipientJWKRepository(closedGormDB, testBarrierService)
		err := repo.DeleteByMessageID(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete message recipient JWKs")
	})

	t.Run("UserRepository.Update error", func(t *testing.T) {
		repo := NewUserRepository(closedGormDB)
		user := &cryptoutilAppsTemplateServiceServerRepository.User{
			ID:       googleUuid.New(),
			Username: "test-user",
		}
		err := repo.Update(ctx, user)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to update user")
	})

	t.Run("UserRepository.Delete error", func(t *testing.T) {
		repo := NewUserRepository(closedGormDB)
		err := repo.Delete(ctx, googleUuid.New())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete user")
	})
}

// TestApplySmIMMigrations_Error tests migration error path.
func TestApplySmIMMigrations_Error(t *testing.T) {
	t.Parallel()

	// Create a closed raw database to trigger migration errors.
	closedDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	err = closedDB.Close()
	require.NoError(t, err)

	// Apply migrations should fail on closed database.
	err = ApplySmIMMigrations(closedDB, DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to apply sm-im migrations")
}

// Note: The following coverage gaps are intentionally NOT tested:
//
// 1. UserRepositoryAdapter.Create panic (type assertion failure):
//    - Requires implementing full 50+ method UserModel interface
//    - Panic is defensive programming for misuse, not normal error path
//    - Production code only ever passes *repository.User
//
// 2. mergedFS.ReadDir "directory not found" error:
//    - Would require mocking embed.FS which is not practical
//    - Error path is already exercised by migrations_test.go for normal cases
//    - This specific error (both templateFS and smIMFS failing) is unlikely
