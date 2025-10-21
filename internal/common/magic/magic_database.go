// Package magic provides commonly used magic numbers and values as named constants.
// This file contains database-related constants.
package magic

import "time"

// Database connection and query timeouts.
const (
	// DBPingAttemptWait - Initial wait time before first database ping attempt.
	DBPingAttemptWait = 750 * time.Millisecond
	// DBLoggerSlowThreshold - Threshold for logging slow database queries.
	DBLoggerSlowThreshold = 200 * time.Millisecond
	// DBPingTimeout - Default timeout for database ping operations.
	DBPingTimeout = 5 * time.Second
	// SQLiteBusyTimeout - Timeout for SQLite busy operations.
	SQLiteBusyTimeout = 30 * time.Second
)

// Database connection retry and pooling.
const (
	// DBMaxPingAttempts - Maximum number of database ping attempts.
	DBMaxPingAttempts = 5
	// DBNextPingAttemptWait - Wait time between database ping attempts.
	DBNextPingAttemptWait = 1 * time.Second
	// SQLiteMaxOpenConnections - Maximum open connections for SQLite.
	SQLiteMaxOpenConnections = 1
)

// Database random suffix generation.
const (
	// DBRandSuffixMax - Maximum value for random database suffix generation.
	DBRandSuffixMax = 10000
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
