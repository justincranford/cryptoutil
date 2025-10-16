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

func ISO8601String2Time(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	converted, err := time.Parse(utcFormat, *value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	return &converted, nil
}
