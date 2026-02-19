// Copyright (c) 2025 Justin Cranford
//
//

package spa

const (
	// SPAUsageMain is the main usage message for the identity spa command.
	SPAUsageMain = `Usage: identity spa <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Single Page Application server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "identity spa <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// SPAUsageServer is the usage message for the server subcommand.
	SPAUsageServer = `Usage: identity spa server [options]

Description:
  Start the Single Page Application server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  identity spa server
  identity spa server --config configs/identity/spa/config.yml`

	// SPAUsageClient is the usage message for the client subcommand.
	SPAUsageClient = `Usage: identity spa client [options]

Description:
  Run client operations for the Single Page Application service.

Options:
  --help, -h    Show this help message

Examples:
  identity spa client`

	// SPAUsageInit is the usage message for the init subcommand.
	SPAUsageInit = `Usage: identity spa init [options]

Description:
  Initialize database schema and configuration for the Single Page Application service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  identity spa init
  identity spa init --config configs/identity/spa/config.yml`

	// SPAUsageHealth is the usage message for the health subcommand.
	SPAUsageHealth = `Usage: identity spa health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:spa)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity spa health
  identity spa health --cacert /path/to/ca.pem`

	// SPAUsageLivez is the usage message for the livez subcommand.
	SPAUsageLivez = `Usage: identity spa livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity spa livez
  identity spa livez --url https://localhost:9090`

	// SPAUsageReadyz is the usage message for the readyz subcommand.
	SPAUsageReadyz = `Usage: identity spa readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity spa readyz
  identity spa readyz --url https://localhost:9090`

	// SPAUsageShutdown is the usage message for the shutdown subcommand.
	SPAUsageShutdown = `Usage: identity spa shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  identity spa shutdown
  identity spa shutdown --url https://localhost:9090
  identity spa shutdown --force`
)
