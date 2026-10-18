// Copyright (c) 2025-2026 Justin Cranford.
//

package apis

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInMemoryAuthorizationRequestStore_GetByCodeMissingIndexedRequest(t *testing.T) {
	t.Parallel()

	store := NewInMemoryAuthorizationRequestStore()
	requestID := googleUuid.Must(googleUuid.NewV7())
	require.NoError(t, store.Store(context.Background(), &AuthorizationRequest{
		RequestID: requestID,
		Code:      "indexed-code",
		ExpiresAt: time.Now().UTC().Add(time.Minute),
	}))

	store.mu.Lock()
	delete(store.requests, requestID)
	store.mu.Unlock()

	_, err := store.GetByCode(context.Background(), "indexed-code")
	require.EqualError(t, err, "authorization request not found")
}
