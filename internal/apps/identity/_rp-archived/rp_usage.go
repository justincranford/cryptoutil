// Copyright (c) 2025 Justin Cranford
//
//

package rp

const (
	// RPUsageMain is the main usage message for the identity rp command.
	RPUsageMain = `Usage: identity rp <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Relying Party server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "identity rp <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// RPUsageServer is the usage message for the server subcommand.
	RPUsageServer = `Usage: identity rp server [options]

Description:
  Start the Relying Party server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  identity rp server
  identity rp server --config configs/identity/rp/config.yml`

	// RPUsageClient is the usage message for the client subcommand.
	RPUsageClient = `Usage: identity rp client [options]

Description:
  Run client operations for the Relying Party service.

Options:
  --help, -h    Show this help message

Examples:
  identity rp client`

	// RPUsageInit is the usage message for the init subcommand.
	RPUsageInit = `Usage: identity rp init [options]

Description:
  Initialize database schema and configuration for the Relying Party service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  identity rp init
  identity rp init --config configs/identity/rp/config.yml`

	// RPUsageHealth is the usage message for the health subcommand.
	RPUsageHealth = `Usage: identity rp health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:rp)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity rp health
  identity rp health --cacert /path/to/ca.pem`

	// RPUsageLivez is the usage message for the livez subcommand.
	RPUsageLivez = `Usage: identity rp livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity rp livez
  identity rp livez --url https://localhost:9090`

	// RPUsageReadyz is the usage message for the readyz subcommand.
	RPUsageReadyz = `Usage: identity rp readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity rp readyz
  identity rp readyz --url https://localhost:9090`

	// RPUsageShutdown is the usage message for the shutdown subcommand.
	RPUsageShutdown = `Usage: identity rp shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  identity rp shutdown
  identity rp shutdown --url https://localhost:9090
  identity rp shutdown --force`
)
