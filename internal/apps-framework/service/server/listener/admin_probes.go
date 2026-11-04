// Copyright (c) 2025-2026 Justin Cranford.
// Package listener provides a reusable template for dual HTTPS server pattern used across all cryptoutil services.
package listener

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func (s *AdminServer) handleStartupz(c *fiber.Ctx) error {
	s.mu.RLock()
	startupComplete := s.startupComplete
	s.mu.RUnlock()

	if !startupComplete {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: "starting",
		}); err != nil {
			return fmt.Errorf("failed to send startupz starting response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "started",
	}); err != nil {
		return fmt.Errorf("failed to send startupz response: %w", err)
	}

	return nil
}

// MarkStartupComplete permanently opens the per-process startup latch.
func (s *AdminServer) MarkStartupComplete() {
	s.mu.Lock()
	s.startupComplete = true
	s.mu.Unlock()
}

// SetReadinessMonitor attaches an in-memory dependency snapshot to readyz.
func (s *AdminServer) SetReadinessMonitor(monitor *cryptoutilAppsFrameworkServiceServerReadiness.Monitor) {
	s.mu.Lock()
	s.readinessMonitor = monitor
	s.mu.Unlock()
}

// Started returns a channel closed after the admin listener is bound.
func (s *AdminServer) Started() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.startedCh == nil {
		s.startedCh = make(chan struct{})
	}

	return s.startedCh
}

func (s *AdminServer) signalStarted() {
	if !s.started {
		if s.startedCh == nil {
			s.startedCh = make(chan struct{})
		}

		close(s.startedCh)
		s.started = true
	}
}

// BeginDrain marks the process unready before graceful request draining.
func (s *AdminServer) BeginDrain() {
	s.mu.Lock()
	s.draining = true
	s.ready = false
	s.mu.Unlock()
}

func (s *AdminServer) handleLivez(c *fiber.Ctx) error {
	s.mu.RLock()
	shutdownComplete := s.shutdownComplete
	s.mu.RUnlock()

	if shutdownComplete {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.StringStatusShuttingDown,
		}); err != nil {
			return fmt.Errorf("failed to send livez shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: "alive",
	}); err != nil {
		return fmt.Errorf("failed to send livez response: %w", err)
	}

	return nil
}
