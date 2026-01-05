// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"net/http"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cipherClient "cryptoutil/internal/cipher/client"
)

// sendMessage sends a message to one or more receivers.
func sendMessage(t *testing.T, sharedHTTPClient *http.Client, baseURL, message, token string, receiverIDs ...googleUuid.UUID) string {
	t.Helper()

	messageID, err := cipherClient.SendMessage(sharedHTTPClient, baseURL, message, token, receiverIDs...)
	require.NoError(t, err)
	return messageID
}
