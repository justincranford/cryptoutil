// Copyright (c) 2025 Justin Cranford
//
//

// Package spa provides the unified command interface for Identity SPA service.
package spa //nolint:wsl_v5

import (
	"context"
	"fmt"
	"log"
	http "net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cryptoutilAppsIdentitySpaServer "cryptoutil/internal/apps/identity/spa/server"
	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity/spa/server/config"
)

const (
	configFlag       = "--config"
	configFlagShort  = "-c"
	defaultAdminPort = 9090
	fileURLPrefix    = "file://"
	httpTimeout      = 5 * time.Second
)

// Execute handles SPA service commands matching other service patterns.
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
	fmt.Println("Usage: cryptoutil identity-spa <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  start    - Start Identity SPA service")
	fmt.Println("  stop     - Stop Identity SPA service")
	fmt.Println("  status   - Check service status")
	fmt.Println("  health   - Check service health")
	fmt.Println("\nOptions:")
	fmt.Println("  --config, -c <path>  - Config file path (default: /app/run/identity-spa-docker.yml)")
}

// startService starts the Identity SPA service.
func startService(parameters []string) {
	configFile := parseConfigFlag(parameters, "/app/run/identity-spa-docker.yml")

	fmt.Fprintf(os.Stderr, "Starting Identity SPA service\n")
	fmt.Fprintf(os.Stderr, "Using config file: %s\n", configFile)

	// Load SPA-specific configuration from YAML file.
	parseArgs := []string{"start", "--config", configFile}

	settings, err := cryptoutilAppsIdentitySpaServerConfig.Parse(parseArgs, false)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configFile, err)
	}

	// Create context.
	ctx := context.Background()

	// Create SPA server using template-based implementation.
	server, err := cryptoutilAppsIdentitySpaServer.NewFromConfig(ctx, settings)
	if err != nil {
		log.Fatalf("Failed to create SPA server: %v", err)
	}

	// Mark server as ready.
	server.SetReady(true)

	fmt.Fprintf(os.Stderr, "Starting SPA server...\n")

	// Start server (blocks until shutdown).
	if err := server.Start(ctx); err != nil {
		log.Fatalf("SPA server error: %v", err)
	}
}

// stopService sends shutdown request to SPA admin endpoint.
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

// statusService checks SPA service readiness.
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

// healthService checks SPA service health.
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

// parseConfigFlag extracts config file path from parameters.
func parseConfigFlag(parameters []string, defaultValue string) string {
	for i, param := range parameters {
		// Handle --config=/path/to/file format (single element with =).
		if strings.HasPrefix(param, configFlag+"=") {
			value := strings.TrimPrefix(param, configFlag+"=")

			return strings.TrimPrefix(value, fileURLPrefix)
		}

		if strings.HasPrefix(param, configFlagShort+"=") {
			value := strings.TrimPrefix(param, configFlagShort+"=")

			return strings.TrimPrefix(value, fileURLPrefix)
		}
		// Handle --config /path/to/file format (two elements).
		if (param == configFlag || param == configFlagShort) && i+1 < len(parameters) {
			// Handle file:// prefix for Docker secrets.
			return strings.TrimPrefix(parameters[i+1], fileURLPrefix)
		}
	}

	return defaultValue
}

// parseAdminPort extracts admin port from parameters.
func parseAdminPort(parameters []string, defaultValue int) int {
	for i, param := range parameters {
		if param == "--admin-port" && i+1 < len(parameters) {
			port, err := strconv.Atoi(parameters[i+1])
			if err == nil {
				return port
			}
		}
	}

	return defaultValue
}
