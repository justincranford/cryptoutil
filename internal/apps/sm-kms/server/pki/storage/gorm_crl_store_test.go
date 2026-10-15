// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testInvalidSQLiteVersionError = "invalid SQLite version"

func TestRetryableCRLSequenceError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dialect string
		err     error
		want    bool
	}{
		{name: "nil error", dialect: cryptoutilSharedMagic.TestDatabaseSQLite, err: nil, want: false},
		{name: "sqlite locked", dialect: cryptoutilSharedMagic.TestDatabaseSQLite, err: errors.New("database is locked"), want: true},
		{name: "sqlite busy", dialect: cryptoutilSharedMagic.TestDatabaseSQLite, err: errors.New("database is busy"), want: true},
		{name: "sqlite unrelated", dialect: cryptoutilSharedMagic.TestDatabaseSQLite, err: errors.New("constraint failed"), want: false},
		{
			name:    "postgres deadlock sqlstate",
			dialect: cryptoutilSharedMagic.DockerServicePostgres,
			err:     &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeDeadlockDetected},
			want:    true,
		},
		{
			name:    "wrapped postgres serialization sqlstate",
			dialect: cryptoutilSharedMagic.DockerServicePostgres,
			err:     fmt.Errorf("query failed: %w", &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeSerializationFailure}),
			want:    true,
		},
		{name: "postgres deadlock text fallback", dialect: cryptoutilSharedMagic.DockerServicePostgres, err: errors.New("deadlock detected"), want: true},
		{name: "postgres serialization text fallback", dialect: cryptoutilSharedMagic.DockerServicePostgres, err: errors.New("serialization failure"), want: true},
		{
			name:    "postgres nonretryable sqlstate",
			dialect: cryptoutilSharedMagic.DockerServicePostgres,
			err:     &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeUniqueViolation},
			want:    false,
		},
		{name: "unknown dialect", dialect: "mysql", err: errors.New("database is locked"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, retryableCRLSequenceError(tc.err, tc.dialect))
		})
	}
}

func TestValidateSQLiteUpdateReturningVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		wantErr string
	}{
		{name: "minimum version", version: cryptoutilSharedMagic.SQLiteMinimumVersionUpdateReturning},
		{name: "newer major version", version: "4.0.0"},
		{name: "newer minor version", version: "3.36.0"},
		{name: "older version", version: "3.34.9", wantErr: "3.35.0 or newer"},
		{name: "malformed version", version: "3.x.0", wantErr: testInvalidSQLiteVersionError},
		{name: "missing version component", version: "3.35", wantErr: testInvalidSQLiteVersionError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateSQLiteUpdateReturningVersion(tc.version)
			if tc.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestAllocateCRLNumberWithRetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dialect      string
		retryErr     error
		failures     int
		wantAttempts int
		wantWaits    int
		wantNumber   int64
		wantErr      string
	}{
		{
			name:         "sqlite retries are bounded",
			dialect:      cryptoutilSharedMagic.TestDatabaseSQLite,
			retryErr:     errors.New("database is locked"),
			failures:     crlSequenceAllocationAttempts,
			wantAttempts: crlSequenceAllocationAttempts,
			wantWaits:    crlSequenceAllocationAttempts - 1,
			wantErr:      fmt.Sprintf("after %d attempts", crlSequenceAllocationAttempts),
		},
		{
			name:         "postgres SQLSTATE retries are bounded",
			dialect:      cryptoutilSharedMagic.DockerServicePostgres,
			retryErr:     &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeDeadlockDetected},
			failures:     crlSequenceAllocationAttempts,
			wantAttempts: crlSequenceAllocationAttempts,
			wantWaits:    crlSequenceAllocationAttempts - 1,
			wantErr:      fmt.Sprintf("after %d attempts", crlSequenceAllocationAttempts),
		},
		{
			name:         "successful allocation after retry",
			dialect:      cryptoutilSharedMagic.DockerServicePostgres,
			retryErr:     &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeSerializationFailure},
			failures:     2,
			wantAttempts: 3,
			wantWaits:    2,
			wantNumber:   42,
		},
		{
			name:         "nonretryable error returns immediately",
			dialect:      cryptoutilSharedMagic.DockerServicePostgres,
			retryErr:     &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeUniqueViolation},
			failures:     1,
			wantAttempts: 1,
			wantErr:      cryptoutilSharedMagic.PGCodeUniqueViolation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			attempts := 0
			waits := 0
			gotNumber, err := allocateCRLNumberWithRetry(
				context.Background(),
				tc.dialect,
				func() (int64, error) {
					attempts++
					if attempts <= tc.failures {
						return 0, tc.retryErr
					}

					return tc.wantNumber, nil
				},
				func(_ context.Context, attempt int) error {
					require.Equal(t, waits, attempt)
					waits++

					return nil
				},
			)

			require.Equal(t, tc.wantAttempts, attempts)
			require.Equal(t, tc.wantWaits, waits)

			if tc.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tc.wantNumber, gotNumber)

				return
			}

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
