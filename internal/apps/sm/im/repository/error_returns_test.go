package repository

import (
	"context"
	"database/sql"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTestdb "cryptoutil/internal/apps/template/service/testing/testdb"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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
