// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseConnection_GORM(t *testing.T) {
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
			conn, err := NewDatabaseConnectionGORM(db)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, conn)
			} else {
				require.NoError(t, err)
				require.NotNil(t, conn)
				require.Equal(t, DatabaseModeGORM, conn.Mode())
				require.NotNil(t, conn.GORM())
				require.True(t, conn.HasGORM())
				require.True(t, conn.HasRawSQL()) // GORM can extract sql.DB

				sqlDB, err := conn.SQL()
				require.NoError(t, err)
				require.NotNil(t, sqlDB)

				// Cleanup.
				err = conn.Close()
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConnection_RawSQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *sql.DB
		wantErr bool
	}{
		{
			name: "valid sql db",
			setup: func() *sql.DB {
				db, _ := sql.Open("sqlite", ":memory:")

				return db
			},
			wantErr: false,
		},
		{
			name:    "nil sql db",
			setup:   func() *sql.DB { return nil },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := tt.setup()
			conn, err := NewDatabaseConnectionRawSQL(db)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, conn)
			} else {
				require.NoError(t, err)
				require.NotNil(t, conn)
				require.Equal(t, DatabaseModeRawSQL, conn.Mode())
				require.Nil(t, conn.GORM())
				require.False(t, conn.HasGORM())
				require.True(t, conn.HasRawSQL())

				sqlDB, err := conn.SQL()
				require.NoError(t, err)
				require.NotNil(t, sqlDB)

				// Cleanup.
				err = conn.Close()
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConnection_Dual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() (*gorm.DB, *sql.DB)
		wantErr bool
	}{
		{
			name: "valid dual dbs",
			setup: func() (*gorm.DB, *sql.DB) {
				gormDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				sqlDB, _ := sql.Open("sqlite", ":memory:")

				return gormDB, sqlDB
			},
			wantErr: false,
		},
		{
			name: "nil gorm db",
			setup: func() (*gorm.DB, *sql.DB) {
				sqlDB, _ := sql.Open("sqlite", ":memory:")

				return nil, sqlDB
			},
			wantErr: true,
		},
		{
			name: "nil sql db",
			setup: func() (*gorm.DB, *sql.DB) {
				gormDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

				return gormDB, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gormDB, sqlDB := tt.setup()
			conn, err := NewDatabaseConnectionDual(gormDB, sqlDB)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, conn)
				// Cleanup if needed.
				if sqlDB != nil {
					_ = sqlDB.Close()
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, conn)
				require.Equal(t, DatabaseModeDual, conn.Mode())
				require.NotNil(t, conn.GORM())
				require.True(t, conn.HasGORM())
				require.True(t, conn.HasRawSQL())

				retSqlDB, err := conn.SQL()
				require.NoError(t, err)
				require.NotNil(t, retSqlDB)

				// Cleanup.
				err = conn.Close()
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfig(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()

		cfg := NewDefaultDatabaseConfig("postgres://localhost/db")
		require.Equal(t, DatabaseModeGORM, cfg.Mode)
		require.Equal(t, "postgres://localhost/db", cfg.URL)
		require.False(t, cfg.VerboseMode)
		require.Equal(t, "disabled", cfg.ContainerMode)
		require.False(t, cfg.SkipTemplateMigrations)
	})

	t.Run("kms config", func(t *testing.T) {
		t.Parallel()

		cfg := NewKMSDatabaseConfig(":memory:")
		require.Equal(t, DatabaseModeRawSQL, cfg.Mode)
		require.Equal(t, ":memory:", cfg.URL)
		require.False(t, cfg.VerboseMode)
		require.Equal(t, "disabled", cfg.ContainerMode)
		require.True(t, cfg.SkipTemplateMigrations)
	})
}
