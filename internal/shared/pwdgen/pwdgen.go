// Copyright (c) 2025 Justin Cranford
//

// Package pwdgen provides secure password generation with configurable policies.
package pwdgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// CharSetConfig defines a character set with minimum and maximum requirements.
type CharSetConfig struct {
	Name       string // Name of this character set (for error messages)
	Characters []rune // UTF-8 characters in this set
	Min        int    // Minimum required characters from this set
	Max        int    // Maximum allowed characters from this set (use MaxInt for unlimited)
}

// PasswordPolicy defines the policy for password generation.
type PasswordPolicy struct {
	Name                 string          // Policy name for identification
	MinLength            int             // Minimum total password length
	MaxLength            int             // Maximum total password length
	AllowDuplicates      bool            // Allow duplicate characters
	AllowAdjacentRepeats bool            // Allow adjacent repeated characters
	StartCharacters      []rune          // Characters allowed at start position
	EndCharacters        []rune          // Characters allowed at end position
	CharSets             []CharSetConfig // List of character sets with requirements
}

// PasswordGenerator generates passwords based on configured policy.
type PasswordGenerator struct {
	policy PasswordPolicy
}

const (
	// MaxInt represents effectively unlimited for Max values.
	MaxInt = int(^uint(0) >> 1)
)

// Password policy length constants.
const (
	basicPolicyMinLength      = 8
	basicPolicyMaxLength      = 16
	strongPolicyMinLength     = 12
	strongPolicyMaxLength     = 24
	enterprisePolicyMinLength = 16
	enterprisePolicyMaxLength = 32
)

// Common character sets.
var (
	LowercaseLetters = []rune("abcdefghijklmnopqrstuvwxyz")
	UppercaseLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Digits           = []rune(cryptoutilSharedMagic.EmailOTPCharset)
	SpecialChars     = []rune("!@#$%^&*()-_=+[]{}|;:,.<>?")
	Alphanumeric     = append(append([]rune{}, LowercaseLetters...), append(UppercaseLetters, Digits...)...)
)

// Default policies.
var (
	// BasicPolicy: 8-16 chars, must have lowercase, uppercase, digit.
	BasicPolicy = PasswordPolicy{
		Name:                 "basic",
		MinLength:            basicPolicyMinLength,
		MaxLength:            basicPolicyMaxLength,
		AllowDuplicates:      true,
		AllowAdjacentRepeats: false,
		StartCharacters:      append([]rune{}, LowercaseLetters...),
		EndCharacters:        append(append([]rune{}, Digits...), SpecialChars...),
		CharSets: []CharSetConfig{
			{Name: "lowercase", Characters: LowercaseLetters, Min: 1, Max: MaxInt},
			{Name: "uppercase", Characters: UppercaseLetters, Min: 1, Max: MaxInt},
			{Name: "digits", Characters: Digits, Min: 1, Max: MaxInt},
		},
	}

	// StrongPolicy: 12-24 chars, must have lowercase, uppercase, digit, special.
	StrongPolicy = PasswordPolicy{
		Name:                 "strong",
		MinLength:            strongPolicyMinLength,
		MaxLength:            strongPolicyMaxLength,
		AllowDuplicates:      true,
		AllowAdjacentRepeats: false,
		StartCharacters:      append(append([]rune{}, LowercaseLetters...), UppercaseLetters...),
		EndCharacters:        append(append([]rune{}, Digits...), SpecialChars...),
		CharSets: []CharSetConfig{
			{Name: "lowercase", Characters: LowercaseLetters, Min: 1, Max: MaxInt},
			{Name: "uppercase", Characters: UppercaseLetters, Min: 1, Max: MaxInt},
			{Name: "digits", Characters: Digits, Min: 1, Max: MaxInt},
			{Name: "special", Characters: SpecialChars, Min: 1, Max: MaxInt},
		},
	}

	// EnterprisePolicy: 16-32 chars, strict requirements.
	EnterprisePolicy = PasswordPolicy{
		Name:                 "enterprise",
		MinLength:            enterprisePolicyMinLength,
		MaxLength:            enterprisePolicyMaxLength,
		AllowDuplicates:      false,
		AllowAdjacentRepeats: false,
		StartCharacters:      append([]rune{}, LowercaseLetters...),
		EndCharacters:        append([]rune{}, Digits...),
		CharSets: []CharSetConfig{
			{Name: "lowercase", Characters: LowercaseLetters, Min: 2, Max: MaxInt},
			{Name: "uppercase", Characters: UppercaseLetters, Min: 2, Max: MaxInt},
			{Name: "digits", Characters: Digits, Min: 2, Max: MaxInt},
			{Name: "special", Characters: SpecialChars, Min: 2, Max: MaxInt},
		},
	}
)

// NewPasswordGenerator creates a new password generator with the given policy.
func NewPasswordGenerator(policy PasswordPolicy) (*PasswordGenerator, error) {
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid password policy: %w", err)
	}

	return &PasswordGenerator{policy: policy}, nil
}

// Validate checks if the policy configuration is valid.
func (p *PasswordPolicy) Validate() error {
	if p.MinLength < 1 {
		return errors.New("MinLength must be at least 1")
	}

	if p.MaxLength < p.MinLength {
		return errors.New("MaxLength must be >= MinLength")
	}

	if len(p.CharSets) == 0 {
		return errors.New("at least one CharSet required")
	}

	totalMinRequired := 0

	for _, cs := range p.CharSets {
		if len(cs.Characters) == 0 {
			return fmt.Errorf("CharSet '%s' has no characters", cs.Name)
		}

		if cs.Min < 0 {
			return fmt.Errorf("CharSet '%s' Min must be >= 0", cs.Name)
		}

		if cs.Max < cs.Min {
			return fmt.Errorf("CharSet '%s' Max must be >= Min", cs.Name)
		}

		totalMinRequired += cs.Min
	}

	if totalMinRequired > p.MaxLength {
		return fmt.Errorf("sum of CharSet minimums (%d) exceeds MaxLength (%d)", totalMinRequired, p.MaxLength)
	}

	return nil
}

// Generate generates a password according to the policy.
func (g *PasswordGenerator) Generate() (string, error) {
	// Determine password length.
	length, err := g.randomLength()
	if err != nil {
		return "", fmt.Errorf("failed to determine password length: %w", err)
	}

	// Build all available characters from char sets.
	allChars := make([]rune, 0)
	for _, cs := range g.policy.CharSets {
		allChars = append(allChars, cs.Characters...)
	}

	// Generate password meeting all requirements.
	for attempts := 0; attempts < cryptoutilSharedMagic.JoseJADefaultListLimit; attempts++ {
		password, err := g.generateCandidate(length, allChars)
		if err != nil {
			return "", err
		}

		if g.meetsRequirements(password) {
			return password, nil
		}
	}

	return "", errors.New("failed to generate password meeting requirements after 1000 attempts")
}

// randomLength picks a random length between MinLength and MaxLength.
// pwdgenCrandIntFn is an injectable var for crypto/rand.Int, used for testing error paths.
var pwdgenCrandIntFn = func(max *big.Int) (*big.Int, error) { //nolint:gochecknoglobals // Injectable for testing.
	return crand.Int(crand.Reader, max)
}

func (g *PasswordGenerator) randomLength() (int, error) {
	if g.policy.MinLength == g.policy.MaxLength {
		return g.policy.MinLength, nil
	}

	rangeSize := g.policy.MaxLength - g.policy.MinLength + 1

	n, err := pwdgenCrandIntFn(big.NewInt(int64(rangeSize)))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random length: %w", err)
	}

	return g.policy.MinLength + int(n.Int64()), nil
}

// generateCandidate generates a candidate password.
func (g *PasswordGenerator) generateCandidate(length int, allChars []rune) (string, error) {
	password := make([]rune, length)

	// First character from StartCharacters if specified.
	if len(g.policy.StartCharacters) > 0 {
		idx, err := pwdgenCrandIntFn(big.NewInt(int64(len(g.policy.StartCharacters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate start character index: %w", err)
		}

		password[0] = g.policy.StartCharacters[idx.Int64()]
	} else {
		idx, err := pwdgenCrandIntFn(big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate first character index: %w", err)
		}

		password[0] = allChars[idx.Int64()]
	}

	// Last character from EndCharacters if specified.
	if len(g.policy.EndCharacters) > 0 {
		idx, err := pwdgenCrandIntFn(big.NewInt(int64(len(g.policy.EndCharacters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate end character index: %w", err)
		}

		password[length-1] = g.policy.EndCharacters[idx.Int64()]
	} else {
		idx, err := pwdgenCrandIntFn(big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate last character index: %w", err)
		}

		password[length-1] = allChars[idx.Int64()]
	}

	// Fill middle characters.
	for i := 1; i < length-1; i++ {
		for attempts := 0; attempts < cryptoutilSharedMagic.JoseJAMaxMaterials; attempts++ {
			idx, err := pwdgenCrandIntFn(big.NewInt(int64(len(allChars))))
			if err != nil {
				return "", fmt.Errorf("failed to generate random middle character index: %w", err)
			}

			char := allChars[idx.Int64()]

			// Check adjacent repeats.
			if !g.policy.AllowAdjacentRepeats && i > 0 && password[i-1] == char {
				continue
			}

			// Check duplicates.
			if !g.policy.AllowDuplicates && g.containsChar(password[:i], char) {
				continue
			}

			password[i] = char

			break
		}
	}

	return string(password), nil
}

// meetsRequirements checks if password meets all CharSet requirements.
func (g *PasswordGenerator) meetsRequirements(password string) bool {
	passwordRunes := []rune(password)

	// Check for duplicates if not allowed.
	if !g.policy.AllowDuplicates {
		seen := make(map[rune]bool)
		for _, r := range passwordRunes {
			if seen[r] {
				return false // Duplicate found.
			}

			seen[r] = true
		}
	}

	// Check for adjacent repeats if not allowed.
	if !g.policy.AllowAdjacentRepeats {
		for i := 1; i < len(passwordRunes); i++ {
			if passwordRunes[i] == passwordRunes[i-1] {
				return false // Adjacent repeat found.
			}
		}
	}

	// Check CharSet requirements.
	for _, cs := range g.policy.CharSets {
		count := 0

		for _, pr := range passwordRunes {
			if g.containsRune(cs.Characters, pr) {
				count++
			}
		}

		if count < cs.Min || count > cs.Max {
			return false
		}
	}

	return true
}

// containsChar checks if a character appears in a rune slice.
func (g *PasswordGenerator) containsChar(runes []rune, char rune) bool {
	for _, r := range runes {
		if r == char {
			return true
		}
	}

	return false
}

// containsRune checks if a rune appears in a slice.
func (g *PasswordGenerator) containsRune(haystack []rune, needle rune) bool {
	for _, r := range haystack {
		if r == needle {
			return true
		}
	}

	return false
}
