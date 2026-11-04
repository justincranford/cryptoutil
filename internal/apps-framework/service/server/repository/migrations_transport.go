// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
)

type migrationQueries interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type migrationConnector struct {
	ctx         context.Context
	db          *sql.DB
	tx          *sql.Tx
	postgres    bool
	mu          sync.Mutex
	closed      bool
	connections []*migrationConnection
}

func (connector *migrationConnector) Connect(ctx context.Context) (driver.Conn, error) {
	connection := &migrationConnection{ctx: connector.ctx, tx: connector.tx, postgres: connector.postgres}
	if connector.tx != nil {
		connection.queries = connector.tx
	} else {
		// Drivers may request their lease with Background. Both the caller and
		// the owning migration operation must bound a busy owner-pool wait.
		borrowCtx, cancel := context.WithCancel(connector.ctx)

		stop := context.AfterFunc(ctx, cancel)
		defer stop()
		defer cancel()

		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("borrow migration context: %w", err)
		}

		borrowed, err := connector.db.Conn(borrowCtx)
		if err != nil {
			return nil, fmt.Errorf("borrow migration connection: %w", err)
		}

		connection.queries = borrowed
		connection.borrowed = borrowed
	}

	connector.mu.Lock()
	defer connector.mu.Unlock()

	if connector.closed {
		_ = connection.Close()

		return nil, errors.New("migration connection scope closed")
	}

	connector.connections = append(connector.connections, connection)

	return connection, nil
}
func (*migrationConnector) Driver() driver.Driver { return migrationSQLDriver{} }
func (connector *migrationConnector) close() error {
	connector.mu.Lock()
	defer connector.mu.Unlock()

	connector.closed = true

	var result error
	for _, connection := range connector.connections {
		result = errors.Join(result, connection.Close())
	}

	return result
}

type migrationSQLDriver struct{}

func (migrationSQLDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("migration connector required")
}

func migrationDatabaseView(ctx context.Context, db *sql.DB, tx *sql.Tx, postgres bool) (*sql.DB, func() error) {
	connector := &migrationConnector{ctx: ctx, db: db, tx: tx, postgres: postgres}
	view := sql.OpenDB(connector)
	view.SetMaxOpenConns(1)

	return view, func() error { return errors.Join(view.Close(), connector.close()) }
}

type migrationConnection struct {
	ctx       context.Context
	queries   migrationQueries
	borrowed  *sql.Conn
	tx        *sql.Tx
	postgres  bool
	once      sync.Once
	closeErr  error
	savepoint uint64
}

func (*migrationConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared migration statements are unsupported")
}

func (connection *migrationConnection) Close() error {
	connection.once.Do(func() {
		if connection.borrowed != nil {
			if connection.postgres && connection.ctx.Err() != nil {
				// A cancelled driver may not have released its session advisory lock.
				// Discard this leased physical connection, never the caller's pool.
				err := connection.borrowed.Raw(func(any) error { return driver.ErrBadConn })
				if err != nil && !errors.Is(err, driver.ErrBadConn) && !errors.Is(err, sql.ErrConnDone) {
					connection.closeErr = err
				}

				return
			}

			connection.closeErr = connection.borrowed.Close()
		}
	})

	return connection.closeErr
}

func (connection *migrationConnection) Begin() (driver.Tx, error) {
	return connection.BeginTx(connection.ctx, driver.TxOptions{})
}

func (connection *migrationConnection) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("migration transaction context: %w", err)
	}

	if connection.tx == nil {
		tx, err := connection.borrowed.BeginTx(connection.ctx, &sql.TxOptions{Isolation: sql.IsolationLevel(options.Isolation), ReadOnly: options.ReadOnly})
		if err != nil {
			return nil, fmt.Errorf("begin migration transaction: %w", err)
		}

		return tx, nil
	}

	connection.savepoint++

	name := "cu_migration_" + strconv.FormatUint(connection.savepoint, 10)
	if _, err := connection.tx.ExecContext(connection.ctx, "SAVEPOINT "+name); err != nil {
		return nil, fmt.Errorf("begin migration savepoint: %w", err)
	}

	return &migrationSavepoint{ctx: connection.ctx, tx: connection.tx, name: name}, nil
}

func migrationArguments(values []driver.NamedValue) []any {
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value.Value
	}

	return args
}

func (connection *migrationConnection) ExecContext(ctx context.Context, query string, values []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("migration query context: %w", err)
	}

	if connection.tx != nil && connection.postgres {
		// Preserve the driver's lock key, but keep a transaction-bound lock
		// until the owner commits or rolls back, including on cancellation.
		if query == "SELECT pg_advisory_lock($1)" {
			query = "SELECT pg_advisory_xact_lock($1)"
		}

		if query == "SELECT pg_advisory_unlock($1)" {
			return driver.RowsAffected(0), nil
		}
	}

	result, err := connection.queries.ExecContext(connection.ctx, query, migrationArguments(values)...)
	if err != nil {
		return nil, fmt.Errorf("execute migration SQL: %w", err)
	}

	return result, nil
}

func (connection *migrationConnection) QueryContext(ctx context.Context, query string, values []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("migration query context: %w", err)
	}

	rows, err := connection.queries.QueryContext(connection.ctx, query, migrationArguments(values)...)
	if err != nil {
		return nil, fmt.Errorf("query migration SQL: %w", err)
	}

	columns, err := rows.Columns()
	if err != nil {
		_ = rows.Close()

		return nil, fmt.Errorf("migration columns: %w", err)
	}

	return &migrationRows{rows: rows, columns: columns}, nil
}

type migrationRows struct {
	rows    *sql.Rows
	columns []string
}

func (rows *migrationRows) Columns() []string { return rows.columns }
func (rows *migrationRows) Close() error {
	if err := rows.rows.Close(); err != nil {
		return fmt.Errorf("close migration rows: %w", err)
	}

	return nil
}

func (rows *migrationRows) Next(destination []driver.Value) error {
	if !rows.rows.Next() {
		if err := rows.rows.Err(); err != nil {
			return fmt.Errorf("read migration rows: %w", err)
		}

		return io.EOF
	}

	values := make([]any, len(destination))
	pointers := make([]any, len(destination))

	for index := range values {
		pointers[index] = &values[index]
	}

	if err := rows.rows.Scan(pointers...); err != nil {
		return fmt.Errorf("scan migration rows: %w", err)
	}

	for index, value := range values {
		converted, err := driver.DefaultParameterConverter.ConvertValue(value)
		if err != nil {
			return fmt.Errorf("convert migration row: %w", err)
		}

		destination[index] = converted
	}

	return nil
}

type migrationSavepoint struct {
	ctx  context.Context
	tx   *sql.Tx
	name string
}

func (savepoint *migrationSavepoint) Commit() error {
	if _, err := savepoint.tx.ExecContext(savepoint.ctx, "RELEASE SAVEPOINT "+savepoint.name); err != nil {
		return fmt.Errorf("commit migration savepoint: %w", err)
	}

	return nil
}

func (savepoint *migrationSavepoint) Rollback() error {
	_, err := savepoint.tx.ExecContext(savepoint.ctx, "ROLLBACK TO SAVEPOINT "+savepoint.name)
	_, releaseErr := savepoint.tx.ExecContext(savepoint.ctx, "RELEASE SAVEPOINT "+savepoint.name)

	if err != nil || releaseErr != nil {
		return fmt.Errorf("rollback migration savepoint: %w", errors.Join(err, releaseErr))
	}

	return nil
}
