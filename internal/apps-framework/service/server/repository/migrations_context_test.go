// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"testing/fstest"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	source "github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
)

type cancelledMigrationDriver struct {
	database.Driver
	cancel    context.CancelFunc
	migration *migrate.Migrate
	runs      int
}

func (d *cancelledMigrationDriver) Run(reader io.Reader) error {
	if err := d.Driver.Run(reader); err != nil {
		return err
	}

	d.runs++
	d.cancel()
	// Observe the actual AfterFunc signal before returning the completed file;
	// requeue that same signal so the runner consumes it at its next boundary.
	select {
	case signal := <-d.migration.GracefulStop:
		d.migration.GracefulStop <- signal
	case <-time.After(time.Second):
		return errors.New("migration context did not signal graceful stop")
	}

	return nil
}

func TestMigrationContextStopsBeforeNextEmbeddedFile(t *testing.T) {
	t.Parallel()
	contextMigrationMu.Lock()
	defer contextMigrationMu.Unlock()

	for _, statement := range []string{"DROP TABLE IF EXISTS context_first", "DROP TABLE IF EXISTS context_second", "DROP TABLE IF EXISTS schema_migrations"} {
		_, err := contextMigrationDB.Exec(statement)
		require.NoError(t, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	files := fstest.MapFS{
		"migrations/000001_first.up.sql":    {Data: []byte("CREATE TABLE context_first (id INTEGER);")},
		"migrations/000001_first.down.sql":  {Data: []byte("DROP TABLE context_first;")},
		"migrations/000002_second.up.sql":   {Data: []byte("CREATE TABLE context_second (id INTEGER);")},
		"migrations/000002_second.down.sql": {Data: []byte("DROP TABLE context_second;")},
	}

	var wrapped *cancelledMigrationDriver

	runner := newMigrationRunnerInternal(files, "migrations", func(sourceName string, sourceDriver source.Driver, dbName string, dbDriver database.Driver) (*migrate.Migrate, error) {
		wrapped = &cancelledMigrationDriver{Driver: dbDriver, cancel: cancel}
		instance, err := migrate.NewWithInstance(sourceName, sourceDriver, dbName, wrapped)
		wrapped.migration = instance

		return instance, err
	})
	require.ErrorIs(t, runner.ApplyContext(ctx, contextMigrationDB, DatabaseTypeSQLite), context.Canceled)
	require.Equal(t, 1, wrapped.runs)

	for _, test := range []struct {
		name  string
		count int
	}{{"context_first", 1}, {"context_second", 0}} {
		var count int
		require.NoError(t, contextMigrationDB.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", test.name).Scan(&count))
		require.Equal(t, test.count, count)
	}
}

func TestMigrationContextRejectsNilAndPrecancelled(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		ctx  context.Context
		db   *sql.DB
	}{
		{"nil context", nil, contextMigrationDB}, {"nil database", context.Background(), nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, NewMigrationRunner(MigrationsFS, "migrations").ApplyContext(test.ctx, test.db, DatabaseTypeSQLite))
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.ErrorIs(t, NewMigrationRunner(MigrationsFS, "migrations").ApplyContext(ctx, contextMigrationDB, DatabaseTypeSQLite), context.Canceled)
}

func TestMigrationContextCancellationDuringConstruction(t *testing.T) {
	t.Parallel()
	contextMigrationMu.Lock()
	defer contextMigrationMu.Unlock()

	_, err := contextMigrationDB.Exec("DROP TABLE IF EXISTS schema_migrations")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := newMigrationRunnerInternal(fstest.MapFS{
		"migrations/000001_constructor.up.sql": {Data: []byte("CREATE TABLE context_constructor (id INTEGER);")},
	}, "migrations", func(sourceName string, sourceDriver source.Driver, dbName string, dbDriver database.Driver) (*migrate.Migrate, error) {
		instance, err := migrate.NewWithInstance(sourceName, sourceDriver, dbName, dbDriver)

		cancel()

		return instance, err
	})
	require.ErrorIs(t, runner.ApplyContext(ctx, contextMigrationDB, DatabaseTypeSQLite), context.Canceled)

	var count int
	require.NoError(t, contextMigrationDB.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='context_constructor'").Scan(&count))
	require.Zero(t, count)
}
