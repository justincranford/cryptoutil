// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
)

// TestInMemoryAuthorizationRequestStore_StoreAndGet validates basic store and retrieval.
func TestInMemoryAuthorizationRequestStore_StoreAndGet(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	requestID := googleUuid.New()
	request := &cryptoutilIdentityAuthz.AuthorizationRequest{
		RequestID:    requestID,
		ClientID:     "test-client",
		RedirectURI:  "https://example.com/callback",
		ResponseType: "code",
		Scope:        "openid profile",
		State:        "state123",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}

	err := store.Store(ctx, request)
	require.NoError(t, err, "Store should succeed")

	retrieved, err := store.GetByRequestID(ctx, requestID)
	require.NoError(t, err, "GetByRequestID should succeed")
	require.Equal(t, request.RequestID, retrieved.RequestID, "Retrieved request should match stored")
	require.Equal(t, request.ClientID, retrieved.ClientID, "Client ID should match")
}

// TestInMemoryAuthorizationRequestStore_GetByCode validates code-based retrieval.
func TestInMemoryAuthorizationRequestStore_GetByCode(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	requestID := googleUuid.New()
	code := "auth-code-12345"
	request := &cryptoutilIdentityAuthz.AuthorizationRequest{
		RequestID:    requestID,
		ClientID:     "test-client",
		RedirectURI:  "https://example.com/callback",
		ResponseType: "code",
		Code:         code,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}

	err := store.Store(ctx, request)
	require.NoError(t, err, "Store should succeed")

	retrieved, err := store.GetByCode(ctx, code)
	require.NoError(t, err, "GetByCode should succeed")
	require.Equal(t, request.RequestID, retrieved.RequestID, "Retrieved request should match stored")
	require.Equal(t, code, retrieved.Code, "Authorization code should match")
}

// TestInMemoryAuthorizationRequestStore_Update validates request updates.
func TestInMemoryAuthorizationRequestStore_Update(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	requestID := googleUuid.New()
	request := &cryptoutilIdentityAuthz.AuthorizationRequest{
		RequestID:    requestID,
		ClientID:     "test-client",
		RedirectURI:  "https://example.com/callback",
		ResponseType: "code",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}

	err := store.Store(ctx, request)
	require.NoError(t, err, "Store should succeed")

	// Update with code.
	request.Code = "updated-code-123"
	err = store.Update(ctx, request)
	require.NoError(t, err, "Update should succeed")

	retrieved, err := store.GetByCode(ctx, "updated-code-123")
	require.NoError(t, err, "GetByCode should succeed after update")
	require.Equal(t, request.RequestID, retrieved.RequestID, "Retrieved request should match updated")
}

// TestInMemoryAuthorizationRequestStore_Delete validates request deletion.
func TestInMemoryAuthorizationRequestStore_Delete(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	requestID := googleUuid.New()
	code := "delete-test-code"
	request := &cryptoutilIdentityAuthz.AuthorizationRequest{
		RequestID:    requestID,
		ClientID:     "test-client",
		RedirectURI:  "https://example.com/callback",
		ResponseType: "code",
		Code:         code,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}

	err := store.Store(ctx, request)
	require.NoError(t, err, "Store should succeed")

	err = store.Delete(ctx, requestID)
	require.NoError(t, err, "Delete should succeed")

	_, err = store.GetByRequestID(ctx, requestID)
	require.Error(t, err, "GetByRequestID should fail after deletion")

	_, err = store.GetByCode(ctx, code)
	require.Error(t, err, "GetByCode should fail after deletion")
}

// TestInMemoryAuthorizationRequestStore_ExpiredRequest validates expiration handling.
func TestInMemoryAuthorizationRequestStore_ExpiredRequest(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	requestID := googleUuid.New()
	request := &cryptoutilIdentityAuthz.AuthorizationRequest{
		RequestID:    requestID,
		ClientID:     "test-client",
		RedirectURI:  "https://example.com/callback",
		ResponseType: "code",
		CreatedAt:    time.Now().UTC().Add(-20 * time.Minute),
		ExpiresAt:    time.Now().UTC().Add(-10 * time.Minute), // Expired.
	}

	err := store.Store(ctx, request)
	require.NoError(t, err, "Store should succeed")

	_, err = store.GetByRequestID(ctx, requestID)
	require.Error(t, err, "GetByRequestID should fail for expired request")
	require.Contains(t, err.Error(), "expired", "Error should indicate expiration")
}

// TestInMemoryAuthorizationRequestStore_NotFound validates missing request handling.
func TestInMemoryAuthorizationRequestStore_NotFound(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	nonExistentID := googleUuid.New()

	_, err := store.GetByRequestID(ctx, nonExistentID)
	require.Error(t, err, "GetByRequestID should fail for non-existent request")
	require.Contains(t, err.Error(), "not found", "Error should indicate not found")

	_, err = store.GetByCode(ctx, "non-existent-code")
	require.Error(t, err, "GetByCode should fail for non-existent code")
	require.Contains(t, err.Error(), "not found", "Error should indicate not found")
}

// TestInMemoryAuthorizationRequestStore_UpdateNonExistent validates update failure.
func TestInMemoryAuthorizationRequestStore_UpdateNonExistent(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityAuthz.NewInMemoryAuthorizationRequestStore()
	ctx := context.Background()

	nonExistentRequest := &cryptoutilIdentityAuthz.AuthorizationRequest{
		RequestID:    googleUuid.New(),
		ClientID:     "test-client",
		RedirectURI:  "https://example.com/callback",
		ResponseType: "code",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}

	err := store.Update(ctx, nonExistentRequest)
	require.Error(t, err, "Update should fail for non-existent request")
	require.Contains(t, err.Error(), "not found", "Error should indicate not found")
}
