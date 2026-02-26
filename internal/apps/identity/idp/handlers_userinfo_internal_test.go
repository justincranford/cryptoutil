// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestAddScopeBasedClaims(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	userID := googleUuid.New()

	baseUser := &cryptoutilIdentityDomain.User{
		ID:                userID,
		Sub:               "testuser",
		Name:              "Test User",
		GivenName:         "Test",
		FamilyName:        "User",
		MiddleName:        "Middle",
		Nickname:          "tester",
		PreferredUsername: "testuser",
		Profile:           "https://example.com/profile",
		Picture:           "https://example.com/picture.png",
		Website:           "https://example.com",
		Gender:            "other",
		Birthdate:         "1990-01-01",
		Zoneinfo:          "America/New_York",
		Locale:            "en-US",
		Email:             "test@example.com",
		EmailVerified:     true,
		PhoneNumber:       "+1234567890",
		PhoneVerified:     true,
		Address: &cryptoutilIdentityDomain.Address{
			Formatted:     "123 Main St, City, State 12345",
			StreetAddress: "123 Main St",
			Locality:      "City",
			Region:        "State",
			PostalCode:    "12345",
			Country:       "US",
		},
		UpdatedAt: now,
	}

	tests := []struct {
		name           string
		scopes         []string
		expectedClaims []string
		missingClaims  []string
	}{
		{
			name:           "profile scope",
			scopes:         []string{cryptoutilSharedMagic.ClaimProfile},
			expectedClaims: []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimGivenName, cryptoutilSharedMagic.ClaimFamilyName, cryptoutilSharedMagic.ClaimMiddleName, cryptoutilSharedMagic.ClaimNickname, cryptoutilSharedMagic.ClaimPreferredUsername, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimPicture, cryptoutilSharedMagic.ClaimWebsite, cryptoutilSharedMagic.ClaimGender, cryptoutilSharedMagic.ClaimBirthdate, cryptoutilSharedMagic.ClaimZoneinfo, cryptoutilSharedMagic.ClaimLocale, cryptoutilSharedMagic.ClaimUpdatedAt},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimEmailVerified, cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ClaimPhoneNumber, cryptoutilSharedMagic.ClaimPhoneVerified},
		},
		{
			name:           "email scope",
			scopes:         []string{cryptoutilSharedMagic.ClaimEmail},
			expectedClaims: []string{cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimEmailVerified},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ClaimPhoneNumber},
		},
		{
			name:           "phone scope",
			scopes:         []string{cryptoutilSharedMagic.ScopePhone},
			expectedClaims: []string{cryptoutilSharedMagic.ClaimPhoneNumber, cryptoutilSharedMagic.ClaimPhoneVerified},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimAddress},
		},
		{
			name:           "address scope",
			scopes:         []string{cryptoutilSharedMagic.ClaimAddress},
			expectedClaims: []string{cryptoutilSharedMagic.ClaimAddress},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimPhoneNumber},
		},
		{
			name:           "multiple scopes",
			scopes:         []string{cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
			expectedClaims: []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimEmailVerified},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimAddress, cryptoutilSharedMagic.ClaimPhoneNumber},
		},
		{
			name:           "all scopes",
			scopes:         []string{cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ScopePhone, cryptoutilSharedMagic.ClaimAddress},
			expectedClaims: []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimPhoneNumber, cryptoutilSharedMagic.ClaimAddress},
			missingClaims:  []string{},
		},
		{
			name:           "empty scopes",
			scopes:         []string{},
			expectedClaims: []string{},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimPhoneNumber, cryptoutilSharedMagic.ClaimAddress},
		},
		{
			name:           "openid only",
			scopes:         []string{cryptoutilSharedMagic.ScopeOpenID},
			expectedClaims: []string{},
			missingClaims:  []string{cryptoutilSharedMagic.ClaimName, cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimPhoneNumber, cryptoutilSharedMagic.ClaimAddress},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userInfo := make(map[string]any)
			addScopeBasedClaims(userInfo, tc.scopes, baseUser)

			for _, claim := range tc.expectedClaims {
				testify.Contains(t, userInfo, claim, "expected claim %s not found", claim)
			}

			for _, claim := range tc.missingClaims {
				testify.NotContains(t, userInfo, claim, "unexpected claim %s found", claim)
			}
		})
	}
}

func TestAddScopeBasedClaimsNilAddress(t *testing.T) {
	t.Parallel()

	userWithNoAddress := &cryptoutilIdentityDomain.User{
		ID:        googleUuid.New(),
		Sub:       "testuser",
		Address:   nil,
		UpdatedAt: time.Now().UTC(),
	}

	userInfo := make(map[string]any)
	addScopeBasedClaims(userInfo, []string{cryptoutilSharedMagic.ClaimAddress}, userWithNoAddress)

	// Address should not be present when user.Address is nil.
	testify.NotContains(t, userInfo, cryptoutilSharedMagic.ClaimAddress)
}

func TestAddScopeBasedClaimsProfileValues(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	user := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.New(),
		Sub:               "testuser",
		Name:              "John Doe",
		GivenName:         "John",
		FamilyName:        "Doe",
		MiddleName:        "William",
		Nickname:          "JD",
		PreferredUsername: "johndoe",
		Profile:           "https://example.com/johndoe",
		Picture:           "https://example.com/johndoe.png",
		Website:           "https://johndoe.com",
		Gender:            "male",
		Birthdate:         "1985-05-15",
		Zoneinfo:          "Europe/London",
		Locale:            "en-GB",
		UpdatedAt:         now,
	}

	userInfo := make(map[string]any)
	addScopeBasedClaims(userInfo, []string{cryptoutilSharedMagic.ClaimProfile}, user)

	testify.Equal(t, "John Doe", userInfo[cryptoutilSharedMagic.ClaimName])
	testify.Equal(t, "John", userInfo[cryptoutilSharedMagic.ClaimGivenName])
	testify.Equal(t, "Doe", userInfo[cryptoutilSharedMagic.ClaimFamilyName])
	testify.Equal(t, "William", userInfo[cryptoutilSharedMagic.ClaimMiddleName])
	testify.Equal(t, "JD", userInfo[cryptoutilSharedMagic.ClaimNickname])
	testify.Equal(t, "johndoe", userInfo[cryptoutilSharedMagic.ClaimPreferredUsername])
	testify.Equal(t, "https://example.com/johndoe", userInfo[cryptoutilSharedMagic.ClaimProfile])
	testify.Equal(t, "https://example.com/johndoe.png", userInfo[cryptoutilSharedMagic.ClaimPicture])
	testify.Equal(t, "https://johndoe.com", userInfo[cryptoutilSharedMagic.ClaimWebsite])
	testify.Equal(t, "male", userInfo[cryptoutilSharedMagic.ClaimGender])
	testify.Equal(t, "1985-05-15", userInfo[cryptoutilSharedMagic.ClaimBirthdate])
	testify.Equal(t, "Europe/London", userInfo[cryptoutilSharedMagic.ClaimZoneinfo])
	testify.Equal(t, "en-GB", userInfo[cryptoutilSharedMagic.ClaimLocale])
	testify.Equal(t, now.Unix(), userInfo[cryptoutilSharedMagic.ClaimUpdatedAt])
}

func TestAddScopeBasedClaimsAddressValues(t *testing.T) {
	t.Parallel()

	user := &cryptoutilIdentityDomain.User{
		ID:  googleUuid.New(),
		Sub: "testuser",
		Address: &cryptoutilIdentityDomain.Address{
			Formatted:     "456 Oak Ave, Suite 100, Metro City, MC 67890, Canada",
			StreetAddress: "456 Oak Ave, Suite 100",
			Locality:      "Metro City",
			Region:        "MC",
			PostalCode:    "67890",
			Country:       "Canada",
		},
		UpdatedAt: time.Now().UTC(),
	}

	userInfo := make(map[string]any)
	addScopeBasedClaims(userInfo, []string{cryptoutilSharedMagic.ClaimAddress}, user)

	addressMap, ok := userInfo[cryptoutilSharedMagic.ClaimAddress].(map[string]any)
	testify.True(t, ok)
	testify.Equal(t, "456 Oak Ave, Suite 100, Metro City, MC 67890, Canada", addressMap[cryptoutilSharedMagic.AddressFormatted])
	testify.Equal(t, "456 Oak Ave, Suite 100", addressMap[cryptoutilSharedMagic.AddressStreetAddress])
	testify.Equal(t, "Metro City", addressMap[cryptoutilSharedMagic.AddressLocality])
	testify.Equal(t, "MC", addressMap[cryptoutilSharedMagic.AddressRegion])
	testify.Equal(t, "67890", addressMap[cryptoutilSharedMagic.AddressPostalCode])
	testify.Equal(t, "Canada", addressMap[cryptoutilSharedMagic.AddressCountry])
}
