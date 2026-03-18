// Copyright (c) 2025 Justin Cranford

// Package e2e_helpers provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from sm-im implementation to support 9-service migration.
package e2e_helpers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ServerWithActualPort defines interface for servers that expose their dynamically allocated port.
// Reusable for all services implementing dynamic port allocation (port 0 pattern).
type ServerWithActualPort interface {
	Start(ctx context.Context) error
	ActualPort() int
}

// ServerWaitParams holds configuration for waiting for server to bind to port.
type ServerWaitParams struct {
	MaxWaitAttempts int
	WaitInterval    time.Duration
}

// DefaultServerWaitParams returns default wait parameters for server binding.
func DefaultServerWaitParams() ServerWaitParams {
	const (
		defaultMaxWaitAttempts = 50
		defaultWaitInterval    = 100 * time.Millisecond
	)

	return ServerWaitParams{
		MaxWaitAttempts: defaultMaxWaitAttempts,
		WaitInterval:    defaultWaitInterval,
	}
}

// WaitForServerPort waits for server to bind to port and returns base URL.
// Reusable for all services implementing dynamic port allocation.
//
// Parameters:
//   - server: Server implementing ActualPort() method
//   - waitParams: Wait configuration (max attempts, interval)
//
// Returns base URL (https://127.0.0.1:PORT) or error if server fails to bind.
func WaitForServerPort(server ServerWithActualPort, waitParams ServerWaitParams) (string, error) {
	actualPort := 0

	for i := 0; i < waitParams.MaxWaitAttempts; i++ {
		actualPort = server.ActualPort()
		if actualPort > 0 {
			break
		}

		time.Sleep(waitParams.WaitInterval)
	}

	if actualPort <= 0 {
		return "", fmt.Errorf("server did not bind to port after %d attempts", waitParams.MaxWaitAttempts)
	}

	baseURL := "https://" + cryptoutilSharedMagic.IPv4Loopback + ":" + strconv.Itoa(actualPort)

	return baseURL, nil
}

// StartServerAsync starts server in background goroutine.
// Reusable for all services implementing Start(ctx) error method.
//
// Parameters:
//   - ctx: Context for server lifecycle
//   - server: Server implementing ServerWithActualPort interface
//
// Returns error channel for monitoring startup failures.
func StartServerAsync(ctx context.Context, server ServerWithActualPort) chan error {
	errChan := make(chan error, 1)

	go func() {
		if startErr := server.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	return errChan
}
