// Copyright (c) 2025-2026 Justin Cranford.
//

package dbtypes

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
)

// BoolInt stores a boolean as an integer for schemas that use INTEGER columns.
type BoolInt bool

// Value returns the integer representation used by SQLite and PostgreSQL.
func (b BoolInt) Value() (driver.Value, error) {
	if b {
		return int64(1), nil
	}

	return int64(0), nil
}

// Scan converts the database integer representation back into a boolean.
func (b *BoolInt) Scan(value any) error {
	if b == nil {
		return errors.New("cannot scan into nil *BoolInt")
	}

	switch v := value.(type) {
	case bool:
		*b = BoolInt(v)
	case int64:
		*b = BoolInt(v != 0)
	case int32:
		*b = BoolInt(v != 0)
	case []byte:
		return b.scanString(string(v))
	case string:
		return b.scanString(v)
	case nil:
		return errors.New("cannot scan NULL into BoolInt")
	default:
		return fmt.Errorf("cannot scan %T into BoolInt", value)
	}

	return nil
}

// GormDataType declares the corresponding database column type.
func (BoolInt) GormDataType() string {
	return "integer"
}

func (b *BoolInt) scanString(value string) error {
	integer, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as BoolInt: %w", value, err)
	}

	*b = BoolInt(integer != 0)

	return nil
}
