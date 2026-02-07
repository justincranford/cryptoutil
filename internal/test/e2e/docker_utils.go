// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// DockerContainer represents a Docker container.
type DockerContainer struct {
	Name string
}

// Docker compose command arguments constants.
var (
	dockerComposeArgsStopServices  = []string{"down", "-v", "--remove-orphans"}
	dockerComposeArgsStartServices = []string{"up", "-d", "--force-recreate"}
	dockerComposeArgsPsServices    = []string{"ps", "-a", "--format", "json"}
)

// Docker compose command description constants.
const (
	dockerComposeDescStopServices  = "Stop services"
	dockerComposeDescStartServices = "Start services"
	dockerComposeDescBatchHealth   = "Batch health check"
)

// runDockerComposeCommand executes a docker compose command with the given arguments.
//
//	Windows: docker compose -f .\deployments\compose\compose.yml <command> <args>
//	Linux:   docker compose -f ./deployments/compose/compose.yml <command> <args>
//
// Always use relative path for cross-platform compatibility in
// in GitHub Actions (Ubuntu runners) and Windows (`act` runner).
func runDockerComposeCommand(ctx context.Context, logger *Logger, description string, args []string) ([]byte, error) {
	// Validate required parameters
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	} else if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	} else if len(args) == 0 {
		return nil, fmt.Errorf("args cannot be empty")
	}

	// Log start message based on description
	// could use tagged switch on description instead of if/elseif/else
	logComposeMessage(logger, description, true)

	composeFile := getComposeFilePath()
	allArgs := append([]string{"docker", "compose", "-f", composeFile}, args...)
	cmd := exec.CommandContext(ctx, allArgs[0], allArgs[1:]...)
	// Enable postgres profile via environment variable (more compatible than --profile flag).
	cmd.Env = append(os.Environ(), "COMPOSE_PROFILES=postgres")
	output, err := cmd.CombinedOutput()
	LogCommand(logger, description, cmd.String(), string(output))

	if err != nil {
		return output, fmt.Errorf("docker command failed: %w", err)
	}

	// Log success message based on description
	// could use tagged switch on description instead of if/elseif/else
	logComposeMessage(logger, description, false)

	return output, nil
}

// getComposeFilePath returns the compose file path appropriate for the current OS.
// getComposeFilePath returns the absolute path to the docker-compose file.
func getComposeFilePath() string {
	// Navigate up from internal/test/e2e/ to project root, then to deployments/compose/compose.yml
	projectRoot := filepath.Join("..", "..", "..")
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

// getContainerLogsOutputDir returns the absolute path to the container logs output directory.
func getContainerLogsOutputDir() string {
	// Navigate up from internal/test/e2e/ to project root, then to workflow-reports/e2e/
	projectRoot := filepath.Join("..", "..", "..")
	outputPath := filepath.Join(projectRoot, "workflow-reports", "e2e")

	// Convert to absolute path to ensure it works regardless of working directory
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		// Fallback to relative path if absolute path fails
		return outputPath
	}

	return absPath
}

// could use tagged switch on description instead of if/elseif/else.
func logComposeMessage(logger *Logger, description string, isStart bool) {
	switch description {
	case dockerComposeDescStopServices:
		if isStart {
			Log(logger, "🧹 Stopping Docker Compose services")
		} else {
			Log(logger, "✅ Existing services stopped successfully")
		}
	case dockerComposeDescStartServices:
		if isStart {
			Log(logger, "🚀 Starting Docker Compose services")
		} else {
			Log(logger, "✅ Docker Compose services started successfully")
		}
	case dockerComposeDescBatchHealth:
		if isStart {
			Log(logger, "🔍 Checking Docker Compose services health")
		} else {
			Log(logger, "✅ Docker Compose services health check completed")
		}
	}
}

// CaptureAndZipContainerLogs captures logs from all Docker containers and creates a zip archive.
func CaptureAndZipContainerLogs(ctx context.Context, logger *Logger, outputDir string) error {
	// Validate required parameters
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	} else if logger == nil {
		return fmt.Errorf("logger cannot be nil")
	} else if outputDir == "" {
		return fmt.Errorf("outputDir cannot be empty")
	}

	Log(logger, "📦 Capturing container logs...")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get list of all containers (including exited ones) that match our compose project
	containers, err := getDockerContainers(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to get Docker containers: %w", err)
	}

	if len(containers) == 0 {
		Log(logger, "⚠️ No Docker containers found, skipping log capture")

		return nil
	}

	// Create zip file with timestamp
	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05")
	zipFileName := filepath.Join(outputDir, fmt.Sprintf("container-logs_%s.zip", timestamp))

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Capture logs for each container
	for _, container := range containers {
		if err := captureContainerLogs(ctx, logger, zipWriter, container); err != nil {
			Log(logger, "⚠️ Failed to capture logs for container %s: %v", container.Name, err)
			// Continue with other containers even if one fails
			continue
		}
	}

	Log(logger, "✅ Container logs captured and zipped to: %s", zipFileName)

	return nil
}

// getDockerContainers returns a list of all Docker containers that match our compose project.
func getDockerContainers(ctx context.Context, logger *Logger) ([]DockerContainer, error) {
	// Get the compose project name from the compose file
	composeFile := getComposeFilePath()

	projectName, err := getComposeProjectName(ctx, logger, composeFile)
	if err != nil {
		Log(logger, "⚠️ Failed to get compose project name, falling back to listing all containers: %v", err)
		// Fall back to getting all containers if we can't determine project name
		return getAllDockerContainers(ctx, logger)
	}

	// Use docker ps -a with label filter to find containers from our compose project
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", fmt.Sprintf("label=com.docker.compose.project=%s", projectName), "--format", "{{.Names}}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		LogCommand(logger, "List containers by project", cmd.String(), string(output))
		Log(logger, "⚠️ Failed to list containers by project, falling back to all containers: %v", err)
		// Fall back to getting all containers
		return getAllDockerContainers(ctx, logger)
	}

	// Parse container names from output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]DockerContainer, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			containers = append(containers, DockerContainer{Name: line})
		}
	}

	Log(logger, "📋 Found %d containers for project %s", len(containers), projectName)

	return containers, nil
}

// getAllDockerContainers returns all Docker containers (fallback method).
func getAllDockerContainers(ctx context.Context, logger *Logger) ([]DockerContainer, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "{{.Names}}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		LogCommand(logger, "List all containers", cmd.String(), string(output))

		return nil, fmt.Errorf("failed to list all containers: %w", err)
	}

	// Parse container names from output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]DockerContainer, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			containers = append(containers, DockerContainer{Name: line})
		}
	}

	Log(logger, "📋 Found %d containers total (fallback method)", len(containers))

	return containers, nil
}

// getComposeProjectName extracts the project name from docker-compose config.
func getComposeProjectName(ctx context.Context, logger *Logger, composeFile string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "config", "--format", "json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		LogCommand(logger, "Get compose config", cmd.String(), string(output))

		return "", fmt.Errorf("failed to get compose config: %w", err)
	}

	// Simple approach: look for the project name in the compose file path
	// Docker Compose uses the directory name as the project name by default
	composeDir := filepath.Dir(composeFile)
	projectName := filepath.Base(composeDir)

	Log(logger, "📋 Determined compose project name: %s", projectName)

	return projectName, nil
}

// captureContainerLogs captures logs for a single container and adds them to the zip archive.
func captureContainerLogs(ctx context.Context, logger *Logger, zipWriter *zip.Writer, container DockerContainer) error {
	Log(logger, "  📋 Capturing logs for container: %s", container.Name)

	cmd := exec.CommandContext(ctx, "docker", "logs", "--timestamps", container.Name)
	output, err := cmd.CombinedOutput()

	// Create entry in zip file
	logFileName := fmt.Sprintf("%s.log", container.Name)

	zipEntry, zipErr := zipWriter.Create(logFileName)
	if zipErr != nil {
		return fmt.Errorf("failed to create zip entry for %s: %w", container.Name, zipErr)
	}

	// Write logs to zip entry (even if command failed, write what we have)
	if len(output) > 0 {
		if _, writeErr := zipEntry.Write(output); writeErr != nil {
			return fmt.Errorf("failed to write logs to zip for %s: %w", container.Name, writeErr)
		}
	} else {
		// Write a message if no logs available
		noLogsMsg := fmt.Sprintf("[No logs available for container %s]\n", container.Name)
		if _, writeErr := io.WriteString(zipEntry, noLogsMsg); writeErr != nil {
			return fmt.Errorf("failed to write no-logs message for %s: %w", container.Name, writeErr)
		}
	}

	if err != nil {
		// Log the error but don't fail - we've already captured what we could
		Log(logger, "  ⚠️ Error capturing logs for %s (captured partial logs): %v", container.Name, err)
	} else {
		Log(logger, "  ✅ Captured logs for container: %s (%d bytes)", container.Name, len(output))
	}

	return nil
}
