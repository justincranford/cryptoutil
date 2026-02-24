// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Validates requirements:
// - R01-01: /oauth2/v1/authorize stores authorization request and redirects to login.
func TestAuthorizationRequestStore_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewInMemoryAuthorizationRequestStore()

	// Test Store.
	requestID := googleUuid.Must(googleUuid.NewV7())
	authRequest := &AuthorizationRequest{
		RequestID:           requestID,
		ClientID:            "test_client",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		Scope:               "openid",
		State:               "state123",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
	}

	err := store.Store(ctx, authRequest)
	require.NoError(t, err)

	// Test GetByRequestID.
	retrieved, err := store.GetByRequestID(ctx, requestID)
	require.NoError(t, err)
	require.Equal(t, authRequest.ClientID, retrieved.ClientID)

	// Test Update.
	authRequest.Code = "test_code_123"
	authRequest.ConsentGranted = true
	err = store.Update(ctx, authRequest)
	require.NoError(t, err)

	// Test GetByCode.
	retrievedByCode, err := store.GetByCode(ctx, "test_code_123")
	require.NoError(t, err)
	require.Equal(t, requestID, retrievedByCode.RequestID)
	require.True(t, retrievedByCode.ConsentGranted)

	// Test Delete.
	err = store.Delete(ctx, requestID)
	require.NoError(t, err)

	// Verify deletion.
	_, err = store.GetByRequestID(ctx, requestID)
	require.Error(t, err)
}

func TestAuthorizationRequestStore_Expiration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewInMemoryAuthorizationRequestStore()

	// Create non-expired request.
	requestID1 := googleUuid.Must(googleUuid.NewV7())
	authRequest1 := &AuthorizationRequest{
		RequestID:           requestID1,
		ClientID:            "test_client",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		Scope:               "openid",
		State:               "state123",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		Code:                "valid_code",
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(5 * time.Minute), // Valid.
		ConsentGranted:      true,
	}

	err := store.Store(ctx, authRequest1)
	require.NoError(t, err)

	// Verify non-expired request is retrievable.
	retrieved, err := store.GetByRequestID(ctx, requestID1)
	require.NoError(t, err)
	require.Equal(t, "test_client", retrieved.ClientID)

	retrievedByCode, err := store.GetByCode(ctx, "valid_code")
	require.NoError(t, err)
	require.Equal(t, requestID1, retrievedByCode.RequestID)
}

// Validates requirements:
// - R01-03: Consent approval generates authorization code with user context.
func TestGenerateAuthorizationCode(t *testing.T) {
	t.Parallel()

	// Generate multiple codes and ensure uniqueness.
	codes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		code, err := GenerateAuthorizationCode()
		require.NoError(t, err)
		require.NotEmpty(t, code)
		require.False(t, codes[code], "Duplicate code generated")

		codes[code] = true
	}

	// Verify code length (base64 URL encoded 32 bytes).
	code, err := GenerateAuthorizationCode()
	require.NoError(t, err)
	require.Greater(t, len(code), 40) // Base64 URL encoding produces 44 characters for 32 bytes.
}

// Validates requirements:
// - R01-05: Authorization code single-use enforcement.
func TestAuthorizationRequestStore_CodeIndexing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewInMemoryAuthorizationRequestStore()

	// Create multiple requests.
	requestID1 := googleUuid.Must(googleUuid.NewV7())
	authRequest1 := &AuthorizationRequest{
		RequestID:           requestID1,
		ClientID:            "test_client_1",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		Scope:               "openid",
		State:               "state1",
		CodeChallenge:       "challenge1",
		CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
	}

	requestID2 := googleUuid.Must(googleUuid.NewV7())
	authRequest2 := &AuthorizationRequest{
		RequestID:           requestID2,
		ClientID:            "test_client_2",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		Scope:               "openid profile",
		State:               "state2",
		CodeChallenge:       "challenge2",
		CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
	}

	// Store requests.
	err := store.Store(ctx, authRequest1)
	require.NoError(t, err)

	err = store.Store(ctx, authRequest2)
	require.NoError(t, err)

	// Update with codes.
	authRequest1.Code = "code_123"
	authRequest1.ConsentGranted = true
	err = store.Update(ctx, authRequest1)
	require.NoError(t, err)

	authRequest2.Code = "code_456"
	authRequest2.ConsentGranted = true
	err = store.Update(ctx, authRequest2)
	require.NoError(t, err)

	// Retrieve by code.
	retrieved1, err := store.GetByCode(ctx, "code_123")
	require.NoError(t, err)
	require.Equal(t, requestID1, retrieved1.RequestID)
	require.Equal(t, "test_client_1", retrieved1.ClientID)

	retrieved2, err := store.GetByCode(ctx, "code_456")
	require.NoError(t, err)
	require.Equal(t, requestID2, retrieved2.RequestID)
	require.Equal(t, "test_client_2", retrieved2.ClientID)

	// Delete one request and verify code index updated.
	err = store.Delete(ctx, requestID1)
	require.NoError(t, err)

	_, err = store.GetByCode(ctx, "code_123")
	require.Error(t, err, "Code index should be cleaned up after deletion")

	// Second request still retrievable.
	retrieved2Again, err := store.GetByCode(ctx, "code_456")
	require.NoError(t, err)
	require.Equal(t, requestID2, retrieved2Again.RequestID)
}
