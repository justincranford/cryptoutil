// Copyright (c) 2025 Justin Cranford
//
//

package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	usernamePrefix    = "u"
	usernameSuffix    = ""
	passwordPrefix    = "p"
	passwordSuffix    = ""
	domainPrefix      = "d"
	domainSuffix      = ".com"
	usernameMinLength = 3
	usernameMaxLength = 64
	passwordMinLength = 8
	passwordMaxLength = 64
	domainMinLength   = 5
	domainMaxLength   = 255
)

// GenerateUsername generates a random username of the specified length for testing.
func GenerateUsername(t *testing.T, length int) *string {
	return generateValue(t, usernamePrefix, usernameSuffix, length, usernameMinLength, usernameMaxLength)
}

// GeneratePassword generates a random password of the specified length for testing.
func GeneratePassword(t *testing.T, length int) *string {
	return generateValue(t, passwordPrefix, passwordSuffix, length, passwordMinLength, passwordMaxLength)
}

// GenerateDomain generates a random domain of the specified length for testing.
func GenerateDomain(t *testing.T, length int) *string {
	return generateValue(t, domainPrefix, domainSuffix, length, domainMinLength, domainMaxLength)
}

// GenerateEmailAddress generates a random email address for testing.
func GenerateEmailAddress(t *testing.T, usernameLength, domainLength int) *string {
	username := GenerateUsername(t, usernameLength)
	domain := GenerateDomain(t, domainLength)

	emailAddress := *username + "@" + *domain

	return &emailAddress
}

func generateValue(t *testing.T, prefix string, suffix string, length, minLength, maxLength int) *string {
	require.GreaterOrEqual(t, len(prefix), 1, "prefix must be at least length 1")
	require.GreaterOrEqual(t, minLength, 1, "minLength must be at least length 1")
	require.GreaterOrEqual(t, maxLength, 1, "maxLength must be at least length 1")
	require.GreaterOrEqual(t, maxLength, minLength, "maxLength must be greater than or equal to minLength")
	require.Greater(t, length, len(prefix), "length must be greater than prefix length %d", len(prefix))
	require.GreaterOrEqual(t, length, minLength, "length must be at least minLength %d", minLength)
	require.LessOrEqual(t, length, maxLength, "length must be at most %d", maxLength)

	randomSuffix, err := GenerateString(length - len(prefix) - len(suffix))
	require.NoError(t, err, "failed to generate random string of %d random characters: %w", length-len(prefix)-len(suffix), err)

	value := prefix + randomSuffix + suffix

	return &value
}
