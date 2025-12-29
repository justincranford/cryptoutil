// Copyright (c) 2025 Justin Cranford
//
//

package random

import (
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2/log"
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

func GenerateUsername(t *testing.T, length int) *string {
	return generateValue(t, usernamePrefix, usernameSuffix, length, usernameMinLength, usernameMaxLength)
}

func GeneratePassword(t *testing.T, length int) *string {
	return generateValue(t, passwordPrefix, passwordSuffix, length, passwordMinLength, passwordMaxLength)
}

func GenerateDomain(t *testing.T, length int) *string {
	return generateValue(t, domainPrefix, domainSuffix, length, domainMinLength, domainMaxLength)
}

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

type UsernameTest struct {
	name       string
	length     int
	wantLength int
}

type PasswordTest struct {
	name       string
	length     int
	wantLength int
}

type DomainTest struct {
	name       string
	length     int
	wantLength int
}

type EmailAddressTest struct {
	name           string
	usernameLength int
	domainLength   int
	wantLength     int
}

var (
	usernameTests = []UsernameTest{
		{
			name:       "minimum username length 3",
			length:     usernameMinLength,
			wantLength: usernameMinLength,
		},
		{
			name:       "maximum username length 64",
			length:     usernameMaxLength,
			wantLength: usernameMaxLength,
		},
	}
	passwordTests = []PasswordTest{
		{
			name:       "minimum password length 8",
			length:     passwordMinLength,
			wantLength: passwordMinLength,
		},
		{
			name:       "maximum password length 64",
			length:     passwordMaxLength,
			wantLength: passwordMaxLength,
		},
	}
	domainTests = []DomainTest{
		{
			name:       "minimum domain length 5",
			length:     domainMinLength,
			wantLength: domainMinLength,
		},
		{
			name:       "maximum domain length 255",
			length:     domainMaxLength,
			wantLength: domainMaxLength,
		},
	}
	emailAddressTests = []EmailAddressTest{
		{
			name:           "minimum email address length 8",
			usernameLength: usernameMinLength,
			domainLength:   domainMinLength,
			wantLength:     usernameMinLength + 1 + domainMinLength,
		},
		{
			name:           "maximum email address length 320",
			usernameLength: usernameMaxLength,
			domainLength:   domainMaxLength,
			wantLength:     usernameMaxLength + 1 + domainMaxLength,
		},
	}
)

func TestGenerateUsername_HappyPath(t *testing.T) {
	t.Parallel()
	for _, tt := range usernameTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validateGeneratedValue(t, GenerateUsername(t, tt.length), usernamePrefix, usernameSuffix, tt.length, tt.wantLength)
		})
	}
}

func TestGeneratePassword_HappyPath(t *testing.T) {
	t.Parallel()
	for _, tt := range passwordTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validateGeneratedValue(t, GeneratePassword(t, tt.length), passwordPrefix, passwordSuffix, tt.length, tt.wantLength)
		})
	}
}

func TestGenerateDomain_HappyPath(t *testing.T) {
	t.Parallel()
	for _, tt := range domainTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validateGeneratedValue(t, GenerateDomain(t, tt.length), domainPrefix, domainSuffix, tt.length, tt.wantLength)
		})
	}
}

func TestGenerateEmailAddress_HappyPath(t *testing.T) {
	t.Parallel()
	for _, tt := range emailAddressTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validateGeneratedValue(t, GenerateEmailAddress(t, tt.usernameLength, tt.domainLength), usernamePrefix, domainSuffix, tt.usernameLength+1+tt.domainLength, tt.wantLength)
		})
	}
}

func validateGeneratedValue(t *testing.T, result *string, wantPrefix string, wantSuffix string, length int, wantLength int) {
	t.Helper()
	require.NotNil(t, result, "GeneratedValue(%d) should return result", length)
	require.Equal(t, wantLength, len(*result), "GeneratedValue(%d) length mismatch", length)
	require.True(t, strings.HasPrefix(*result, wantPrefix), "GeneratedValue(%d) should start with prefix", length)
	require.True(t, strings.HasSuffix(*result, wantSuffix), "GeneratedValue(%d) should end with suffix", length)
	log.Info("Length: ", len(*result), " Value: ", *result)
}
