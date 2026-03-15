// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"database/sql/driver"
	"fmt"
)

// IntBool is a boolean type that stores as INTEGER (0/1) in the database.
// This type implements sql.Scanner and driver.Valuer to correctly serialize/deserialize
// booleans to/from INTEGER columns for cross-database compatibility (SQLite + PostgreSQL).
//
// Usage in GORM models:
//
//	ConsentGranted IntBool `gorm:"type:integer;not null;default:0"`
type IntBool bool

// Scan implements sql.Scanner interface.
// Accepts: int64, int, bool, nil (treated as false).
func (b *IntBool) Scan(value any) error {
	if value == nil {
		*b = false

		return nil
	}

	switch v := value.(type) {
	case int64:
		*b = v != 0

		return nil
	case int:
		*b = v != 0

		return nil
	case bool:
		*b = IntBool(v)

		return nil
	default:
		return fmt.Errorf("cannot scan type %T into IntBool", value)
	}
}

// Value implements driver.Valuer interface.
// Returns int64 (0 or 1) for database storage.
func (b IntBool) Value() (driver.Value, error) {
	if b {
		return int64(1), nil
	}

	return int64(0), nil
}

// Bool returns the underlying bool value.
func (b IntBool) Bool() bool {
	return bool(b)
}
