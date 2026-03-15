// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	cryptoutilIdentityJobs "cryptoutil/internal/apps/identity/jobs"
)

// ServerManager manages the lifecycle of all identity servers.
type ServerManager struct {
	authzServer *AuthZServer
	idpServer   *IDPServer
	rsServer    *RSServer
	wg          sync.WaitGroup
	errChan     chan error
	cleanupJob  *cryptoutilIdentityJobs.CleanupJob
	logger      *slog.Logger
}

// NewServerManager creates a new server manager.
func NewServerManager(
	authzServer *AuthZServer,
	idpServer *IDPServer,
	rsServer *RSServer,
	cleanupJob *cryptoutilIdentityJobs.CleanupJob,
	logger *slog.Logger,
) *ServerManager {
	return &ServerManager{
		authzServer: authzServer,
		idpServer:   idpServer,
		rsServer:    rsServer,
		errChan:     make(chan error, 3),
		cleanupJob:  cleanupJob,
		logger:      logger,
	}
}

// Start starts all servers concurrently.
func (m *ServerManager) Start(ctx context.Context) error {
	// Start cleanup job.
	if m.cleanupJob != nil {
		m.logger.Info("Starting cleanup job")
		m.wg.Add(1)

		go func() {
			defer m.wg.Done()

			m.cleanupJob.Start(ctx)
		}()
	}

	// Start AuthZ server.
	if m.authzServer != nil {
		m.wg.Add(1)

		go func() {
			defer m.wg.Done()

			if err := m.authzServer.Start(ctx); err != nil {
				m.errChan <- fmt.Errorf("authz server error: %w", err)
			}
		}()
	}

	// Start IdP server.
	if m.idpServer != nil {
		m.wg.Add(1)

		go func() {
			defer m.wg.Done()

			if err := m.idpServer.Start(ctx); err != nil {
				m.errChan <- fmt.Errorf("idp server error: %w", err)
			}
		}()
	}

	// Start RS server.
	if m.rsServer != nil {
		m.wg.Add(1)

		go func() {
			defer m.wg.Done()

			if err := m.rsServer.Start(ctx); err != nil {
				m.errChan <- fmt.Errorf("rs server error: %w", err)
			}
		}()
	}

	// Wait for any server to fail.
	select {
	case err := <-m.errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("server manager context cancelled: %w", ctx.Err())
	}
}

// Stop stops all servers gracefully.
func (m *ServerManager) Stop(ctx context.Context) error {
	var errors []error

	// Stop cleanup job.
	if m.cleanupJob != nil {
		m.logger.Info("Stopping cleanup job")
		m.cleanupJob.Stop()
	}

	// Stop AuthZ server.
	if m.authzServer != nil {
		if err := m.authzServer.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("authz server stop error: %w", err))
		}
	}

	// Stop IdP server.
	if m.idpServer != nil {
		if err := m.idpServer.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("idp server stop error: %w", err))
		}
	}

	// Stop RS server.
	if m.rsServer != nil {
		if err := m.rsServer.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("rs server stop error: %w", err))
		}
	}

	// Wait for all servers to stop.
	m.wg.Wait()

	// Return combined errors.
	if len(errors) > 0 {
		return fmt.Errorf("server shutdown errors: %v", errors)
	}

	return nil
}

// GetCleanupMetrics returns current cleanup job metrics.
func (m *ServerManager) GetCleanupMetrics() cryptoutilIdentityJobs.CleanupJobMetrics {
	if m.cleanupJob == nil {
		return cryptoutilIdentityJobs.CleanupJobMetrics{}
	}

	return m.cleanupJob.GetMetrics()
}

// IsCleanupHealthy checks if the cleanup job is running successfully.
func (m *ServerManager) IsCleanupHealthy() bool {
	if m.cleanupJob == nil {
		return true // No cleanup job configured, consider healthy.
	}

	return m.cleanupJob.IsHealthy()
}
