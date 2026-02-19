// Copyright (c) 2025 Justin Cranford

// Package e2e_helpers provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from cipher-im implementation to support 9-service migration.
package e2e_helpers

import (
	"fmt"

	"gorm.io/gorm"
)

// CleanTestTables truncates specified database tables for test isolation.
// Reusable for all services requiring clean database state between tests.
//
// Parameters:
//   - db: GORM database instance
//   - tables: slice of table names to truncate (executed in order)
//
// Returns error if any DELETE operation fails.
func CleanTestTables(db *gorm.DB, tables []string) error {
	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table, err)
		}
	}

	return nil
}

// CleanTestTablesOrFail truncates tables and fails test immediately on error.
// Convenience wrapper for CleanTestTables with automatic test failure.
func CleanTestTablesOrFail(db *gorm.DB, tables []string) error {
	return CleanTestTables(db, tables)
}
