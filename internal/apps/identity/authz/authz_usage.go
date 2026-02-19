// Copyright (c) 2025 Justin Cranford
//
//

package authz

const (
	// AUTHZUsageMain is the main usage message for the identity authz command.
	AUTHZUsageMain = `Usage: identity authz <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Authorization Server server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "identity authz <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// AUTHZUsageServer is the usage message for the server subcommand.
	AUTHZUsageServer = `Usage: identity authz server [options]

Description:
  Start the Authorization Server server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  identity authz server
  identity authz server --config configs/identity/authz/config.yml`

	// AUTHZUsageClient is the usage message for the client subcommand.
	AUTHZUsageClient = `Usage: identity authz client [options]

Description:
  Run client operations for the Authorization Server service.

Options:
  --help, -h    Show this help message

Examples:
  identity authz client`

	// AUTHZUsageInit is the usage message for the init subcommand.
	AUTHZUsageInit = `Usage: identity authz init [options]

Description:
  Initialize database schema and configuration for the Authorization Server service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  identity authz init
  identity authz init --config configs/identity/authz/config.yml`

	// AUTHZUsageHealth is the usage message for the health subcommand.
	AUTHZUsageHealth = `Usage: identity authz health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:authz)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity authz health
  identity authz health --cacert /path/to/ca.pem`

	// AUTHZUsageLivez is the usage message for the livez subcommand.
	AUTHZUsageLivez = `Usage: identity authz livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity authz livez
  identity authz livez --url https://localhost:9090`

	// AUTHZUsageReadyz is the usage message for the readyz subcommand.
	AUTHZUsageReadyz = `Usage: identity authz readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity authz readyz
  identity authz readyz --url https://localhost:9090`

	// AUTHZUsageShutdown is the usage message for the shutdown subcommand.
	AUTHZUsageShutdown = `Usage: identity authz shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  identity authz shutdown
  identity authz shutdown --url https://localhost:9090
  identity authz shutdown --force`
)
