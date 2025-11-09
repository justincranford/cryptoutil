package tests

import (
	"context"
	"errors"
	"testing"

	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Register CGO-free SQLite driver
)

func TestTransactionRollback(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()
	repoFactory := setupTestRepositoryFactory(t, ctx)
	defer repoFactory.Close()

	// Test transaction rollback on error
	err := repoFactory.Transaction(ctx, func(ctx context.Context) error {
		// Create a user within transaction
		user := &cryptoutilIdentityDomain.User{
			Sub:   "rollback-test-user",
			Email: "rollback@example.com",
			Name:  "Rollback Test User",
		}

		userRepo := repoFactory.UserRepository()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create a client within transaction
		client := &cryptoutilIdentityDomain.Client{
			ClientID:   "rollback-test-client",
			ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:       "Rollback Test Client",
		}

		clientRepo := repoFactory.ClientRepository()
		err = clientRepo.Create(ctx, client)
		require.NoError(t, err)

		// Simulate an error to trigger rollback
		return errors.New("simulated error for rollback test")

		// The transaction should rollback, so neither user nor client should exist
	})

	// Transaction should have failed
	require.Error(t, err)

	// Verify rollback - neither entity should exist
	userRepo := repoFactory.UserRepository()
	_, err = userRepo.GetBySub(ctx, "rollback-test-user")
	require.Error(t, err) // Should not find the user

	clientRepo := repoFactory.ClientRepository()
	_, err = clientRepo.GetByClientID(ctx, "rollback-test-client")
	require.Error(t, err) // Should not find the client
}

func TestTransactionCommit(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()
	repoFactory := setupTestRepositoryFactory(t, ctx)
	defer repoFactory.Close()

	// Test successful transaction commit
	err := repoFactory.Transaction(ctx, func(ctx context.Context) error {
		// Create a user within transaction
		user := &cryptoutilIdentityDomain.User{
			Sub:   "commit-test-user",
			Email: "commit@example.com",
			Name:  "Commit Test User",
		}

		userRepo := repoFactory.UserRepository()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Create a client within transaction
		client := &cryptoutilIdentityDomain.Client{
			ClientID:   "commit-test-client",
			ClientType: cryptoutilIdentityDomain.ClientTypeConfidential,
			Name:       "Commit Test Client",
		}

		clientRepo := repoFactory.ClientRepository()
		err = clientRepo.Create(ctx, client)
		require.NoError(t, err)

		return nil // Success - transaction should commit
	})

	// Transaction should have succeeded
	require.NoError(t, err)

	// Verify commit - both entities should exist
	userRepo := repoFactory.UserRepository()
	user, err := userRepo.GetBySub(ctx, "commit-test-user")
	require.NoError(t, err)
	require.Equal(t, "commit-test-user", user.Sub)

	clientRepo := repoFactory.ClientRepository()
	client, err := clientRepo.GetByClientID(ctx, "commit-test-client")
	require.NoError(t, err)
	require.Equal(t, "commit-test-client", client.ClientID)
}

func TestTransactionIsolation(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()
	repoFactory := setupTestRepositoryFactory(t, ctx)
	defer repoFactory.Close()

	// Create a user outside transaction
	externalUser := &cryptoutilIdentityDomain.User{
		Sub:   "external-user",
		Email: "external@example.com",
		Name:  "External User",
	}

	userRepo := repoFactory.UserRepository()
	err := userRepo.Create(ctx, externalUser)
	require.NoError(t, err)

	// Test transaction isolation - changes inside transaction shouldn't be visible outside until commit
	err = repoFactory.Transaction(ctx, func(ctx context.Context) error {
		// Create a user within transaction
		user := &cryptoutilIdentityDomain.User{
			Sub:   "isolated-user",
			Email: "isolated@example.com",
			Name:  "Isolated User",
		}

		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Verify the external user is visible within transaction
		_, err = userRepo.GetByID(ctx, externalUser.ID)
		require.NoError(t, err)

		return nil
	})

	require.NoError(t, err)

	// After commit, the transaction user should be visible
	_, err = userRepo.GetBySub(ctx, "isolated-user")
	require.NoError(t, err)
}

func TestConcurrentTransactions(t *testing.T) {
	t.Parallel()

	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()
	repoFactory := setupTestRepositoryFactory(t, ctx)
	defer repoFactory.Close()

	// Test concurrent transactions
	done := make(chan bool, 2)

	// Transaction 1
	go func() {
		err := repoFactory.Transaction(ctx, func(ctx context.Context) error {
			user := &cryptoutilIdentityDomain.User{
				Sub:   "concurrent-user-1",
				Email: "concurrent1@example.com",
				Name:  "Concurrent User 1",
			}

			userRepo := repoFactory.UserRepository()
			return userRepo.Create(ctx, user)
		})
		require.NoError(t, err)
		done <- true
	}()

	// Transaction 2
	go func() {
		err := repoFactory.Transaction(ctx, func(ctx context.Context) error {
			user := &cryptoutilIdentityDomain.User{
				Sub:   "concurrent-user-2",
				Email: "concurrent2@example.com",
				Name:  "Concurrent User 2",
			}

			userRepo := repoFactory.UserRepository()
			return userRepo.Create(ctx, user)
		})
		require.NoError(t, err)
		done <- true
	}()

	// Wait for both transactions to complete
	<-done
	<-done

	// Verify both users were created
	userRepo := repoFactory.UserRepository()
	user1, err := userRepo.GetBySub(ctx, "concurrent-user-1")
	require.NoError(t, err)
	require.Equal(t, "concurrent-user-1", user1.Sub)

	user2, err := userRepo.GetBySub(ctx, "concurrent-user-2")
	require.NoError(t, err)
	require.Equal(t, "concurrent-user-2", user2.Sub)
}