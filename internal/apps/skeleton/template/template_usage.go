// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use.

package template

const (
	// TemplateUsageMain is the main usage message for the skeleton template command.
	TemplateUsageMain = `Usage: skeleton template <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Skeleton Template server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "skeleton template <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

	// TemplateUsageServer is the usage message for the server subcommand.
	TemplateUsageServer = `Usage: skeleton template server [options]

Description:
  Start the Skeleton Template server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL       Database URL (default: SQLite in-memory)
                           SQLite: file::memory:?cache=shared
                           PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --help, -h               Show this help message

Examples:
  skeleton template server
  skeleton template server --database-url file:/tmp/skeleton.db`

	// TemplateUsageClient is the usage message for the client subcommand.
	TemplateUsageClient = `Usage: skeleton template client [options]

Description:
  Run client operations for the Skeleton Template service.

Options:
  --help, -h    Show this help message

Examples:
  skeleton template client`

	// TemplateUsageInit is the usage message for the init subcommand.
	TemplateUsageInit = `Usage: skeleton template init [options]

Description:
  Initialize database schema and configuration for the Skeleton Template service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  skeleton template init
  skeleton template init --config configs/skeleton/template/config.yml`

	// TemplateUsageHealth is the usage message for the health subcommand.
	TemplateUsageHealth = `Usage: skeleton template health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8900)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  skeleton template health
  skeleton template health --url https://localhost:8900
  skeleton template health --cacert /path/to/ca.pem`

	// TemplateUsageLivez is the usage message for the livez subcommand.
	TemplateUsageLivez = `Usage: skeleton template livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  skeleton template livez
  skeleton template livez --url https://localhost:9090`

	// TemplateUsageReadyz is the usage message for the readyz subcommand.
	TemplateUsageReadyz = `Usage: skeleton template readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  skeleton template readyz
  skeleton template readyz --url https://localhost:9090`

	// TemplateUsageShutdown is the usage message for the shutdown subcommand.
	TemplateUsageShutdown = `Usage: skeleton template shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  skeleton template shutdown
  skeleton template shutdown --url https://localhost:9090
  skeleton template shutdown --force`
)
