// Copyright (c) 2025 Justin Cranford
//

package pwdgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// impossiblePolicy has a charset where Max=0, making meetsRequirements always fail.
// Used to test the 1000-attempts-exhausted path.
var impossiblePolicy = PasswordPolicy{
	Name:      "impossible",
	MinLength: 4,
	MaxLength: 4,
	CharSets: []CharSetConfig{
		{Name: "letters", Characters: []rune("abcd"), Min: 0, Max: 0},
	},
}

// noStartEndPolicy has no StartCharacters or EndCharacters to trigger else branches.
var noStartEndPolicy = PasswordPolicy{
	Name:      "no-start-end",
	MinLength: 4,
	MaxLength: 4,
	CharSets: []CharSetConfig{
		{Name: "letters", Characters: []rune("abcdefghijklmnopqrstuvwxyz"), Min: 4, Max: MaxInt},
	},
}

// TestGenerate_NoStartEndCharacters tests password generation where StartCharacters
// and EndCharacters are not specified, exercising the else branches.
func TestGenerate_NoStartEndCharacters(t *testing.T) {
	t.Parallel()

	gen, err := NewPasswordGenerator(noStartEndPolicy)
	require.NoError(t, err)

	password, err := gen.Generate()
	require.NoError(t, err)
	require.Len(t, []rune(password), 4)

	// All chars should be lowercase letters.
	for _, r := range password {
		require.GreaterOrEqual(t, r, rune('a'))
		require.LessOrEqual(t, r, rune('z'))
	}
}

// TestGenerate_1000AttemptsExhausted tests the error path when all attempts fail.
// Cannot be parallel because it modifies package-level vars.
// Sequential: uses shared state (not safe for parallel execution).
func TestGenerate_1000AttemptsExhausted(t *testing.T) {
	gen, err := NewPasswordGenerator(impossiblePolicy)
	require.NoError(t, err)

	_, err = gen.Generate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate password meeting requirements after 1000 attempts")
}

// TestGenerate_RandomLengthError tests error when crand.Int fails in randomLength.
// Cannot be parallel because it modifies package-level vars.
// Sequential: uses shared state (not safe for parallel execution).
func TestGenerate_RandomLengthError(t *testing.T) {
	orig := pwdgenCrandIntFn
	pwdgenCrandIntFn = func(_ *big.Int) (*big.Int, error) {
		return nil, errors.New("injected crand failure")
	}

	defer func() { pwdgenCrandIntFn = orig }()

	// Use a policy with MinLength != MaxLength to trigger randomLength -> crand.Int.
	pol := PasswordPolicy{
		Name:      "var-length",
		MinLength: cryptoutilSharedMagic.IMMinPasswordLength,
		MaxLength: cryptoutilSharedMagic.HashPrefixLength,
		CharSets:  []CharSetConfig{{Name: "letters", Characters: []rune("abcdefgh"), Min: cryptoutilSharedMagic.IMMinPasswordLength, Max: MaxInt}},
	}
	gen, err := NewPasswordGenerator(pol)
	require.NoError(t, err)

	_, err = gen.Generate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to determine password length")
}

// TestGenerateCandidate_StartCharIndexError tests crand.Int error in start char selection.
// Cannot be parallel because it modifies package-level vars.
// Sequential: uses shared state (not safe for parallel execution).
func TestGenerateCandidate_StartCharIndexError(t *testing.T) {
	orig := pwdgenCrandIntFn
	pwdgenCrandIntFn = func(_ *big.Int) (*big.Int, error) {
		return nil, errors.New("injected crand failure")
	}

	defer func() { pwdgenCrandIntFn = orig }()

	pol := PasswordPolicy{
		Name:            "with-start",
		MinLength:       4,
		MaxLength:       4,
		StartCharacters: []rune("ABCD"),
		CharSets:        []CharSetConfig{{Name: "letters", Characters: []rune("abcd"), Min: 4, Max: MaxInt}},
	}
	gen, err := NewPasswordGenerator(pol)
	require.NoError(t, err)

	_, err = gen.Generate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate start character index")
}

// TestGenerateCandidate_FirstCharIndexError tests crand.Int error when no StartCharacters.
// Cannot be parallel because it modifies package-level vars.
// Sequential: uses shared state (not safe for parallel execution).
func TestGenerateCandidate_FirstCharIndexError(t *testing.T) {
	orig := pwdgenCrandIntFn
	pwdgenCrandIntFn = func(_ *big.Int) (*big.Int, error) {
		return nil, errors.New("injected crand failure")
	}

	defer func() { pwdgenCrandIntFn = orig }()

	gen, err := NewPasswordGenerator(noStartEndPolicy)
	require.NoError(t, err)

	_, err = gen.Generate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate first character index")
}

// TestGenerateCandidate_EndCharIndexError tests crand.Int error in end char selection.
// Cannot be parallel because it modifies package-level vars.
// Sequential: uses shared state (not safe for parallel execution).
func TestGenerateCandidate_EndCharIndexError(t *testing.T) {
	// First call succeeds (for start), subsequent calls fail.
	callCount := 0
	orig := pwdgenCrandIntFn
	pwdgenCrandIntFn = func(max *big.Int) (*big.Int, error) {
		callCount++
		if callCount < 2 {
			// Let first crand.Int succeed for start character.
			return big.NewInt(0), nil
		}

		return nil, errors.New("injected crand failure for end char")
	}

	defer func() { pwdgenCrandIntFn = orig }()

	pol := PasswordPolicy{
		Name:          "with-end",
		MinLength:     4,
		MaxLength:     4,
		EndCharacters: []rune("XYZ"),
		CharSets:      []CharSetConfig{{Name: "letters", Characters: []rune("abcd"), Min: 4, Max: MaxInt}},
	}
	gen, err := NewPasswordGenerator(pol)
	require.NoError(t, err)

	_, err = gen.Generate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate end character index")
}

// TestGenerateCandidate_MiddleCharIndexError tests crand.Int error in middle char selection.
// Cannot be parallel because it modifies package-level vars.
// Sequential: uses shared state (not safe for parallel execution).
func TestGenerateCandidate_MiddleCharIndexError(t *testing.T) {
	// First few calls succeed, then fail.
	callCount := 0
	orig := pwdgenCrandIntFn
	pwdgenCrandIntFn = func(max *big.Int) (*big.Int, error) {
		callCount++
		if callCount <= 2 {
			// Let first+second crand.Int succeed for start and end/first.
			return big.NewInt(0), nil
		}

		return nil, errors.New("injected crand failure for middle char")
	}

	defer func() { pwdgenCrandIntFn = orig }()

	gen, err := NewPasswordGenerator(noStartEndPolicy)
	require.NoError(t, err)

	_, err = gen.Generate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate random middle character index")
}
