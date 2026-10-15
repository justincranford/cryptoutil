// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"fmt"
	"strconv"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func validateSQLiteUpdateReturningVersion(version string) error {
	actual, err := parseSQLiteVersion(version)
	if err != nil {
		return fmt.Errorf("invalid SQLite version %q: %w", version, err)
	}

	minimum, err := parseSQLiteVersion(cryptoutilSharedMagic.SQLiteMinimumVersionUpdateReturning)
	if err != nil {
		return fmt.Errorf("invalid configured minimum SQLite version: %w", err)
	}

	for index := range actual {
		switch {
		case actual[index] > minimum[index]:
			return nil
		case actual[index] < minimum[index]:
			return fmt.Errorf(
				"SQLite %s or newer is required for UPDATE ... RETURNING; found %s",
				cryptoutilSharedMagic.SQLiteMinimumVersionUpdateReturning,
				version,
			)
		}
	}

	return nil
}

func parseSQLiteVersion(version string) ([3]int, error) {
	var parsed [3]int

	components := strings.Split(version, ".")
	if len(components) != len(parsed) {
		return parsed, fmt.Errorf("expected three numeric components")
	}

	for index, component := range components {
		value, err := strconv.Atoi(component)
		if err != nil || value < 0 {
			return parsed, fmt.Errorf("invalid numeric component %q", component)
		}

		parsed[index] = value
	}

	return parsed, nil
}
