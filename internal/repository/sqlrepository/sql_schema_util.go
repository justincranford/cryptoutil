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
	rows, err := sqlRepository.sqlDB.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';`)
	if err != nil {
		return fmt.Errorf("failed to query SQLite schema: %w", err)
	}
	defer rows.Close()

	fmt.Println("SQLite Schema:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		fmt.Printf("Table: %s\n", tableName)

		columns, err := sqlRepository.sqlDB.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
		if err != nil {
			return fmt.Errorf("failed to query table info for %s: %w", tableName, err)
		}
		defer columns.Close()

		for columns.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := columns.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				return fmt.Errorf("failed to scan column info: %w", err)
			}
			fmt.Printf("  Column: %s, Type: %s, NotNull: %d, PrimaryKey: %d\n", name, ctype, notnull, pk)
		}
	}
	return nil
}

func logPostgresSchema(sqlRepository *SqlRepository) error {
	rows, err := sqlRepository.sqlDB.Query(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';`)
	if err != nil {
		return fmt.Errorf("failed to query PostgreSQL schema: %w", err)
	}
	defer rows.Close()

	fmt.Println("PostgreSQL Schema:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		fmt.Printf("Table: %s\n", tableName)

		columns, err := sqlRepository.sqlDB.Query(fmt.Sprintf(`SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = '%s';`, tableName))
		if err != nil {
			return fmt.Errorf("failed to query column info for %s: %w", tableName, err)
		}
		defer columns.Close()

		for columns.Next() {
			var columnName, dataType, isNullable string
			if err := columns.Scan(&columnName, &dataType, &isNullable); err != nil {
				return fmt.Errorf("failed to scan column info: %w", err)
			}
			fmt.Printf("  Column: %s, Type: %s, Nullable: %s\n", columnName, dataType, isNullable)
		}
	}
	return nil
}
