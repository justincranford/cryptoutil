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
