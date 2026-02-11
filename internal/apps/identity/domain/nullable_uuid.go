// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"database/sql/driver"
	"fmt"

	googleUuid "github.com/google/uuid"
)

// NullableUUID is a nullable UUID type that properly handles SQLite TEXT columns.
// This type implements sql.Scanner and driver.Valuer to correctly serialize/deserialize
// UUIDs to/from TEXT columns in SQLite databases.
type NullableUUID struct {
	UUID  googleUuid.UUID
	Valid bool // Valid is true if UUID is not NULL
}

// NewNullableUUID creates a new NullableUUID from a UUID pointer.
func NewNullableUUID(id *googleUuid.UUID) NullableUUID {
	if id == nil || *id == googleUuid.Nil {
		return NullableUUID{Valid: false}
	}

	return NullableUUID{UUID: *id, Valid: true}
}

// Ptr returns a pointer to the UUID if valid, nil otherwise.
func (n NullableUUID) Ptr() *googleUuid.UUID {
	if !n.Valid {
		return nil
	}

	id := n.UUID

	return &id
}

// Scan implements sql.Scanner interface.
func (n *NullableUUID) Scan(value any) error {
	if value == nil {
		n.UUID, n.Valid = googleUuid.Nil, false

		return nil
	}

	switch v := value.(type) {
	case string:
		id, err := googleUuid.Parse(v)
		if err != nil {
			return fmt.Errorf("failed to parse UUID from string: %w", err)
		}

		n.UUID, n.Valid = id, true

		return nil
	case []byte:
		id, err := googleUuid.ParseBytes(v)
		if err != nil {
			return fmt.Errorf("failed to parse UUID from bytes: %w", err)
		}

		n.UUID, n.Valid = id, true

		return nil
	default:
		return fmt.Errorf("cannot scan type %T into NullableUUID", value)
	}
}

// Value implements driver.Valuer interface.
func (n NullableUUID) Value() (driver.Value, error) {
	if !n.Valid {
		//nolint:nilnil // database/sql requires (nil, nil) for SQL NULL values
		return nil, nil
	}

	return n.UUID.String(), nil
}
