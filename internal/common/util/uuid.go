package util

import (
	"fmt"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"

	googleUuid "github.com/google/uuid"
)

func GenerateUUIDv7Function() func() (*googleUuid.UUID, error) {
	return func() (*googleUuid.UUID, error) { return GenerateUUIDv7() }
}

func GenerateUUIDv7() (*googleUuid.UUID, error) {
	uuidV7, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	return &uuidV7, nil
}

func ValidateUUID(uuid *googleUuid.UUID, msg *string) error {
	if uuid == nil {
		return fmt.Errorf("%s: %w", *msg, cryptoutilAppErr.ErrUUIDCantBeNil)
	} else if *uuid == googleUuid.Nil {
		return fmt.Errorf("%s: %w", *msg, cryptoutilAppErr.ErrUUIDCantBeZero)
	} else if *uuid == googleUuid.Max {
		return fmt.Errorf("%s: %w", *msg, cryptoutilAppErr.ErrUUIDCantBeMax)
	}

	return nil
}

func ValidateUUIDs(uuids []googleUuid.UUID, msg *string) error {
	if uuids == nil {
		return fmt.Errorf("%s: %w", *msg, cryptoutilAppErr.ErrUUIDsCantBeNil)
	} else if len(uuids) == 0 {
		return fmt.Errorf("%s: %w", *msg, cryptoutilAppErr.ErrUUIDsCantBeEmpty)
	}

	for i, uuid := range uuids {
		if err := ValidateUUID(&uuid, msg); err != nil {
			return fmt.Errorf("%s, offset %d: %w", *msg, i, cryptoutilAppErr.ErrUUIDsCantBeNil)
		}
	}

	return nil
}
