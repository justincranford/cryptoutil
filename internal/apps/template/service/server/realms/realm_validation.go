// Copyright (c) 2025 Justin Cranford
//
//

package realms

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	// MinUsernameLength is the minimum username length constraint.
	MinUsernameLength = 3
	// MaxUsernameLength is the maximum username length constraint.
	MaxUsernameLength = 64
)

// ValidatePasswordForRealm validates a password against realm-specific rules.
// Returns nil if valid, error with specific violation if invalid.
func ValidatePasswordForRealm(password string, realm *RealmConfig) error {
	if realm == nil {
		return fmt.Errorf("realm configuration is nil")
	}

	// Check minimum length.
	if len(password) < realm.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", realm.PasswordMinLength)
	}

	// Track character types and unique characters.
	var (
		hasUppercase    bool
		hasLowercase    bool
		hasDigit        bool
		hasSpecial      bool
		uniqueChars     = make(map[rune]bool)
		repeatedCount   int
		maxRepeatedSeen int
		prevRune        rune
	)

	for i, r := range password {
		uniqueChars[r] = true

		// Check character types.
		switch {
		case unicode.IsUpper(r):
			hasUppercase = true
		case unicode.IsLower(r):
			hasLowercase = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}

		// Track consecutive repeated characters.
		if i > 0 && r == prevRune {
			repeatedCount++
			if repeatedCount > maxRepeatedSeen {
				maxRepeatedSeen = repeatedCount
			}
		} else {
			repeatedCount = 0
		}

		prevRune = r
	}

	// Validate character type requirements.
	if realm.PasswordRequireUppercase && !hasUppercase {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if realm.PasswordRequireLowercase && !hasLowercase {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if realm.PasswordRequireDigits && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if realm.PasswordRequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Validate unique characters requirement.
	if len(uniqueChars) < realm.PasswordMinUniqueChars {
		return fmt.Errorf("password must contain at least %d unique characters", realm.PasswordMinUniqueChars)
	}

	// Validate maximum consecutive repeated characters.
	if maxRepeatedSeen >= realm.PasswordMaxRepeatedChars {
		return fmt.Errorf("password must not contain more than %d consecutive repeated characters", realm.PasswordMaxRepeatedChars)
	}

	return nil
}

// ValidateUsernameForRealm validates a username (currently just length, future: realm-specific rules).
func ValidateUsernameForRealm(username string, realm *RealmConfig) error { //nolint:revive // Future realm-specific rules.
	if realm == nil {
		return fmt.Errorf("realm configuration is nil")
	}

	// Basic validation (length).
	username = strings.TrimSpace(username)
	if len(username) < MinUsernameLength {
		return fmt.Errorf("username must be at least %d characters long", MinUsernameLength)
	}

	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username must not exceed %d characters", MaxUsernameLength)
	}

	// Future: Realm-specific username rules (allowed characters, patterns, etc.).
	return nil
}
