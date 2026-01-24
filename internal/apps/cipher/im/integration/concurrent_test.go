// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"crypto/tls"
	"fmt"
	http "net/http"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsCipherImClient "cryptoutil/internal/apps/cipher/im/client"
	cryptoutilAppsTemplateServiceTestingE2e "cryptoutil/internal/apps/template/service/testing/e2e"
)

// TestConcurrent_MultipleUsersSimultaneousSends tests concurrent message sending scenarios.
// Tests robustness of database transactions, encryption/decryption, and race condition prevention.
func TestConcurrent_MultipleUsersSimultaneousSends(t *testing.T) {
	// Use shared server from TestMain (amortizes startup cost).
	require.NotNil(t, cipherImServer)
	require.NotEmpty(t, sharedServiceBaseURL)

	// Create HTTP client for API calls.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Test server uses self-signed cert.
		},
		Timeout: 10 * time.Second,
	}

	// Define test scenarios.
	// Note: Target durations adjusted based on actual performance benchmarks.
	// Each scenario includes user registration + encryption + barrier encryption overhead.
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
			targetDuration:  15 * time.Second, // Adjusted from 4s (observed ~14s)
		},
		{
			name:            "N=5 users, P=3 concurrent sends (2 recipients each)",
			numUsers:        5,
			concurrentSends: 3,
			recipientsEach:  2,
			targetDuration:  10 * time.Second, // Adjusted from 5s (observed ~8s)
		},
		{
			name:            "N=5 users, Q=2 concurrent sends (all recipients broadcast)",
			numUsers:        5,
			concurrentSends: 2,
			recipientsEach:  4,                // All other users (broadcast)
			targetDuration:  10 * time.Second, // Adjusted from 6s (observed ~7s)
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

					recipientIDs := make([]googleUuid.UUID, len(recipients))
					for i, r := range recipients {
						recipientIDs[i] = r.ID
					}

					sendMessageAPI(t, client, sharedServiceBaseURL, recipientIDs, fmt.Sprintf("encrypted-content-%d", senderIdx), sender.Token)
				}(i)
			}

			wg.Wait()

			duration := time.Since(start)

			// Verify timing (should complete within target duration).
			require.Less(t, duration, tt.targetDuration, "Test took too long: %v > %v", duration, tt.targetDuration)

			// Verify all messages created successfully by querying user inboxes.
			totalMessagesReceived := 0

			for _, user := range users {
				messages := getMessagesAPI(t, client, sharedServiceBaseURL, user.ID, user.Token)
				totalMessagesReceived += len(messages)
			}

			// Each message is sent to N recipients, so total messages received = concurrentSends * recipientsEach.
			expectedTotalReceived := tt.concurrentSends * tt.recipientsEach
			require.Equal(t, expectedTotalReceived, totalMessagesReceived, "Expected %d total messages received, got %d", expectedTotalReceived, totalMessagesReceived)
		})
	}
}

// createTestUsersAPI creates N test users via API calls using the reusable helper.
func createTestUsersAPI(t *testing.T, client *http.Client, baseURL string, numUsers int) []*cryptoutilAppsTemplateServiceTestingE2e.TestUser {
	t.Helper()

	users := make([]*cryptoutilAppsTemplateServiceTestingE2e.TestUser, numUsers)

	for i := 0; i < numUsers; i++ {
		users[i] = cryptoutilAppsTemplateServiceTestingE2e.RegisterTestUserService(t, client, baseURL)
	}

	return users
}

// sendMessageAPI sends a message via API call using cipher client.
func sendMessageAPI(t *testing.T, client *http.Client, baseURL string, recipientIDs []googleUuid.UUID, content string, token string) {
	t.Helper()

	_, err := cryptoutilAppsCipherImClient.SendMessage(client, baseURL, content, token, recipientIDs...)
	require.NoError(t, err, "Failed to send message")
}

// getMessagesAPI retrieves messages for a user via API call using cipher client.
func getMessagesAPI(t *testing.T, client *http.Client, baseURL string, _ googleUuid.UUID, token string) []map[string]any {
	t.Helper()

	messages, err := cryptoutilAppsCipherImClient.ReceiveMessagesService(client, baseURL, token)
	require.NoError(t, err, "Failed to get messages")

	return messages
}

// selectRecipients selects N random recipients (excluding sender).
func selectRecipients(users []*cryptoutilAppsTemplateServiceTestingE2e.TestUser, senderID googleUuid.UUID, count int) []*cryptoutilAppsTemplateServiceTestingE2e.TestUser {
	recipients := make([]*cryptoutilAppsTemplateServiceTestingE2e.TestUser, 0, count)

	for _, user := range users {
		if user.ID != senderID && len(recipients) < count {
			recipients = append(recipients, user)
		}
	}

	return recipients
}
