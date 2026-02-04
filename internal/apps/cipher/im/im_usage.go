// Copyright (c) 2025 Justin Cranford
//
//

package im

const (
	// IMUsageMain is the main usage message for the cipher im command.
	IMUsageMain = `Usage: cipher im <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the instant messaging server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "learn im <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// IMUsageServer is the usage message for the server subcommand.
	IMUsageServer = `Usage: cipher im server [options]

Description:
  Start the instant messaging server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: sqliteInMemoryURL
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --help, -h            Show this help message

Examples:
  learn im server
  learn im server --database-url file:/tmp/cipher.db`

	// IMUsageClient is the usage message for the client subcommand.
	IMUsageClient = `Usage: cipher im client [options]

Description:
  Run client operations for instant messaging service.

Options:
  --help, -h    Show this help message

Examples:
  learn im client`

	// IMUsageInit is the usage message for the init subcommand.
	IMUsageInit = `Usage: cipher im init [options]

Description:
  Initialize database schema and configuration for instant messaging service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  learn im init
  learn im init --config configs/learn/im/config.yml`

	// IMUsageHealth is the usage message for the health subcommand.
	IMUsageHealth = `Usage: cipher im health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8070)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  learn im health
  learn im health --url https://localhost:8070
  learn im health --cacert /path/to/ca.pem`

	// IMUsageLivez is the usage message for the livez subcommand.
	IMUsageLivez = `Usage: cipher im livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  learn im livez
  learn im livez --url https://localhost:9090`

	// IMUsageReadyz is the usage message for the readyz subcommand.
	IMUsageReadyz = `Usage: cipher im readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  learn im readyz
  learn im readyz --url https://localhost:9090`

	// IMUsageShutdown is the usage message for the shutdown subcommand.
	IMUsageShutdown = `Usage: cipher im shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  learn im shutdown
  learn im shutdown --url https://localhost:9090
  learn im shutdown --force`
)
