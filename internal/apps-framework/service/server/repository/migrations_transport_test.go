// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	transportCaseScalar                   = "scalar"
	transportCaseBorrowError              = "borrow error"
	transportCaseClosedScope              = "closed scope"
	transportCaseBeginError               = "begin error"
	transportCaseCancelledBegin           = "cancelled begin"
	transportCaseCancelledExec            = "cancelled exec"
	transportCaseCancelledQuery           = "cancelled query"
	transportCaseExecFailure              = "exec failure"
	transportCaseQueryFailure             = "query failure"
	transportCaseColumnsFailure           = "columns failure"
	transportCaseRowsEOF                  = "rows EOF"
	transportCaseRowsReadFailure          = "rows read failure"
	transportCaseRowsScanFailure          = "rows scan failure"
	transportCaseRowsConversionFailure    = "rows conversion failure"
	transportCaseRowsCloseFailure         = "rows close failure"
	transportCaseSavepointBeginFailure    = "savepoint begin failure"
	transportCaseSavepointCommitFailure   = "savepoint commit failure"
	transportCaseSavepointRollbackFailure = "savepoint rollback failure"
	transportCaseSavepointSuccess         = "savepoint success"
	transportCaseTransactionLock          = "transaction lock"
	transportCaseDiscardPGLease           = "discard cancelled PostgreSQL lease"
	transportCaseConnectorGuards          = "connector guards"
	transportIgnoredQuery                 = "ignored"
	transportUnused                       = "unused"
	transportXactLock                     = "SELECT pg_advisory_xact_lock($1)"
)

type transportTestConnector struct{ connection *transportTestConnection }

func (c transportTestConnector) Connect(context.Context) (driver.Conn, error) {
	return c.connection, nil
}
func (c transportTestConnector) Driver() driver.Driver { return migrationSQLDriver{} }

type transportTestConnection struct {
	initializationTestConnection
	rows                            [][]driver.Value
	queryErr, errorClose, errorNext bool
	lastQuery                       string
}

func (c *transportTestConnection) Begin() (driver.Tx, error) { return transportTestTransaction{}, nil }

func (c *transportTestConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if c.queryErr {
		return nil, errors.New("synthetic query error")
	}

	return &transportTestRows{data: c.rows, closeErr: c.errorClose, nextErr: c.errorNext}, nil
}

func (c *transportTestConnection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.lastQuery = query

	return c.initializationTestConnection.ExecContext(ctx, query, args)
}

type transportTestTransaction struct{}

func (transportTestTransaction) Commit() error   { return nil }
func (transportTestTransaction) Rollback() error { return nil }

type transportTestRows struct {
	data              [][]driver.Value
	closeErr, nextErr bool
}

func (r *transportTestRows) Columns() []string { return []string{transportCaseScalar} }
func (r *transportTestRows) Close() error {
	if r.closeErr {
		return errors.New("synthetic row close error")
	}

	return nil
}

func (r *transportTestRows) Next(destination []driver.Value) error {
	if r.nextErr {
		return errors.New("synthetic row read error")
	}

	if len(r.data) == 0 {
		return io.EOF
	}

	copy(destination, r.data[0])
	r.data = r.data[1:]

	return nil
}

type transportQueryFixture struct{ rows *sql.Rows }

func (f transportQueryFixture) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return driver.RowsAffected(1), nil
}

func (f transportQueryFixture) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return f.rows, nil
}

func TestMigrationTransportErrorAndOwnershipPaths(t *testing.T) {
	t.Parallel()

	for _, test := range []string{
		transportCaseBorrowError, transportCaseClosedScope, transportCaseBeginError, transportCaseCancelledBegin, transportCaseCancelledExec, transportCaseCancelledQuery,
		transportCaseExecFailure, transportCaseQueryFailure, transportCaseColumnsFailure, transportCaseRowsEOF, transportCaseRowsReadFailure, transportCaseRowsScanFailure, transportCaseRowsConversionFailure, transportCaseRowsCloseFailure,
		transportCaseSavepointBeginFailure, transportCaseSavepointCommitFailure, transportCaseSavepointRollbackFailure, transportCaseSavepointSuccess, transportCaseTransactionLock,
		transportCaseDiscardPGLease, transportCaseConnectorGuards,
	} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			physical := &transportTestConnection{rows: [][]driver.Value{{int64(9)}}}
			db := sql.OpenDB(transportTestConnector{connection: physical})

			defer func() { _ = db.Close() }()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			switch test {
			case transportCaseBorrowError:
				require.NoError(t, db.Close())
				connector := &migrationConnector{ctx: ctx, db: db}
				_, err := connector.Connect(ctx)
				require.Error(t, err)
			case transportCaseClosedScope:
				connector := &migrationConnector{ctx: ctx, tx: new(sql.Tx)}
				require.NoError(t, connector.close())
				_, err := connector.Connect(ctx)
				require.Error(t, err)
			case transportCaseBeginError:
				borrowed, err := db.Conn(ctx)
				require.NoError(t, err)
				require.NoError(t, borrowed.Close())
				connection := &migrationConnection{ctx: ctx, borrowed: borrowed}
				_, err = connection.Begin()
				require.Error(t, err)
				require.Error(t, connection.Close())
			case transportCaseCancelledBegin, transportCaseCancelledExec, transportCaseCancelledQuery:
				cancel()

				connection := &migrationConnection{ctx: ctx}

				var err error

				switch test {
				case transportCaseCancelledBegin:
					_, err = connection.BeginTx(ctx, driver.TxOptions{})
				case transportCaseCancelledExec:
					_, err = connection.ExecContext(ctx, transportIgnoredQuery, nil)
				default:
					_, err = connection.QueryContext(ctx, transportIgnoredQuery, nil)
				}

				require.ErrorIs(t, err, context.Canceled)
			case transportCaseExecFailure, transportCaseQueryFailure:
				connection := &migrationConnection{ctx: ctx, queries: initializationTestDBConnector{}}

				var err error
				if test == transportCaseExecFailure {
					_, err = connection.ExecContext(ctx, transportIgnoredQuery, nil)
				} else {
					_, err = connection.QueryContext(ctx, transportIgnoredQuery, nil)
				}

				require.Error(t, err)
			case transportCaseColumnsFailure:
				rows, err := db.QueryContext(ctx, transportIgnoredQuery)
				require.NoError(t, err)
				require.NoError(t, rows.Close())
				connection := &migrationConnection{ctx: ctx, queries: transportQueryFixture{rows: rows}}
				_, err = connection.QueryContext(ctx, transportIgnoredQuery, nil)
				require.Error(t, err)
			case transportCaseRowsEOF, transportCaseRowsReadFailure, transportCaseRowsScanFailure, transportCaseRowsConversionFailure, transportCaseRowsCloseFailure:
				physical.errorNext = test == transportCaseRowsReadFailure
				physical.errorClose = test == transportCaseRowsCloseFailure

				if test == transportCaseRowsConversionFailure {
					physical.rows = [][]driver.Value{{make(chan int)}}
				}

				connection := &migrationConnection{ctx: ctx, queries: db}
				rows, err := connection.QueryContext(ctx, transportIgnoredQuery, nil)
				require.NoError(t, err)
				require.Equal(t, []string{transportCaseScalar}, rows.Columns())

				if test == transportCaseRowsCloseFailure {
					require.Error(t, rows.Close())

					return
				}

				destination := make([]driver.Value, 1)
				if test == transportCaseRowsScanFailure {
					destination = nil
				}

				err = rows.Next(destination)
				if test == transportCaseRowsEOF {
					require.NoError(t, err)
					require.ErrorIs(t, rows.Next(destination), io.EOF)
				} else {
					require.Error(t, err)
				}

				_ = rows.Close()
			case transportCaseSavepointBeginFailure, transportCaseSavepointCommitFailure, transportCaseSavepointRollbackFailure, transportCaseSavepointSuccess, transportCaseTransactionLock:
				tx, err := db.BeginTx(ctx, nil)
				require.NoError(t, err)

				defer func() { _ = tx.Rollback() }()

				connection := &migrationConnection{ctx: ctx, tx: tx, queries: tx, postgres: true}
				if test == transportCaseTransactionLock {
					_, err = connection.ExecContext(ctx, "SELECT pg_advisory_lock($1)", []driver.NamedValue{{Ordinal: 1, Value: int64(9)}})
					require.NoError(t, err)
					require.Equal(t, transportXactLock, physical.lastQuery)

					_, err = connection.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", nil)
					require.NoError(t, err)
					require.Equal(t, transportXactLock, physical.lastQuery)

					return
				}

				physical.execErr = errors.New("synthetic savepoint error")
				if test == transportCaseSavepointBeginFailure {
					physical.execErrAt = 1
				}

				nested, err := connection.Begin()
				if test == transportCaseSavepointBeginFailure {
					require.Error(t, err)

					return
				}

				require.NoError(t, err)

				if test == transportCaseSavepointSuccess {
					require.NoError(t, nested.Commit())
					nested, err = connection.Begin()
					require.NoError(t, err)
					require.NoError(t, nested.Rollback())
				} else {
					physical.execErrAt = 2

					if test == transportCaseSavepointCommitFailure {
						require.Error(t, nested.Commit())
					} else {
						require.Error(t, nested.Rollback())
					}
				}
			case transportCaseDiscardPGLease:
				connector := &migrationConnector{ctx: ctx, db: db, postgres: true}
				lease, err := connector.Connect(ctx)
				require.NoError(t, err)
				cancel()
				require.NoError(t, lease.Close())
				require.NoError(t, connector.close())
				require.Zero(t, db.Stats().InUse)
				require.NoError(t, db.PingContext(context.Background()), "the owner pool is not closed")
			case transportCaseConnectorGuards:
				_, err := migrationSQLDriver{}.Open(transportUnused)
				require.Error(t, err)

				connection := &migrationConnection{}
				_, err = connection.Prepare(transportUnused)
				require.Error(t, err)
				require.NotNil(t, (&migrationConnector{}).Driver())
			}
		})
	}
}
