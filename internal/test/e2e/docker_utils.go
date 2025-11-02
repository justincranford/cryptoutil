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

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

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
// Since E2E tests run from internal/cmd/e2e/ directory, we need to navigate up to project root.
func getComposeFilePath() string {
	// Navigate up from internal/cmd/e2e/ to project root, then to deployments/compose/compose.yml
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

	// Get list of all services
	services, err := getDockerComposeServices(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to get Docker Compose services: %w", err)
	}

	if len(services) == 0 {
		Log(logger, "⚠️ No Docker Compose services found, skipping log capture")

		return nil
	}

	// Create zip file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	zipFileName := filepath.Join(outputDir, fmt.Sprintf("container-logs_%s.zip", timestamp))

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Capture logs for each service
	for _, service := range services {
		if err := captureServiceLogs(ctx, logger, zipWriter, service); err != nil {
			Log(logger, "⚠️ Failed to capture logs for service %s: %v", service, err)
			// Continue with other services even if one fails
			continue
		}
	}

	Log(logger, "✅ Container logs captured and zipped to: %s", zipFileName)

	return nil
}

// getDockerComposeServices returns a list of all Docker Compose services.
func getDockerComposeServices(ctx context.Context, logger *Logger) ([]string, error) {
	composeFile := getComposeFilePath()
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "ps", "-a", "--services")
	output, err := cmd.CombinedOutput()
	if err != nil {
		LogCommand(logger, "List services", cmd.String(), string(output))

		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Parse service names from output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	services := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			services = append(services, line)
		}
	}

	return services, nil
}

// captureServiceLogs captures logs for a single service and adds them to the zip archive.
func captureServiceLogs(ctx context.Context, logger *Logger, zipWriter *zip.Writer, service string) error {
	Log(logger, "  📋 Capturing logs for service: %s", service)

	composeFile := getComposeFilePath()
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "logs", "--no-color", "--timestamps", service)
	output, err := cmd.CombinedOutput()

	// Create entry in zip file
	logFileName := fmt.Sprintf("%s.log", service)

	zipEntry, zipErr := zipWriter.Create(logFileName)
	if zipErr != nil {
		return fmt.Errorf("failed to create zip entry for %s: %w", service, zipErr)
	}

	// Write logs to zip entry (even if command failed, write what we have)
	if len(output) > 0 {
		if _, writeErr := zipEntry.Write(output); writeErr != nil {
			return fmt.Errorf("failed to write logs to zip for %s: %w", service, writeErr)
		}
	} else {
		// Write a message if no logs available
		noLogsMsg := fmt.Sprintf("[No logs available for service %s]\n", service)
		if _, writeErr := io.WriteString(zipEntry, noLogsMsg); writeErr != nil {
			return fmt.Errorf("failed to write no-logs message for %s: %w", service, writeErr)
		}
	}

	if err != nil {
		// Log the error but don't fail - we've already captured what we could
		Log(logger, "  ⚠️ Error capturing logs for %s (captured partial logs): %v", service, err)
	} else {
		Log(logger, "  ✅ Captured logs for service: %s (%d bytes)", service, len(output))
	}

	return nil
}
