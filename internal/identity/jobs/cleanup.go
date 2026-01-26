// Copyright (c) 2025 Justin Cranford
//
//

// Package jobs provides background job scheduling and execution for the identity service.
package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

const (
	defaultCleanupInterval = 1 * time.Hour
)

// CleanupJobMetrics tracks cleanup job execution metrics.
type CleanupJobMetrics struct {
	LastRunTime        time.Time // Last time the cleanup job ran successfully.
	TokensDeleted      int       // Total tokens deleted by last run.
	SessionsDeleted    int       // Total sessions deleted by last run.
	ErrorCount         int       // Number of errors encountered.
	LastError          error     // Last error encountered.
	TotalRunCount      int       // Total number of successful runs.
	TotalTokensDeleted int       // Cumulative tokens deleted.
	TotalSessionsDel   int       // Cumulative sessions deleted.
}

// CleanupJob handles periodic cleanup of expired tokens and sessions.
type CleanupJob struct {
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	logger      *slog.Logger
	interval    time.Duration
	stopChan    chan struct{}
	metrics     *CleanupJobMetrics
}

// NewCleanupJob creates a new cleanup job.
func NewCleanupJob(repoFactory *cryptoutilIdentityRepository.RepositoryFactory, logger *slog.Logger, interval time.Duration) *CleanupJob {
	if interval <= 0 {
		interval = defaultCleanupInterval
	}

	return &CleanupJob{
		repoFactory: repoFactory,
		logger:      logger,
		interval:    interval,
		stopChan:    make(chan struct{}),
		metrics:     &CleanupJobMetrics{},
	}
}

// Start begins the cleanup job loop.
func (j *CleanupJob) Start(ctx context.Context) {
	j.logger.Info("Starting cleanup job", "interval", j.interval)

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Run cleanup immediately on start.
	j.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("Cleanup job stopped due to context cancellation")

			return
		case <-j.stopChan:
			j.logger.Info("Cleanup job stopped")

			return
		case <-ticker.C:
			j.cleanup(ctx)
		}
	}
}

// Stop stops the cleanup job.
func (j *CleanupJob) Stop() {
	close(j.stopChan)
}

// cleanup performs the actual cleanup work.
func (j *CleanupJob) cleanup(ctx context.Context) {
	j.logger.Debug("Running cleanup tasks")

	now := time.Now().UTC()

	// Cleanup expired tokens.
	tokensDeleted, err := j.cleanupExpiredTokens(ctx, now)
	if err != nil {
		j.logger.Error("Failed to cleanup expired tokens", "error", err)

		j.metrics.ErrorCount++
		j.metrics.LastError = fmt.Errorf("token cleanup failed: %w", err)

		return
	}

	// Cleanup expired sessions.
	sessionsDeleted, err := j.cleanupExpiredSessions(ctx, now)
	if err != nil {
		j.logger.Error("Failed to cleanup expired sessions", "error", err)

		j.metrics.ErrorCount++
		j.metrics.LastError = fmt.Errorf("session cleanup failed: %w", err)

		return
	}

	// Update metrics on successful cleanup.
	j.metrics.LastRunTime = now
	j.metrics.TokensDeleted = tokensDeleted
	j.metrics.SessionsDeleted = sessionsDeleted
	j.metrics.TotalRunCount++
	j.metrics.TotalTokensDeleted += tokensDeleted
	j.metrics.TotalSessionsDel += sessionsDeleted

	j.logger.Debug("Cleanup tasks completed", "tokens_deleted", tokensDeleted, "sessions_deleted", sessionsDeleted)
}

// cleanupExpiredTokens removes expired access tokens from the database.
func (j *CleanupJob) cleanupExpiredTokens(ctx context.Context, beforeTime time.Time) (int, error) {
	tokenRepo := j.repoFactory.TokenRepository()

	j.logger.Debug("Cleaning up expired tokens", "before_time", beforeTime)

	// Delete all tokens expired before the given time.
	deletedCount, err := tokenRepo.DeleteExpiredBefore(ctx, beforeTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return deletedCount, nil
}

// cleanupExpiredSessions removes expired sessions from the database.
func (j *CleanupJob) cleanupExpiredSessions(ctx context.Context, beforeTime time.Time) (int, error) {
	sessionRepo := j.repoFactory.SessionRepository()

	j.logger.Debug("Cleaning up expired sessions", "before_time", beforeTime)

	// Delete all sessions expired before the given time.
	deletedCount, err := sessionRepo.DeleteExpiredBefore(ctx, beforeTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return deletedCount, nil
}

// GetMetrics returns the current cleanup job metrics.
func (j *CleanupJob) GetMetrics() CleanupJobMetrics {
	return *j.metrics
}

// IsHealthy checks if the cleanup job is running successfully.
func (j *CleanupJob) IsHealthy() bool {
	// Job is healthy if last run was within 2x the interval.
	maxAge := j.interval * 2
	timeSinceLastRun := time.Since(j.metrics.LastRunTime)

	return timeSinceLastRun < maxAge && j.metrics.LastError == nil
}
