// Package magic provides commonly used magic numbers and values as named constants.
// This file contains database-related constants.
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
