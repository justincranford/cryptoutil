// Copyright (c) 2025 Justin Cranford

// Package usage provides shared builder functions for PS-ID CLI usage strings.
// All 10 PS-ID services and their product-level wrappers use these helpers to
// eliminate const-redefine linter violations that arise when the same usage
// string is declared as a const in both the PS-ID package and its product
// package.
package usage

import "fmt"

// BuildUsageMain returns the main usage string for a PS-ID CLI entry point.
// productName is the top-level product command (e.g., "sm", "jose", "pki").
// serviceName is the service subcommand (e.g., "kms", "im", "ja", "ca").
// serviceDisplayName is the human-readable service name (e.g., "Key Management Service").
func BuildUsageMain(productName, serviceName, serviceDisplayName string) string {
	return fmt.Sprintf(`Usage: %s %s <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the %s server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "%s %s <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`,
		productName, serviceName, serviceDisplayName,
		productName, serviceName)
}

// BuildUsageServer returns the server subcommand usage string.
// configFilePath is the example config file path (e.g., "configs/sm-kms/sm-kms-framework.yml").
func BuildUsageServer(productName, serviceName, serviceDisplayName, configFilePath string) string {
	return fmt.Sprintf(`Usage: %s %s server [options]

Description:
  Start the %s server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  %s %s server
  %s %s server --config %s`,
		productName, serviceName, serviceDisplayName,
		productName, serviceName,
		productName, serviceName, configFilePath)
}

// BuildUsageClient returns the client subcommand usage string.
func BuildUsageClient(productName, serviceName, serviceDisplayName string) string {
	return fmt.Sprintf(`Usage: %s %s client [options]

Description:
  Run client operations for the %s.

Options:
  --help, -h    Show this help message

Examples:
  %s %s client`,
		productName, serviceName, serviceDisplayName,
		productName, serviceName)
}

// BuildUsageInit returns the init subcommand usage string.
// configFilePath is the example config file path (e.g., "configs/sm-kms/sm-kms-framework.yml").
func BuildUsageInit(productName, serviceName, serviceDisplayName, configFilePath string) string {
	return fmt.Sprintf(`Usage: %s %s init [options]

Description:
  Initialize database schema and configuration for the %s.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  %s %s init
  %s %s init --config %s`,
		productName, serviceName, serviceDisplayName,
		productName, serviceName,
		productName, serviceName, configFilePath)
}

// BuildUsageHealth returns the health subcommand usage string.
// defaultPublicPort is the service's default public port (e.g., "8000" for sm-kms).
func BuildUsageHealth(productName, serviceName, defaultPublicPort string) string {
	return fmt.Sprintf(`Usage: %s %s health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /service/api/v1/health endpoint by service-to-service clients.
  Calls GET /browser/api/v1/health endpoint by browser clients.

Options:
  --url URL      Service URL (default: https://127.0.0.1:%s)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  %s %s health
  %s %s health --url https://localhost:%s
  %s %s health --cacert /path/to/ca.pem`,
		productName, serviceName, defaultPublicPort,
		productName, serviceName,
		productName, serviceName, defaultPublicPort,
		productName, serviceName)
}

// BuildUsageLivez returns the livez subcommand usage string.
// The admin port is always 9090 per the architecture.
func BuildUsageLivez(productName, serviceName string) string {
	return fmt.Sprintf(`Usage: %s %s livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  %s %s livez
  %s %s livez --url https://localhost:9090`,
		productName, serviceName,
		productName, serviceName,
		productName, serviceName)
}

// BuildUsageReadyz returns the readyz subcommand usage string.
// The admin port is always 9090 per the architecture.
func BuildUsageReadyz(productName, serviceName string) string {
	return fmt.Sprintf(`Usage: %s %s readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  %s %s readyz
  %s %s readyz --url https://localhost:9090`,
		productName, serviceName,
		productName, serviceName,
		productName, serviceName)
}

// BuildUsageShutdown returns the shutdown subcommand usage string.
// The admin port is always 9090 per the architecture.
func BuildUsageShutdown(productName, serviceName string) string {
	return fmt.Sprintf(`Usage: %s %s shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  %s %s shutdown
  %s %s shutdown --url https://localhost:9090
  %s %s shutdown --force`,
		productName, serviceName,
		productName, serviceName,
		productName, serviceName,
		productName, serviceName)
}
