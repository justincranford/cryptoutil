// Copyright (c) 2025 Justin Cranford
//

// Package testdb_test provides black-box tests for the testdb package.
package testdb_test

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilTestingTestdb "cryptoutil/internal/apps/framework/service/testing/testdb"
)

// simpleModel is used by AutoMigrate tests to verify schema creation.
type simpleModel struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

func TestNewInMemorySQLiteDB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "creates working db"},
		{name: "creates second independent db"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilTestingTestdb.NewInMemorySQLiteDB(t)
			require.NotNil(t, db)

			sqlDB, err := db.DB()
			require.NoError(t, err)

			require.NoError(t, sqlDB.PingContext(t.Context()))
		})
	}
}

func TestNewInMemorySQLiteDB_Independence(t *testing.T) {
	t.Parallel()

	db1 := cryptoutilTestingTestdb.NewInMemorySQLiteDB(t)
	db2 := cryptoutilTestingTestdb.NewInMemorySQLiteDB(t)

	require.NotNil(t, db1)
	require.NotNil(t, db2)

	require.NoError(t, db1.AutoMigrate(&simpleModel{}))

	var count int64

	err := db2.Model(&simpleModel{}).Count(&count).Error
	require.Error(t, err, "db2 should not have the table created in db1")
}

func TestRequireNewInMemorySQLiteDB_NoModels(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	require.NoError(t, sqlDB.PingContext(t.Context()))
}

func TestRequireNewInMemorySQLiteDB_WithModels(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.RequireNewInMemorySQLiteDB(t, &simpleModel{})
	require.NotNil(t, db)

	result := db.Create(&simpleModel{ID: "test-id", Name: "test-name"})
	require.NoError(t, result.Error)

	var found simpleModel

	require.NoError(t, db.First(&found, "id = ?", "test-id").Error)
	require.Equal(t, "test-name", found.Name)
}

func TestFormatDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		port    string
		user    string
		pass    string
		dbName  string
		wantDSN string
	}{
		{
			name:    "standard postgres dsn",
			host:    cryptoutilSharedMagic.HostnameLocalhost,
			port:    "5432",
			user:    "admin",
			pass:    "secret",
			dbName:  "mydb",
			wantDSN: "postgres://admin:secret@localhost:5432/mydb?sslmode=disable",
		},
		{
			name:    "empty password",
			host:    cryptoutilSharedMagic.IPv4Loopback,
			port:    "5433",
			user:    "testuser",
			pass:    "",
			dbName:  "testdb",
			wantDSN: "postgres://testuser:@127.0.0.1:5433/testdb?sslmode=disable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := cryptoutilTestingTestdb.FormatDSN(tc.host, tc.port, tc.user, tc.pass, tc.dbName)
			require.Equal(t, tc.wantDSN, got)
		})
	}
}

func TestNewPostgresTestContainer_SkipsWhenUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// When Docker IS available: container starts successfully, db returned.
	// When Docker is NOT available: t.Skipf is called, test is skipped.
	db := cryptoutilTestingTestdb.NewPostgresTestContainer(ctx, t)
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	require.NoError(t, err)

	require.NoError(t, sqlDB.PingContext(ctx))
}

func TestRequireNewPostgresTestContainer_SkipsWhenUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := cryptoutilTestingTestdb.RequireNewPostgresTestContainer(ctx, t, &simpleModel{})
	if db == nil {
		return
	}

	result := db.Create(&simpleModel{ID: "pg-test-id", Name: "pg-test-name"})
	require.NoError(t, result.Error)

	var found simpleModel

	require.NoError(t, db.First(&found, "id = ?", "pg-test-id").Error)
	require.Equal(t, "pg-test-name", found.Name)
}

func TestNewClosedSQLiteDB_NoMigrations(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestingTestdb.NewClosedSQLiteDB(t, nil)
	require.NotNil(t, db)

	// All operations should fail because the underlying connection is closed.
	require.Error(t, db.Exec("SELECT 1").Error)
}

func TestNewClosedSQLiteDB_WithMigrations(t *testing.T) {
	t.Parallel()

	migrationsCalled := false

	db := cryptoutilTestingTestdb.NewClosedSQLiteDB(t, func(_ *sql.DB) error {
		migrationsCalled = true

		return nil
	})
	require.NotNil(t, db)
	require.True(t, migrationsCalled)

	// All operations should fail because the underlying connection is closed.
	require.Error(t, db.Exec("SELECT 1").Error)
}
