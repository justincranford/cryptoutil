package migrations

import (
	"cryptoutil/database"
	"database/sql"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyMigrationsHappyPath(t *testing.T) {
	dbService, err := database.NewService()
	require.Nil(t, err)
	require.NotNil(t, dbService)
	defer dbService.Shutdown()

	err = ApplyMigrations(dbService.DB())
	require.Nil(t, err)

	rows, err := dbService.DB().Query("SELECT name FROM sqlite_master WHERE type='table'")
	require.Nil(t, err)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		require.Nil(t, err)
	}(rows)

	var names []string
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		require.Nil(t, err)
		names = append(names, name)
	}
	require.NotEmpty(t, names)
	log.Printf("names: %v", names)
}

func TestApplyMigrationsSadPath(t *testing.T) {
	err := ApplyMigrations(nil)
	require.NotNil(t, err)
}
