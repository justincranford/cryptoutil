package checker

import (
	"context"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/config"
	"fmt"
	"regexp"
	"strings"
)

// Checker defines the interface for checking dependency installation status.
type Checker interface {
	// Check verifies if a dependency is installed and returns (installed, version, error)
	Check(ctx context.Context, dep *config.Dependency) (bool, string, error)
}

// SystemChecker executes system commands to verify installation.
type SystemChecker struct {
	executor Executor
}

// Executor defines the interface for executing external commands.
type Executor interface {
	// Execute runs a command and returns stdout, stderr, and error
	Execute(ctx context.Context, cmd string, args ...string) (stdout, stderr string, err error)
}

// NewDependencyChecker creates a checker with a real system executor.
func NewDependencyChecker(executor Executor) Checker {
	return &SystemChecker{
		executor: executor,
	}
}

// Check verifies installation by running the CheckCmd.
func (sc *SystemChecker) Check(ctx context.Context, dep *config.Dependency) (bool, string, error) {
	if dep.CheckCmd == "" {
		return false, "", fmt.Errorf("no check command defined for %s", dep.Name)
	}

	// Parse command and arguments
	parts := strings.Fields(dep.CheckCmd)
	if len(parts) == 0 {
		return false, "", fmt.Errorf("invalid check command: %s", dep.CheckCmd)
	}

	cmd := parts[0]
	args := parts[1:]

	// Execute check command
	stdout, _, err := sc.executor.Execute(ctx, cmd, args...)
	// If command succeeded, tool is installed
	if err == nil {
		// Extract version if detector is defined
		version := ""
		if dep.DetectVersionCmd != "" {
			version = extractVersion(stdout)
		}

		return true, version, nil
	}

	// Tool is not installed (command failed) - this is not an error condition
	return false, "", nil
}

// extractVersion parses version from output
// Expects patterns like: "version X.Y.Z", "X.Y.Z", etc.
func extractVersion(output string) string {
	// Try various patterns
	patterns := []string{
		`(\d+\.\d+\.\d+(?:[.-]\w+)?)`,  // 1.2.3, 1.2.3-alpha, 1.2.3.4
		`v(\d+\.\d+\.\d+(?:[.-]\w+)?)`, // v1.2.3
		`(\d+\.\d+)`,                   // 1.2
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)

		matches := re.FindStringSubmatch(strings.TrimSpace(output))
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return strings.TrimSpace(output)
}
