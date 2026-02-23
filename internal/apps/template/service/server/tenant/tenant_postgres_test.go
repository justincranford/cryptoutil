// Copyright (c) 2025 Justin Cranford
//

package tenant

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"

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

func (d *mockPGDriver) setFail(fail bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.failExec = fail
}

func (d *mockPGDriver) setFailScan(fail bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.failScan = fail
}

func (d *mockPGDriver) setFailIter(fail bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.failIter = fail
}

func (d *mockPGDriver) setRows(cols []string, rows [][]driver.Value) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.columns = cols
	d.rowsData = rows
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

// pgDriver is the package-level mock driver registered once.
var pgDriver = &mockPGDriver{} //nolint:gochecknoglobals // Required for sql.Register.

func init() { //nolint:gochecknoinits // Required for sql.Register at package init.
	sql.Register("mockpg", pgDriver)
}

// newMockPGSchemaManager creates a SchemaManager with the mock Postgres driver.
func newMockPGSchemaManager(t *testing.T) *SchemaManager {
	t.Helper()

	db, err := sql.Open("mockpg", "dummy")
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

func TestCreateSchema_UnsupportedType(t *testing.T) {
	t.Parallel()

	sm := &SchemaManager{dbType: "unsupported"}
	err := sm.CreateSchema(context.Background(), "tenant-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}

func TestDropSchema_UnsupportedType(t *testing.T) {
	t.Parallel()

	sm := &SchemaManager{dbType: "unsupported"}
	err := sm.DropSchema(context.Background(), "tenant-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}

func TestSchemaExists_UnsupportedType(t *testing.T) {
	t.Parallel()

	sm := &SchemaManager{dbType: "unsupported"}
	_, err := sm.SchemaExists(context.Background(), "tenant-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}

func TestListSchemas_UnsupportedType(t *testing.T) {
	t.Parallel()

	sm := &SchemaManager{dbType: "unsupported"}
	_, err := sm.ListSchemas(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}

// ----------------------------------------------------------------------------
// Tests for PostgreSQL operations via mock driver.
// ----------------------------------------------------------------------------

// TestCreatePostgresSchema_Success tests successful schema creation.
// Cannot be parallel because it configures the global pgDriver.
func TestCreatePostgresSchema_Success(t *testing.T) {
	pgDriver.setFail(false)

	sm := newMockPGSchemaManager(t)

	err := sm.CreateSchema(context.Background(), "test-tenant-id")
	require.NoError(t, err)
}

// TestCreatePostgresSchema_Error tests error path in schema creation.
// Cannot be parallel because it configures the global pgDriver.
func TestCreatePostgresSchema_Error(t *testing.T) {
	pgDriver.setFail(true)

	defer pgDriver.setFail(false)

	sm := newMockPGSchemaManager(t)

	err := sm.CreateSchema(context.Background(), "test-tenant-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create PostgreSQL schema")
}

// TestDropPostgresSchema_Success tests successful schema drop.
// Cannot be parallel because it configures the global pgDriver.
func TestDropPostgresSchema_Success(t *testing.T) {
	pgDriver.setFail(false)

	sm := newMockPGSchemaManager(t)

	err := sm.DropSchema(context.Background(), "test-tenant-id")
	require.NoError(t, err)
}

// TestDropPostgresSchema_Error tests error path in schema drop.
// Cannot be parallel because it configures the global pgDriver.
func TestDropPostgresSchema_Error(t *testing.T) {
	pgDriver.setFail(true)

	defer pgDriver.setFail(false)

	sm := newMockPGSchemaManager(t)

	err := sm.DropSchema(context.Background(), "test-tenant-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to drop PostgreSQL schema")
}

// TestPostgresSchemaExists_ExistsTrue tests schema existence returning true.
// Cannot be parallel because it configures the global pgDriver.
func TestPostgresSchemaExists_ExistsTrue(t *testing.T) {
	pgDriver.setFail(false)
	pgDriver.setRows([]string{"exists"}, [][]driver.Value{{true}})

	defer pgDriver.setRows(nil, nil)

	sm := newMockPGSchemaManager(t)

	exists, err := sm.SchemaExists(context.Background(), "test-tenant-id")
	require.NoError(t, err)
	require.True(t, exists)
}

// TestPostgresSchemaExists_ExistsFalse tests schema existence returning false.
// Cannot be parallel because it configures the global pgDriver.
func TestPostgresSchemaExists_ExistsFalse(t *testing.T) {
	pgDriver.setFail(false)
	pgDriver.setRows([]string{"exists"}, [][]driver.Value{{false}})

	defer pgDriver.setRows(nil, nil)

	sm := newMockPGSchemaManager(t)

	exists, err := sm.SchemaExists(context.Background(), "test-tenant-id")
	require.NoError(t, err)
	require.False(t, exists)
}

// TestPostgresSchemaExists_Error tests error path in schema existence check.
// Cannot be parallel because it configures the global pgDriver.
func TestPostgresSchemaExists_Error(t *testing.T) {
	pgDriver.setFail(true)

	defer pgDriver.setFail(false)

	sm := newMockPGSchemaManager(t)

	_, err := sm.SchemaExists(context.Background(), "test-tenant-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check PostgreSQL schema existence")
}

// TestListPostgresSchemas_Success tests listing schemas with results.
// Cannot be parallel because it configures the global pgDriver.
func TestListPostgresSchemas_Success(t *testing.T) {
	pgDriver.setFail(false)
	pgDriver.setRows([]string{"schema_name"}, [][]driver.Value{
		{"tenant_abc"},
		{"tenant_def"},
	})

	defer pgDriver.setRows(nil, nil)

	sm := newMockPGSchemaManager(t)

	schemas, err := sm.ListSchemas(context.Background())
	require.NoError(t, err)
	require.Len(t, schemas, 2)
	require.Contains(t, schemas, "tenant_abc")
}

// TestListPostgresSchemas_Error tests error path when listing schemas fails.
// Cannot be parallel because it configures the global pgDriver.
func TestListPostgresSchemas_Error(t *testing.T) {
	pgDriver.setFail(true)

	defer pgDriver.setFail(false)

	sm := newMockPGSchemaManager(t)

	_, err := sm.ListSchemas(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list PostgreSQL schemas")
}

// TestListPostgresSchemas_ScanError tests error path when rows.Scan fails.
// Cannot be parallel because it configures the global pgDriver.
func TestListPostgresSchemas_ScanError(t *testing.T) {
	pgDriver.setFail(false)
	pgDriver.setFailScan(true)
	pgDriver.setRows([]string{"schema_name"}, [][]driver.Value{
		{"tenant_bad"},
	})

	defer func() {
		pgDriver.setFailScan(false)
		pgDriver.setRows(nil, nil)
	}()

	sm := newMockPGSchemaManager(t)

	_, err := sm.ListSchemas(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to scan schema name")
}

// TestListPostgresSchemas_IterError tests error path when rows.Err() returns an error.
// Cannot be parallel because it configures the global pgDriver.
func TestListPostgresSchemas_IterError(t *testing.T) {
	pgDriver.setFail(false)
	pgDriver.setFailIter(true)
	pgDriver.setRows([]string{"schema_name"}, [][]driver.Value{
		{"tenant_ok"},
		{"tenant_fail"},
	})

	defer func() {
		pgDriver.setFailIter(false)
		pgDriver.setRows(nil, nil)
	}()

	sm := newMockPGSchemaManager(t)

	_, err := sm.ListSchemas(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "error iterating schema rows")
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
	rawDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Close the underlying connection to force errors going forward.
	require.NoError(t, rawDB.Close())

	// Build a fresh raw DB that is closed immediately for the sqlDB field.
	closedDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	require.NoError(t, closedDB.Close())

	return &SchemaManager{
		db:     gormDB,
		sqlDB:  closedDB,
		dbType: DBTypeSQLite,
	}
}

func TestCreateSQLiteSchema_Error(t *testing.T) {
	t.Parallel()

	sm := newClosedSQLiteSchemaManager(t)
	err := sm.CreateSchema(context.Background(), "error-tenant")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to attach SQLite database")
}

func TestDropSQLiteSchema_Error(t *testing.T) {
	t.Parallel()

	sm := newClosedSQLiteSchemaManager(t)
	err := sm.DropSchema(context.Background(), "error-tenant")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to detach SQLite database")
}

func TestListSQLiteSchemas_QueryError(t *testing.T) {
	t.Parallel()

	sm := newClosedSQLiteSchemaManager(t)
	_, err := sm.ListSchemas(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list SQLite schemas")
}

func TestSQLiteSchemaExists_ScanError(t *testing.T) {
	t.Parallel()

	sm := newClosedSQLiteSchemaManager(t)
	_, err := sm.SchemaExists(context.Background(), "error-tenant")
	require.Error(t, err)
}
