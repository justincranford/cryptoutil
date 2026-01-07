// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntBool_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     any
		expected  IntBool
		wantError bool
	}{
		{
			name:      "nil_value",
			input:     nil,
			expected:  false,
			wantError: false,
		},
		{
			name:      "int64_zero",
			input:     int64(0),
			expected:  false,
			wantError: false,
		},
		{
			name:      "int64_one",
			input:     int64(1),
			expected:  true,
			wantError: false,
		},
		{
			name:      "int64_nonzero",
			input:     int64(42),
			expected:  true,
			wantError: false,
		},
		{
			name:      "int_zero",
			input:     0,
			expected:  false,
			wantError: false,
		},
		{
			name:      "int_one",
			input:     1,
			expected:  true,
			wantError: false,
		},
		{
			name:      "int_nonzero",
			input:     99,
			expected:  true,
			wantError: false,
		},
		{
			name:      "bool_false",
			input:     false,
			expected:  false,
			wantError: false,
		},
		{
			name:      "bool_true",
			input:     true,
			expected:  true,
			wantError: false,
		},
		{
			name:      "invalid_string",
			input:     "invalid",
			expected:  false,
			wantError: true,
		},
		{
			name:      "invalid_float",
			input:     1.5,
			expected:  false,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var b IntBool

			err := b.Scan(tc.input)

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "cannot scan type")
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			}
		})
	}
}

func TestIntBool_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		intBool  IntBool
		expected driver.Value
	}{
		{
			name:     "false_returns_0",
			intBool:  false,
			expected: int64(0),
		},
		{
			name:     "true_returns_1",
			intBool:  true,
			expected: int64(1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			value, err := tc.intBool.Value()
			require.NoError(t, err)
			require.Equal(t, tc.expected, value)
		})
	}
}

func TestIntBool_Bool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		intBool  IntBool
		expected bool
	}{
		{
			name:     "false_to_bool",
			intBool:  false,
			expected: false,
		},
		{
			name:     "true_to_bool",
			intBool:  true,
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.intBool.Bool()
			require.Equal(t, tc.expected, result)
		})
	}
}
