// Copyright (c) 2025 Justin Cranford
//
//

package ca

const (
	// CAUsageMain is the main usage message for the pki ca command.
	CAUsageMain = `Usage: pki ca <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Certificate Authority server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "pki ca <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// CAUsageServer is the usage message for the server subcommand.
	CAUsageServer = `Usage: pki ca server [options]

Description:
  Start the Certificate Authority server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --dev                 Enable development mode (relaxed security)
  --help, -h            Show this help message

Examples:
  pki ca server
  pki ca server --config configs/pki/ca/config.yml
  pki ca server --dev`

	// CAUsageClient is the usage message for the client subcommand.
	CAUsageClient = `Usage: pki ca client [options]

Description:
  Run client operations for the Certificate Authority service.

Options:
  --help, -h    Show this help message

Examples:
  pki ca client`

	// CAUsageInit is the usage message for the init subcommand.
	CAUsageInit = `Usage: pki ca init [options]

Description:
  Initialize database schema and configuration for the Certificate Authority service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  pki ca init
  pki ca init --config configs/pki/ca/config.yml`

	// CAUsageHealth is the usage message for the health subcommand.
	CAUsageHealth = `Usage: pki ca health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8100)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  pki ca health
  pki ca health --url https://localhost:8100
  pki ca health --cacert /path/to/ca.pem`

	// CAUsageLivez is the usage message for the livez subcommand.
	CAUsageLivez = `Usage: pki ca livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  pki ca livez
  pki ca livez --url https://localhost:9090`

	// CAUsageReadyz is the usage message for the readyz subcommand.
	CAUsageReadyz = `Usage: pki ca readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  pki ca readyz
  pki ca readyz --url https://localhost:9090`

	// CAUsageShutdown is the usage message for the shutdown subcommand.
	CAUsageShutdown = `Usage: pki ca shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  pki ca shutdown
  pki ca shutdown --url https://localhost:9090
  pki ca shutdown --force`
)
