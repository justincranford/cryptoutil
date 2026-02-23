package tls_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// TestDefaultDurationConstants validates that CA and end entity duration constants
// have the expected values. This kills ARITHMETIC_BASE mutants on the constant
// arithmetic expressions (10 * 365 * 24 * time.Hour, 365 * 24 * time.Hour).
func TestDefaultDurationConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   time.Duration
		expected time.Duration
	}{
		{
			name:     "default CA duration is 10 years",
			actual:   cryptoutilSharedCryptoTls.DefaultCADuration,
			expected: 87600 * time.Hour, // 10 * 365 * 24 = 87600 hours.
		},
		{
			name:     "default end entity duration is 1 year",
			actual:   cryptoutilSharedCryptoTls.DefaultEndEntityDuration,
			expected: 8760 * time.Hour, // 365 * 24 = 8760 hours.
		},
		{
			name:     "default CA chain length is 3",
			actual:   time.Duration(cryptoutilSharedCryptoTls.DefaultCAChainLength),
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, tc.actual)
		})
	}
}
