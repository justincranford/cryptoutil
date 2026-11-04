// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type errorOutput struct{}

func (errorOutput) Write([]byte) (int, error) { return 0, errors.New("DO-NOT-LOG") }

func TestDatabaseCLIProtectedBoundary(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testHelp, testHelpWrite, testFlags, testFixedPath, testPositionals, testRead, testDecodeFailure, testValidate, testExecute, testOutput, testSuccess} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			input := inputFixture(Migrate)
			if test == testValidate {
				input.Backend = testUnknown
			}

			raw, err := json.Marshal(input)
			require.NoError(t, err)

			args := []string{"--operation=" + Migrate, "--input=" + SecretDirectory + "owned/" + InputName}
			if test == testHelp || test == testHelpWrite {
				args = []string{testHelpFlag}
			}

			if test == testFlags {
				args = []string{"--DO-NOT-LOG=value"}
			}

			if test == testFixedPath {
				args = []string{"--input=/etc/secret"}
			}

			if test == testPositionals {
				args = append(args, testUnowned)
			}

			var (
				stdout, stderr bytes.Buffer
				output         io.Writer = &stdout
			)
			if test == testOutput || test == testHelpWrite {
				output = errorOutput{}
			}

			calls := 0

			code := MainWith(args, output, &stderr, func(string) ([]byte, error) {
				if test == testRead {
					return nil, errors.New("DO-NOT-LOG")
				}

				if test == testDecodeFailure {
					return []byte(testInvalid), nil
				}

				return raw, nil
			}, func(ctx context.Context, operation string, got Input) error {
				calls++

				require.Equal(t, Migrate, operation)
				require.Equal(t, input, got)

				_, deadline := ctx.Deadline()
				require.True(t, deadline)

				if test == testExecute {
					return errors.New("DO-NOT-LOG")
				}

				return nil
			})
			if test == testHelp || test == testSuccess {
				require.Zero(t, code)
			} else {
				require.Equal(t, 1, code)
			}

			require.NotContains(t, stderr.String(), "DO-NOT-LOG")

			if test == testSuccess || test == testOutput || test == testExecute {
				require.Equal(t, 1, calls)
			} else {
				require.Zero(t, calls)
			}
		})
	}
}

type failingReadCloser struct{ readErr, closeErr bool }

func (r failingReadCloser) Read([]byte) (int, error) {
	if r.readErr {
		return 0, errors.New("DO-NOT-LOG")
	}

	return 0, io.EOF
}

func (r failingReadCloser) Close() error {
	if r.closeErr {
		return errors.New("DO-NOT-LOG")
	}

	return nil
}

func TestDatabaseProtectedReads(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testSuccess, testInvalid, testOpen, testRead, testClose, testEmpty, testOversized, testLimit} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			file := SecretDirectory + "owned/key"
			if test == testInvalid {
				file = testRelative
			}

			got, err := readProtectedWith(file, func(string) (io.ReadCloser, error) {
				switch test {
				case testOpen:
					return nil, errors.New("DO-NOT-LOG")
				case testRead:
					return failingReadCloser{readErr: true}, nil
				case testClose:
					return failingReadCloser{closeErr: true}, nil
				case testEmpty:
					return io.NopCloser(strings.NewReader("")), nil
				case testOversized:
					return io.NopCloser(strings.NewReader(strings.Repeat("x", maximumInput+1))), nil
				case testLimit:
					return io.NopCloser(strings.NewReader(strings.Repeat("x", maximumInput))), nil
				default:
					return io.NopCloser(strings.NewReader("protected")), nil
				}
			})
			if test == testSuccess || test == testLimit {
				require.NoError(t, err)
				require.NotEmpty(t, got)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
			}
		})
	}

	require.Error(t, func() error {
		_, err := readProtected("not-absolute")

		return err
	}())
}

func TestDatabaseSuiteHandlerHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	require.Zero(t, Main([]string{testHelpFlag}, nil, &stdout, &stderr))
	require.Contains(t, stdout.String(), Command)
	require.Equal(t, 1, Main([]string{testMigrationFlag, "--input=" + SecretDirectory + "absent/" + InputName}, nil, &stdout, &stderr))
}

type borrowedSQLite struct{ *SQLClient }

func (borrowedSQLite) Close() error { return nil }

func TestDatabaseCLIThroughRealSQLiteMigrations(t *testing.T) {
	t.Parallel()
	sqliteMu.Lock()
	defer sqliteMu.Unlock()

	resetSharedSQLite(t)

	input := inputFixture(Migrate)
	input.Backend = SQLite
	input.Targets[0].Leader = Connection{DSNFile: SecretDirectory + "owned/dsn"}
	raw, err := json.Marshal(input)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer

	code := MainWith([]string{testMigrationFlag, "--input=" + SecretDirectory + "owned/" + InputName}, &stdout, &stderr,
		func(string) ([]byte, error) { return raw, nil }, func(ctx context.Context, operation string, got Input) error {
			return Execute(ctx, operation, got, func(context.Context, Connection, string) (Client, error) {
				return borrowedSQLite{&SQLClient{DB: sharedSQLite, backend: SQLite}}, nil
			})
		})
	require.Zero(t, code, stderr.String())

	var dirty bool
	require.NoError(t, sharedSQLite.QueryRow("SELECT dirty FROM schema_migrations").Scan(&dirty))
	require.False(t, dirty)
}
