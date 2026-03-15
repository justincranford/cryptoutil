// Copyright (c) 2025 Justin Cranford
//
//

package rs

const (
	// RSUsageMain is the main usage message for the identity rs command.
	RSUsageMain = `Usage: identity rs <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Resource Server server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "identity rs <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// RSUsageServer is the usage message for the server subcommand.
	RSUsageServer = `Usage: identity rs server [options]

Description:
  Start the Resource Server server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  identity rs server
  identity rs server --config configs/identity/rs/config.yml`

	// RSUsageClient is the usage message for the client subcommand.
	RSUsageClient = `Usage: identity rs client [options]

Description:
  Run client operations for the Resource Server service.

Options:
  --help, -h    Show this help message

Examples:
  identity rs client`

	// RSUsageInit is the usage message for the init subcommand.
	RSUsageInit = `Usage: identity rs init [options]

Description:
  Initialize database schema and configuration for the Resource Server service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  identity rs init
  identity rs init --config configs/identity/rs/config.yml`

	// RSUsageHealth is the usage message for the health subcommand.
	RSUsageHealth = `Usage: identity rs health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:rs)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity rs health
  identity rs health --cacert /path/to/ca.pem`

	// RSUsageLivez is the usage message for the livez subcommand.
	RSUsageLivez = `Usage: identity rs livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity rs livez
  identity rs livez --url https://localhost:9090`

	// RSUsageReadyz is the usage message for the readyz subcommand.
	RSUsageReadyz = `Usage: identity rs readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity rs readyz
  identity rs readyz --url https://localhost:9090`

	// RSUsageShutdown is the usage message for the shutdown subcommand.
	RSUsageShutdown = `Usage: identity rs shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  identity rs shutdown
  identity rs shutdown --url https://localhost:9090
  identity rs shutdown --force`
)
