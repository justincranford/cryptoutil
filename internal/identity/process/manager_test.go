// Copyright (c) 2025 Justin Cranford

//go:build !windows

package process

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManagerStartStop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serviceName    string
		binary         string
		args           []string
		forceKill      bool
		expectStartErr bool
		expectStopErr  bool
	}{
		{
			name:           "start and stop sleep process",
			serviceName:    "test-sleep",
			binary:         "timeout",
			args:           []string{"60"},
			forceKill:      false,
			expectStartErr: false,
			expectStopErr:  false,
		},
		{
			name:           "start and force kill sleep process",
			serviceName:    "test-sleep-force",
			binary:         "timeout",
			args:           []string{"60"},
			forceKill:      true,
			expectStartErr: false,
			expectStopErr:  false,
		},
		{
			name:           "start nonexistent binary",
			serviceName:    "test-nonexistent",
			binary:         "nonexistent-binary-xyz",
			args:           []string{},
			forceKill:      false,
			expectStartErr: true,
			expectStopErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pidDir := t.TempDir()
			manager, err := NewManager(pidDir)
			require.NoError(t, err)

			ctx := context.Background()

			err = manager.Start(ctx, tc.serviceName, tc.binary, tc.args)
			if tc.expectStartErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Verify PID file exists
			pidFile := filepath.Join(pidDir, tc.serviceName+".pid")
			require.FileExists(t, pidFile)

			// Verify service is reported as running
			require.True(t, manager.IsRunning(tc.serviceName))

			// Get PID
			pid, err := manager.GetPID(tc.serviceName)
			require.NoError(t, err)
			require.Greater(t, pid, 0)

			// Stop the service
			err = manager.Stop(tc.serviceName, tc.forceKill, 5*time.Second)
			if tc.expectStopErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify PID file is removed
			_, statErr := os.Stat(pidFile)
			require.True(t, os.IsNotExist(statErr))

			// Verify service is no longer running
			require.False(t, manager.IsRunning(tc.serviceName))
		})
	}
}

func TestManagerStopAll(t *testing.T) {
	t.Parallel()

	pidDir := t.TempDir()
	manager, err := NewManager(pidDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Start multiple services
	services := []string{"service1", "service2", "service3"}
	for _, svc := range services {
		err := manager.Start(ctx, svc, "timeout", []string{"60"})
		require.NoError(t, err)
		require.True(t, manager.IsRunning(svc))
	}

	// Stop all services
	err = manager.StopAll(false, 5*time.Second)
	require.NoError(t, err)

	// Verify all services stopped
	for _, svc := range services {
		require.False(t, manager.IsRunning(svc))
	}
}

func TestManagerDoubleStart(t *testing.T) {
	t.Parallel()

	pidDir := t.TempDir()
	manager, err := NewManager(pidDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Start service once
	err = manager.Start(ctx, "test-double", "timeout", []string{"60"})
	require.NoError(t, err)

	// Try to start again - should fail
	err = manager.Start(ctx, "test-double", "timeout", []string{"60"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "already running")

	// Cleanup
	err = manager.Stop("test-double", true, 5*time.Second)
	require.NoError(t, err)
}

func TestManagerStopNonRunning(t *testing.T) {
	t.Parallel()

	pidDir := t.TempDir()
	manager, err := NewManager(pidDir)
	require.NoError(t, err)

	// Try to stop non-running service
	err = manager.Stop("nonexistent", false, 5*time.Second)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not running")
}
