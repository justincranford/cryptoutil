// Copyright (c) 2025 Justin Cranford

package realms

import (
	"bytes"
	json "encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestValidatePasswordForRealm_BoundaryUniqueChars kills CONDITIONALS_BOUNDARY
// at realm_validation.go:90 (len(uniqueChars) < PasswordMinUniqueChars → <=).
func TestValidatePasswordForRealm_BoundaryUniqueChars(t *testing.T) {
	t.Parallel()

	realm := DefaultRealm()
	// DefaultRealm PasswordMinUniqueChars = 8.
	// Build a 12-char password with EXACTLY 8 unique characters: A,b,1,!,C,d,2,@.
	password := "Ab1!Cd2@Ab1!" // pragma: allowlist secret - Test vector for realm validation.
	uniqueChars := make(map[rune]bool)

	for _, r := range password {
		uniqueChars[r] = true
	}

	require.Equal(t, cryptoutilSharedMagic.CipherDefaultPasswordMinUniqueChars, len(uniqueChars),
		"test password must have exactly MinUniqueChars unique characters")

	err := ValidatePasswordForRealm(password, realm)
	require.NoError(t, err, "password with exactly MinUniqueChars unique chars should be valid")
}

// TestValidatePasswordForRealm_BoundaryRepeatedChars kills CONDITIONALS_BOUNDARY
// at realm_validation.go:62 (maxRepeatedSeen >= PasswordMaxRepeatedChars → >).
func TestValidatePasswordForRealm_BoundaryRepeatedChars(t *testing.T) {
	t.Parallel()

	realm := DefaultRealm()
	// DefaultRealm PasswordMaxRepeatedChars = 3.
	// Code: repeatedCount tracks consecutive repeats AFTER first char.
	// "aaa" = repeatedCount=2, "aaaa" = repeatedCount=3.
	// Check `maxRepeatedSeen >= 3` means 3 repeats (4 consecutive chars) fails.
	// Boundary test: "aaa" (2 repeats) should PASS, "aaaa" (3 repeats) should FAIL.

	// Exactly at boundary: 3 consecutive same chars (2 repeats) = should pass.
	passBoundary := "Ab1!aaa2Cd3@" // pragma: allowlist secret - Test vector for realm validation.
	err := ValidatePasswordForRealm(passBoundary, realm)
	require.NoError(t, err, "password with exactly MaxRepeatedChars-1 consecutive repeats should be valid")

	// One above boundary: 4 consecutive same chars (3 repeats) = should fail.
	failBoundary := "Ab1!aaaa2Cd@" // pragma: allowlist secret - Test vector for realm validation.
	err = ValidatePasswordForRealm(failBoundary, realm)
	require.Error(t, err, "password with MaxRepeatedChars consecutive repeats should be invalid")
	require.Contains(t, err.Error(), "consecutive repeated characters")
}

// TestRegisterUser_BoundaryUsernameLengths kills CONDITIONALS_BOUNDARY mutants
// at service.go:86 (len(username) < minUsernameLength → <=)
// and service.go:90 (len(username) > maxUsernameLength → >=).
func TestRegisterUser_BoundaryUsernameLengths(t *testing.T) {
	t.Parallel()

	// Create a service with mock deps — testing validation before any DB calls.
	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	s := NewUserService(repo, factory)

	tests := []struct {
		name     string
		username string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "exact min username length",
			username: strings.Repeat("a", 3), // minUsernameLength=3.
			wantErr:  false,
		},
		{
			name:     "below min username length",
			username: strings.Repeat("a", 2),
			wantErr:  true,
			errMsg:   "username must be at least",
		},
		{
			name:     "exact max username length",
			username: strings.Repeat("a", 50), // maxUsernameLength=50.
			wantErr:  false,
		},
		{
			name:     "above max username length",
			username: strings.Repeat("a", 51),
			wantErr:  true,
			errMsg:   "username cannot exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// RegisterUser validates username first, then password.
			// Use valid password so username validation is the focus.
			_, err := s.RegisterUser(nil, tt.username, "ValidPass123!") //nolint:staticcheck // pragma: allowlist secret - nil ctx OK for unit test.

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				// Username validation passed, may fail on password hash or DB — that's fine.
				if err != nil {
					// Verify it's NOT a username validation error.
					require.NotContains(t, err.Error(), "username must be at least")
					require.NotContains(t, err.Error(), "username cannot exceed")
				}
			}
		})
	}
}

// TestRegisterUser_BoundaryPasswordLength kills CONDITIONALS_BOUNDARY mutant
// at service.go:99 (len(password) < DefaultPasswordMinLengthChars → <=).
func TestRegisterUser_BoundaryPasswordLength(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	s := NewUserService(repo, factory)

	// Exact boundary: password with exactly DefaultPasswordMinLengthChars (8) chars.
	password := strings.Repeat("p", cryptoutilSharedMagic.DefaultPasswordMinLengthChars) // pragma: allowlist secret
	_, err := s.RegisterUser(nil, "testuser", password)                                  //nolint:staticcheck // nil ctx OK for unit test — hits mock repo.
	// Password length validation should pass; may succeed or fail on hash — verify no length error.
	if err != nil {
		require.NotContains(t, err.Error(), "password must be at least")
	}
}

// TestValidateUsernameForRealm_BoundaryLengths kills CONDITIONALS_BOUNDARY
// at realm_validation.go:110 (len(username) < MinUsernameLength → <=)
// and realm_validation.go:115 (len(username) > MaxUsernameLength → >=).
func TestValidateUsernameForRealm_BoundaryLengths(t *testing.T) {
	t.Parallel()

	realm := DefaultRealm()

	// Exact min boundary (3 = MinUsernameLength) — should pass.
	err := ValidateUsernameForRealm(strings.Repeat("a", MinUsernameLength), realm)
	require.NoError(t, err, "username with exactly MinUsernameLength should be valid")

	// One below min (2) — should fail.
	err = ValidateUsernameForRealm(strings.Repeat("a", MinUsernameLength-1), realm)
	require.Error(t, err)

	// Exact max boundary (64 = MaxUsernameLength) — should pass.
	err = ValidateUsernameForRealm(strings.Repeat("a", MaxUsernameLength), realm)
	require.NoError(t, err, "username with exactly MaxUsernameLength should be valid")

	// One above max (65) — should fail.
	err = ValidateUsernameForRealm(strings.Repeat("a", MaxUsernameLength+1), realm)
	require.Error(t, err)
}

// TestHandleRegisterUser_BoundaryValues kills CONDITIONALS_BOUNDARY mutants
// at handlers.go:83 (len < CipherMinUsernameLength → <=),
// handlers.go:84 (len > CipherMaxUsernameLength → >=),
// and handlers.go:91 (len < CipherMinPasswordLength → <=).
func TestHandleRegisterUser_BoundaryValues(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/register", svc.HandleRegisterUser())

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
	}{
		{
			name:       "username exact min length passes",
			username:   strings.Repeat("u", cryptoutilSharedMagic.CipherMinUsernameLength),
			password:   "ValidPass123!", // pragma: allowlist secret
			wantStatus: fiber.StatusCreated,
		},
		{
			name:       "username exact max length passes",
			username:   strings.Repeat("u", cryptoutilSharedMagic.CipherMaxUsernameLength),
			password:   "ValidPass123!", // pragma: allowlist secret
			wantStatus: fiber.StatusCreated,
		},
		{
			name:       "password exact min length passes",
			username:   "validuser",
			password:   strings.Repeat("p", cryptoutilSharedMagic.CipherMinPasswordLength),
			wantStatus: fiber.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bodyBytes, err := json.Marshal(map[string]string{
				"username": tt.username,
				"password": tt.password,
			})
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Boundary values should pass handler validation (201 = created by mock repo).
			require.Equal(t, tt.wantStatus, resp.StatusCode,
				"unexpected status for %s: body=%s", tt.name, string(body))
		})
	}
}
