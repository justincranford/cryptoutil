// Copyright (c) 2025 Justin Cranford

//go:build windows

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

// Manager handles starting, stopping, and tracking service processes.
type Manager struct {
	pidDir string
	mu     sync.Mutex
}

// NewManager creates a new process manager with the specified PID directory.
func NewManager(pidDir string) (*Manager, error) {
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create PID directory: %w", err)
	}

	return &Manager{pidDir: pidDir}, nil
}

// StartService starts a service process in the background.
func (m *Manager) StartService(ctx context.Context, serviceName, binary string, args ...string) error {
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
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// StopService stops a running service process.
func (m *Manager) StopService(serviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pid, err := m.readPID(serviceName)
	if err != nil {
		return fmt.Errorf("service %s is not running: %w", serviceName, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

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

	const gracefulShutdownTimeout = 10 * time.Second

	select {
	case <-time.After(gracefulShutdownTimeout):
		// Force kill if process doesn't exit within timeout
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to force kill process %d: %w", pid, err)
		}
	case err := <-done:
		if err != nil && err.Error() != "os: process already finished" {
			return fmt.Errorf("error waiting for process %d: %w", pid, err)
		}
	}

	// Remove PID file
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	return nil
}

// StopAll stops all managed service processes.
func (m *Manager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := os.ReadDir(m.pidDir)
	if err != nil {
		return fmt.Errorf("failed to read PID directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".pid" {
			continue
		}

		serviceName := file.Name()[:len(file.Name())-4]

		m.mu.Unlock()
		if err := m.StopService(serviceName); err != nil {
			m.mu.Lock()

			return err
		}

		m.mu.Lock()
	}

	return nil
}

// IsRunning checks if a service is currently running.
func (m *Manager) IsRunning(serviceName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.isRunning(serviceName)
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
	// Try to get process state
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}

// readPID reads the PID from a PID file (caller must hold lock).
func (m *Manager) readPID(serviceName string) (int, error) {
	pidFile := filepath.Join(m.pidDir, serviceName+".pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file content: %w", err)
	}

	return pid, nil
}
