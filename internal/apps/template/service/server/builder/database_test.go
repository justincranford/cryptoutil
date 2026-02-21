// Copyright (c) 2025 Justin Cranford
// SPDX-License-Identifier: MIT

package builder

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseConnection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *gorm.DB
		wantErr bool
	}{
		{
			name: "valid gorm db",
			setup: func() *gorm.DB {
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

				return db
			},
			wantErr: false,
		},
		{
			name:    "nil gorm db",
			setup:   func() *gorm.DB { return nil },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := tt.setup()
			conn, err := NewDatabaseConnection(db)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, conn)
			} else {
				require.NoError(t, err)
				require.NotNil(t, conn)
				require.NotNil(t, conn.GORM())

				sqlDB, err := conn.SQL()
				require.NoError(t, err)
				require.NotNil(t, sqlDB)

				err = conn.Close()
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConnection_SQL_NilDB(t *testing.T) {
	t.Parallel()

	conn := &DatabaseConnection{gormDB: nil}
	sqlDB, err := conn.SQL()
	require.Error(t, err)
	require.Nil(t, sqlDB)
	require.Contains(t, err.Error(), "gorm.DB is nil")
}

func TestDatabaseConnection_Close_NilDB(t *testing.T) {
	t.Parallel()

	conn := &DatabaseConnection{gormDB: nil}
	err := conn.Close()
	require.NoError(t, err)
}

func TestDatabaseConfig(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()

		cfg := NewDatabaseConfig("postgres://localhost/db")
		require.Equal(t, "postgres://localhost/db", cfg.URL)
		require.False(t, cfg.VerboseMode)
		require.Equal(t, "disabled", cfg.ContainerMode)
	})

	t.Run("memory config", func(t *testing.T) {
		t.Parallel()

		cfg := NewDatabaseConfig(":memory:")
		require.Equal(t, ":memory:", cfg.URL)
		require.False(t, cfg.VerboseMode)
		require.Equal(t, "disabled", cfg.ContainerMode)
	})
}
// fakeConnPool implements gorm.ConnPool but is NOT a *sql.DB.
// This causes gorm.DB.DB() to return ErrInvalidDB, enabling error path testing.
type fakeConnPool struct{}

func (f *fakeConnPool) PrepareContext(_ context.Context, _ string) (*sql.Stmt, error) {
	return nil, nil
}

func (f *fakeConnPool) ExecContext(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (f *fakeConnPool) QueryContext(_ context.Context, _ string, _ ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (f *fakeConnPool) QueryRowContext(_ context.Context, _ string, _ ...interface{}) *sql.Row {
	return nil
}

func TestDatabaseConnection_SQL_DBError(t *testing.T) {
	t.Parallel()

	// Use a GORM DB with a fake ConnPool (not *sql.DB) to trigger gormDB.DB() error.
	fakeDB := &gorm.DB{
		Config: &gorm.Config{
			ConnPool: &fakeConnPool{},
		},
	}

	conn := &DatabaseConnection{gormDB: fakeDB}
	sqlDB, err := conn.SQL()
	require.Error(t, err)
	require.Nil(t, sqlDB)
}

func TestDatabaseConnection_Close_DBError(t *testing.T) {
	t.Parallel()

	// Use a GORM DB with a fake ConnPool (not *sql.DB) to trigger gormDB.DB() error.
	fakeDB := &gorm.DB{
		Config: &gorm.Config{
			ConnPool: &fakeConnPool{},
		},
	}

	conn := &DatabaseConnection{gormDB: fakeDB}
	err := conn.Close()
	require.Error(t, err)
}