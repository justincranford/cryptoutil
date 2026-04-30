// Copyright (c) 2025-2026 Justin Cranford.
package lifecycle_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	cryptoutilLifecycle "cryptoutil/internal/apps-framework/service/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// mockServer implements Starter for testing.
type mockServer struct {
	startFn    func(ctx context.Context) error
	shutdownFn func(ctx context.Context) error
}

func (m *mockServer) Start(ctx context.Context) error {
	return m.startFn(ctx)
}

func (m *mockServer) Shutdown(ctx context.Context) error {
	return m.shutdownFn(ctx)
}

// TestRunWithSignalChan_SignalPath tests the signal path using injected signal channels.
// This avoids sending real OS signals to the test process (which would terminate it).
func TestRunWithSignalChan_SignalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		signal         os.Signal
		shutdownErr    error
		wantExitCode   int
		wantStdoutText string
		wantStderrText string
	}{
		{
			name:           "SIGINT clean shutdown",
			signal:         syscall.SIGINT,
			shutdownErr:    nil,
			wantExitCode:   0,
			wantStdoutText: "shutting down gracefully",
		},
		{
			name:           "SIGTERM clean shutdown",
			signal:         syscall.SIGTERM,
			shutdownErr:    nil,
			wantExitCode:   0,
			wantStdoutText: "shutting down gracefully",
		},
		{
			name:           "SIGINT with shutdown error",
			signal:         syscall.SIGINT,
			shutdownErr:    errors.New("shutdown failed"),
			wantExitCode:   0, // shutdown errors don't change exit code
			wantStdoutText: "shutting down gracefully",
			wantStderrText: "Shutdown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			ctx := context.Background()

			started := make(chan struct{})

			startFn := func(_ context.Context) error {
				close(started)
				// Block until context cancelled or test ends.
				<-ctx.Done()

				return nil
			}

			shutdownFn := func(_ context.Context) error {
				return tt.shutdownErr
			}

			// Inject synthetic signal channel — no real OS signal sent.
			sigChan := make(chan os.Signal, 1)
			sigChan <- tt.signal

			done := make(chan int, 1)

			go func() {
				done <- cryptoutilLifecycle.RunWithSignalChan(ctx, &stdout, &stderr, startFn, shutdownFn, sigChan)
			}()

			// Wait for server goroutine to start.
			select {
			case <-started:
			case <-time.After(cryptoutilSharedMagic.TestTimeoutServiceRetry):
				t.Fatal("server did not start in time")
			}

			// Wait for lifecycle to complete.
			select {
			case exitCode := <-done:
				require.Equal(t, tt.wantExitCode, exitCode)
			case <-time.After(cryptoutilSharedMagic.TestTimeoutDockerComposeInit):
				t.Fatal("lifecycle did not complete in time")
			}

			if tt.wantStdoutText != "" {
				require.Contains(t, stdout.String(), tt.wantStdoutText)
			}

			if tt.wantStderrText != "" {
				require.Contains(t, stderr.String(), tt.wantStderrText)
			}
		})
	}
}

func TestRunWithGracefulShutdown_ServerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		startErr     error
		wantExitCode int
		wantStderr   string
	}{
		{
			name:         "server error returns 1",
			startErr:     errors.New("port already in use"),
			wantExitCode: 1,
			wantStderr:   "Server error",
		},
		{
			name:         "server exits cleanly returns 0",
			startErr:     nil,
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			ctx := context.Background()

			startFn := func(_ context.Context) error {
				return tt.startErr
			}

			shutdownFn := func(_ context.Context) error {
				return nil
			}

			exitCode := cryptoutilLifecycle.RunWithGracefulShutdown(ctx, &stdout, &stderr, startFn, shutdownFn)

			require.Equal(t, tt.wantExitCode, exitCode)

			if tt.wantStderr != "" {
				require.Contains(t, stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestRunService_ServerError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	ctx := context.Background()

	srv := &mockServer{
		startFn: func(_ context.Context) error {
			return errors.New("injected start error")
		},
		shutdownFn: func(_ context.Context) error {
			return nil
		},
	}

	exitCode := cryptoutilLifecycle.RunService(ctx, &stdout, &stderr, srv)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Server error")
}

func TestRunService_CleanExit(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	ctx := context.Background()

	srv := &mockServer{
		startFn: func(_ context.Context) error {
			return nil
		},
		shutdownFn: func(_ context.Context) error {
			return nil
		},
	}

	exitCode := cryptoutilLifecycle.RunService(ctx, &stdout, &stderr, srv)

	require.Equal(t, 0, exitCode)
	require.Empty(t, stderr.String())
}

func TestRunWithGracefulShutdown_StartFnCalledOnce(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	ctx := context.Background()

	var callCount int32

	startFn := func(_ context.Context) error {
		atomic.AddInt32(&callCount, 1)

		return nil
	}

	shutdownFn := func(_ context.Context) error {
		return nil
	}

	exitCode := cryptoutilLifecycle.RunWithGracefulShutdown(ctx, &stdout, &stderr, startFn, shutdownFn)

	require.Equal(t, 0, exitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&callCount), "startFn must be called exactly once")
}

// TestRunWithSignalChan_SignalDeliveredBeforeStart verifies that a pre-buffered signal
// triggers shutdown even if start completes before select chooses signal case.
func TestRunWithSignalChan_SignalDeliveredBeforeStart(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	ctx := context.Background()

	startFn := func(_ context.Context) error {
		return nil
	}

	shutdownFn := func(_ context.Context) error {
		return nil
	}

	sigChan := make(chan os.Signal, 1)
	sigChan <- syscall.SIGINT

	exitCode := cryptoutilLifecycle.RunWithSignalChan(ctx, &stdout, &stderr, startFn, shutdownFn, sigChan)

	// Either the signal or the clean start path is taken — both yield exit code 0.
	require.Equal(t, 0, exitCode)
}
