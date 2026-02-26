// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"bytes"
	"context"
	json "encoding/json"
	http "net/http"
	"testing"

	cryptoutilAppsSmImClient "cryptoutil/internal/apps/sm/im/client"
	cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	"github.com/stretchr/testify/require"
)

// TestE2E_RotateRootKey tests manual root key rotation via admin API.
func TestE2E_RotateRootKey(t *testing.T) {
	t.Parallel()

	// Step 1: Send baseline message before rotation.
	user1 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterServiceUser(t, sharedHTTPClient, publicBaseURL, "user1_rotate_root", *cryptoutilSharedUtilRandom.GeneratePassword(t, cryptoutilSharedMagic.DefaultCodeChallengeLength))
	user2 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterServiceUser(t, sharedHTTPClient, publicBaseURL, "user2_rotate_root", *cryptoutilSharedUtilRandom.GeneratePassword(t, cryptoutilSharedMagic.DefaultCodeChallengeLength))

	plaintext1 := "Message before root key rotation"

	messageID1, err := cryptoutilAppsSmImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext1, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID1, "baseline message ID should not be empty")

	// Step 2: Get initial barrier keys status.
	initialStatus := getBarrierKeysStatus(t, sharedHTTPClient, adminBaseURL)

	initialRootKeyUUID, ok := initialStatus["root_key"].(map[string]any)[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "initial root_key uuid should be string")
	require.NotEmpty(t, initialRootKeyUUID, "initial root_key uuid should not be empty")

	// Step 3: Rotate root key via admin API.
	rotationReason := "E2E test: manual root key rotation"

	rotateResponse := rotateKey(t, sharedHTTPClient, adminBaseURL, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+"/barrier/rotate/root", rotationReason)

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
	updatedStatus := getBarrierKeysStatus(t, sharedHTTPClient, adminBaseURL)

	updatedRootKeyUUID, ok := updatedStatus["root_key"].(map[string]any)[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "updated root_key uuid should be string")
	require.Equal(t, newKeyUUID, updatedRootKeyUUID, "status should reflect new root key UUID")

	// Step 5: Send new message after rotation (uses new root key chain).
	plaintext2 := "Message after root key rotation"

	messageID2, err := cryptoutilAppsSmImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext2, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID2, "post-rotation message ID should not be empty")

	// Step 6: Verify user2 can decrypt BOTH old and new messages (backward compatibility).
	messages, err := cryptoutilAppsSmImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
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

	// Step 1: Send baseline message before rotation.
	user1 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterServiceUser(t, sharedHTTPClient, publicBaseURL, "user1_rotate_intermediate", *cryptoutilSharedUtilRandom.GeneratePassword(t, cryptoutilSharedMagic.DefaultCodeChallengeLength))
	user2 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterServiceUser(t, sharedHTTPClient, publicBaseURL, "user2_rotate_intermediate", *cryptoutilSharedUtilRandom.GeneratePassword(t, cryptoutilSharedMagic.DefaultCodeChallengeLength))

	plaintext1 := "Message before intermediate key rotation"

	messageID1, err := cryptoutilAppsSmImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext1, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID1, "baseline message ID should not be empty")

	// Step 2: Get initial intermediate key status.
	initialStatus := getBarrierKeysStatus(t, sharedHTTPClient, adminBaseURL)

	initialIntermediateKeyUUID, ok := initialStatus["intermediate_key"].(map[string]any)[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "initial intermediate_key uuid should be string")
	require.NotEmpty(t, initialIntermediateKeyUUID, "initial intermediate_key uuid should not be empty")

	// Step 3: Rotate intermediate key via admin API.
	rotationReason := "E2E test: manual intermediate key rotation"

	rotateResponse := rotateKey(t, sharedHTTPClient, adminBaseURL, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+"/barrier/rotate/intermediate", rotationReason)

	oldKeyUUID, ok := rotateResponse["old_key_uuid"].(string)
	require.True(t, ok, "old_key_uuid should be string")
	require.Equal(t, initialIntermediateKeyUUID, oldKeyUUID, "old_key_uuid should match initial intermediate key")

	newKeyUUID, ok := rotateResponse["new_key_uuid"].(string)
	require.True(t, ok, "new_key_uuid should be string")
	require.NotEqual(t, oldKeyUUID, newKeyUUID, "new intermediate key should be different from old key")

	// Step 4: Verify status reflects new intermediate key.
	updatedStatus := getBarrierKeysStatus(t, sharedHTTPClient, adminBaseURL)

	updatedIntermediateKeyUUID, ok := updatedStatus["intermediate_key"].(map[string]any)[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "updated intermediate_key uuid should be string")
	require.Equal(t, newKeyUUID, updatedIntermediateKeyUUID, "status should reflect new intermediate key")

	// Step 5: Send new message after rotation.
	plaintext2 := "Message after intermediate key rotation"

	messageID2, err := cryptoutilAppsSmImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext2, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID2, "post-rotation message ID should not be empty")

	// Step 6: Verify backward compatibility.
	messages, err := cryptoutilAppsSmImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
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

	// Step 1: Send baseline message (creates first content key).
	user1 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterServiceUser(t, sharedHTTPClient, publicBaseURL, "user1_rotate_content", *cryptoutilSharedUtilRandom.GeneratePassword(t, cryptoutilSharedMagic.DefaultCodeChallengeLength))
	user2 := cryptoutilAppsTemplateServiceTestingE2eHelpers.RegisterServiceUser(t, sharedHTTPClient, publicBaseURL, "user2_rotate_content", *cryptoutilSharedUtilRandom.GeneratePassword(t, cryptoutilSharedMagic.DefaultCodeChallengeLength))

	plaintext1 := "Message before content key rotation"

	messageID1, err := cryptoutilAppsSmImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext1, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID1, "baseline message ID should not be empty")

	// Step 2: Rotate content key (elastic rotation - creates new key, keeps old).
	rotationReason := "E2E test: manual content key rotation"

	rotateResponse := rotateKey(t, sharedHTTPClient, adminBaseURL, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+"/barrier/rotate/content", rotationReason)

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

	messageID2, err := cryptoutilAppsSmImClient.SendMessage(sharedHTTPClient, publicBaseURL, plaintext2, user1.Token, user2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, messageID2, "post-rotation message ID should not be empty")

	// Step 4: Verify both messages decrypt correctly (elastic rotation preserves old keys).
	messages, err := cryptoutilAppsSmImClient.ReceiveMessagesService(sharedHTTPClient, publicBaseURL, user2.Token)
	require.NoError(t, err)
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

// TestE2E_GetBarrierKeysStatus tests GET /admin/api/v1/barrier/keys/status endpoint.
func TestE2E_GetBarrierKeysStatus(t *testing.T) {
	t.Parallel()

	// Step 1: Get initial status (root + intermediate keys auto-initialized).
	initialStatus := getBarrierKeysStatus(t, sharedHTTPClient, adminBaseURL)

	// Verify root_key fields.
	rootKey, ok := initialStatus["root_key"].(map[string]any)
	require.True(t, ok, "root_key should be object")

	rootKeyUUID, ok := rootKey[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "root_key uuid should be string")
	require.NotEmpty(t, rootKeyUUID, "root_key uuid should not be empty")

	rootKeyCreatedAt, ok := rootKey["created_at"].(float64)
	require.True(t, ok, "root_key created_at should be number")
	require.Greater(t, rootKeyCreatedAt, float64(0), "root_key created_at should be positive timestamp")

	rootKeyUpdatedAt, ok := rootKey[cryptoutilSharedMagic.ClaimUpdatedAt].(float64)
	require.True(t, ok, "root_key updated_at should be number")
	require.Greater(t, rootKeyUpdatedAt, float64(0), "root_key updated_at should be positive timestamp")

	// Verify intermediate_key fields.
	intermediateKey, ok := initialStatus["intermediate_key"].(map[string]any)
	require.True(t, ok, "intermediate_key should be object")

	intermediateKeyUUID, ok := intermediateKey[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "intermediate_key uuid should be string")
	require.NotEmpty(t, intermediateKeyUUID, "intermediate_key uuid should not be empty")

	intermediateKeyCreatedAt, ok := intermediateKey["created_at"].(float64)
	require.True(t, ok, "intermediate_key created_at should be number")
	require.Greater(t, intermediateKeyCreatedAt, float64(0), "intermediate_key created_at should be positive timestamp")

	// Step 2: Rotate root key.
	rotateKey(t, sharedHTTPClient, adminBaseURL, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+"/barrier/rotate/root", "E2E test: verify status update after rotation")

	// Step 3: Get updated status.
	updatedStatus := getBarrierKeysStatus(t, sharedHTTPClient, adminBaseURL)

	// Verify root_key UUID changed.
	updatedRootKey, ok := updatedStatus["root_key"].(map[string]any)
	require.True(t, ok, "updated root_key should be object")

	updatedRootKeyUUID, ok := updatedRootKey[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "updated root_key uuid should be string")
	require.NotEqual(t, rootKeyUUID, updatedRootKeyUUID, "root_key UUID should change after rotation")

	// Verify intermediate_key UUID unchanged (only root key rotated).
	updatedIntermediateKey, ok := updatedStatus["intermediate_key"].(map[string]any)
	require.True(t, ok, "updated intermediate_key should be object")

	updatedIntermediateKeyUUID, ok := updatedIntermediateKey[cryptoutilSharedMagic.IdentityTokenFormatUUID].(string)
	require.True(t, ok, "updated intermediate_key uuid should be string")
	require.Equal(t, intermediateKeyUUID, updatedIntermediateKeyUUID, "intermediate_key UUID should remain unchanged after root rotation")
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, adminURL+cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+"/barrier/keys/status", nil)
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
