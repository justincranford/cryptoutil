// Copyright (c) 2025 Justin Cranford

package combinations

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	testify "github.com/stretchr/testify/require"
)

// TestToString_M tests M.ToString() method.
func TestToString_M(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        M
		expected string
	}{
		{
			name:     "Empty M",
			m:        M{},
			expected: "[]",
		},
		{
			name:     "Single value",
			m:        M{[]byte("A")},
			expected: "[A]",
		},
		{
			name:     "Multiple values",
			m:        M{[]byte("A"), []byte("B"), []byte("C")},
			expected: "[A, B, C]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.m.ToString()
			testify.Equal(t, tc.expected, result, "ToString output should match expected")
		})
	}
}

// TestToString_value tests value.ToString() method.
func TestToString_value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v        value
		expected string
	}{
		{
			name:     "Empty value",
			v:        value(""),
			expected: `""`,
		},
		{
			name:     "Single char",
			v:        value("A"),
			expected: `"A"`,
		},
		{
			name:     "Multi char",
			v:        value("ABC"),
			expected: `"ABC"`,
		},
		{
			name:     "Binary value",
			v:        value{0x00, 0x01, 0xFF},
			expected: `"\x00\x01\xff"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.v.ToString()
			testify.Equal(t, tc.expected, result, "ToString output should match expected")
		})
	}
}

// TestToString_combination tests combination.ToString() method.
func TestToString_combination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		c        combination
		expected string
	}{
		{
			name:     "Empty combination",
			c:        combination{},
			expected: "[]",
		},
		{
			name:     "Single value",
			c:        combination{value("A")},
			expected: `["A"]`,
		},
		{
			name:     "Multiple values",
			c:        combination{value("A"), value("B"), value("C")},
			expected: `["A", "B", "C"]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.c.ToString()
			testify.Equal(t, tc.expected, result, "ToString output should match expected")
		})
	}
}

// TestToString_combinations tests combinations.ToString() method.
func TestToString_combinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		c        Combinations
		expected string
	}{
		{
			name:     "Empty combinations",
			c:        Combinations{},
			expected: "[]",
		},
		{
			name:     "Single combination",
			c:        Combinations{combination{value("A")}},
			expected: `[["A"]]`,
		},
		{
			name:     "Multiple combinations",
			c:        Combinations{combination{value("A")}, combination{value("B"), value("C")}},
			expected: `[["A"], ["B", "C"]]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.c.ToString()
			testify.Equal(t, tc.expected, result, "ToString output should match expected")
		})
	}
}

// TestEncode_Panic_CombinationTooLarge tests panic when combination length exceeds uint8.
func TestEncode_Panic_CombinationTooLarge(t *testing.T) {
	t.Parallel()

	// Create combination with 256 values (exceeds uint8 max of 255)
	largeCombination := make(combination, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	for i := range largeCombination {
		largeCombination[i] = value("X")
	}

	testify.Panics(t, func() {
		_ = largeCombination.Encode()
	}, "Encode should panic when combination length > 255")
}

// TestEncode_NoPanic_CombinationExactly255 tests no panic at exact boundary of 255 elements.
// Kills CONDITIONALS_BOUNDARY mutant changing `> 255` to `>= 255` at combinations.go:82.
func TestEncode_NoPanic_CombinationExactly255(t *testing.T) {
	t.Parallel()

	exactCombination := make(combination, maxUint8Value)
	for i := range exactCombination {
		exactCombination[i] = value("X")
	}

	testify.NotPanics(t, func() {
		encoded := exactCombination.Encode()
		testify.NotEmpty(t, encoded, "Encoding 255-element combination should produce non-empty output")
	}, "Encode should NOT panic at exactly 255 elements")
}

// TestEncode_Panic_ValueTooLarge tests panic when value length exceeds uint8.
func TestEncode_Panic_ValueTooLarge(t *testing.T) {
	t.Parallel()

	// Create value with 256 bytes (exceeds uint8 max of 255)
	largeValue := make(value, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	for i := range largeValue {
		largeValue[i] = 'X'
	}

	c := combination{largeValue}

	testify.Panics(t, func() {
		_ = c.Encode()
	}, "Encode should panic when value length > 255")
}

// TestEncode_NoPanic_ValueExactly255 tests no panic at exact boundary of 255-byte value.
// Kills CONDITIONALS_BOUNDARY mutant changing `> 255` to `>= 255` at combinations.go:90.
func TestEncode_NoPanic_ValueExactly255(t *testing.T) {
	t.Parallel()

	exactValue := make(value, maxUint8Value) // exactly 255 bytes
	for i := range exactValue {
		exactValue[i] = 'X'
	}

	c := combination{exactValue}

	testify.NotPanics(t, func() {
		encoded := c.Encode()
		testify.NotEmpty(t, encoded, "Encoding 255-byte value should produce non-empty output")
	}, "Encode should NOT panic at exactly 255-byte value")
}

// TestComputeCombinations_LargeMError tests error when M length exceeds uint8.
func TestComputeCombinations_LargeMError(t *testing.T) {
	t.Parallel()

	// Create M with 256 values (exceeds uint8 max of 255)
	largeM := make(M, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	for i := range largeM {
		largeM[i] = []byte("X")
	}

	_, err := ComputeCombinations(largeM, 1)

	testify.Error(t, err, "ComputeCombinations should error when M length >= 255")
	testify.Contains(t, err.Error(), "can't be greater than", "Error should mention bounds")
}

// TestComputeCombinations_ExactlyMaxMError tests error at exact boundary of 255 M elements.
// Kills CONDITIONALS_BOUNDARY mutant changing `>= 255` to `> 255` at combinations.go:31.
func TestComputeCombinations_ExactlyMaxMError(t *testing.T) {
	t.Parallel()

	exactM := make(M, maxUint8Value) // exactly 255
	for i := range exactM {
		exactM[i] = []byte("X")
	}

	_, err := ComputeCombinations(exactM, 1)

	testify.Error(t, err, "ComputeCombinations should error when M length == 255")
	testify.Contains(t, err.Error(), "can't be greater than", "Error should mention bounds")
}

// TestComputeCombinations_JustBelowMaxM tests that M with 254 elements succeeds.
func TestComputeCombinations_JustBelowMaxM(t *testing.T) {
	t.Parallel()

	belowMaxM := make(M, maxUint8Value-1) // exactly 254
	for i := range belowMaxM {
		belowMaxM[i] = []byte("X")
	}

	result, err := ComputeCombinations(belowMaxM, 1)

	testify.NoError(t, err, "ComputeCombinations should succeed with M length 254")
	testify.Len(t, result, maxUint8Value-1, "Should return 254 single-element combinations")
}

// TestEncode_EmptyCombination tests encoding empty combination.
func TestEncode_EmptyCombination(t *testing.T) {
	t.Parallel()

	c := combination{}

	encoded := c.Encode()

	testify.Equal(t, []byte{0}, encoded, "Empty combination should encode as [0]")
}

// TestEncode_EmptyValue tests encoding combination with empty value.
func TestEncode_EmptyValue(t *testing.T) {
	t.Parallel()

	c := combination{value("")}

	encoded := c.Encode()

	testify.Equal(t, []byte{1, 0}, encoded, "Combination with empty value should encode as [1, 0]")
}

// TestEncode_BinaryValue tests encoding combination with binary value.
func TestEncode_BinaryValue(t *testing.T) {
	t.Parallel()

	binaryVal := value{0x00, 0x01, 0xFF}
	c := combination{binaryVal}

	encoded := c.Encode()

	expected := []byte{1, 3, 0x00, 0x01, 0xFF}
	testify.Equal(t, expected, encoded, "Binary value should be encoded correctly")
}

// TestEncode_MultipleCombinations tests encoding multiple combinations.
func TestEncode_MultipleCombinations(t *testing.T) {
	t.Parallel()

	c1 := combination{value("A")}
	c2 := combination{value("B"), value("C")}
	c3 := combination{value("D"), value("E"), value("F")}

	combos := Combinations{c1, c2, c3}

	encodings := combos.Encode()

	testify.Len(t, encodings, 3, "Should have 3 encodings")

	// Verify first combination encoding
	testify.Equal(t, []byte{1, 1, 'A'}, encodings[0], "First combination encoding should match")

	// Verify second combination encoding
	testify.Equal(t, []byte{2, 1, 'B', 1, 'C'}, encodings[1], "Second combination encoding should match")

	// Verify third combination encoding
	testify.Equal(t, []byte{3, 1, 'D', 1, 'E', 1, 'F'}, encodings[2], "Third combination encoding should match")
}

// TestToString_Integration tests ToString on complex combinations.
func TestToString_Integration(t *testing.T) {
	t.Parallel()

	m := M{[]byte("AA"), []byte("BB"), []byte("CC")}
	combos, err := ComputeCombinations(m, 2)
	testify.NoError(t, err, "ComputeCombinations should succeed")

	result := combos.ToString()

	testify.Contains(t, result, `"AA"`, "Should contain AA")
	testify.Contains(t, result, `"BB"`, "Should contain BB")
	testify.Contains(t, result, `"CC"`, "Should contain CC")
	testify.True(t, strings.HasPrefix(result, "["), "Should start with [")
	testify.True(t, strings.HasSuffix(result, "]"), "Should end with ]")
}
