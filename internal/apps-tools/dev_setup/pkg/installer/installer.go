package installer

import (
	"context"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/checker"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/config"
	"fmt"
	"os/exec"
	"strings"
)

// Installer defines the interface for installing dependencies.
type Installer interface {
	// Install attempts to install a dependency
	Install(ctx context.Context, dep *config.Dependency) error
}

// DependencyInstaller handles installing dependencies.
type DependencyInstaller struct {
	executor    checker.Executor
	projectRoot string
}

// NewDependencyInstaller creates an installer with a real system executor.
func NewDependencyInstaller(executor checker.Executor, projectRoot string) Installer {
	return &DependencyInstaller{
		executor:    executor,
		projectRoot: projectRoot,
	}
}

// Install executes the installation command for a dependency.
func (di *DependencyInstaller) Install(ctx context.Context, dep *config.Dependency) error {
	if dep.InstallCmd == "" {
		return fmt.Errorf("no install command defined for %s", dep.Name)
	}

	// Parse command and arguments
	parts := strings.Fields(dep.InstallCmd)
	if len(parts) == 0 {
		return fmt.Errorf("invalid install command: %s", dep.InstallCmd)
	}

	cmd := parts[0]
	args := parts[1:]

	// Execute installation
	stdout, stderr, err := di.executor.Execute(ctx, cmd, args...)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nstdout: %s\nstderr: %s", dep.Name, err, stdout, stderr)
	}

	return nil
}

// SystemExecutor implements the Executor interface using os/exec.
type SystemExecutor struct{}

// NewSystemExecutor creates a real system executor.
func NewSystemExecutor() checker.Executor {
	return &SystemExecutor{}
}

// Execute runs a command via os/exec.
func (se *SystemExecutor) Execute(ctx context.Context, cmd string, args ...string) (string, string, error) {
	execCmd := exec.CommandContext(ctx, cmd, args...)
	output, err := execCmd.CombinedOutput()

	// Return combined output as both stdout/stderr for simplicity
	return string(output), string(output), err
}
