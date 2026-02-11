// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"sync"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsCipherImDomain "cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TestConcurrentAccess_ParallelCreates tests concurrent Create operations for race conditions.
func TestConcurrentAccess_ParallelCreates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)
	userRepo := NewUserRepository(testDB)

	// Create test user.
	senderID := googleUuid.New()
	testUser := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       senderID,
		Username: "concurrent_sender_" + googleUuid.New().String(),
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// Create 10 messages concurrently.
	const numMessages = 10

	var wg sync.WaitGroup
	wg.Add(numMessages)

	errors := make([]error, numMessages)

	for i := 0; i < numMessages; i++ {
		go func(index int) {
			defer wg.Done()

			messageID := googleUuid.New()
			message := &cryptoutilAppsCipherImDomain.Message{
				ID:       messageID,
				SenderID: senderID,
				JWE:      "test-jwe-concurrent-" + googleUuid.New().String(),
			}

			errors[index] = messageRepo.Create(ctx, message)
		}(i)
	}

	wg.Wait()

	// Verify all creates succeeded.
	for i, err := range errors {
		require.NoError(t, err, "Create %d failed", i)
	}

	// Verify all messages exist in database.
	var count int64
	require.NoError(t, testDB.Model(&cryptoutilAppsCipherImDomain.Message{}).Where("sender_id = ?", senderID).Count(&count).Error)
	require.Equal(t, int64(numMessages), count)
}

// TestConcurrentAccess_ParallelReadsAndWrites tests concurrent read/write operations.
func TestConcurrentAccess_ParallelReadsAndWrites(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)
	userRepo := NewUserRepository(testDB)

	// Create test user and message.
	senderID := googleUuid.New()
	testUser := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       senderID,
		Username: "rw_sender_" + googleUuid.New().String(),
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	messageID := googleUuid.New()
	testMessage := &cryptoutilAppsCipherImDomain.Message{
		ID:       messageID,
		SenderID: senderID,
		JWE:      "test-jwe-read-write",
	}
	require.NoError(t, messageRepo.Create(ctx, testMessage))

	// Start 5 readers and 5 writers concurrently.
	const (
		numReaders = 5
		numWriters = 5
	)

	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	readErrors := make([]error, numReaders)
	writeErrors := make([]error, numWriters)

	// Start readers.
	for i := 0; i < numReaders; i++ {
		go func(index int) {
			defer wg.Done()

			_, readErrors[index] = messageRepo.FindByID(ctx, messageID)
		}(i)
	}

	// Start writers (MarkAsRead operations).
	for i := 0; i < numWriters; i++ {
		go func(index int) {
			defer wg.Done()

			writeErrors[index] = messageRepo.MarkAsRead(ctx, messageID)
		}(i)
	}

	wg.Wait()

	// Verify all operations succeeded.
	for i, err := range readErrors {
		require.NoError(t, err, "Read %d failed", i)
	}

	for i, err := range writeErrors {
		require.NoError(t, err, "Write %d failed", i)
	}

	// Verify message was marked as read.
	updatedMessage, err := messageRepo.FindByID(ctx, messageID)
	require.NoError(t, err)
	require.NotNil(t, updatedMessage.ReadAt)
}

// TestConcurrentAccess_ParallelFindByRecipientID tests concurrent FindByRecipientID queries.
func TestConcurrentAccess_ParallelFindByRecipientID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)
	userRepo := NewUserRepository(testDB)
	jwkRepo := NewMessageRecipientJWKRepository(testDB, testBarrierService)

	// Create sender and recipient.
	senderID := googleUuid.New()
	recipientID := googleUuid.New()

	sender := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       senderID,
		Username: "sender_" + googleUuid.New().String(),
	}
	require.NoError(t, userRepo.Create(ctx, sender))

	recipient := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       recipientID,
		Username: "recipient_" + googleUuid.New().String(),
	}
	require.NoError(t, userRepo.Create(ctx, recipient))

	// Create 3 messages for recipient.
	for i := 0; i < 3; i++ {
		messageID := googleUuid.New()
		message := &cryptoutilAppsCipherImDomain.Message{
			ID:       messageID,
			SenderID: senderID,
			JWE:      "test-jwe-recipient-" + googleUuid.New().String(),
		}
		require.NoError(t, messageRepo.Create(ctx, message))

		// Create recipient JWK association.
		jwkEntry := &cryptoutilAppsCipherImDomain.MessageRecipientJWK{
			ID:           googleUuid.New(),
			MessageID:    messageID,
			RecipientID:  recipientID,
			EncryptedJWK: "test-encrypted-jwk",
		}
		require.NoError(t, jwkRepo.Create(ctx, jwkEntry))
	}

	// Query messages concurrently 10 times.
	const numQueries = 10

	var wg sync.WaitGroup
	wg.Add(numQueries)

	results := make([][]cryptoutilAppsCipherImDomain.Message, numQueries)
	errors := make([]error, numQueries)

	for i := 0; i < numQueries; i++ {
		go func(index int) {
			defer wg.Done()

			results[index], errors[index] = messageRepo.FindByRecipientID(ctx, recipientID)
		}(i)
	}

	wg.Wait()

	// Verify all queries succeeded and returned 3 messages.
	for i := 0; i < numQueries; i++ {
		require.NoError(t, errors[i], "Query %d failed", i)
		require.Len(t, results[i], 3, "Query %d returned wrong count", i)
	}
}

// TestConcurrentAccess_ParallelDeletes tests concurrent Delete operations.
func TestConcurrentAccess_ParallelDeletes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	messageRepo := NewMessageRepository(testDB)
	userRepo := NewUserRepository(testDB)

	// Create test user.
	senderID := googleUuid.New()
	testUser := &cryptoutilAppsTemplateServiceServerRepository.User{
		ID:       senderID,
		Username: "delete_sender_" + googleUuid.New().String(),
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// Create 5 messages.
	const numMessages = 5

	messageIDs := make([]googleUuid.UUID, numMessages)

	for i := 0; i < numMessages; i++ {
		messageIDs[i] = googleUuid.New()
		message := &cryptoutilAppsCipherImDomain.Message{
			ID:       messageIDs[i],
			SenderID: senderID,
			JWE:      "test-jwe-delete-" + googleUuid.New().String(),
		}
		require.NoError(t, messageRepo.Create(ctx, message))
	}

	// Delete all messages concurrently.
	var wg sync.WaitGroup
	wg.Add(numMessages)

	errors := make([]error, numMessages)

	for i := 0; i < numMessages; i++ {
		go func(index int) {
			defer wg.Done()

			errors[index] = messageRepo.Delete(ctx, messageIDs[index])
		}(i)
	}

	wg.Wait()

	// Verify all deletes succeeded.
	for i, err := range errors {
		require.NoError(t, err, "Delete %d failed", i)
	}

	// Verify all messages deleted.
	var count int64
	require.NoError(t, testDB.Model(&cryptoutilAppsCipherImDomain.Message{}).Where("sender_id = ?", senderID).Count(&count).Error)
	require.Equal(t, int64(0), count)
}
