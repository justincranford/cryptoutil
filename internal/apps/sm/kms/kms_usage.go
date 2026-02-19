// Copyright (c) 2025 Justin Cranford
//
//

package kms

const (
// KMSUsageMain is the main usage message for the sm kms command.
KMSUsageMain = `Usage: sm kms <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the Key Management Service server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "sm kms <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

// KMSUsageServer is the usage message for the server subcommand.
KMSUsageServer = `Usage: sm kms server [options]

Description:
  Start the Key Management Service server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL    Database URL (default: SQLite in-memory)
                        SQLite: file::memory:?cache=shared
                        PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --config PATH         Configuration file path
  --help, -h            Show this help message

Examples:
  sm kms server
  sm kms server --config configs/sm/kms/config.yml`

// KMSUsageClient is the usage message for the client subcommand.
KMSUsageClient = `Usage: sm kms client [options]

Description:
  Run client operations for the Key Management Service.

Options:
  --help, -h    Show this help message

Examples:
  sm kms client`

// KMSUsageInit is the usage message for the init subcommand.
KMSUsageInit = `Usage: sm kms init [options]

Description:
  Initialize database schema and configuration for the Key Management Service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  sm kms init
  sm kms init --config configs/sm/kms/config.yml`

// KMSUsageHealth is the usage message for the health subcommand.
KMSUsageHealth = `Usage: sm kms health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8000)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  sm kms health
  sm kms health --url https://localhost:8000
  sm kms health --cacert /path/to/ca.pem`

// KMSUsageLivez is the usage message for the livez subcommand.
KMSUsageLivez = `Usage: sm kms livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  sm kms livez
  sm kms livez --url https://localhost:9090`

// KMSUsageReadyz is the usage message for the readyz subcommand.
KMSUsageReadyz = `Usage: sm kms readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  sm kms readyz
  sm kms readyz --url https://localhost:9090`

// KMSUsageShutdown is the usage message for the shutdown subcommand.
KMSUsageShutdown = `Usage: sm kms shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  sm kms shutdown
  sm kms shutdown --url https://localhost:9090
  sm kms shutdown --force`
)
