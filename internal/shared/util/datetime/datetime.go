// Copyright (c) 2025 Justin Cranford
//
//

// Package datetime provides date and time utility functions.
package datetime

import (
	"fmt"
	"time"
)

const utcFormat = "2006-01-02T15:04:05Z"

func ISO8601Time2String(value *time.Time) *string {
	if value == nil {
		return nil
	}

	converted := (*value).Format(utcFormat)

	return &converted
}

// ISO8601String2Time converts ISO8601 UTC string to Time pointer.
// Returns nil if input is nil (valid sentinel value).
func ISO8601String2Time(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // nil is valid sentinel value
	}

	converted, err := time.Parse(utcFormat, *value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	return &converted, nil
}
