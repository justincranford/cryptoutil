// Copyright (c) 2025 Justin Cranford

//go:build !windows

// Package process provides process management functionality for Unix-based systems.
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
)

const (
	dirPermission755  = 0o755
	filePermission600 = 0o600
)

// Manager handles starting, stopping, and tracking service processes.
type Manager struct {
	pidDir string
	mu     sync.Mutex
}

// NewManager creates a new process manager.
func NewManager(pidDir string) (*Manager, error) {
	// Create PID directory if it doesn't exist
	if err := os.MkdirAll(pidDir, dirPermission755); err != nil {
		return nil, fmt.Errorf("failed to create PID directory: %w", err)
	}

	return &Manager{
		pidDir: pidDir,
	}, nil
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

	// Set process group (Unix/Linux)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", serviceName, err)
	}

	// Write PID file
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), filePermission600); err != nil {
		// Kill the process if we can't write PID
		err2 := cmd.Process.Kill()
		if err2 != nil {
			return fmt.Errorf("failed to write PID file: %w; additionally failed to kill process: %w", err, err2)
		}

		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// Stop terminates a running service process.
func (m *Manager) Stop(serviceName string, force bool, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pid, err := m.readPID(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not running: %w", serviceName, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		// Clean up stale PID file
		err2 := m.removePIDFile(serviceName)
		if err2 != nil {
			return fmt.Errorf("failed to find process: %w; additionally failed to remove stale PID file: %w", err, err2)
		}

		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	if force {
		// SIGKILL equivalent on Windows (TerminateProcess)
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	} else {
		// Windows: Send Ctrl+Break signal for graceful shutdown
		// Note: SIGTERM not supported on Windows, using process.Kill() for both cases
		// Production code should use GenerateConsoleCtrlEvent for graceful shutdown
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to terminate process %d: %w", pid, err)
		}
	}

	// Remove PID file
	return m.removePIDFile(serviceName)
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

// isRunning checks if a service is running (caller must hold mutex).
func (m *Manager) isRunning(serviceName string) bool {
	pid, err := m.readPID(serviceName)
	if err != nil {
		return false
	}

	// Check if process exists and is alive
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Windows: FindProcess always succeeds, so we need to test if process is actually running
	// Try to get process state - if it fails, process doesn't exist
	// Note: On Windows, Signal(0) doesn't work reliably, so we just assume FindProcess success means running
	// This is a simplification - production code should use WMI or tasklist
	_ = process // Avoid unused variable warning

	return true // If we can find the process, assume it's running (Windows limitation)
}

// readPID reads the PID from a service's PID file (caller must hold mutex).
func (m *Manager) readPID(serviceName string) (int, error) {
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
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
