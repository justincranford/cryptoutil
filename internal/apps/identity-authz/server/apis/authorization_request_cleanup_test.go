// Copyright (c) 2025-2026 Justin Cranford.
package apis

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInMemoryAuthorizationRequestStore_CleanupExpired(t *testing.T) {
	t.Parallel()

	store := &InMemoryAuthorizationRequestStore{
		requests:         make(map[googleUuid.UUID]*AuthorizationRequest),
		codeIdx:          make(map[string]googleUuid.UUID),
		requestCodeIndex: make(map[googleUuid.UUID]string),
	}
	now := time.Now().UTC()
	expiredRequestID := googleUuid.Must(googleUuid.NewV7())
	currentRequestID := googleUuid.Must(googleUuid.NewV7())
	expiredUniqueCodeRequestID := googleUuid.Must(googleUuid.NewV7())
	activeRequestID := googleUuid.Must(googleUuid.NewV7())

	require.NoError(t, store.Store(context.Background(), &AuthorizationRequest{
		RequestID: expiredRequestID,
		Code:      testSharedAuthorizationCode,
		ExpiresAt: now.Add(-time.Minute),
	}))
	require.NoError(t, store.Store(context.Background(), &AuthorizationRequest{
		RequestID: currentRequestID,
		Code:      testSharedAuthorizationCode,
		ExpiresAt: now.Add(time.Hour),
	}))
	require.NoError(t, store.Store(context.Background(), &AuthorizationRequest{
		RequestID: expiredUniqueCodeRequestID,
		Code:      "expired-code",
		ExpiresAt: now.Add(-time.Minute),
	}))
	require.NoError(t, store.Store(context.Background(), &AuthorizationRequest{
		RequestID: activeRequestID,
		ExpiresAt: now.Add(time.Hour),
	}))

	store.cleanupExpired(now)

	require.NotContains(t, store.requests, expiredRequestID)
	require.Contains(t, store.requests, currentRequestID)
	require.NotContains(t, store.requests, expiredUniqueCodeRequestID)
	require.Contains(t, store.requests, activeRequestID)
	require.Equal(t, currentRequestID, store.codeIdx[testSharedAuthorizationCode])
	require.NotContains(t, store.codeIdx, "expired-code")
	require.NotContains(t, store.requestCodeIndex, expiredRequestID)
	require.NotContains(t, store.requestCodeIndex, expiredUniqueCodeRequestID)
}
