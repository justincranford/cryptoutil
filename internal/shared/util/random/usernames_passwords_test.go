// Copyright (c) 2025 Justin Cranford
//
//

package random

import (
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
			length:     cryptoutilSharedMagic.IMMinUsernameLength,
			wantLength: cryptoutilSharedMagic.IMMinUsernameLength,
		},
		{
			name:       "maximum username length 50",
			length:     cryptoutilSharedMagic.IMMaxUsernameLength,
			wantLength: cryptoutilSharedMagic.IMMaxUsernameLength,
		},
	}
	passwordTests = []PasswordTest{
		{
			name:       "minimum password length 8",
			length:     cryptoutilSharedMagic.IMMinPasswordLength,
			wantLength: cryptoutilSharedMagic.IMMinPasswordLength,
		},
		{
			name:       "maximum password length 128",
			length:     cryptoutilSharedMagic.MaxPasswordLength,
			wantLength: cryptoutilSharedMagic.MaxPasswordLength,
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
			length:     cryptoutilSharedMagic.EmailDomainMaxLength,
			wantLength: cryptoutilSharedMagic.EmailDomainMaxLength,
		},
	}
	emailAddressTests = []EmailAddressTest{
		{
			name:           "minimum email address length 8",
			usernameLength: cryptoutilSharedMagic.IMMinUsernameLength,
			domainLength:   domainMinLength,
			wantLength:     cryptoutilSharedMagic.IMMinUsernameLength + 1 + domainMinLength,
		},
		{
			name:           "maximum email address length 306",
			usernameLength: cryptoutilSharedMagic.IMMaxUsernameLength,
			domainLength:   cryptoutilSharedMagic.EmailDomainMaxLength,
			wantLength:     cryptoutilSharedMagic.IMMaxUsernameLength + 1 + cryptoutilSharedMagic.EmailDomainMaxLength,
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
