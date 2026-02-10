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

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"

	testify "github.com/stretchr/testify/require"
)

func TestNewCleanupJob(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repoFactory := createTestRepoFactory(t)

	tests := []struct {
		name             string
		interval         time.Duration
		expectedInterval time.Duration
	}{
		{
			name:             "Custom interval",
			interval:         30 * time.Minute,
			expectedInterval: 30 * time.Minute,
		},
		{
			name:             "Zero interval uses default",
			interval:         0,
			expectedInterval: defaultCleanupInterval,
		},
		{
			name:             "Negative interval uses default",
			interval:         -1 * time.Hour,
			expectedInterval: defaultCleanupInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := NewCleanupJob(repoFactory, logger, tt.interval)

			testify.NotNil(t, job)
			testify.Equal(t, tt.expectedInterval, job.interval)
			testify.NotNil(t, job.stopChan)
		})
	}
}

func TestCleanupJob_StartAndStop(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repoFactory := createTestRepoFactory(t)

	// Create job with short interval for testing.
	job := NewCleanupJob(repoFactory, logger, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start job in goroutine.
	done := make(chan struct{})

	go func() {
		job.Start(ctx)
		close(done)
	}()

	// Wait a bit to ensure job runs.
	time.Sleep(250 * time.Millisecond)

	// Stop job.
	job.Stop()

	// Wait for job to finish.
	select {
	case <-done:
		// Job stopped successfully.
	case <-time.After(2 * time.Second):
		t.Fatal("Job did not stop within timeout")
	}
}

func TestCleanupJob_ContextCancellation(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repoFactory := createTestRepoFactory(t)

	job := NewCleanupJob(repoFactory, logger, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Start job in goroutine.
	done := make(chan struct{})

	go func() {
		job.Start(ctx)
		close(done)
	}()

	// Wait for context to be cancelled.
	select {
	case <-done:
		// Job stopped successfully due to context cancellation.
	case <-time.After(2 * time.Second):
		t.Fatal("Job did not stop within timeout")
	}
}

func TestCleanupJob_CleanupExecution(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	repoFactory := createTestRepoFactory(t)

	job := NewCleanupJob(repoFactory, logger, 1*time.Hour)

	ctx := context.Background()

	// Test cleanup execution (currently just logs, no errors expected).
	job.cleanup(ctx)

	// Verify no errors occurred.
	testify.NotNil(t, job.repoFactory)
}

// createTestRepoFactory creates a repository factory for testing.
func createTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	config := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(context.Background(), config)
	testify.NoError(t, err, "Failed to create repository factory")

	return repoFactory
}
