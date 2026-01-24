// Copyright (c) 2025 Justin Cranford
//
//

// Package jose provides the unified command interface for JOSE Authority service.
package jose

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilJoseServer "cryptoutil/internal/jose/server"
)

const (
	configFlag       = "--config"
	configFlagShort  = "-c"
	defaultAdminPort = 9090
	fileURLPrefix    = "file://"
)

// Execute handles JOSE service commands matching KMS/Identity pattern.
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
	fmt.Println("Usage: cryptoutil jose <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  start    - Start JOSE Authority service")
	fmt.Println("  stop     - Stop JOSE Authority service")
	fmt.Println("  status   - Check service status")
	fmt.Println("  health   - Check service health")
	fmt.Println("\nOptions:")
	fmt.Println("  --config, -c <path>  - Config file path (default: /app/run/jose-docker.yml)")
}

// startService starts the JOSE Authority service.
func startService(parameters []string) {
	configFile := parseConfigFlag(parameters, "/app/run/jose-docker.yml")

	fmt.Fprintf(os.Stderr, "Starting JOSE Authority service\n")
	fmt.Fprintf(os.Stderr, "Using config file: %s\n", configFile)

	// Load configuration from YAML file
	parseArgs := []string{"start", "--config", configFile}

	settings, err := cryptoutilAppsTemplateServiceConfig.Parse(parseArgs, false)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configFile, err)
	}

	// Create context
	ctx := context.Background()

	// Create JOSE application
	app, err := cryptoutilJoseServer.NewApplication(ctx, settings)
	if err != nil {
		log.Fatalf("Failed to create JOSE application: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting JOSE server...\n")

	// Start application (blocks until shutdown)
	if err := app.Start(ctx); err != nil {
		log.Fatalf("JOSE server error: %v", err)
	}
}

// stopService sends shutdown request to JOSE admin endpoint.
func stopService(parameters []string) {
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, "Sending shutdown request to admin endpoint (port %d)...\n", adminPort)

	// TODO: Implement HTTP POST to https://127.0.0.1:<adminPort>/admin/v1/shutdown
	fmt.Fprintf(os.Stderr, "TODO: HTTP client implementation pending\n")
	os.Exit(1)
}

// statusService checks JOSE service readiness.
func statusService(parameters []string) {
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, "Checking service status (port %d)...\n", adminPort)

	// TODO: Implement HTTP GET to https://127.0.0.1:<adminPort>/admin/v1/readyz
	fmt.Fprintf(os.Stderr, "TODO: HTTP client implementation pending\n")
	os.Exit(1)
}

// healthService checks JOSE service health.
func healthService(parameters []string) {
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, "Checking service health (port %d)...\n", adminPort)

	// TODO: Implement HTTP GET to https://127.0.0.1:<adminPort>/admin/v1/livez
	fmt.Fprintf(os.Stderr, "TODO: HTTP client implementation pending\n")
	os.Exit(1)
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
