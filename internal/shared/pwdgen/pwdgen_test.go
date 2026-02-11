// Copyright (c) 2025 Justin Cranford
//

package pwdgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPasswordPolicy_Validate tests policy validation.
func TestPasswordPolicy_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		policy  PasswordPolicy
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_basic_policy",
			policy:  BasicPolicy,
			wantErr: false,
		},
		{
			name:    "valid_strong_policy",
			policy:  StrongPolicy,
			wantErr: false,
		},
		{
			name:    "valid_enterprise_policy",
			policy:  EnterprisePolicy,
			wantErr: false,
		},
		{
			name: "invalid_min_length_zero",
			policy: PasswordPolicy{
				MinLength: 0,
				MaxLength: 10,
				CharSets:  []CharSetConfig{{Name: "test", Characters: []rune("abc"), Min: 1, Max: MaxInt}},
			},
			wantErr: true,
			errMsg:  "MinLength must be at least 1",
		},
		{
			name: "invalid_max_less_than_min",
			policy: PasswordPolicy{
				MinLength: 10,
				MaxLength: 5,
				CharSets:  []CharSetConfig{{Name: "test", Characters: []rune("abc"), Min: 1, Max: MaxInt}},
			},
			wantErr: true,
			errMsg:  "MaxLength must be >= MinLength",
		},
		{
			name: "invalid_no_charsets",
			policy: PasswordPolicy{
				MinLength: 8,
				MaxLength: 16,
				CharSets:  []CharSetConfig{},
			},
			wantErr: true,
			errMsg:  "at least one CharSet required",
		},
		{
			name: "invalid_charset_no_characters",
			policy: PasswordPolicy{
				MinLength: 8,
				MaxLength: 16,
				CharSets:  []CharSetConfig{{Name: "empty", Characters: []rune{}, Min: 1, Max: MaxInt}},
			},
			wantErr: true,
			errMsg:  "has no characters",
		},
		{
			name: "invalid_charset_negative_min",
			policy: PasswordPolicy{
				MinLength: 8,
				MaxLength: 16,
				CharSets:  []CharSetConfig{{Name: "test", Characters: []rune("abc"), Min: -1, Max: MaxInt}},
			},
			wantErr: true,
			errMsg:  "Min must be >= 0",
		},
		{
			name: "invalid_charset_max_less_than_min",
			policy: PasswordPolicy{
				MinLength: 8,
				MaxLength: 16,
				CharSets:  []CharSetConfig{{Name: "test", Characters: []rune("abc"), Min: 5, Max: 2}},
			},
			wantErr: true,
			errMsg:  "Max must be >= Min",
		},
		{
			name: "invalid_sum_of_mins_exceeds_max_length",
			policy: PasswordPolicy{
				MinLength: 8,
				MaxLength: 10,
				CharSets: []CharSetConfig{
					{Name: "lower", Characters: LowercaseLetters, Min: 5, Max: MaxInt},
					{Name: "upper", Characters: UppercaseLetters, Min: 5, Max: MaxInt},
					{Name: "digit", Characters: Digits, Min: 5, Max: MaxInt},
				},
			},
			wantErr: true,
			errMsg:  "sum of CharSet minimums",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.policy.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestNewPasswordGenerator tests generator construction.
func TestNewPasswordGenerator(t *testing.T) {
	t.Parallel()

	t.Run("valid_policy", func(t *testing.T) {
		t.Parallel()

		gen, err := NewPasswordGenerator(BasicPolicy)
		require.NoError(t, err)
		require.NotNil(t, gen)
	})

	t.Run("invalid_policy", func(t *testing.T) {
		t.Parallel()

		invalidPolicy := PasswordPolicy{
			MinLength: 0,
			MaxLength: 10,
		}

		_, err := NewPasswordGenerator(invalidPolicy)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid password policy")
	})
}

// TestGenerate_BasicPolicy tests password generation with basic policy.
func TestGenerate_BasicPolicy(t *testing.T) {
	t.Parallel()

	gen, err := NewPasswordGenerator(BasicPolicy)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		password, err := gen.Generate()
		require.NoError(t, err)
		require.NotEmpty(t, password)

		// Check length.
		require.GreaterOrEqual(t, len(password), BasicPolicy.MinLength)
		require.LessOrEqual(t, len(password), BasicPolicy.MaxLength)

		// Check contains at least one lowercase, uppercase, digit.
		require.True(t, containsAny(password, LowercaseLetters), "password must contain lowercase: %s", password)
		require.True(t, containsAny(password, UppercaseLetters), "password must contain uppercase: %s", password)
		require.True(t, containsAny(password, Digits), "password must contain digit: %s", password)

		// Check start character constraint.
		require.True(t, containsRune(BasicPolicy.StartCharacters, rune(password[0])), "first char must be from StartCharacters: %s", password)

		// Check end character constraint.
		require.True(t, containsRune(BasicPolicy.EndCharacters, rune(password[len(password)-1])), "last char must be from EndCharacters: %s", password)
	}
}

// TestGenerate_StrongPolicy tests password generation with strong policy.
func TestGenerate_StrongPolicy(t *testing.T) {
	t.Parallel()

	gen, err := NewPasswordGenerator(StrongPolicy)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		password, err := gen.Generate()
		require.NoError(t, err)
		require.NotEmpty(t, password)

		// Check length.
		require.GreaterOrEqual(t, len(password), StrongPolicy.MinLength)
		require.LessOrEqual(t, len(password), StrongPolicy.MaxLength)

		// Check contains at least one lowercase, uppercase, digit, special.
		require.True(t, containsAny(password, LowercaseLetters), "password must contain lowercase: %s", password)
		require.True(t, containsAny(password, UppercaseLetters), "password must contain uppercase: %s", password)
		require.True(t, containsAny(password, Digits), "password must contain digit: %s", password)
		require.True(t, containsAny(password, SpecialChars), "password must contain special: %s", password)
	}
}

// TestGenerate_EnterprisePolicy tests password generation with enterprise policy.
func TestGenerate_EnterprisePolicy(t *testing.T) {
	t.Parallel()

	gen, err := NewPasswordGenerator(EnterprisePolicy)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		password, err := gen.Generate()
		require.NoError(t, err)
		require.NotEmpty(t, password)

		// Check length.
		require.GreaterOrEqual(t, len(password), EnterprisePolicy.MinLength)
		require.LessOrEqual(t, len(password), EnterprisePolicy.MaxLength)

		// Check minimum 2 of each type.
		require.GreaterOrEqual(t, countChars(password, LowercaseLetters), 2, "must have at least 2 lowercase: %s", password)
		require.GreaterOrEqual(t, countChars(password, UppercaseLetters), 2, "must have at least 2 uppercase: %s", password)
		require.GreaterOrEqual(t, countChars(password, Digits), 2, "must have at least 2 digits: %s", password)
		require.GreaterOrEqual(t, countChars(password, SpecialChars), 2, "must have at least 2 special: %s", password)

		// Check no duplicates (AllowDuplicates = false).
		require.True(t, hasNoDuplicates(password), "password must not have duplicates: %s", password)

		// Check no adjacent repeats (AllowAdjacentRepeats = false).
		require.True(t, hasNoAdjacentRepeats(password), "password must not have adjacent repeats: %s", password)
	}
}

// TestGenerate_CustomPolicy tests password generation with custom policy.
func TestGenerate_CustomPolicy(t *testing.T) {
	t.Parallel()

	customPolicy := PasswordPolicy{
		Name:                 "custom",
		MinLength:            10,
		MaxLength:            10, // Fixed length.
		AllowDuplicates:      true,
		AllowAdjacentRepeats: true,
		StartCharacters:      []rune("ABC"),
		EndCharacters:        []rune("123"),
		CharSets: []CharSetConfig{
			{Name: "uppercase", Characters: []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"), Min: 3, Max: 5},
			{Name: "digits", Characters: []rune("0123456789"), Min: 3, Max: 5},
		},
	}

	gen, err := NewPasswordGenerator(customPolicy)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		password, err := gen.Generate()
		require.NoError(t, err)

		// Check fixed length.
		require.Equal(t, 10, len(password))

		// Check start/end constraints.
		require.True(t, containsRune(customPolicy.StartCharacters, rune(password[0])))
		require.True(t, containsRune(customPolicy.EndCharacters, rune(password[len(password)-1])))

		// Check character set requirements.
		upperCount := countChars(password, []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
		digitCount := countChars(password, []rune("0123456789"))

		require.GreaterOrEqual(t, upperCount, 3)
		require.LessOrEqual(t, upperCount, 5)
		require.GreaterOrEqual(t, digitCount, 3)
		require.LessOrEqual(t, digitCount, 5)
	}
}

// TestGenerate_Uniqueness tests that generated passwords are unique.
func TestGenerate_Uniqueness(t *testing.T) {
	t.Parallel()

	gen, err := NewPasswordGenerator(StrongPolicy)
	require.NoError(t, err)

	passwords := make(map[string]bool)

	for i := 0; i < 100; i++ {
		password, err := gen.Generate()
		require.NoError(t, err)
		require.False(t, passwords[password], "generated duplicate password: %s", password)
		passwords[password] = true
	}
}

// Helper functions.

func containsAny(password string, chars []rune) bool {
	for _, c := range password {
		for _, allowed := range chars {
			if c == allowed {
				return true
			}
		}
	}

	return false
}

func containsRune(chars []rune, r rune) bool {
	for _, c := range chars {
		if c == r {
			return true
		}
	}

	return false
}

func countChars(password string, chars []rune) int {
	count := 0

	for _, c := range password {
		for _, allowed := range chars {
			if c == allowed {
				count++

				break
			}
		}
	}

	return count
}

func hasNoDuplicates(password string) bool {
	seen := make(map[rune]bool)
	for _, c := range password {
		if seen[c] {
			return false
		}

		seen[c] = true
	}

	return true
}

func hasNoAdjacentRepeats(password string) bool {
	runes := []rune(password)
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			return false
		}
	}

	return true
}
