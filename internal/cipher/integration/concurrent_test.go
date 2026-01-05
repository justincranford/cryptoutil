// Copyright (c) 2025 Justin Cranford
//
//

//go:build !windows

package integration

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"cryptoutil/internal/cipher/domain"
)

// TestConcurrent_MultipleUsersSimultaneousSends tests concurrent message sending scenarios.
// Tests robustness of database transactions, encryption/decryption, and race condition prevention.
func TestConcurrent_MultipleUsersSimultaneousSends(t *testing.T) {
	// Use shared server from TestMain (amortizes startup cost).
	require.NotNil(t, sharedServer)
	require.NotEmpty(t, sharedServiceBaseURL)

	// Create HTTP client for API calls.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Test server uses self-signed cert.
		},
		Timeout: 10 * time.Second,
	}

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

			// Create test users via API.
			users := createTestUsersAPI(t, client, sharedServiceBaseURL, tt.numUsers)

			// Send messages concurrently.
			var wg sync.WaitGroup
			for i := 0; i < tt.concurrentSends; i++ {
				wg.Add(1)

				go func(senderIdx int) {
					defer wg.Done()

					sender := users[senderIdx%len(users)]
					recipients := selectRecipients(users, sender.ID, tt.recipientsEach)

					// Create message via API.
					messageID := googleUuid.New()

					recipientIDs := make([]googleUuid.UUID, len(recipients))
					for i, r := range recipients {
						recipientIDs[i] = r.ID
					}

					sendMessageAPI(t, client, sharedServiceBaseURL, sender.ID, messageID, recipientIDs, fmt.Sprintf("encrypted-content-%d", senderIdx))
				}(i)
			}

			wg.Wait()

			duration := time.Since(start)

			// Verify timing (should complete within target duration).
			require.Less(t, duration, tt.targetDuration, "Test took too long: %v > %v", duration, tt.targetDuration)

			// Verify all messages created successfully by querying user inboxes.
			totalMessagesReceived := 0

			for _, user := range users {
				messages := getMessagesAPI(t, client, sharedServiceBaseURL, user.ID)
				totalMessagesReceived += len(messages)
			}

			// Each message is sent to N recipients, so total messages received = concurrentSends * recipientsEach.
			expectedTotalReceived := tt.concurrentSends * tt.recipientsEach
			require.Equal(t, expectedTotalReceived, totalMessagesReceived, "Expected %d total messages received, got %d", expectedTotalReceived, totalMessagesReceived)
		})
	}
}

// createTestUsersAPI creates N test users via API calls.
func createTestUsersAPI(t *testing.T, client *http.Client, baseURL string, numUsers int) []*domain.User {
	t.Helper()

	users := make([]*domain.User, numUsers)

	for i := 0; i < numUsers; i++ {
		userID := googleUuid.New()
		username := fmt.Sprintf("user%d_%s", i, googleUuid.NewString()[:8])

		// Register user via API.
		reqBody := map[string]interface{}{
			"id":       userID.String(),
			"username": username,
			"password": "test-password-123",
		}

		body, _ := json.Marshal(reqBody)
		resp, err := client.Post(baseURL+"/users/register", "application/json", bytes.NewReader(body))
		require.NoError(t, err)

		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create user %s", username)

		// Parse response.
		var user domain.User

		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		users[i] = &user
	}

	return users
}

// sendMessageAPI sends a message via API call.
func sendMessageAPI(t *testing.T, client *http.Client, baseURL string, senderID, messageID googleUuid.UUID, recipientIDs []googleUuid.UUID, content string) {
	t.Helper()

	reqBody := map[string]interface{}{
		"id":            messageID.String(),
		"sender_id":     senderID.String(),
		"recipient_ids": recipientIDs,
		"jwe":           content,
	}

	body, _ := json.Marshal(reqBody)
	resp, err := client.Post(baseURL+"/messages", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to send message")
}

// getMessagesAPI retrieves messages for a user via API call.
func getMessagesAPI(t *testing.T, client *http.Client, baseURL string, userID googleUuid.UUID) []domain.Message {
	t.Helper()

	resp, err := client.Get(fmt.Sprintf("%s/users/%s/messages", baseURL, userID.String()))
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get messages")

	var messages []domain.Message

	err = json.NewDecoder(resp.Body).Decode(&messages)
	require.NoError(t, err)

	return messages
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
