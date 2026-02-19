// Copyright (c) 2025 Justin Cranford
//
//

package ja

const (
// JAUsageMain is the main usage message for the jose ja command.
JAUsageMain = `Usage: jose ja <subcommand> [options]

Available subcommands:
  version     Print version information
  server      Start the JWK Authority server (default)
  client      Run client operations
  init        Initialize database and configuration
  health      Check service health (public API)
  livez       Check service liveness (admin API)
  readyz      Check service readiness (admin API)
  shutdown    Trigger graceful shutdown (admin API)

Use "jose ja <subcommand> help" for subcommand-specific help.
Version information is available via Docker image tags.`

// JAUsageServer is the usage message for the server subcommand.
JAUsageServer = `Usage: jose ja server [options]

Description:
  Start the JWK Authority server with database initialization.
  Supports both SQLite (default) and PostgreSQL databases.

Options:
  --database-url URL       Database URL (default: SQLite in-memory)
                           SQLite: file::memory:?cache=shared
                           PostgreSQL: postgres://user:pass@host:port/dbname?sslmode=disable
  --max-materials INT      Default maximum material keys per elastic key
  --audit-enabled BOOL     Enable audit logging for JWK operations
  --audit-sampling-rate INT Audit sampling rate (0-100, percentage)
  --help, -h               Show this help message

Examples:
  jose ja server
  jose ja server --database-url file:/tmp/jose.db`

// JAUsageClient is the usage message for the client subcommand.
JAUsageClient = `Usage: jose ja client [options]

Description:
  Run client operations for the JWK Authority service.

Options:
  --help, -h    Show this help message

Examples:
  jose ja client`

// JAUsageInit is the usage message for the init subcommand.
JAUsageInit = `Usage: jose ja init [options]

Description:
  Initialize database schema and configuration for the JWK Authority service.

Options:
  --config PATH    Configuration file path
  --help, -h       Show this help message

Examples:
  jose ja init
  jose ja init --config configs/jose/ja/config.yml`

// JAUsageHealth is the usage message for the health subcommand.
JAUsageHealth = `Usage: jose ja health [options]

Description:
  Check service health via public API endpoint.
  Calls GET /browser/api/v1/health endpoint on the public server.

Options:
  --url URL      Service URL (default: https://127.0.0.1:8800)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  jose ja health
  jose ja health --url https://localhost:8800
  jose ja health --cacert /path/to/ca.pem`

// JAUsageLivez is the usage message for the livez subcommand.
JAUsageLivez = `Usage: jose ja livez [options]

Description:
  Check service liveness via admin API endpoint.
  Calls GET /admin/api/v1/livez endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  jose ja livez
  jose ja livez --url https://localhost:9090`

// JAUsageReadyz is the usage message for the readyz subcommand.
JAUsageReadyz = `Usage: jose ja readyz [options]

Description:
  Check service readiness via admin API endpoint.
  Calls GET /admin/api/v1/readyz endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --help, -h     Show this help message

Examples:
  jose ja readyz
  jose ja readyz --url https://localhost:9090`

// JAUsageShutdown is the usage message for the shutdown subcommand.
JAUsageShutdown = `Usage: jose ja shutdown [options]

Description:
  Trigger graceful shutdown via admin API endpoint.
  Calls POST /admin/api/v1/shutdown endpoint on the admin server.

Options:
  --url URL      Admin URL (default: https://127.0.0.1:9090)
  --cacert FILE  CA certificate file for TLS validation
  --force        Force shutdown without graceful drain
  --help, -h     Show this help message

Examples:
  jose ja shutdown
  jose ja shutdown --url https://localhost:9090
  jose ja shutdown --force`
)
