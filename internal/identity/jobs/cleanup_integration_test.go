// Copyright (c) 2025 Justin Cranford
//
//

package jobs

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"

	testify "github.com/stretchr/testify/require"
)

func TestCleanupJob_Integration_TokenDeletion(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	repoFactory := createTestRepoFactory(t)

	ctx := context.Background()

	// Create expired token (expires in the past).
	tokenRepo := repoFactory.TokenRepository()
	expiredToken := &cryptoutilIdentityDomain.Token{
		TokenValue:   googleUuid.NewString(),
		TokenType:    cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:  cryptoutilIdentityDomain.TokenFormatUUID,
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired 1 hour ago.
		IssuedAt:     time.Now().Add(-2 * time.Hour),
		Scopes:       []string{"read", "write"},
		ClientID:     googleUuid.Must(googleUuid.NewV7()),
		UserID:       cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		CodeChallenge: "",
	}

	err := tokenRepo.Create(ctx, expiredToken)
	testify.NoError(t, err, "Failed to create expired token")

	// Create non-expired token (expires in the future).
	validToken := &cryptoutilIdentityDomain.Token{
		TokenValue:   googleUuid.NewString(),
		TokenType:    cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat:  cryptoutilIdentityDomain.TokenFormatUUID,
		ExpiresAt:    time.Now().Add(1 * time.Hour), // Expires 1 hour from now.
		IssuedAt:     time.Now(),
		Scopes:       []string{"read", "write"},
		ClientID:     googleUuid.Must(googleUuid.NewV7()),
		UserID:       cryptoutilIdentityDomain.NullableUUID{UUID: googleUuid.Must(googleUuid.NewV7()), Valid: true},
		CodeChallenge: "",
	}

	err = tokenRepo.Create(ctx, validToken)
	testify.NoError(t, err, "Failed to create valid token")

	// Create cleanup job with very short interval for testing.
	job := NewCleanupJob(repoFactory, logger, 100*time.Millisecond)

	// Run cleanup once.
	job.cleanup(ctx)

	// Verify expired token was deleted.
	deletedToken, err := tokenRepo.GetByID(ctx, expiredToken.ID)
	testify.Error(t, err, "Expired token should be deleted")
	testify.Nil(t, deletedToken, "Expired token should not exist")

	// Verify valid token still exists.
	retrievedToken, err := tokenRepo.GetByID(ctx, validToken.ID)
	testify.NoError(t, err, "Valid token should still exist")
	testify.NotNil(t, retrievedToken, "Valid token should be retrievable")
	testify.Equal(t, validToken.ID, retrievedToken.ID)

	// Verify metrics.
	metrics := job.GetMetrics()
	testify.Equal(t, 1, metrics.TokensDeleted, "Expected 1 token to be deleted")
	testify.Equal(t, 0, metrics.ErrorCount, "Expected no errors")
	testify.Nil(t, metrics.LastError, "Expected no last error")
}

func TestCleanupJob_Integration_SessionDeletion(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	repoFactory := createTestRepoFactory(t)

	ctx := context.Background()

	// Create expired session (expires in the past).
	sessionRepo := repoFactory.SessionRepository()
	expiredSession := &cryptoutilIdentityDomain.Session{
		SessionID:          googleUuid.NewString(),
		UserID:             googleUuid.Must(googleUuid.NewV7()),
		IssuedAt:           time.Now().Add(-2 * time.Hour),
		LastSeenAt:         time.Now().Add(-2 * time.Hour),
		ExpiresAt:          time.Now().Add(-1 * time.Hour), // Expired 1 hour ago.
		IPAddress:          "192.168.1.100",
		UserAgent:          "Mozilla/5.0",
		AuthenticationTime: time.Now().Add(-2 * time.Hour),
		Active:             true,
	}

	err := sessionRepo.Create(ctx, expiredSession)
	testify.NoError(t, err, "Failed to create expired session")

	// Create non-expired session (expires in the future).
	validSession := &cryptoutilIdentityDomain.Session{
		SessionID:          googleUuid.NewString(),
		UserID:             googleUuid.Must(googleUuid.NewV7()),
		IssuedAt:           time.Now(),
		LastSeenAt:         time.Now(),
		ExpiresAt:          time.Now().Add(1 * time.Hour), // Expires 1 hour from now.
		IPAddress:          "192.168.1.101",
		UserAgent:          "Mozilla/5.0",
		AuthenticationTime: time.Now(),
		Active:             true,
	}

	err = sessionRepo.Create(ctx, validSession)
	testify.NoError(t, err, "Failed to create valid session")

	// Create cleanup job with very short interval for testing.
	job := NewCleanupJob(repoFactory, logger, 100*time.Millisecond)

	// Run cleanup once.
	job.cleanup(ctx)

	// Verify expired session was deleted.
	deletedSession, err := sessionRepo.GetBySessionID(ctx, expiredSession.SessionID)
	testify.Error(t, err, "Expired session should be deleted")
	testify.Nil(t, deletedSession, "Expired session should not exist")

	// Verify valid session still exists.
	retrievedSession, err := sessionRepo.GetBySessionID(ctx, validSession.SessionID)
	testify.NoError(t, err, "Valid session should still exist")
	testify.NotNil(t, retrievedSession, "Valid session should be retrievable")
	testify.Equal(t, validSession.SessionID, retrievedSession.SessionID)

	// Verify metrics.
	metrics := job.GetMetrics()
	testify.Equal(t, 1, metrics.SessionsDeleted, "Expected 1 session to be deleted")
	testify.Equal(t, 0, metrics.ErrorCount, "Expected no errors")
	testify.Nil(t, metrics.LastError, "Expected no last error")
}

func TestCleanupJob_Integration_ScheduledExecution(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	repoFactory := createTestRepoFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Create cleanup job with very short interval.
	job := NewCleanupJob(repoFactory, logger, 200*time.Millisecond)

	// Start job in goroutine.
	done := make(chan struct{})

	go func() {
		job.Start(ctx)
		close(done)
	}()

	// Wait for job to run at least 2 times (initial + 1 scheduled).
	time.Sleep(500 * time.Millisecond)

	// Stop job.
	job.Stop()

	// Wait for job to finish.
	select {
	case <-done:
		// Job stopped successfully.
	case <-time.After(2 * time.Second):
		t.Fatal("Job did not stop within timeout")
	}

	// Verify metrics show multiple runs.
	metrics := job.GetMetrics()
	testify.GreaterOrEqual(t, metrics.TotalRunCount, 2, "Expected at least 2 cleanup runs")
	testify.Equal(t, 0, metrics.ErrorCount, "Expected no errors during scheduled execution")
}

func TestCleanupJob_Integration_HealthCheck(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repoFactory := createTestRepoFactory(t)

	ctx := context.Background()

	// Create cleanup job with 1 hour interval.
	job := NewCleanupJob(repoFactory, logger, 1*time.Hour)

	// Before first run, job is unhealthy (LastRunTime is zero).
	testify.False(t, job.IsHealthy(), "Job should be unhealthy before first run")

	// Run cleanup once.
	job.cleanup(ctx)

	// After successful run, job is healthy.
	testify.True(t, job.IsHealthy(), "Job should be healthy after successful run")

	// Verify metrics.
	metrics := job.GetMetrics()
	testify.Equal(t, 1, metrics.TotalRunCount, "Expected 1 cleanup run")
	testify.Equal(t, 0, metrics.ErrorCount, "Expected no errors")
	testify.True(t, metrics.LastRunTime.After(time.Time{}), "LastRunTime should be set")
}
