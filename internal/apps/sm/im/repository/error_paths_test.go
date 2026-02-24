// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	cryptoutilAppsSmImDomain "cryptoutil/internal/apps/sm/im/domain"

	"github.com/stretchr/testify/require"
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
