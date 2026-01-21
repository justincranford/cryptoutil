// Copyright (c) 2025 Justin Cranford

//go:build windows

// Package process provides process management utilities for Windows.
package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	processFinishedError = "os: process already finished"
)

// Manager handles starting, stopping, and tracking service processes.
type Manager struct {
	pidDir string
	mu     sync.Mutex
}

// NewManager creates a new process manager with the specified PID directory.
func NewManager(pidDir string) (*Manager, error) {
	if err := os.MkdirAll(pidDir, cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
		return nil, fmt.Errorf("failed to create PID directory: %w", err)
	}

	return &Manager{pidDir: pidDir}, nil
}

// Start launches a service process in the background.
func (m *Manager) Start(ctx context.Context, serviceName string, binary string, args []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if service is already running
	if m.isRunning(serviceName) {
		return fmt.Errorf("service %s is already running", serviceName)
	}

	// Create command with context
	cmd := exec.CommandContext(ctx, binary, args...)

	// Set process group (Windows)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", serviceName, err)
	}

	// Write PID file
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), cryptoutilMagic.FilePermOwnerReadWriteGroupRead); err != nil {
		// Kill the process if we can't write PID
		if killErr := cmd.Process.Kill(); killErr != nil {
			// Log but continue with original error
			_ = killErr
		}

		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// Stop terminates a running service process.
func (m *Manager) Stop(serviceName string, force bool, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pid, err := m.readPID(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not running: %w", serviceName, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		// Clean up stale PID file
		if removeErr := m.removePIDFile(serviceName); removeErr != nil {
			// Log but continue with original error
			_ = removeErr
		}

		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	if force {
		// Force kill immediately
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	} else {
		// Send SIGTERM signal (graceful shutdown)
		if err := process.Signal(syscall.SIGTERM); err != nil {
			// If SIGTERM fails, force kill
			if killErr := process.Kill(); killErr != nil {
				return fmt.Errorf("failed to kill process %d: %w", pid, killErr)
			}
		}

		// Wait for process to exit (with timeout)
		done := make(chan error, 1)

		go func() {
			_, waitErr := process.Wait()
			done <- waitErr
		}()

		select {
		case <-time.After(timeout):
			// Force kill if process doesn't exit within timeout
			if err := process.Kill(); err != nil {
				return fmt.Errorf("failed to force kill process %d: %w", pid, err)
			}
		case err := <-done:
			if err != nil && err.Error() != processFinishedError {
				return fmt.Errorf("error waiting for process %d: %w", pid, err)
			}
		}
	}

	// Remove PID file
	return m.removePIDFile(serviceName)
}

// StopAll stops all tracked services.
func (m *Manager) StopAll(force bool, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := os.ReadDir(m.pidDir)
	if err != nil {
		return fmt.Errorf("failed to read PID directory: %w", err)
	}

	var errs []error

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pid" {
			serviceName := file.Name()[:len(file.Name())-4] // Remove .pid extension

			m.mu.Unlock() // Unlock before recursive call

			if stopErr := m.Stop(serviceName, force, timeout); stopErr != nil {
				errs = append(errs, fmt.Errorf("%s: %w", serviceName, stopErr))
			}

			m.mu.Lock() // Re-lock
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping services: %v", errs)
	}

	return nil
}

// IsRunning checks if a service is currently running.
func (m *Manager) IsRunning(serviceName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.isRunning(serviceName)
}

// GetPID returns the PID of a running service.
func (m *Manager) GetPID(serviceName string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.readPID(serviceName)
}

// isRunning checks if a service is running (caller must hold lock).
func (m *Manager) isRunning(serviceName string) bool {
	pid, err := m.readPID(serviceName)
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Windows, FindProcess always succeeds, so we need to check if process is actually running
	// Signal(0) doesn't work on Windows - attempt to get process state via Wait with timeout
	done := make(chan error, 1)

	go func() {
		_, waitErr := process.Wait()
		done <- waitErr
	}()

	const processStatusCheckTimeout = 10 * time.Millisecond

	select {
	case <-time.After(processStatusCheckTimeout):
		// Process hasn't exited yet - it's running
		return true
	case err := <-done:
		// Process exited or error occurred
		if err != nil && err.Error() == processFinishedError {
			return false
		}

		return false
	}
}

// readPID reads the PID from a PID file (caller must hold lock).
func (m *Manager) readPID(serviceName string) (int, error) {
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file content: %w", err)
	}

	return pid, nil
}

// removePIDFile deletes a service's PID file (caller must hold mutex).
func (m *Manager) removePIDFile(serviceName string) error {
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	return nil
}
