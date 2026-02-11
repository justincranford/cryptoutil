// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// CleanupService manages background token and session cleanup operations.
type CleanupService struct {
	service  *Service
	interval time.Duration
	stopChan chan struct{}
	doneChan chan struct{}
}

// NewCleanupService creates a new cleanup service with specified interval.
func NewCleanupService(service *Service) *CleanupService {
	return &CleanupService{
		service:  service,
		interval: cryptoutilIdentityMagic.DefaultTokenCleanupInterval,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// Start begins the background cleanup process.
func (c *CleanupService) Start(ctx context.Context) {
	go c.cleanupLoop(ctx)
}

// Stop gracefully stops the cleanup service.
func (c *CleanupService) Stop() {
	close(c.stopChan)
	<-c.doneChan
}

// cleanupLoop runs periodic cleanup operations.
func (c *CleanupService) cleanupLoop(ctx context.Context) {
	defer close(c.doneChan)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.runCleanup(ctx)
		case <-c.stopChan:
			slog.InfoContext(ctx, "Cleanup service stopping")

			return
		}
	}
}

// runCleanup executes cleanup operations for expired tokens.
func (c *CleanupService) runCleanup(ctx context.Context) {
	tokenRepo := c.service.repoFactory.TokenRepository()

	slog.InfoContext(ctx, "Starting token cleanup")

	if err := tokenRepo.DeleteExpired(ctx); err != nil {
		slog.ErrorContext(ctx, "Token cleanup failed", "error", err)

		return
	}

	slog.InfoContext(ctx, "Token cleanup completed successfully")
}

// WithInterval sets a custom cleanup interval and returns the service.
func (c *CleanupService) WithInterval(interval time.Duration) *CleanupService {
	if interval <= 0 {
		interval = cryptoutilIdentityMagic.DefaultTokenCleanupInterval
	}

	c.interval = interval

	return c
}

// MigrateClientSecrets runs one-time migration to hash all plaintext client secrets.
func (s *Service) MigrateClientSecrets(ctx context.Context) error {
	clientRepo := s.repoFactory.ClientRepository()
	hasher := s.clientAuth.GetHasher()

	slog.InfoContext(ctx, "Starting client secret migration")

	migrated, err := hasher.MigrateSecrets(ctx, clientRepo)
	if err != nil {
		return fmt.Errorf("client secret migration failed: %w", err)
	}

	slog.InfoContext(ctx, "Client secret migration completed", "migrated_count", migrated)

	return nil
}
