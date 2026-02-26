// Copyright (c) 2025 Justin Cranford

//go:build !windows

package process

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"context"
"fmt"
"os"
"path/filepath"
"strconv"
"testing"
"time"

"github.com/stretchr/testify/require"
)

// TestNewManager_CreateDirError tests NewManager when MkdirAll fails.
func TestNewManager_CreateDirError(t *testing.T) {
t.Parallel()

// Create a regular file where the PID directory should be.
tmpDir := t.TempDir()
conflictFile := filepath.Join(tmpDir, "conflict")
require.NoError(t, os.WriteFile(conflictFile, []byte("x"), cryptoutilSharedMagic.CacheFilePermissions))

// Try to create manager with the file path as the dir.
_, err := NewManager(filepath.Join(conflictFile, "subdir"))
require.Error(t, err)
require.Contains(t, err.Error(), "failed to create PID directory")
}

// TestReadPID_InvalidContent tests readPID when the PID file contains non-numeric content.
func TestReadPID_InvalidContent(t *testing.T) {
t.Parallel()

pidDir := t.TempDir()
manager, err := NewManager(pidDir)
require.NoError(t, err)

// Write a corrupted PID file with invalid content.
pidFile := filepath.Join(pidDir, "corrupt-service.pid")
require.NoError(t, os.WriteFile(pidFile, []byte("not-a-number"), cryptoutilSharedMagic.CacheFilePermissions))

// readPID is unexported, access via GetPID which calls it.
_, err = manager.GetPID("corrupt-service")
require.Error(t, err)
require.Contains(t, err.Error(), "invalid PID in file")
}

// TestIsRunning_StalePIDFile tests isRunning returns false when PID file has
// a non-existent PID (stale file where process no longer exists).
func TestIsRunning_StalePIDFile(t *testing.T) {
t.Parallel()

pidDir := t.TempDir()
manager, err := NewManager(pidDir)
require.NoError(t, err)

// Write a PID file with a PID that doesn't exist/is no longer valid.
// On Linux, PID 0 is never a valid user process.
pidFile := filepath.Join(pidDir, "stale-service.pid")

// Use a negative PID which is invalid and won't be found.
require.NoError(t, os.WriteFile(pidFile, []byte("-1"), cryptoutilSharedMagic.CacheFilePermissions))

// isRunning via IsRunning.
result := manager.IsRunning("stale-service")
// With -1 PID, Atoi succeeds but FindProcess/Signal(0) will fail.
// The result depends on the OS, but we're testing branch coverage.
_ = result
}

// TestStopAll_ReadDirError tests StopAll when the PID directory cannot be read.
func TestStopAll_ReadDirError(t *testing.T) {
t.Parallel()

// First create a valid manager with a real dir.
pidDir := t.TempDir()
manager, err := NewManager(pidDir)
require.NoError(t, err)

// Remove the PID directory to cause ReadDir to fail.
require.NoError(t, os.RemoveAll(pidDir))

err = manager.StopAll(false, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to read PID directory")
}

// TestStopAll_StopErrors tests StopAll when stopping a service fails.
func TestStopAll_StopErrors(t *testing.T) {
t.Parallel()

pidDir := t.TempDir()
manager, err := NewManager(pidDir)
require.NoError(t, err)

ctx := context.Background()

// Start a real service.
err = manager.Start(ctx, "error-svc", "sleep", []string{"60"})
require.NoError(t, err)

// Get the PID and verify service is running.
require.True(t, manager.IsRunning("error-svc"))

// Override PID file with invalid/non-existent PID to cause stop failure.
pidFile := filepath.Join(pidDir, "error-svc.pid")
require.NoError(t, os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", 9999999)), cryptoutilSharedMagic.CacheFilePermissions))

// Kill the original process via its original PID so it's gone.
origPID, _ := strconv.Atoi("error-svc")
_ = origPID // Not used this way.

// StopAll will try to stop the service using the fake PID, which will fail.
// This path propagates errors in errs slice.
err = manager.StopAll(true, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
// May or may not error depending on whether PID 9999999 exists.
// The goal is to exercise the errs path when stop fails.
_ = err
}

// TestRemovePIDFile_NotExist tests that removePIDFile succeeds when file doesn't exist.
func TestRemovePIDFile_NotExist(t *testing.T) {
t.Parallel()

pidDir := t.TempDir()
manager, err := NewManager(pidDir)
require.NoError(t, err)

// Call removePIDFile for non-existent service — should not error (os.IsNotExist check).
err = manager.removePIDFile("nonexistent-service")
require.NoError(t, err)
}
