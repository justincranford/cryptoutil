// Copyright (c) 2025 Justin Cranford
//
//

// Package ca provides the unified command interface for Certificate Authority service.
package ca //nolint:wsl_v5

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cryptoutilCAServer "cryptoutil/internal/apps/ca/server"
	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/ca/server/config"
)

const (
	configFlag       = "--config"
	configFlagShort  = "-c"
	defaultAdminPort = 9090
	fileURLPrefix    = "file://"
	httpTimeout      = 5 * time.Second
)

// Execute handles CA service commands matching KMS/Identity/JOSE pattern.
// Supports: start, stop, status, health.
func Execute(parameters []string) {
	if len(parameters) < 1 {
		printUsage()
		os.Exit(1)
	}

	subcommand := parameters[0]
	cmdParams := parameters[1:]

	switch subcommand {
	case "start":
		startService(cmdParams)
	case "stop":
		stopService(cmdParams)
	case "status":
		statusService(cmdParams)
	case "health":
		healthService(cmdParams)
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cryptoutil ca <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  start    - Start Certificate Authority service")
	fmt.Println("  stop     - Stop Certificate Authority service")
	fmt.Println("  status   - Check service status")
	fmt.Println("  health   - Check service health")
	fmt.Println("\nOptions:")
	fmt.Println("  --config, -c <path>  - Config file path (default: /app/run/ca-docker.yml)")
}

// startService starts the Certificate Authority service.
func startService(parameters []string) {
	configFile := parseConfigFlag(parameters, "/app/run/ca-docker.yml")

	fmt.Fprintf(os.Stderr, "Starting Certificate Authority service\n")
	fmt.Fprintf(os.Stderr, "Using config file: %s\n", configFile)

	// Load CA-specific configuration from YAML file.
	parseArgs := []string{"start", "--config", configFile}

	settings, err := cryptoutilAppsCaServerConfig.Parse(parseArgs, false)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configFile, err)
	}

	// Create context.
	ctx := context.Background()

	// Create CA server using new template-based implementation.
	server, err := cryptoutilCAServer.NewFromConfig(ctx, settings)
	if err != nil {
		log.Fatalf("Failed to create CA server: %v", err)
	}

	// Mark server as ready.
	server.SetReady(true)

	fmt.Fprintf(os.Stderr, "Starting CA server...\n")

	// Start server (blocks until shutdown).
	if err := server.Start(ctx); err != nil {
		log.Fatalf("CA server error: %v", err)
	}
}

// stopService sends shutdown request to CA admin endpoint.
func stopService(parameters []string) { //nolint:wsl_v5
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, "Sending shutdown request to admin endpoint (port %d)...\n", adminPort)

	url := fmt.Sprintf("https://127.0.0.1:%d/admin/v1/shutdown", adminPort)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send shutdown request: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Shutdown request failed with status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Shutdown request sent successfully\n")
}

// statusService checks CA service readiness.
func statusService(parameters []string) { //nolint:wsl_v5
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, "Checking service status (port %d)...\n", adminPort)

	url := fmt.Sprintf("https://127.0.0.1:%d/admin/v1/readyz", adminPort)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Do(req) // #nosec G107 - URL constructed from localhost and controlled adminPort
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check service status: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		fmt.Fprintf(os.Stderr, "Service is ready\n")
	} else {
		fmt.Fprintf(os.Stderr, "Service is not ready (status: %d)\n", resp.StatusCode)
		os.Exit(1)
	}
}

// healthService checks CA service health.
func healthService(parameters []string) { //nolint:wsl_v5
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, "Checking service health (port %d)...\n", adminPort)

	url := fmt.Sprintf("https://127.0.0.1:%d/admin/v1/livez", adminPort)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Do(req) // #nosec G107 - URL constructed from localhost and controlled adminPort
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check service health: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		fmt.Fprintf(os.Stderr, "Service is healthy\n")
	} else {
		fmt.Fprintf(os.Stderr, "Service is not healthy (status: %d)\n", resp.StatusCode)
		os.Exit(1)
	}
}

// parseConfigFlag extracts --config or -c flag value.
func parseConfigFlag(parameters []string, defaultValue string) string {
	for i := 0; i < len(parameters); i++ {
		param := parameters[i]

		if param == configFlag || param == configFlagShort {
			if i+1 < len(parameters) {
				return parameters[i+1]
			}
		}

		if strings.HasPrefix(param, configFlag+"=") {
			return strings.TrimPrefix(param, configFlag+"=")
		}

		if strings.HasPrefix(param, configFlagShort+"=") {
			return strings.TrimPrefix(param, configFlagShort+"=")
		}
	}

	return defaultValue
}

// parseAdminPort extracts admin port from parameters.
func parseAdminPort(parameters []string, defaultValue int) int {
	for i, param := range parameters {
		if param == "--admin-port" && i+1 < len(parameters) {
			if port, err := strconv.Atoi(parameters[i+1]); err == nil && port > 0 && port <= 65535 {
				return port
			}
		}
	}

	return defaultValue
}
