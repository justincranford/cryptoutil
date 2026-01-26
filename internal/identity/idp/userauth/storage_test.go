// Copyright (c) 2025 Justin Cranford
//
//

package userauth_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/identity/idp/userauth"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// InMemoryChallengeStore tests.

func TestInMemoryChallengeStore_NewStore(t *testing.T) {
	t.Parallel()

	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()
	require.NotNil(t, store, "NewInMemoryChallengeStore should return non-nil store")
}

func TestInMemoryChallengeStore_StoreAndRetrieve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()

	challengeID := googleUuid.Must(googleUuid.NewV7())
	testSecret := "test-secret-12345"
	challenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        challengeID,
		UserID:    googleUuid.NewString(),
		Method:    cryptoutilIdentityMagic.AuthMethodSMSOTP,
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}

	// Store the challenge.
	err := store.Store(ctx, challenge, testSecret)
	require.NoError(t, err, "Store should succeed")

	// Retrieve the challenge.
	retrieved, secret, err := store.Retrieve(ctx, challengeID)
	require.NoError(t, err, "Retrieve should succeed")
	require.NotNil(t, retrieved, "Retrieved challenge should not be nil")
	require.Equal(t, challengeID, retrieved.ID, "Challenge ID should match")
	require.Equal(t, testSecret, secret, "Secret should match")
}

func TestInMemoryChallengeStore_RetrieveNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()

	nonExistentID := googleUuid.Must(googleUuid.NewV7())

	_, _, err := store.Retrieve(ctx, nonExistentID)
	require.Error(t, err, "Retrieve should fail for non-existent challenge")
	require.Contains(t, err.Error(), "not found", "Error should mention not found")
}

func TestInMemoryChallengeStore_RetrieveExpired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()

	challengeID := googleUuid.Must(googleUuid.NewV7())
	testSecret := "test-secret-expired"
	challenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        challengeID,
		UserID:    googleUuid.NewString(),
		Method:    cryptoutilIdentityMagic.AuthMethodSMSOTP,
		ExpiresAt: time.Now().UTC().Add(-1 * time.Minute), // Already expired.
	}

	// Store the expired challenge.
	err := store.Store(ctx, challenge, testSecret)
	require.NoError(t, err, "Store should succeed even for expired challenge")

	// Retrieve should fail.
	_, _, err = store.Retrieve(ctx, challengeID)
	require.Error(t, err, "Retrieve should fail for expired challenge")
	require.Contains(t, err.Error(), "expired", "Error should mention expired")
}

func TestInMemoryChallengeStore_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()

	challengeID := googleUuid.Must(googleUuid.NewV7())
	testSecret := "test-secret-delete"
	challenge := &cryptoutilIdentityIdpUserauth.AuthChallenge{
		ID:        challengeID,
		UserID:    googleUuid.NewString(),
		Method:    cryptoutilIdentityMagic.AuthMethodSMSOTP,
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}

	// Store the challenge.
	err := store.Store(ctx, challenge, testSecret)
	require.NoError(t, err, "Store should succeed")

	// Delete the challenge.
	err = store.Delete(ctx, challengeID)
	require.NoError(t, err, "Delete should succeed")

	// Retrieve should fail after deletion.
	_, _, err = store.Retrieve(ctx, challengeID)
	require.Error(t, err, "Retrieve should fail after deletion")
}

func TestInMemoryChallengeStore_DeleteNonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()

	nonExistentID := googleUuid.Must(googleUuid.NewV7())

	// Delete should not error for non-existent challenge.
	err := store.Delete(ctx, nonExistentID)
	require.NoError(t, err, "Delete should succeed even for non-existent challenge")
}

func TestInMemoryChallengeStore_MultipleChallenges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := cryptoutilIdentityIdpUserauth.NewInMemoryChallengeStore()

	// Store multiple challenges.
	challenges := make([]*cryptoutilIdentityIdpUserauth.AuthChallenge, 3)
	secrets := make([]string, 3)

	for i := range 3 {
		challenges[i] = &cryptoutilIdentityIdpUserauth.AuthChallenge{
			ID:        googleUuid.Must(googleUuid.NewV7()),
			UserID:    googleUuid.NewString(),
			Method:    cryptoutilIdentityMagic.AuthMethodSMSOTP,
			ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
		}
		secrets[i] = googleUuid.NewString()

		err := store.Store(ctx, challenges[i], secrets[i])
		require.NoError(t, err, "Store should succeed for challenge %d", i)
	}

	// Verify all can be retrieved.
	for i := range 3 {
		retrieved, secret, err := store.Retrieve(ctx, challenges[i].ID)
		require.NoError(t, err, "Retrieve should succeed for challenge %d", i)
		require.Equal(t, challenges[i].ID, retrieved.ID, "ID should match for challenge %d", i)
		require.Equal(t, secrets[i], secret, "Secret should match for challenge %d", i)
	}
}

// InMemoryRateLimiter tests.

func TestInMemoryRateLimiter_NewLimiter(t *testing.T) {
	t.Parallel()

	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()
	require.NotNil(t, limiter, "NewInMemoryRateLimiter should return non-nil limiter")
}

func TestInMemoryRateLimiter_CheckLimitNoRecord(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	// Check limit for identifier with no record.
	err := limiter.CheckLimit(ctx, "new-identifier")
	require.NoError(t, err, "CheckLimit should succeed for new identifier")
}

func TestInMemoryRateLimiter_RecordFailedAttempt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	identifier := "test-user-failed"

	// Record a failed attempt.
	err := limiter.RecordAttempt(ctx, identifier, false)
	require.NoError(t, err, "RecordAttempt should succeed")

	// Check limit should still pass after one failure.
	err = limiter.CheckLimit(ctx, identifier)
	require.NoError(t, err, "CheckLimit should succeed after one failed attempt")
}

func TestInMemoryRateLimiter_RecordSuccessfulAttempt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	identifier := "test-user-success"

	// Record failed attempts.
	for range 2 {
		err := limiter.RecordAttempt(ctx, identifier, false)
		require.NoError(t, err, "RecordAttempt should succeed")
	}

	// Record a successful attempt (should reset counter).
	err := limiter.RecordAttempt(ctx, identifier, true)
	require.NoError(t, err, "RecordAttempt should succeed for success")

	// Check limit should pass after reset.
	err = limiter.CheckLimit(ctx, identifier)
	require.NoError(t, err, "CheckLimit should succeed after successful attempt reset")
}

func TestInMemoryRateLimiter_ExceedMaxAttempts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	identifier := "test-user-lockout"

	// Record max failed attempts (default is 5).
	for range 5 {
		err := limiter.RecordAttempt(ctx, identifier, false)
		require.NoError(t, err, "RecordAttempt should succeed")
	}

	// Check limit should now fail.
	err := limiter.CheckLimit(ctx, identifier)
	require.Error(t, err, "CheckLimit should fail after max attempts exceeded")
	require.Contains(t, err.Error(), "rate limit exceeded", "Error should mention rate limit")
}

func TestInMemoryRateLimiter_LockoutWithRemainingTime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	identifier := "test-user-lockout-time"

	// Exceed max attempts to trigger lockout.
	for range 5 {
		err := limiter.RecordAttempt(ctx, identifier, false)
		require.NoError(t, err, "RecordAttempt should succeed")
	}

	// Check limit should fail with remaining time info.
	err := limiter.CheckLimit(ctx, identifier)
	require.Error(t, err, "CheckLimit should fail during lockout")
	require.Contains(t, err.Error(), "try again in", "Error should include retry time")
}

func TestInMemoryRateLimiter_MultipleIdentifiers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	limiter := cryptoutilIdentityIdpUserauth.NewInMemoryRateLimiter()

	identifier1 := "user-1"
	identifier2 := "user-2"

	// Lock out identifier1.
	for range 5 {
		err := limiter.RecordAttempt(ctx, identifier1, false)
		require.NoError(t, err, "RecordAttempt should succeed for identifier1")
	}

	// identifier2 should still be allowed.
	err := limiter.CheckLimit(ctx, identifier2)
	require.NoError(t, err, "CheckLimit should succeed for identifier2")

	// identifier1 should be locked out.
	err = limiter.CheckLimit(ctx, identifier1)
	require.Error(t, err, "CheckLimit should fail for locked out identifier1")
}
