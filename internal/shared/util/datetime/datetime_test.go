// Copyright (c) 2025 Justin Cranford
//
//

package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestISO8601Time2String(t *testing.T) {
	t.Parallel()
	// Happy path
	now := time.Now().UTC()
	expected := now.Format(utcFormat)
	result := ISO8601Time2String(&now)

	require.NotNil(t, result)
	require.Equal(t, expected, *result)

	// Sad path
	var nilTime *time.Time

	result = ISO8601Time2String(nilTime)
	require.Nil(t, result)
}

func TestISO8601String2Time(t *testing.T) {
	t.Parallel()
	// Happy path
	now := time.Now().UTC().Format(utcFormat)

	expected, err := time.Parse(utcFormat, now)
	require.NoError(t, err)

	result, err := ISO8601String2Time(&now)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, expected, *result)

	// Sad path: Invalid format
	invalidDate := "not-a-date"

	result, err = ISO8601String2Time(&invalidDate)
	require.Error(t, err)
	require.Nil(t, result)

	// Sad path: Nil string
	var nilString *string

	result, err = ISO8601String2Time(nilString)
	require.NoError(t, err)
	require.Nil(t, result)
}
