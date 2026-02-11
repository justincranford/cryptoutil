// Copyright (c) 2025 Justin Cranford
//
//

package idp

import (
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
			scopes:         []string{"profile"},
			expectedClaims: []string{"name", "given_name", "family_name", "middle_name", "nickname", "preferred_username", "profile", "picture", "website", "gender", "birthdate", "zoneinfo", "locale", "updated_at"},
			missingClaims:  []string{"email", "email_verified", "address", "phone_number", "phone_number_verified"},
		},
		{
			name:           "email scope",
			scopes:         []string{"email"},
			expectedClaims: []string{"email", "email_verified"},
			missingClaims:  []string{"name", "address", "phone_number"},
		},
		{
			name:           "phone scope",
			scopes:         []string{"phone"},
			expectedClaims: []string{"phone_number", "phone_number_verified"},
			missingClaims:  []string{"name", "email", "address"},
		},
		{
			name:           "address scope",
			scopes:         []string{"address"},
			expectedClaims: []string{"address"},
			missingClaims:  []string{"name", "email", "phone_number"},
		},
		{
			name:           "multiple scopes",
			scopes:         []string{"profile", "email"},
			expectedClaims: []string{"name", "email", "email_verified"},
			missingClaims:  []string{"address", "phone_number"},
		},
		{
			name:           "all scopes",
			scopes:         []string{"profile", "email", "phone", "address"},
			expectedClaims: []string{"name", "email", "phone_number", "address"},
			missingClaims:  []string{},
		},
		{
			name:           "empty scopes",
			scopes:         []string{},
			expectedClaims: []string{},
			missingClaims:  []string{"name", "email", "phone_number", "address"},
		},
		{
			name:           "openid only",
			scopes:         []string{"openid"},
			expectedClaims: []string{},
			missingClaims:  []string{"name", "email", "phone_number", "address"},
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
	addScopeBasedClaims(userInfo, []string{"address"}, userWithNoAddress)

	// Address should not be present when user.Address is nil.
	testify.NotContains(t, userInfo, "address")
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
	addScopeBasedClaims(userInfo, []string{"profile"}, user)

	testify.Equal(t, "John Doe", userInfo["name"])
	testify.Equal(t, "John", userInfo["given_name"])
	testify.Equal(t, "Doe", userInfo["family_name"])
	testify.Equal(t, "William", userInfo["middle_name"])
	testify.Equal(t, "JD", userInfo["nickname"])
	testify.Equal(t, "johndoe", userInfo["preferred_username"])
	testify.Equal(t, "https://example.com/johndoe", userInfo["profile"])
	testify.Equal(t, "https://example.com/johndoe.png", userInfo["picture"])
	testify.Equal(t, "https://johndoe.com", userInfo["website"])
	testify.Equal(t, "male", userInfo["gender"])
	testify.Equal(t, "1985-05-15", userInfo["birthdate"])
	testify.Equal(t, "Europe/London", userInfo["zoneinfo"])
	testify.Equal(t, "en-GB", userInfo["locale"])
	testify.Equal(t, now.Unix(), userInfo["updated_at"])
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
	addScopeBasedClaims(userInfo, []string{"address"}, user)

	addressMap, ok := userInfo["address"].(map[string]any)
	testify.True(t, ok)
	testify.Equal(t, "456 Oak Ave, Suite 100, Metro City, MC 67890, Canada", addressMap["formatted"])
	testify.Equal(t, "456 Oak Ave, Suite 100", addressMap["street_address"])
	testify.Equal(t, "Metro City", addressMap["locality"])
	testify.Equal(t, "MC", addressMap["region"])
	testify.Equal(t, "67890", addressMap["postal_code"])
	testify.Equal(t, "Canada", addressMap["country"])
}
