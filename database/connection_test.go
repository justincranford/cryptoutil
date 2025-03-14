package database

import (
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSadPathOpenDatabaseBlankDriverName(t *testing.T) {
	_, err := openDatabase("", databaseUrlSqlite)
	assert.NotNil(t, err)
}

func TestSadPathOpenDatabaseBlankDatabaseName(t *testing.T) {
	_, err := openDatabase(driverNameSqlite, "")
	assert.NotNil(t, err)
}

func TestHappyPathOpenDatabase(t *testing.T) {
	databaseService, err := openDatabase(driverNameSqlite, databaseUrlSqlite)
	require.Nil(t, err)
	require.NotNil(t, databaseService)
	defer databaseService.Shutdown()
	require.NotNil(t, databaseService.DB())
	require.NotNil(t, databaseService.GormDB())
	queryResult, err := databaseService.DB().Query("SELECT 1")
	require.NotNil(t, queryResult)
	require.Nil(t, err)

	defer func(result *sql.Rows) {
		err := result.Close()
		if err != nil {
			t.Logf("Error closing query queryResult: %v", err)
		}
	}(queryResult) // Always close the queryResult set to avoid resource leaks
	for queryResult.Next() {
		var value int
		err = queryResult.Scan(&value)
		require.Nil(t, err)
		t.Logf("Query queryResult: %d", value) // Log the value
	}
	err = queryResult.Err()
	require.Nil(t, err)
}
