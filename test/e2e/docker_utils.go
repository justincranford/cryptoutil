//go:build e2e

package test

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// Docker compose command arguments constants.
var (
	dockerComposeArgsStopServices  = []string{"down", "-v", "--remove-orphans"}
	dockerComposeArgsStartServices = []string{"up", "-d", "--force-recreate", "--build"}
	dockerComposeArgsPsServices    = []string{"ps", "-a", "--format", "json"}
)

// Docker compose command description constants.
const (
	dockerComposeDescStopServices  = "Stop services"
	dockerComposeDescStartServices = "Start services"
	dockerComposeDescBatchHealth   = "Batch health check"
)

// getComposeFilePath returns the compose file path appropriate for the current OS.
// Since E2E tests run from test/e2e/ directory, we need to navigate up to project root.
func getComposeFilePath() string {
	// Navigate up from test/e2e/ to project root, then to deployments/compose/compose.yml
	projectRoot := filepath.Join("..", "..")
	composePath := filepath.Join(projectRoot, "deployments", "compose", "compose.yml")

	// Convert to absolute path to ensure it works regardless of working directory
	absPath, err := filepath.Abs(composePath)
	if err != nil {
		// Fallback to relative path if absolute path fails
		if runtime.GOOS == "windows" {
			return cryptoutilMagic.DockerComposeRelativeFilePathWindows
		}

		return cryptoutilMagic.DockerComposeRelativeFilePathLinux
	}

	return absPath
}

// runDockerComposeCommand executes a docker compose command with the given arguments.
//
//	Windows: docker compose -f .\deployments\compose\compose.yml <command> <args>
//	Linux:   docker compose -f ./deployments/compose/compose.yml <command> <args>
//
// Always use relative path for cross-platform compatibility in
// in GitHub Actions (Ubuntu runners) and Windows (`act` runner).
func runDockerComposeCommand(ctx context.Context, logger *Logger, description string, args []string) ([]byte, error) {
	// Log start message based on description
	if logger != nil {
		switch description {
		case dockerComposeDescStopServices:
			Log(logger, "🧹 Stopping Docker Compose services")
		case dockerComposeDescStartServices:
			Log(logger, "🚀 Starting Docker Compose services")
		}
	}

	composeFile := getComposeFilePath()
	allArgs := append([]string{"docker", "compose", "-f", composeFile}, args...)
	cmd := exec.CommandContext(ctx, allArgs[0], allArgs[1:]...)
	output, err := cmd.CombinedOutput()
	LogCommand(logger, description, cmd.String(), string(output))

	if err != nil {
		return output, err
	}

	// Log success message based on description
	switch description {
	case dockerComposeDescStopServices:
		Log(logger, "✅ Existing services stopped successfully")
	case dockerComposeDescStartServices:
		Log(logger, "✅ Docker Compose services started successfully")
	}

	return output, nil
}
