package jobs

import (
	"context"
	"log/slog"
	"time"

	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

const (
	defaultCleanupInterval = 1 * time.Hour
	defaultTokenExpiration = 24 * time.Hour
)

// CleanupJob handles periodic cleanup of expired tokens and sessions.
type CleanupJob struct {
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	logger      *slog.Logger
	interval    time.Duration
	stopChan    chan struct{}
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

	// Cleanup expired tokens.
	if err := j.cleanupExpiredTokens(ctx); err != nil {
		j.logger.Error("Failed to cleanup expired tokens", "error", err)
	}

	// Cleanup expired sessions.
	if err := j.cleanupExpiredSessions(ctx); err != nil {
		j.logger.Error("Failed to cleanup expired sessions", "error", err)
	}

	j.logger.Debug("Cleanup tasks completed")
}

// cleanupExpiredTokens removes expired access tokens from the database.
func (j *CleanupJob) cleanupExpiredTokens(ctx context.Context) error {
	tokenRepo := j.repoFactory.TokenRepository()

	// Calculate expiration cutoff time.
	cutoffTime := time.Now().Add(-defaultTokenExpiration)

	j.logger.Debug("Cleaning up expired tokens", "cutoff_time", cutoffTime)

	// In a real implementation, this would call a repository method to delete expired tokens.
	// For now, we log the operation.
	_ = tokenRepo
	_ = cutoffTime

	// TODO: Implement actual token cleanup when TokenRepository has DeleteExpiredBefore method.
	// return tokenRepo.DeleteExpiredBefore(ctx, cutoffTime)

	return nil
}

// cleanupExpiredSessions removes expired sessions from the database.
func (j *CleanupJob) cleanupExpiredSessions(ctx context.Context) error {
	sessionRepo := j.repoFactory.SessionRepository()

	// Calculate expiration cutoff time.
	cutoffTime := time.Now().Add(-defaultTokenExpiration)

	j.logger.Debug("Cleaning up expired sessions", "cutoff_time", cutoffTime)

	// In a real implementation, this would call a repository method to delete expired sessions.
	// For now, we log the operation.
	_ = sessionRepo
	_ = cutoffTime

	// TODO: Implement actual session cleanup when SessionRepository has DeleteExpiredBefore method.
	// return sessionRepo.DeleteExpiredBefore(ctx, cutoffTime)

	return nil
}
