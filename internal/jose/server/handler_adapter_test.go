// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIsNotFoundError tests the isNotFoundError helper function.
func TestIsNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "nil error returns false",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "contains not found",
			errMsg:   "elastic key not found",
			expected: true,
		},
		{
			name:     "contains NOT FOUND uppercase",
			errMsg:   "elastic key NOT FOUND",
			expected: true,
		},
		{
			name:     "contains does not exist",
			errMsg:   "key does not exist in database",
			expected: true,
		},
		{
			name:     "unrelated error",
			errMsg:   "database connection failed",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tc.errMsg != "" {
				err = &testError{msg: tc.errMsg}
			}

			result := isNotFoundError(err)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestIsSymmetricKeyError tests the isSymmetricKeyError helper function.
func TestIsSymmetricKeyError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "nil error returns false",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "contains symmetric",
			errMsg:   "cannot get public key for symmetric algorithm",
			expected: true,
		},
		{
			name:     "contains SYMMETRIC uppercase",
			errMsg:   "SYMMETRIC keys do not have public keys",
			expected: true,
		},
		{
			name:     "contains no public key",
			errMsg:   "no public key available for this key type",
			expected: true,
		},
		{
			name:     "unrelated error",
			errMsg:   "encryption failed",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tc.errMsg != "" {
				err = &testError{msg: tc.errMsg}
			}

			result := isSymmetricKeyError(err)
			require.Equal(t, tc.expected, result)
		})
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
