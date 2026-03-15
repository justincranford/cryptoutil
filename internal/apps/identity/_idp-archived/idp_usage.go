// Copyright (c) 2025 Justin Cranford
//
//

package idp

const (
	// IDPUsageMain is the main usage message for the identity idp command.
	IDPUsageMain = `Usage: identity idp <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Identity Provider server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "identity idp <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// IDPUsageServer is the usage message for the server subcommand.
	IDPUsageServer = `Usage: identity idp server [options]

Description:
  Start the Identity Provider server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  identity idp server
  identity idp server --config configs/identity/idp/config.yml`

	// IDPUsageClient is the usage message for the client subcommand.
	IDPUsageClient = `Usage: identity idp client [options]

Description:
  Run client operations for the Identity Provider service.

Options:
  --help, -h    Show this help message

Examples:
  identity idp client`

	// IDPUsageInit is the usage message for the init subcommand.
	IDPUsageInit = `Usage: identity idp init [options]

Description:
  Initialize database schema and configuration for the Identity Provider service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  identity idp init
  identity idp init --config configs/identity/idp/config.yml`

	// IDPUsageHealth is the usage message for the health subcommand.
	IDPUsageHealth = `Usage: identity idp health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:idp)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity idp health
  identity idp health --cacert /path/to/ca.pem`

	// IDPUsageLivez is the usage message for the livez subcommand.
	IDPUsageLivez = `Usage: identity idp livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity idp livez
  identity idp livez --url https://localhost:9090`

	// IDPUsageReadyz is the usage message for the readyz subcommand.
	IDPUsageReadyz = `Usage: identity idp readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  identity idp readyz
  identity idp readyz --url https://localhost:9090`

	// IDPUsageShutdown is the usage message for the shutdown subcommand.
	IDPUsageShutdown = `Usage: identity idp shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  identity idp shutdown
  identity idp shutdown --url https://localhost:9090
  identity idp shutdown --force`
)
