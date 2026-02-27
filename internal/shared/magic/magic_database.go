// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// Pagination defaults.
const (
	// DefaultPageSize - Default page size for pagination.
	DefaultPageSize = 25
)

// Database connection and query timeouts.
const (
	// DBMaxPingAttempts - Maximum number of database ping attempts.
	DBMaxPingAttempts = 5
	// DBPingTimeout - Default timeout for database ping operations.
	DBPingTimeout = 5 * time.Second

	// DBPingFirstAttemptWait - Initial wait time before first database ping attempt.
	DBPingFirstAttemptWait = 750 * time.Millisecond
	// DBPingNextAttemptWait - Wait time between database ping attempts.
	DBPingNextAttemptWait = 1 * time.Second

	// DBLoggerSlowThreshold - Threshold for logging slow database queries.
	DBLoggerSlowThreshold = 200 * time.Millisecond

	// DBSQLiteBusyTimeout - Timeout for SQLite busy operations.
	DBSQLiteBusyTimeout = 30 * time.Second
	// SQLiteMaxOpenConnections - Maximum open connections for SQLite.
	SQLiteMaxOpenConnections = 1

	// SQLiteMaxOpenConnectionsForGORM - Maximum open connections for SQLite when GORM
	// transactions nest barrier operations. GORM services need separate connections for
	// nested transactions (e.g., ORM ReadOnly + barrier DecryptContentWithContext).
	SQLiteMaxOpenConnectionsForGORM = 5

	// PostgreSQLMaxOpenConns - Maximum open connections for PostgreSQL.
	PostgreSQLMaxOpenConns = 25
	// PostgreSQLMaxIdleConns - Maximum idle connections for PostgreSQL.
	PostgreSQLMaxIdleConns = 10
	// PostgreSQLConnMaxLifetime - Maximum connection lifetime for PostgreSQL.
	PostgreSQLConnMaxLifetime = 1 * time.Hour

	// DBPostgresContainerStartupTimeout - PostgreSQL container startup timeout.
	DBPostgresContainerStartupTimeout = 30 * time.Second
	// DefaultDatabaseInitTotalTimeout - Total timeout for database initialization (5 minutes).
	DefaultDatabaseInitTotalTimeout = 5 * time.Minute
	// DefaultDataInitRetryWait - Retry wait time for database initialization (1 second).
	DefaultDataInitRetryWait = 1 * time.Second
	// DefaultDataServerShutdownTimeout - Server shutdown timeout (5 seconds).
	DefaultDataServerShutdownTimeout = 5 * time.Second
)

// Database random suffix generation.
const (
	// DBContainerRandSuffixMax - Maximum value for random database suffix generation.
	DBContainerRandSuffixMax = 10000
)

// Rate limiting configuration.
const (
	// RateLimitRetentionMultiplier - Multiplier for rate limit retention period (e.g., 2x window).
	RateLimitRetentionMultiplier = 2
)

// SQLite error codes.
const (
	// SQLiteErrUniqueConstraint - SQLite unique constraint violation error code.
	SQLiteErrUniqueConstraint = 2067
	// SQLiteErrForeignKey - SQLite foreign key constraint violation error code.
	SQLiteErrForeignKey = 787
	// SQLiteErrCheckConstraint - SQLite check constraint violation error code.
	SQLiteErrCheckConstraint = 1299
)

// PostgreSQL error codes.
const (
	// PGCodeUniqueViolation - PostgreSQL unique violation error code.
	PGCodeUniqueViolation = "23505"
	// PGCodeForeignKeyViolation - PostgreSQL foreign key violation error code.
	PGCodeForeignKeyViolation = "23503"
	// PGCodeCheckViolation - PostgreSQL check constraint violation error code.
	PGCodeCheckViolation = "23514"
	// PGCodeStringDataTruncation - PostgreSQL string data truncation error code.
	PGCodeStringDataTruncation = "22001"
)

const (
	// DefaultDatabaseContainerDisabled - Disabled database container mode.
	DefaultDatabaseContainerDisabled = "disabled"
	// DefaultDatabaseURL - Default database URL with placeholder credentials.
	DefaultDatabaseURL = "postgres://USR:PWD@localhost:5432/DB?sslmode=disable" // pragma: allowlist secret
)

// GORM logger configuration.
const (
	// GormLogModeInfo - GORM logger info level.
	GormLogModeInfo = 4
)

// Test database configurations.
const (
	// TestDatabaseSQLite - SQLite test database configuration name.
	TestDatabaseSQLite = "sqlite"
	// TestDatabasePostgres1 - PostgreSQL test database configuration name 1.
	TestDatabasePostgres1 = "postgres1"
	// TestDatabasePostgres2 - PostgreSQL test database configuration name 2.
	TestDatabasePostgres2 = "postgres2"
)

// SQLite DSN patterns.
const (
	// SQLiteInMemoryDSN - SQLite in-memory DSN for testing.
	SQLiteInMemoryDSN = "file::memory:?cache=shared"
	// SQLiteMemoryPlaceholder - SQLite memory placeholder pattern.
	SQLiteMemoryPlaceholder = ":memory:"
)
