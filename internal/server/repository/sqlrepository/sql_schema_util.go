package sqlrepository

import (
	"fmt"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func LogSchema(sqlRepository *SqlRepository) error {
	switch sqlRepository.dbType {
	case DBTypeSQLite:
		return logSqliteSchema(sqlRepository)
	case DBTypePostgres:
		return logPostgresSchema(sqlRepository)
	default:
		return fmt.Errorf("unsupported database type: %s", sqlRepository.dbType)
	}
}

func logSqliteSchema(sqlRepository *SqlRepository) error {
	tableNames, err := func() ([]string, error) {
		queryResults, err := sqlRepository.sqlDB.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';`)
		if err != nil {
			return nil, fmt.Errorf("failed to query SQLite table names: %w", err)
		}
		defer queryResults.Close() // Ensure query results are closed before for first loop body

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
			queryResults, err := sqlRepository.sqlDB.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
			if err != nil {
				return fmt.Errorf("failed to query table info for %s: %w", tableName, err)
			}
			defer queryResults.Close() // Ensure query results are closed before next loop body

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

func logPostgresSchema(sqlRepository *SqlRepository) error {
	tableNames, err := func() ([]string, error) {
		queryResults, err := sqlRepository.sqlDB.Query(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';`)
		if err != nil {
			return nil, fmt.Errorf("failed to query PostgreSQL table names: %w", err)
		}
		defer queryResults.Close() // Ensure query results are closed before for first loop body

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
			queryResults, err := sqlRepository.sqlDB.Query(fmt.Sprintf(`SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '%s';`, tableName))
			if err != nil {
				return fmt.Errorf("failed to query column info for %s: %w", tableName, err)
			}
			defer queryResults.Close() // Ensure query results are closed before next loop body

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
