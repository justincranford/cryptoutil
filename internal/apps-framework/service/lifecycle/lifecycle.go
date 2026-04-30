// Copyright (c) 2025-2026 Justin Cranford.
// Package lifecycle provides a reusable graceful-shutdown lifecycle helper for service entry points.
// It encapsulates the signal.Notify / errChan / select / signal.Stop pattern that all PS-ID server
// subcommands share, eliminating ~25 lines of identical boilerplate per service.
package lifecycle

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Starter is the minimal interface required by RunWithGracefulShutdown.
// Both Start and Shutdown take a context and return an error.
type Starter interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// RunService runs a service with graceful shutdown handling.
// It starts srv.Start in a goroutine, waits for either a server error or SIGINT/SIGTERM,
// and performs a graceful shutdown if a signal is received.
// Returns 0 on clean exit, 1 on server error or shutdown error.
func RunService(ctx context.Context, stdout, stderr io.Writer, srv Starter) int {
	return RunWithGracefulShutdown(ctx, stdout, stderr, srv.Start, srv.Shutdown)
}

// RunWithGracefulShutdown runs startFn in a goroutine and waits for either a server error
// or an OS signal (SIGINT or SIGTERM). On signal, calls shutdownFn with a timeout-bounded
// context. Returns 0 on clean exit, 1 on error.
//
// This function handles the full signal lifecycle: signal.Notify, errChan select, signal.Stop,
// close(sigChan). Callers do NOT need to manage any of these.
func RunWithGracefulShutdown(
	ctx context.Context,
	stdout, stderr io.Writer,
	startFn func(ctx context.Context) error,
	shutdownFn func(ctx context.Context) error,
) int {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(sigChan)
	defer close(sigChan)

	return runWithSignalChan(ctx, stdout, stderr, startFn, shutdownFn, sigChan)
}

// runWithSignalChan is the internal implementation that accepts an already-prepared signal channel.
// This enables test seam injection: tests pass a synthetic channel pre-loaded with the desired
// signal, avoiding real OS signal delivery to the test process.
func runWithSignalChan(
	ctx context.Context,
	stdout, stderr io.Writer,
	startFn func(ctx context.Context) error,
	shutdownFn func(ctx context.Context) error,
	sigChan <-chan os.Signal,
) int {
	errChan := make(chan error, 1)

	go func() {
		errChan <- startFn(ctx)
	}()

	exitCode := 0

	select {
	case err := <-errChan:
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Server error: %v\n", err)

			exitCode = 1
		}
	case sig := <-sigChan:
		_, _ = fmt.Fprintf(stdout, "\nReceived signal %v, shutting down gracefully...\n", sig)

		shutdownCtx, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
		defer cancel()

		if shutdownErr := shutdownFn(shutdownCtx); shutdownErr != nil {
			_, _ = fmt.Fprintf(stderr, "Shutdown error: %v\n", shutdownErr)
		}
	}

	return exitCode
}
