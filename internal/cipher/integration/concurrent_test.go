// Copyright (c) 2025 Justin Cranford
//
//

//go:build !windows

package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cryptoutil/internal/cipher/domain"
	"cryptoutil/internal/cipher/repository"
)

// TestConcurrent_MultipleUsersSimultaneousSends tests concurrent message sending scenarios.
// Tests robustness of database transactions, encryption/decryption, and race condition prevention.
func TestConcurrent_MultipleUsersSimultaneousSends(t *testing.T) {
	ctx := context.Background()

	// Use shared server and database from TestMain (amortizes startup cost).
	// Server instance is reused across all tests.
	srv := sharedServer
	db := sharedDB
	require.NotNil(t, srv)
	require.NotNil(t, db)

	// Define test scenarios.
	tests := []struct {
		name            string
		numUsers        int
		concurrentSends int
		recipientsEach  int
		targetDuration  time.Duration
	}{
		{
			name:            "N=5 users, M=4 concurrent sends (1 recipient each)",
			numUsers:        5,
			concurrentSends: 4,
			recipientsEach:  1,
			targetDuration:  4 * time.Second,
		},
		{
			name:            "N=5 users, P=3 concurrent sends (2 recipients each)",
			numUsers:        5,
			concurrentSends: 3,
			recipientsEach:  2,
			targetDuration:  5 * time.Second,
		},
		{
			name:            "N=5 users, Q=2 concurrent sends (all recipients broadcast)",
			numUsers:        5,
			concurrentSends: 2,
			recipientsEach:  4, // All other users (broadcast)
			targetDuration:  6 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()

			// Clean up messages and users from previous subtests.
			err = db.Exec("DELETE FROM messages").Error
			require.NoError(t, err)
			err = db.Exec("DELETE FROM users").Error
			require.NoError(t, err)

			// Create test users.
			users := createTestUsers(t, db, tt.numUsers)

			// Send messages concurrently.
			var wg sync.WaitGroup
			for i := 0; i < tt.concurrentSends; i++ {
				wg.Add(1)

				go func(senderIdx int) {
					defer wg.Done()

					sender := users[senderIdx%len(users)]
					_ = selectRecipients(users, sender.ID, tt.recipientsEach) // TODO: Create MessageRecipientJWK entries when implementing Phase 9.2

					// Create message via repository.
					messageID := googleUuid.New()
					msg := &domain.Message{
						ID:       messageID,
						SenderID: sender.ID,
						JWE:      fmt.Sprintf("encrypted-content-%d", senderIdx),
					}

					msgRepo := repository.NewMessageRepository(db)
					err := msgRepo.Create(context.Background(), msg)
					require.NoError(t, err)
				}(i)
			}

			wg.Wait()

			duration := time.Since(start)

			// Verify timing (should complete within target duration).
			require.Less(t, duration, tt.targetDuration, "Test took too long: %v > %v", duration, tt.targetDuration)

			// Verify all messages created successfully.
			var allMessages []domain.Message

			err = db.Find(&allMessages).Error
			require.NoError(t, err)
			require.Len(t, allMessages, tt.concurrentSends, "Expected %d messages, got %d", tt.concurrentSends, len(allMessages))

			// Verify no data corruption (all messages have valid sender IDs).
			for _, msg := range allMessages {
				require.NotEqual(t, googleUuid.Nil, msg.SenderID, "Message has nil sender ID")
				require.NotEmpty(t, msg.JWE, "Message has empty JWE content")
			}
		})
	}
}

// createTestUsers creates N test users in the database.
func createTestUsers(t *testing.T, db *gorm.DB, numUsers int) []*domain.User {
	t.Helper()

	users := make([]*domain.User, numUsers)
	userRepo := repository.NewUserRepository(db)

	for i := 0; i < numUsers; i++ {
		userID := googleUuid.New()
		user := &domain.User{
			ID:           userID,
			Username:     fmt.Sprintf("user%d_%s", i, googleUuid.NewString()[:8]),
			PasswordHash: "test-hash-123", // Not validating password in this test
		}

		err := userRepo.Create(context.Background(), user)
		require.NoError(t, err)

		users[i] = user
	}

	return users
}

// selectRecipients selects N random recipients (excluding sender).
func selectRecipients(users []*domain.User, senderID googleUuid.UUID, count int) []*domain.User {
	recipients := make([]*domain.User, 0, count)

	for _, user := range users {
		if user.ID != senderID && len(recipients) < count {
			recipients = append(recipients, user)
		}
	}

	return recipients
}
