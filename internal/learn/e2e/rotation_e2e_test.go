// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	"cryptoutil/internal/learn/server/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

const adminPort = 0 // Dynamic port allocation for admin server.

// rotationTestServers holds both public and admin server instances for rotation tests.
type rotationTestServers struct {
	learnIMServer *server.LearnIMServer
	publicURL     string
	adminURL      string
	httpClient    *http.Client
}

// setupRotationTestServers creates a full learn-im server (public + admin) for rotation E2E tests.
func setupRotationTestServers(t *testing.T) *rotationTestServers {
	t.Helper()

	ctx := context.Background()

	// Create test database.
	db, err := initTestDB()
	require.NoError(t, err, "failed to create test database")

	// Generate random JWT secret for this test.
	randomJWTSecret, err := cryptoutilRandom.GenerateString(32)
	require.NoError(t, err, "failed to generate random JWT secret")

	// Create AppConfig for learn-im server (embedded ServerSettings).
	cfg := &config.AppConfig{
		JWTSecret: randomJWTSecret,
	}
	cfg.OTLPService = "learn-im-e2e-rotation-test"
	cfg.LogLevel = "error" // Suppress logs during tests.
	cfg.OTLPEndpoint = "grpc://" + cryptoutilMagic.HostnameLocalhost + ":" + strconv.Itoa(int(cryptoutilMagic.DefaultPublicPortOtelCollectorGRPC))

	// Create learn-im server (public + admin servers).
	learnIMServer, err := server.New(ctx, cfg, db, repository.DatabaseTypeSQLite)
	require.NoError(t, err, "failed to create learn-im server")

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := learnIMServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for server to bind to ports (public).
	const maxWaitAttempts = 50
	const waitInterval = 100 * time.Millisecond

	publicActualPort := 0

	for i := 0; i < maxWaitAttempts; i++ {
		publicActualPort = learnIMServer.PublicPort()
		if publicActualPort > 0 {
			break
		}

		time.Sleep(waitInterval)
	}

	require.Greater(t, publicActualPort, 0, "public server did not bind to port")

	// Wait for admin server to bind.
	adminActualPort := 0

	for i := 0; i < maxWaitAttempts; i++ {
		port, portErr := learnIMServer.AdminPort()
		if portErr == nil && port > 0 {
			adminActualPort = port
			break
		}

		time.Sleep(waitInterval)
	}

	require.Greater(t, adminActualPort, 0, "admin server did not bind to port")

	// Cleanup on test completion.
	t.Cleanup(func() {
		_ = learnIMServer.Shutdown(context.Background())
	})

	publicURL := "https://" + cryptoutilMagic.IPv4Loopback + ":" + strconv.Itoa(publicActualPort)
	adminURL := "https://" + cryptoutilMagic.IPv4Loopback + ":" + strconv.Itoa(adminActualPort)

	// Create HTTP client for admin requests (insecure TLS for test environment).
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.LearnDefaultTimeout,
	}

	return &rotationTestServers{
		learnIMServer: learnIMServer,
		publicURL:     publicURL,
		adminURL:      adminURL,
		httpClient:    httpClient,
	}
}

// TestE2E_RotateRootKey tests manual root key rotation via admin API.
func TestE2E_RotateRootKey(t *testing.T) {
	t.Parallel()

	testServers := setupRotationTestServers(t)

	// Step 1: Send baseline message before rotation.
	user1 := registerServiceUser(t, testServers.httpClient, testServers.publicURL, "user1_rotate_root", "Pass1234!")
	user2 := registerServiceUser(t, testServers.httpClient, testServers.publicURL, "user2_rotate_root", "Pass1234!")

	plaintext1 := "Message before root key rotation"

	messageID1 := sendMessage(t, testServers.httpClient, testServers.publicURL, plaintext1, user1.Token, user2.ID)
	require.NotEmpty(t, messageID1, "baseline message ID should not be empty")

	// Step 2: Get initial barrier keys status.
	initialStatus := getBarrierKeysStatus(t, testServers.httpClient, testServers.adminURL)

	initialRootKeyUUID, ok := initialStatus["root_key"].(map[string]any)["uuid"].(string)
	require.True(t, ok, "initial root_key uuid should be string")
	require.NotEmpty(t, initialRootKeyUUID, "initial root_key uuid should not be empty")

	// Step 3: Rotate root key via admin API.
	rotationReason := "E2E test: manual root key rotation"

	rotateResponse := rotateRootKey(t, testServers.httpClient, testServers.adminURL, rotationReason)

	oldKeyUUID, ok := rotateResponse["old_key_uuid"].(string)
	require.True(t, ok, "old_key_uuid should be string")
	require.Equal(t, initialRootKeyUUID, oldKeyUUID, "old_key_uuid should match initial root key")

	newKeyUUID, ok := rotateResponse["new_key_uuid"].(string)
	require.True(t, ok, "new_key_uuid should be string")
	require.NotEqual(t, oldKeyUUID, newKeyUUID, "new root key should be different from old key")

	returnedReason, ok := rotateResponse["reason"].(string)
	require.True(t, ok, "reason should be string")
	require.Equal(t, rotationReason, returnedReason, "returned reason should match request")

	rotatedAt, ok := rotateResponse["rotated_at"].(float64)
	require.True(t, ok, "rotated_at should be number")
	require.Greater(t, rotatedAt, float64(0), "rotated_at timestamp should be positive")

	// Step 4: Verify status endpoint reflects new root key.
	updatedStatus := getBarrierKeysStatus(t, testServers.httpClient, testServers.adminURL)

	updatedRootKeyUUID, ok := updatedStatus["root_key"].(map[string]any)["uuid"].(string)
	require.True(t, ok, "updated root_key uuid should be string")
	require.Equal(t, newKeyUUID, updatedRootKeyUUID, "status should reflect new root key UUID")

	// Step 5: Send new message after rotation (uses new root key chain).
	plaintext2 := "Message after root key rotation"

	messageID2 := sendMessage(t, testServers.httpClient, testServers.publicURL, plaintext2, user1.Token, user2.ID)
	require.NotEmpty(t, messageID2, "post-rotation message ID should not be empty")

	// Step 6: Verify user2 can decrypt BOTH old and new messages (backward compatibility).
	messages := receiveMessagesService(t, testServers.httpClient, testServers.publicURL, user2.Token)
	require.GreaterOrEqual(t, len(messages), 2, "user2 should have at least 2 messages")

	// Find both messages in received set.
	var foundOldMessage, foundNewMessage bool

	for _, msg := range messages {
		content, ok := msg["encrypted_content"].(string)
		require.True(t, ok, "encrypted_content should be string")

		if content == plaintext1 {
			foundOldMessage = true
		}

		if content == plaintext2 {
			foundNewMessage = true
		}
	}

	require.True(t, foundOldMessage, "old message (pre-rotation) should decrypt correctly")
	require.True(t, foundNewMessage, "new message (post-rotation) should decrypt correctly")
}

// TestE2E_RotateIntermediateKey tests manual intermediate key rotation via admin API.
func TestE2E_RotateIntermediateKey(t *testing.T) {
	t.Parallel()

	testServers := setupRotationTestServers(t)

	// Step 1: Send baseline message before rotation.
	user1 := registerServiceUser(t, testServers.httpClient, testServers.publicURL, "user1_rotate_intermediate", "Pass1234!")
	user2 := registerServiceUser(t, testServers.httpClient, testServers.publicURL, "user2_rotate_intermediate", "Pass1234!")

	plaintext1 := "Message before intermediate key rotation"

	messageID1 := sendMessage(t, testServers.httpClient, testServers.publicURL, plaintext1, user1.Token, user2.ID)
	require.NotEmpty(t, messageID1, "baseline message ID should not be empty")

	// Step 2: Get initial intermediate key status.
	initialStatus := getBarrierKeysStatus(t, testServers.httpClient, testServers.adminURL)

	initialIntermediateKeyUUID, ok := initialStatus["intermediate_key"].(map[string]any)["uuid"].(string)
	require.True(t, ok, "initial intermediate_key uuid should be string")
	require.NotEmpty(t, initialIntermediateKeyUUID, "initial intermediate_key uuid should not be empty")

	// Step 3: Rotate intermediate key via admin API.
	rotationReason := "E2E test: manual intermediate key rotation"

	rotateResponse := rotateIntermediateKey(t, testServers.httpClient, testServers.adminURL, rotationReason)

	oldKeyUUID, ok := rotateResponse["old_key_uuid"].(string)
	require.True(t, ok, "old_key_uuid should be string")
	require.Equal(t, initialIntermediateKeyUUID, oldKeyUUID, "old_key_uuid should match initial intermediate key")

	newKeyUUID, ok := rotateResponse["new_key_uuid"].(string)
	require.True(t, ok, "new_key_uuid should be string")
	require.NotEqual(t, oldKeyUUID, newKeyUUID, "new intermediate key should be different from old key")

	// Step 4: Verify status reflects new intermediate key.
	updatedStatus := getBarrierKeysStatus(t, testServers.httpClient, testServers.adminURL)

	updatedIntermediateKeyUUID, ok := updatedStatus["intermediate_key"].(map[string]any)["uuid"].(string)
	require.True(t, ok, "updated intermediate_key uuid should be string")
	require.Equal(t, newKeyUUID, updatedIntermediateKeyUUID, "status should reflect new intermediate key")

	// Step 5: Send new message after rotation.
	plaintext2 := "Message after intermediate key rotation"

	messageID2 := sendMessage(t, testServers.httpClient, testServers.publicURL, plaintext2, user1.Token, user2.ID)
	require.NotEmpty(t, messageID2, "post-rotation message ID should not be empty")

	// Step 6: Verify backward compatibility.
	messages := receiveMessagesService(t, testServers.httpClient, testServers.publicURL, user2.Token)
	require.GreaterOrEqual(t, len(messages), 2, "user2 should have at least 2 messages")

	var foundOldMessage, foundNewMessage bool

	for _, msg := range messages {
		content, ok := msg["encrypted_content"].(string)
		require.True(t, ok, "encrypted_content should be string")

		if content == plaintext1 {
			foundOldMessage = true
		}

		if content == plaintext2 {
			foundNewMessage = true
		}
	}

	require.True(t, foundOldMessage, "old message (pre-rotation) should decrypt correctly")
	require.True(t, foundNewMessage, "new message (post-rotation) should decrypt correctly")
}

// TestE2E_RotateContentKey tests manual content key rotation (elastic rotation).
func TestE2E_RotateContentKey(t *testing.T) {
	t.Parallel()

	testServers := setupRotationTestServers(t)

	// Step 1: Send baseline message (creates first content key).
	user1 := registerServiceUser(t, testServers.httpClient, testServers.publicURL, "user1_rotate_content", "Pass1234!")
	user2 := registerServiceUser(t, testServers.httpClient, testServers.publicURL, "user2_rotate_content", "Pass1234!")

	plaintext1 := "Message before content key rotation"

	messageID1 := sendMessage(t, testServers.httpClient, testServers.publicURL, plaintext1, user1.Token, user2.ID)
	require.NotEmpty(t, messageID1, "baseline message ID should not be empty")

	// Step 2: Rotate content key (elastic rotation - creates new key, keeps old).
	rotationReason := "E2E test: manual content key rotation"

	rotateResponse := rotateContentKey(t, testServers.httpClient, testServers.adminURL, rotationReason)

	// Content key rotation returns new_key_uuid only (no old_key_uuid - elastic rotation).
	newKeyUUID, ok := rotateResponse["new_key_uuid"].(string)
	require.True(t, ok, "new_key_uuid should be string")
	require.NotEmpty(t, newKeyUUID, "new content key UUID should not be empty")

	_, hasOldKeyUUID := rotateResponse["old_key_uuid"]
	require.False(t, hasOldKeyUUID, "content rotation should NOT return old_key_uuid (elastic rotation)")

	returnedReason, ok := rotateResponse["reason"].(string)
	require.True(t, ok, "reason should be string")
	require.Equal(t, rotationReason, returnedReason, "returned reason should match request")

	// Step 3: Send new message after rotation (uses new content key).
	plaintext2 := "Message after content key rotation"

	messageID2 := sendMessage(t, testServers.httpClient, testServers.publicURL, plaintext2, user1.Token, user2.ID)
	require.NotEmpty(t, messageID2, "post-rotation message ID should not be empty")

	// Step 4: Verify both messages decrypt correctly (elastic rotation preserves old keys).
	messages := receiveMessagesService(t, testServers.httpClient, testServers.publicURL, user2.Token)
	require.GreaterOrEqual(t, len(messages), 2, "user2 should have at least 2 messages")

	var foundOldMessage, foundNewMessage bool

	for _, msg := range messages {
		content, ok := msg["encrypted_content"].(string)
		require.True(t, ok, "encrypted_content should be string")

		if content == plaintext1 {
			foundOldMessage = true
		}

		if content == plaintext2 {
			foundNewMessage = true
		}
	}

	require.True(t, foundOldMessage, "old message (pre-rotation) should decrypt with old content key")
	require.True(t, foundNewMessage, "new message (post-rotation) should decrypt with new content key")
}

// TestE2E_GetBarrierKeysStatus tests GET /admin/v1/barrier/keys/status endpoint.
func TestE2E_GetBarrierKeysStatus(t *testing.T) {
	t.Parallel()

	testServers := setupRotationTestServers(t)

	// Step 1: Get initial status (root + intermediate keys auto-initialized).
	initialStatus := getBarrierKeysStatus(t, testServers.httpClient, testServers.adminURL)

	// Verify root_key fields.
	rootKey, ok := initialStatus["root_key"].(map[string]any)
	require.True(t, ok, "root_key should be object")

	rootKeyUUID, ok := rootKey["uuid"].(string)
	require.True(t, ok, "root_key uuid should be string")
	require.NotEmpty(t, rootKeyUUID, "root_key uuid should not be empty")

	rootKeyCreatedAt, ok := rootKey["created_at"].(float64)
	require.True(t, ok, "root_key created_at should be number")
	require.Greater(t, rootKeyCreatedAt, float64(0), "root_key created_at should be positive timestamp")

	rootKeyUpdatedAt, ok := rootKey["updated_at"].(float64)
	require.True(t, ok, "root_key updated_at should be number")
	require.Greater(t, rootKeyUpdatedAt, float64(0), "root_key updated_at should be positive timestamp")

	// Verify intermediate_key fields.
	intermediateKey, ok := initialStatus["intermediate_key"].(map[string]any)
	require.True(t, ok, "intermediate_key should be object")

	intermediateKeyUUID, ok := intermediateKey["uuid"].(string)
	require.True(t, ok, "intermediate_key uuid should be string")
	require.NotEmpty(t, intermediateKeyUUID, "intermediate_key uuid should not be empty")

	intermediateKeyCreatedAt, ok := intermediateKey["created_at"].(float64)
	require.True(t, ok, "intermediate_key created_at should be number")
	require.Greater(t, intermediateKeyCreatedAt, float64(0), "intermediate_key created_at should be positive timestamp")

	// Step 2: Rotate root key.
	rotateRootKey(t, testServers.httpClient, testServers.adminURL, "E2E test: verify status update after rotation")

	// Step 3: Get updated status.
	updatedStatus := getBarrierKeysStatus(t, testServers.httpClient, testServers.adminURL)

	// Verify root_key UUID changed.
	updatedRootKey, ok := updatedStatus["root_key"].(map[string]any)
	require.True(t, ok, "updated root_key should be object")

	updatedRootKeyUUID, ok := updatedRootKey["uuid"].(string)
	require.True(t, ok, "updated root_key uuid should be string")
	require.NotEqual(t, rootKeyUUID, updatedRootKeyUUID, "root_key UUID should change after rotation")

	// Verify intermediate_key UUID unchanged (only root key rotated).
	updatedIntermediateKey, ok := updatedStatus["intermediate_key"].(map[string]any)
	require.True(t, ok, "updated intermediate_key should be object")

	updatedIntermediateKeyUUID, ok := updatedIntermediateKey["uuid"].(string)
	require.True(t, ok, "updated intermediate_key uuid should be string")
	require.Equal(t, intermediateKeyUUID, updatedIntermediateKeyUUID, "intermediate_key UUID should remain unchanged after root rotation")
}

// rotateRootKey rotates root encryption key via admin API.
func rotateRootKey(t *testing.T, client *http.Client, adminURL, reason string) map[string]any {
	t.Helper()

	return rotateKey(t, client, adminURL, "/admin/v1/barrier/rotate/root", reason)
}

// rotateIntermediateKey rotates intermediate encryption key via admin API.
func rotateIntermediateKey(t *testing.T, client *http.Client, adminURL, reason string) map[string]any {
	t.Helper()

	return rotateKey(t, client, adminURL, "/admin/v1/barrier/rotate/intermediate", reason)
}

// rotateContentKey rotates content encryption key via admin API (elastic rotation).
func rotateContentKey(t *testing.T, client *http.Client, adminURL, reason string) map[string]any {
	t.Helper()

	return rotateKey(t, client, adminURL, "/admin/v1/barrier/rotate/content", reason)
}

// rotateKey is a helper function for rotation endpoints.
func rotateKey(t *testing.T, client *http.Client, adminURL, endpoint, reason string) map[string]any {
	t.Helper()

	reqBody := map[string]string{
		"reason": reason,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err, "failed to marshal rotation request")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, adminURL+endpoint, bytes.NewReader(reqJSON))
	require.NoError(t, err, "failed to create rotation request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err, "failed to send rotation request")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode, "rotation request should return 200 OK")

	var respBody map[string]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err, "failed to decode rotation response")

	return respBody
}

// getBarrierKeysStatus retrieves current barrier keys status via admin API.
func getBarrierKeysStatus(t *testing.T, client *http.Client, adminURL string) map[string]any {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, adminURL+"/admin/v1/barrier/keys/status", nil)
	require.NoError(t, err, "failed to create status request")

	resp, err := client.Do(req)
	require.NoError(t, err, "failed to send status request")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode, "status request should return 200 OK")

	var respBody map[string]any

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err, "failed to decode status response")

	return respBody
}
