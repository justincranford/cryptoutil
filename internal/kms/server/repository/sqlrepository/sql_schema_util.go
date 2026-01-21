// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver registration
	_ "modernc.org/sqlite"             // SQLite driver registration (CGO-free)
)

// LogSchema logs the database schema for the given repository.
func LogSchema(sqlRepository *SQLRepository) error {
	switch sqlRepository.dbType {
	case DBTypeSQLite:
		return logSQLiteSchema(sqlRepository)
	case DBTypePostgres:
		return logPostgresSchema(sqlRepository)
	default:
		return fmt.Errorf("unsupported database type: %s", sqlRepository.dbType)
	}
}

func logSQLiteSchema(sqlRepository *SQLRepository) error {
	tableNames, err := func() ([]string, error) {
		queryResults, err := sqlRepository.sqlDB.QueryContext(context.Background(), `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';`)
		if err != nil {
			return nil, fmt.Errorf("failed to query SQLite table names: %w", err)
		}

		defer func() {
			if closeErr := queryResults.Close(); closeErr != nil {
				sqlRepository.telemetryService.Slogger.Error("failed to close query results", "error", closeErr)
			}
		}() // Ensure query results are closed before for first loop body

		var tableNames []string

		for queryResults.Next() {
			var tableName string
			if err := queryResults.Scan(&tableName); err != nil {
				return nil, fmt.Errorf("failed to scan table name: %w", err)
			}

			tableNames = append(tableNames, tableName)
		}

		return tableNames, nil
	}()
	if err != nil {
		return fmt.Errorf("failed to query table names: %w", err)
	}

	fmt.Println("SQLite Schema:")

	for _, tableName := range tableNames {
		fmt.Printf("Table: %s\n", tableName)

		err = func() error {
			queryResults, err := sqlRepository.sqlDB.QueryContext(context.Background(), fmt.Sprintf("PRAGMA table_info(%s);", tableName))
			if err != nil {
				return fmt.Errorf("failed to query table info for %s: %w", tableName, err)
			}

			defer func() {
				if closeErr := queryResults.Close(); closeErr != nil {
					sqlRepository.telemetryService.Slogger.Error("failed to close query results", "error", closeErr)
				}
			}() // Ensure query results are closed before next loop body

			for queryResults.Next() {
				var cid int

				var name, ctype string

				var notnull, pk int

				var dfltValue any
				if err := queryResults.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
					return fmt.Errorf("failed to scan column info: %w", err)
				}

				fmt.Printf("  Column: %s, Type: %s, NotNull: %d, PrimaryKey: %d\n", name, ctype, notnull, pk)
			}

			return nil
		}()
		if err != nil {
			return fmt.Errorf("failed to log columns for table %s: %w", tableName, err)
		}
	}

	return nil
}

func logPostgresSchema(sqlRepository *SQLRepository) error {
	tableNames, err := func() ([]string, error) {
		queryResults, err := sqlRepository.sqlDB.QueryContext(context.Background(), `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';`)
		if err != nil {
			return nil, fmt.Errorf("failed to query PostgreSQL table names: %w", err)
		}

		defer func() {
			if closeErr := queryResults.Close(); closeErr != nil {
				sqlRepository.telemetryService.Slogger.Error("failed to close query results", "error", closeErr)
			}
		}() // Ensure query results are closed before for first loop body

		var tableNames []string

		for queryResults.Next() {
			var tableName string
			if err := queryResults.Scan(&tableName); err != nil {
				return nil, fmt.Errorf("failed to scan table name: %w", err)
			}

			tableNames = append(tableNames, tableName)
		}

		return tableNames, nil
	}()
	if err != nil {
		return fmt.Errorf("failed to query table names: %w", err)
	}

	for _, tableName := range tableNames {
		fmt.Printf("Table: %s\n", tableName)

		err = func() error {
			queryResults, err := sqlRepository.sqlDB.QueryContext(context.Background(), fmt.Sprintf(`SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '%s';`, tableName))
			if err != nil {
				return fmt.Errorf("failed to query column info for %s: %w", tableName, err)
			}

			defer func() {
				if closeErr := queryResults.Close(); closeErr != nil {
					sqlRepository.telemetryService.Slogger.Error("failed to close query results", "error", closeErr)
				}
			}() // Ensure query results are closed before next loop body

			for queryResults.Next() {
				var columnName, dataType, isNullable string
				if err := queryResults.Scan(&columnName, &dataType, &isNullable); err != nil {
					return fmt.Errorf("failed to scan column info: %w", err)
				}

				fmt.Printf("  Column: %s, Type: %s, Nullable: %s\n", columnName, dataType, isNullable)
			}

			return nil
		}()
		if err != nil {
			return fmt.Errorf("failed to log columns for table %s: %w", tableName, err)
		}
	}

	return nil
}
