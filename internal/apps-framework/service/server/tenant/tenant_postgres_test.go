// Copyright (c) 2025 Justin Cranford
//

package tenant

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

// ----------------------------------------------------------------------------
// Minimal mock SQL driver for testing PostgreSQL code paths without Postgres.
// ----------------------------------------------------------------------------

// mockPGDriver is a configurable in-process SQL driver.
type mockPGDriver struct {
	mu       sync.Mutex
	failExec bool
	failScan bool // When true, rows return non-string values to trigger Scan error.
	failIter bool // When true, rows.Next() returns an error after first row.
	rowsData [][]driver.Value
	columns  []string
}

func (d *mockPGDriver) Open(_ string) (driver.Conn, error) {
	return &mockPGConn{driver: d}, nil
}

type mockPGConn struct {
	driver *mockPGDriver
}

func (c *mockPGConn) Prepare(query string) (driver.Stmt, error) {
	return &mockPGStmt{conn: c, query: query}, nil
}

func (c *mockPGConn) Close() error { return nil }

func (c *mockPGConn) Begin() (driver.Tx, error) { return &mockPGTx{}, nil }

type mockPGTx struct{}

func (t *mockPGTx) Commit() error   { return nil }
func (t *mockPGTx) Rollback() error { return nil }

type mockPGStmt struct {
	conn  *mockPGConn
	query string
}

func (s *mockPGStmt) Close() error  { return nil }
func (s *mockPGStmt) NumInput() int { return -1 }

func (s *mockPGStmt) Exec(_ []driver.Value) (driver.Result, error) {
	s.conn.driver.mu.Lock()
	defer s.conn.driver.mu.Unlock()

	if s.conn.driver.failExec {
		return nil, errors.New("mock exec failure")
	}

	return driver.RowsAffected(1), nil
}

func (s *mockPGStmt) Query(_ []driver.Value) (driver.Rows, error) {
	s.conn.driver.mu.Lock()
	defer s.conn.driver.mu.Unlock()

	if s.conn.driver.failExec {
		return nil, errors.New("mock query failure")
	}

	return &mockPGRows{
		columns:  append([]string{}, s.conn.driver.columns...),
		data:     append([][]driver.Value{}, s.conn.driver.rowsData...),
		failScan: s.conn.driver.failScan,
		failIter: s.conn.driver.failIter,
	}, nil
}

type mockPGRows struct {
	columns  []string
	data     [][]driver.Value
	pos      int
	failScan bool
	failIter bool
}

func (r *mockPGRows) Columns() []string { return r.columns }
func (r *mockPGRows) Close() error      { return nil }

func (r *mockPGRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}

	// Simulate iteration error after first row.
	if r.failIter && r.pos > 0 {
		return errors.New("mock iteration error")
	}

	row := r.data[r.pos]
	r.pos++

	if r.failScan {
		// Return an unsupported driver.Value type to trigger Scan error.
		for i := range dest {
			dest[i] = struct{}{}
		}

		return nil
	}

	copy(dest, row)

	return nil
}

// pgDriverCounter generates unique driver names to allow parallel testing.
var pgDriverCounter atomic.Uint64 //nolint:gochecknoglobals // Required for unique sql.Register names per test.

// newMockPGSchemaManager creates a SchemaManager with a fresh per-test mock Postgres driver.
// Each call registers a uniquely-named driver so tests can run in parallel without sharing state.
func newMockPGSchemaManager(t *testing.T, d *mockPGDriver) *SchemaManager {
	t.Helper()

	driverName := fmt.Sprintf("mockpg_%d", pgDriverCounter.Add(1))
	sql.Register(driverName, d)

	db, err := sql.Open(driverName, "dummy")
	require.NoError(t, err)

	// Need a GORM db for NewSchemaManager — use SQLite.
	gormDB := setupTestDB(t)

	return &SchemaManager{
		db:     gormDB,
		sqlDB:  db,
		dbType: DBTypePostgres,
	}
}

// ----------------------------------------------------------------------------
// Tests for unsupported database type (default switch cases).
// ----------------------------------------------------------------------------

func TestSchemaManager_UnsupportedType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		op      func(sm *SchemaManager) error
		wantErr string
	}{
		{name: "CreateSchema unsupported", op: func(sm *SchemaManager) error { return sm.CreateSchema(context.Background(), "tenant-id") }, wantErr: "unsupported database type"},
		{name: "DropSchema unsupported", op: func(sm *SchemaManager) error { return sm.DropSchema(context.Background(), "tenant-id") }, wantErr: "unsupported database type"},
		{name: "SchemaExists unsupported", op: func(sm *SchemaManager) error {
			_, err := sm.SchemaExists(context.Background(), "tenant-id")

			return err
		}, wantErr: "unsupported database type"},
		{name: "ListSchemas unsupported", op: func(sm *SchemaManager) error {
			_, err := sm.ListSchemas(context.Background())

			return err
		}, wantErr: "unsupported database type"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm := &SchemaManager{dbType: "unsupported"}
			err := tc.op(sm)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

// ----------------------------------------------------------------------------
// Tests for PostgreSQL Create/Drop operations via mock driver.
// ----------------------------------------------------------------------------

func TestPostgresSchema_CreateDrop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		failExec bool
		op       func(sm *SchemaManager) error
		wantErr  string
	}{
		{name: "create success", failExec: false, op: func(sm *SchemaManager) error { return sm.CreateSchema(context.Background(), "test-tenant-id") }},
		{name: "create error", failExec: true, op: func(sm *SchemaManager) error { return sm.CreateSchema(context.Background(), "test-tenant-id") }, wantErr: "failed to create PostgreSQL schema"},
		{name: "drop success", failExec: false, op: func(sm *SchemaManager) error { return sm.DropSchema(context.Background(), "test-tenant-id") }},
		{name: "drop error", failExec: true, op: func(sm *SchemaManager) error { return sm.DropSchema(context.Background(), "test-tenant-id") }, wantErr: "failed to drop PostgreSQL schema"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := &mockPGDriver{failExec: tc.failExec}
			sm := newMockPGSchemaManager(t, d)
			err := tc.op(sm)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Tests for PostgreSQL SchemaExists via mock driver.
// ----------------------------------------------------------------------------

func TestPostgresSchemaExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		driver  *mockPGDriver
		want    bool
		wantErr string
	}{
		{name: "exists true", driver: &mockPGDriver{columns: []string{"exists"}, rowsData: [][]driver.Value{{true}}}, want: true},
		{name: "exists false", driver: &mockPGDriver{columns: []string{"exists"}, rowsData: [][]driver.Value{{false}}}, want: false},
		{name: "query error", driver: &mockPGDriver{failExec: true}, wantErr: "failed to check PostgreSQL schema existence"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm := newMockPGSchemaManager(t, tc.driver)
			exists, err := sm.SchemaExists(context.Background(), "test-tenant-id")

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, exists)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Tests for PostgreSQL ListSchemas via mock driver.
// ----------------------------------------------------------------------------

func TestPostgresListSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		driver    *mockPGDriver
		wantLen   int
		wantEntry string
		wantErr   string
	}{
		{
			name:      "success",
			driver:    &mockPGDriver{columns: []string{"schema_name"}, rowsData: [][]driver.Value{{"tenant_abc"}, {"tenant_def"}}},
			wantLen:   2,
			wantEntry: "tenant_abc",
		},
		{name: "query error", driver: &mockPGDriver{failExec: true}, wantErr: "failed to list PostgreSQL schemas"},
		{name: "scan error", driver: &mockPGDriver{failScan: true, columns: []string{"schema_name"}, rowsData: [][]driver.Value{{"tenant_bad"}}}, wantErr: "failed to scan schema name"},
		{name: "iteration error", driver: &mockPGDriver{failIter: true, columns: []string{"schema_name"}, rowsData: [][]driver.Value{{"tenant_ok"}, {"tenant_fail"}}}, wantErr: "error iterating schema rows"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm := newMockPGSchemaManager(t, tc.driver)
			schemas, err := sm.ListSchemas(context.Background())

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.Len(t, schemas, tc.wantLen)
				require.Contains(t, schemas, tc.wantEntry)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Tests for GetScopedDB PostgreSQL branch.
// ----------------------------------------------------------------------------

func TestGetScopedDB_Postgres(t *testing.T) {
	t.Parallel()

	gormDB := setupTestDB(t)
	sm := &SchemaManager{
		db:     gormDB,
		dbType: DBTypePostgres,
	}

	scoped := sm.GetScopedDB("test-tenant-pg")
	require.NotNil(t, scoped)
}

// ----------------------------------------------------------------------------
// Tests for SQLite error paths using a closed database.
// ----------------------------------------------------------------------------

// newClosedSQLiteSchemaManager creates a SchemaManager with a closed SQLite DB.
func newClosedSQLiteSchemaManager(t *testing.T) *SchemaManager {
	t.Helper()

	// Create a normal SQLite db then close it to force errors.
	rawDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Close the underlying connection to force errors going forward.
	require.NoError(t, rawDB.Close())

	// Build a fresh raw DB that is closed immediately for the sqlDB field.
	closedDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
	require.NoError(t, err)
	require.NoError(t, closedDB.Close())

	return &SchemaManager{
		db:     gormDB,
		sqlDB:  closedDB,
		dbType: DBTypeSQLite,
	}
}

func TestClosedSQLiteSchema_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		op      func(sm *SchemaManager) error
		wantErr string
	}{
		{name: "create error", op: func(sm *SchemaManager) error { return sm.CreateSchema(context.Background(), "error-tenant") }, wantErr: "failed to attach SQLite database"},
		{name: "drop error", op: func(sm *SchemaManager) error { return sm.DropSchema(context.Background(), "error-tenant") }, wantErr: "failed to detach SQLite database"},
		{name: "list query error", op: func(sm *SchemaManager) error {
			_, err := sm.ListSchemas(context.Background())

			return err
		}, wantErr: "failed to list SQLite schemas"},
		{name: "exists scan error", op: func(sm *SchemaManager) error {
			_, err := sm.SchemaExists(context.Background(), "error-tenant")

			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm := newClosedSQLiteSchemaManager(t)
			err := tc.op(sm)
			require.Error(t, err)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
