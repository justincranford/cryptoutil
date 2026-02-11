// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsCipherImDomain "cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

func TestMessageRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	// Create test user for sender relation.
	userRepo := NewUserRepository(testDB)
	sender := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-sender-" + testJWKGenService.GenerateUUIDv7().String(),
	}
	require.NoError(t, userRepo.Create(ctx, sender))

	defer func() { _ = userRepo.Delete(ctx, sender.ID) }()

	tests := []struct {
		name    string
		message *cryptoutilAppsCipherImDomain.Message
		wantErr bool
	}{
		{
			name: "valid message creation",
			message: &cryptoutilAppsCipherImDomain.Message{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				SenderID: sender.ID,
				JWE:      `{"protected":"...","encrypted_key":"...","iv":"...","ciphertext":"...","tag":"..."}`,
			},
			wantErr: false,
		},
		{
			name: "message with empty JWE",
			message: &cryptoutilAppsCipherImDomain.Message{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				SenderID: sender.ID,
				JWE:      "",
			},
			wantErr: false, // Repository doesn't validate JWE content (validation in handler)
		},
		{
			name: "message with large JWE payload",
			message: &cryptoutilAppsCipherImDomain.Message{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				SenderID: sender.ID,
				JWE:      `{"protected":"...","ciphertext":"` + string(make([]byte, 4096)) + `"}`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create unique copy for this test to avoid shared mutations.
			testMessage := &cryptoutilAppsCipherImDomain.Message{
				ID:       tt.message.ID,
				SenderID: tt.message.SenderID,
				JWE:      tt.message.JWE,
			}

			err := repo.Create(ctx, testMessage)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify retrieval works.
			retrieved, err := repo.FindByID(ctx, testMessage.ID)
			require.NoError(t, err)
			require.Equal(t, testMessage.ID, retrieved.ID)
			require.Equal(t, testMessage.SenderID, retrieved.SenderID)
			require.Equal(t, testMessage.JWE, retrieved.JWE)
			require.NotZero(t, retrieved.CreatedAt, "CreatedAt should be set automatically")

			// Cleanup.
			require.NoError(t, repo.Delete(ctx, testMessage.ID))
		})
	}
}

func TestMessageRepository_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "found existing message",
			wantErr: false,
		},
		{
			name:    "nonexistent message",
			wantErr: true, // GORM returns gorm.ErrRecordNotFound
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantErr {
				// Test nonexistent message.
				nonexistentID := *testJWKGenService.GenerateUUIDv7()
				retrieved, err := repo.FindByID(ctx, nonexistentID)
				require.Error(t, err)
				require.Nil(t, retrieved)

				return
			}

			// Test found existing message.
			userRepo := NewUserRepository(testDB)
			sender := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-sender-" + testJWKGenService.GenerateUUIDv7().String(),
			}
			require.NoError(t, userRepo.Create(ctx, sender))

			defer func() { _ = userRepo.Delete(ctx, sender.ID) }()

			message := &cryptoutilAppsCipherImDomain.Message{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				SenderID: sender.ID,
				JWE:      `{"protected":"test"}`,
			}
			require.NoError(t, repo.Create(ctx, message))

			defer func() { _ = repo.Delete(ctx, message.ID) }()

			retrieved, err := repo.FindByID(ctx, message.ID)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			require.Equal(t, message.ID, retrieved.ID)
			require.Equal(t, message.SenderID, retrieved.SenderID)
			require.Equal(t, message.JWE, retrieved.JWE)
		})
	}
}

func TestMessageRepository_FindByRecipientID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	tests := []struct {
		name      string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "find all messages for recipient (3 messages)",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "nonexistent recipient (0 messages)",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test users for sender and recipient.
			userRepo := NewUserRepository(testDB)
			sender := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-sender-" + testJWKGenService.GenerateUUIDv7().String(),
			}
			require.NoError(t, userRepo.Create(ctx, sender))

			defer func() { _ = userRepo.Delete(ctx, sender.ID) }()

			recipient := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-recipient-" + testJWKGenService.GenerateUUIDv7().String(),
			}
			require.NoError(t, userRepo.Create(ctx, recipient))

			defer func() { _ = userRepo.Delete(ctx, recipient.ID) }()

			if tt.name == "nonexistent recipient (0 messages)" {
				// Test nonexistent recipient.
				nonexistentID := *testJWKGenService.GenerateUUIDv7()
				retrieved, err := repo.FindByRecipientID(ctx, nonexistentID)
				require.NoError(t, err)
				require.Len(t, retrieved, 0)

				return
			}

			// Test finding messages for recipient.
			messages := []*cryptoutilAppsCipherImDomain.Message{
				{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					SenderID: sender.ID,
					JWE:      `{"protected":"message1"}`,
				},
				{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					SenderID: sender.ID,
					JWE:      `{"protected":"message2"}`,
				},
				{
					ID:       *testJWKGenService.GenerateUUIDv7(),
					SenderID: sender.ID,
					JWE:      `{"protected":"message3"}`,
				},
			}

			for _, msg := range messages {
				require.NoError(t, repo.Create(ctx, msg))
			}

			defer func() {
				for _, msg := range messages {
					_ = repo.Delete(ctx, msg.ID)
				}
			}()

			// Create message_recipient_jwks for this recipient (3-table schema: messages + messages_recipient_jwks + users).
			jwkRepo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

			for _, msg := range messages {
				jwk := &cryptoutilAppsCipherImDomain.MessageRecipientJWK{
					ID:           *testJWKGenService.GenerateUUIDv7(),
					RecipientID:  recipient.ID,
					MessageID:    msg.ID,
					EncryptedJWK: generateTestJWK(t),
				}
				require.NoError(t, jwkRepo.Create(ctx, jwk))

				defer func(id googleUuid.UUID) { _ = jwkRepo.Delete(ctx, id) }(jwk.ID)
			}

			retrieved, err := repo.FindByRecipientID(ctx, recipient.ID)
			require.NoError(t, err)
			require.Len(t, retrieved, tt.wantCount)

			// Verify messages ordered by created_at DESC.
			if tt.wantCount > 0 {
				for i := 0; i < len(retrieved)-1; i++ {
					require.GreaterOrEqual(t, retrieved[i].CreatedAt, retrieved[i+1].CreatedAt, "Messages should be ordered by created_at DESC")
				}

				// Verify sender preloaded.
				for _, msg := range retrieved {
					require.NotNil(t, msg.Sender, "Sender should be preloaded")
					require.Equal(t, sender.ID, msg.Sender.ID)
				}
			}
		})
	}
}

func TestMessageRepository_MarkAsRead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "mark existing message as read",
			wantErr: false,
		},
		{
			name:    "mark nonexistent message (no error, 0 rows affected)",
			wantErr: false, // GORM doesn't error on 0 rows updated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.name == "mark nonexistent message (no error, 0 rows affected)" {
				// Test nonexistent message.
				nonexistentID := *testJWKGenService.GenerateUUIDv7()
				err := repo.MarkAsRead(ctx, nonexistentID)
				require.NoError(t, err) // GORM doesn't error on 0 rows updated

				return
			}

			// Test marking existing message as read.
			userRepo := NewUserRepository(testDB)
			sender := &cryptoutilAppsTemplateServiceServerRepository.User{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				Username: "test-sender-" + testJWKGenService.GenerateUUIDv7().String(),
			}
			require.NoError(t, userRepo.Create(ctx, sender))

			defer func() { _ = userRepo.Delete(ctx, sender.ID) }()

			message := &cryptoutilAppsCipherImDomain.Message{
				ID:       *testJWKGenService.GenerateUUIDv7(),
				SenderID: sender.ID,
				JWE:      `{"protected":"test"}`,
			}
			require.NoError(t, repo.Create(ctx, message))

			defer func() { _ = repo.Delete(ctx, message.ID) }()

			err := repo.MarkAsRead(ctx, message.ID)
			require.NoError(t, err)

			// Verify read_at is set.
			retrieved, err := repo.FindByID(ctx, message.ID)
			require.NoError(t, err)
			require.NotNil(t, retrieved.ReadAt, "ReadAt should be set after marking as read")
		})
	}
}

func TestMessageRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	// Create test user for sender relation.
	userRepo := NewUserRepository(testDB)
	sender := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-sender-" + testJWKGenService.GenerateUUIDv7().String(),
	}
	require.NoError(t, userRepo.Create(ctx, sender))

	defer func() { _ = userRepo.Delete(ctx, sender.ID) }()

	// Create test message.
	message := &cryptoutilAppsCipherImDomain.Message{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		SenderID: sender.ID,
		JWE:      `{"protected":"test"}`,
	}
	require.NoError(t, repo.Create(ctx, message))

	tests := []struct {
		name    string
		id      googleUuid.UUID
		wantErr bool
	}{
		{
			name:    "delete existing message",
			id:      message.ID,
			wantErr: false,
		},
		{
			name:    "delete nonexistent message (idempotent)",
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
			if tt.id == message.ID {
				_, err := repo.FindByID(ctx, tt.id)
				require.Error(t, err, "Should not find deleted message")
				require.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

func TestMessageRepository_TransactionContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMessageRepository(testDB)

	// Create test user for sender relation.
	userRepo := NewUserRepository(testDB)
	sender := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		Username: "test-sender-" + testJWKGenService.GenerateUUIDv7().String(),
	}
	require.NoError(t, userRepo.Create(ctx, sender))

	defer func() { _ = userRepo.Delete(ctx, sender.ID) }()

	// Test transaction rollback.
	tx := testDB.Begin()
	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	message := &cryptoutilAppsCipherImDomain.Message{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		SenderID: sender.ID,
		JWE:      `{"protected":"test"}`,
	}

	// Create message within transaction.
	require.NoError(t, repo.Create(txCtx, message))

	// Rollback transaction.
	require.NoError(t, tx.Rollback().Error)

	// Verify message was NOT persisted (transaction rolled back).
	_, err := repo.FindByID(ctx, message.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Test transaction commit.
	tx = testDB.Begin()
	txCtx = cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	message2 := &cryptoutilAppsCipherImDomain.Message{
		ID:       *testJWKGenService.GenerateUUIDv7(),
		SenderID: sender.ID,
		JWE:      `{"protected":"test2"}`,
	}

	// Create message within transaction.
	require.NoError(t, repo.Create(txCtx, message2))

	// Commit transaction.
	require.NoError(t, tx.Commit().Error)

	defer func() { _ = repo.Delete(ctx, message2.ID) }()

	// Verify message WAS persisted (transaction committed).
	retrieved, err := repo.FindByID(ctx, message2.ID)
	require.NoError(t, err)
	require.Equal(t, message2.ID, retrieved.ID)
}
