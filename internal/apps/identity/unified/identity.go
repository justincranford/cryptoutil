// Copyright (c) 2025 Justin Cranford
//
//

// Package identity provides the unified command interface for Identity services.
package identity

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	cryptoutilIdentityAuthzServer "cryptoutil/internal/identity/authz/server"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIdpServer "cryptoutil/internal/identity/idp/server"
	cryptoutilIdentityRsServer "cryptoutil/internal/identity/rs/server"
)

const (
	configFlag      = "--config"
	configFlagShort = "-c"
	serviceFlag     = "--service"
	dsnFlag         = "-u"
	dsnFlagLong     = "--database-url"
	fileURLPrefix   = "file://"
)

// Unified handles Identity service commands matching KMS pattern.
// Supports: start, stop, status, health.
func Unified(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	parameters := args[1:]
	if len(parameters) < 1 {
		printUsage()

		return 1
	}

	subcommand := parameters[0]
	cmdParams := parameters[1:]

	switch subcommand {
	case "start":
		startServices(cmdParams)

		return 0
	case "stop":
		stopServices(cmdParams)

		return 0
	case "status":
		statusServices(cmdParams)

		return 0
	case "health":
		healthServices(cmdParams)

		return 0
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()

		return 1
	}
}

func printUsage() {
	fmt.Println("Usage: cryptoutil identity <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  start    - Start Identity services (authz, idp, rs)")
	fmt.Println("  stop     - Stop Identity services")
	fmt.Println("  status   - Check service status")
	fmt.Println("  health   - Check service health")
	fmt.Println("\nOptions:")
	fmt.Println("  --config, -c <path>  - Config file path (default: /app/run/*.yml)")
	fmt.Println("  --service <name>     - Specific service (authz, idp, rs, all)")
	fmt.Println("  -u, --database-url   - Database DSN (supports file:// for Docker secrets)")
}

// startServices starts one or all Identity services.
func startServices(parameters []string) {
	service := parseServiceFlag(parameters, "authz") // Default to authz for backward compatibility
	configFile := parseConfigFlag(parameters, fmt.Sprintf("/app/run/%s-docker.yml", service))

	fmt.Fprintf(os.Stderr, "Starting Identity service: %s\n", service)
	fmt.Fprintf(os.Stderr, "Using config file: %s\n", configFile)

	// Load configuration from YAML file
	config, err := cryptoutilIdentityConfig.LoadFromFile(configFile)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configFile, err)
	}

	// Override DSN from command line if provided (-u flag for Docker secrets)
	if dsn := parseDSNFlag(parameters); dsn != "" {
		fmt.Fprintf(os.Stderr, "Using DSN from command line flag\n")

		config.Database.DSN = dsn
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create context
	ctx := context.Background()

	// Start the appropriate service (Application handles its own initialization)
	switch service {
	case "authz":
		startAuthzServer(ctx, config)
	case "idp":
		startIdpServer(ctx, config)
	case "rs":
		startRsServer(ctx, config)
	case "all":
		// TODO: Implement all-in-one mode (requires goroutines and signal handling)
		log.Fatal("All-in-one mode not yet implemented")
	default:
		log.Fatalf("Unknown service: %s (valid: authz, idp, rs, all)", service)
	}
}

func startAuthzServer(ctx context.Context, config *cryptoutilIdentityConfig.Config) {
	app, err := cryptoutilIdentityAuthzServer.NewApplication(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create AuthZ application: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting AuthZ server...\n")

	if err := app.Start(ctx); err != nil {
		log.Fatalf("AuthZ server error: %v", err)
	}
}

func startIdpServer(ctx context.Context, config *cryptoutilIdentityConfig.Config) {
	app, err := cryptoutilIdentityIdpServer.NewApplication(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create IdP application: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting IdP server...\n")

	if err := app.Start(ctx); err != nil {
		log.Fatalf("IdP server error: %v", err)
	}
}

func startRsServer(ctx context.Context, config *cryptoutilIdentityConfig.Config) {
	app, err := cryptoutilIdentityRsServer.NewApplication(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create RS application: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting RS server...\n")

	if err := app.Start(ctx); err != nil {
		log.Fatalf("RS server error: %v", err)
	}
}

const (
	defaultAdminPort = 9090 // Default admin port for Identity services
)

// stopServices stops Identity services via admin endpoint.
func stopServices(parameters []string) {
	service := parseServiceFlag(parameters, "authz")
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Printf("Stopping Identity service: %s (admin port %d)\n", service, adminPort)

	// TODO: Implement admin endpoint shutdown call
	log.Fatal("Stop command not yet implemented - use admin endpoint: POST https://127.0.0.1:9090/admin/v1/shutdown")
}

// statusServices checks Identity service status via admin endpoint.
func statusServices(parameters []string) {
	service := parseServiceFlag(parameters, "authz")
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Printf("Checking status for Identity service: %s (admin port %d)\n", service, adminPort)

	// TODO: Implement admin endpoint status call
	log.Fatal("Status command not yet implemented - use admin endpoint: GET https://127.0.0.1:9090/admin/v1/readyz")
}

// healthServices checks Identity service health via admin endpoint.
func healthServices(parameters []string) {
	service := parseServiceFlag(parameters, "authz")
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Printf("Checking health for Identity service: %s (admin port %d)\n", service, adminPort)

	// TODO: Implement admin endpoint health call
	log.Fatal("Health command not yet implemented - use admin endpoint: GET https://127.0.0.1:9090/admin/v1/healthz")
}

// parseConfigFlag extracts config file path from parameters.
// Supports both "--config /path" and "--config=/path" formats.
func parseConfigFlag(parameters []string, defaultConfig string) string {
	for i, param := range parameters {
		// Support --config /path format
		if param == configFlag || param == configFlagShort {
			if i+1 < len(parameters) {
				return parameters[i+1]
			}
		}

		// Support --config=/path format
		if len(param) > len(configFlag) && param[:len(configFlag)+1] == configFlag+"=" {
			return param[len(configFlag)+1:]
		}

		// Support -c=/path format
		if len(param) > len(configFlagShort) && param[:len(configFlagShort)+1] == configFlagShort+"=" {
			return param[len(configFlagShort)+1:]
		}
	}

	return defaultConfig
}

// parseServiceFlag extracts service name from parameters.
func parseServiceFlag(parameters []string, defaultService string) string {
	for i, param := range parameters {
		// Support --service <name> format
		if param == serviceFlag {
			if i+1 < len(parameters) {
				return parameters[i+1]
			}
		}

		// Support --service=<name> format
		if len(param) > len(serviceFlag) && param[:len(serviceFlag)+1] == serviceFlag+"=" {
			return param[len(serviceFlag)+1:]
		}
	}

	return defaultService
}

// parseAdminPort extracts admin port from parameters.
func parseAdminPort(parameters []string, defaultPort int) int {
	for i, param := range parameters {
		if param == "--admin-port" && i+1 < len(parameters) {
			if port, err := strconv.Atoi(parameters[i+1]); err == nil && port > 0 && port <= 65535 {
				return port
			}
		}
	}

	return defaultPort
}

// parseDSNFlag extracts database URL from parameters.
// Supports both "-u value" and "-u=value" formats.
// If the value starts with "file://", it reads the DSN from that file path.
func parseDSNFlag(parameters []string) string {
	for i, param := range parameters {
		// Support -u /path or --database-url /path format
		if param == dsnFlag || param == dsnFlagLong {
			if i+1 < len(parameters) {
				return resolveDSNValue(parameters[i+1])
			}
		}

		// Support -u=/path format
		if len(param) > len(dsnFlag) && param[:len(dsnFlag)+1] == dsnFlag+"=" {
			return resolveDSNValue(param[len(dsnFlag)+1:])
		}

		// Support --database-url=/path format
		if len(param) > len(dsnFlagLong) && param[:len(dsnFlagLong)+1] == dsnFlagLong+"=" {
			return resolveDSNValue(param[len(dsnFlagLong)+1:])
		}
	}

	return ""
}

// resolveDSNValue resolves a DSN value, reading from file if it's a file:// URL.
func resolveDSNValue(value string) string {
	if strings.HasPrefix(value, fileURLPrefix) {
		filePath := strings.TrimPrefix(value, fileURLPrefix)

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read DSN from file %s: %v\n", filePath, err)

			return ""
		}

		// Trim whitespace (including newlines) from the DSN
		return strings.TrimSpace(string(data))
	}

	return value
}
