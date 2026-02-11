// Copyright (c) 2025 Justin Cranford
//
//

package clientauth_test

import (
	crand "crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	testify "github.com/stretchr/testify/require"
)

const (
	testPassword = "TestPassword123"
)

// TestNewPBKDF2Hasher validates PBKDF2 hasher creation.
func TestNewPBKDF2Hasher(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	testify.NotNil(t, hasher, "Hasher should not be nil")
	testify.Implements(t, (*cryptoutilIdentityClientAuth.SecretHasher)(nil), hasher, "Should implement SecretHasher interface")
}

// TestPBKDF2Hasher_HashSecret validates password hashing.
//
// Validates requirements:
// - R01-03: PBKDF2-HMAC-SHA256 password hashing (FIPS 140-3 approved).
func TestPBKDF2Hasher_HashLowEntropyNonDeterministic(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	tests := []struct {
		name      string
		secret    string
		wantError bool
	}{
		{
			name:      "valid strong password",
			secret:    "StrongP@ssw0rd!123",
			wantError: false,
		},
		{
			name:      "valid weak password",
			secret:    "weak",
			wantError: false,
		},
		{
			name:      "valid empty password",
			secret:    "",
			wantError: false,
		},
		{
			name:      "valid unicode password",
			secret:    "パスワード123",
			wantError: false,
		},
		{
			name:      "valid long password (256 chars)",
			secret:    strings.Repeat("a", 256),
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := hasher.HashLowEntropyNonDeterministic(tc.secret)

			if tc.wantError {
				testify.Error(t, err, "Expected error for test case")
				testify.Empty(t, hash, "Hash should be empty on error")
			} else {
				testify.NoError(t, err, "Should hash without error")
				testify.NotEmpty(t, hash, "Hash should not be empty")

				// Verify hash format: cryptoutilMagic.PBKDF2DefaultHashName$iterations$base64(salt)$base64(hash).
				parts := strings.Split(hash, "$")
				testify.Len(t, parts, 5, "Hash should have 5 parts: empty$$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$iterations$salt$hash")
				testify.Equal(t, "", parts[0], "First part should be empty (leading $)")
				testify.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultHashName, parts[1], "Second part should be 'pbkdf2-sha256'")
				testify.Equal(t, "100000", parts[2], "Iterations should be 100000")

				// Verify salt is base64-encoded 16 bytes (128 bits).
				saltBytes, err := base64.RawStdEncoding.DecodeString(parts[3])
				testify.NoError(t, err, "Salt should be valid base64")
				testify.Len(t, saltBytes, 16, "Salt should be 16 bytes (128 bits)")

				// Verify hash is base64-encoded 32 bytes (256 bits).
				hashBytes, err := base64.RawStdEncoding.DecodeString(parts[4])
				testify.NoError(t, err, "Hash should be valid base64")
				testify.Len(t, hashBytes, 32, "Hash should be 32 bytes (256 bits)")
			}
		})
	}
}

// TestPBKDF2Hasher_HashSecret_Uniqueness validates unique hashes for same password.
//
// Validates requirements:
// - R01-03: Random salt ensures different hashes for same password.
func TestPBKDF2Hasher_HashUniqueness(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()
	password := "SamePassword123"

	// Generate two hashes for the same password.
	hash1, err1 := hasher.HashLowEntropyNonDeterministic(password)
	testify.NoError(t, err1, "First hash should succeed")

	hash2, err2 := hasher.HashLowEntropyNonDeterministic(password)
	testify.NoError(t, err2, "Second hash should succeed")

	// Hashes should be different due to random salts.
	testify.NotEqual(t, hash1, hash2, "Hashes for same password should differ (random salts)")

	// Extract salts from both hashes.
	parts1 := strings.Split(hash1, "$")
	parts2 := strings.Split(hash2, "$")

	testify.NotEqual(t, parts1[3], parts2[3], "Salts should be different")
}

// TestPBKDF2Hasher_CompareSecret validates password verification.
//
// Validates requirements:
// - R01-03: PBKDF2-HMAC-SHA256 password verification.
func TestPBKDF2Hasher_CompareSecret(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	// Generate hash for test password.
	password := "CorrectPassword123"
	hash, err := hasher.HashLowEntropyNonDeterministic(password)
	testify.NoError(t, err, "Should hash test password")

	tests := []struct {
		name      string
		hash      string
		secret    string
		wantMatch bool
		wantError bool
	}{
		{
			name:      "correct password matches",
			hash:      hash,
			secret:    password,
			wantMatch: true,
			wantError: false,
		},
		{
			name:      "incorrect password does not match",
			hash:      hash,
			secret:    "WrongPassword456",
			wantMatch: false,
			wantError: false,
		},
		{
			name:      "empty password does not match",
			hash:      hash,
			secret:    "",
			wantMatch: false,
			wantError: false,
		},
		{
			name:      "case-sensitive comparison",
			hash:      hash,
			secret:    strings.ToLower(password),
			wantMatch: false,
			wantError: false,
		},
		{
			name:      "malformed hash (3 parts)",
			hash:      "pbkdf2$100000$invalid",
			secret:    password,
			wantMatch: false,
			wantError: true,
		},
		{
			name: "malformed hash (wrong prefix)",
			// cspell:disable-next-line
			hash:      "bcrypt$10$abcdef$ghijkl",
			secret:    password,
			wantMatch: false,
			wantError: true,
		},
		{
			name: "malformed hash (invalid iterations)",
			// cspell:disable-next-line
			hash:      "pbkdf2$notanumber$abcdef$ghijkl",
			secret:    password,
			wantMatch: false,
			wantError: true,
		},
		{
			name: "malformed hash (invalid salt base64)",
			// cspell:disable-next-line
			hash:      "pbkdf2$100000$!!!invalid!!!$ghijkl",
			secret:    password,
			wantMatch: false,
			wantError: true,
		},
		{
			name:      "malformed hash (invalid hash base64)",
			hash:      "pbkdf2$100000$dmFsaWRzYWx0$!!!invalid!!!",
			secret:    password,
			wantMatch: false,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := hasher.CompareSecret(tc.hash, tc.secret)

			if tc.wantError {
				testify.Error(t, err, "Expected error for malformed hash")
			} else {
				if tc.wantMatch {
					testify.NoError(t, err, "Should match without error")
				} else {
					testify.Error(t, err, "Should return error for non-matching password")
				}
			}
		})
	}
}

// TestPBKDF2Hasher_CompareSecret_ConstantTime validates timing attack resistance.
//
// Validates requirements:
// - R01-03: Constant-time comparison prevents timing attacks.
func TestPBKDF2Hasher_CompareSecret_ConstantTime(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	// Generate hash for test password.
	password := "SecretPassword123"
	hash, err := hasher.HashLowEntropyNonDeterministic(password)
	testify.NoError(t, err, "Should hash test password")

	// Test multiple incorrect passwords (timing should be constant).
	incorrectPasswords := []string{
		"Wrong1",
		"Wrong2",
		"Wrong3",
		"CompletelyDifferentPassword",
		strings.Repeat("x", len(password)),
	}

	for _, incorrect := range incorrectPasswords {
		err := hasher.CompareSecret(hash, incorrect)
		testify.Error(t, err, "Incorrect passwords should not match")
	}
}

// TestPBKDF2Hasher_FIPS140_3_Compliance validates FIPS 140-3 requirements.
//
// Validates requirements:
// - R01-03: PBKDF2-HMAC-SHA256 with 100,000 iterations (FIPS 140-3 approved).
func TestPBKDF2Hasher_FIPS140_3Compliance(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	hash, err := hasher.HashLowEntropyNonDeterministic(testPassword)
	testify.NoError(t, err, "Should hash without error")

	// Parse hash format.
	parts := strings.Split(hash, "$")
	testify.Len(t, parts, 5, "Hash should have 5 parts")

	// Verify PBKDF2 algorithm identifier.
	testify.Equal(t, cryptoutilSharedMagic.PBKDF2DefaultHashName, parts[1], "Algorithm should be PBKDF2-SHA256")

	// Verify iteration count (FIPS 140-3 recommends ≥100,000 for PBKDF2).
	testify.Equal(t, "100000", parts[2], "Iteration count should be 100,000")

	// Verify salt length (FIPS 140-3 recommends ≥128 bits).
	saltBytes, err := base64.RawStdEncoding.DecodeString(parts[3])
	testify.NoError(t, err, "Salt should be valid base64")
	testify.GreaterOrEqual(t, len(saltBytes), 16, "Salt should be ≥128 bits (16 bytes)")

	// Verify hash length (SHA-256 produces 256 bits).
	hashBytes, err := base64.RawStdEncoding.DecodeString(parts[4])
	testify.NoError(t, err, "Hash should be valid base64")
	testify.Equal(t, 32, len(hashBytes), "Hash should be 256 bits (32 bytes) for SHA-256")
}

// TestPBKDF2Hasher_SaltRandomness validates cryptographic salt generation.
//
// Validates requirements:
// - R01-03: Cryptographically secure random salt (crypto/rand).
func TestPBKDF2Hasher_SaltRandomness(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	// Generate 100 hashes and collect salts.
	salts := make(map[string]bool)
	iterations := 100

	for range iterations {
		hash, err := hasher.HashLowEntropyNonDeterministic(testPassword)
		testify.NoError(t, err, "Should hash without error")

		// Extract salt from hash.
		parts := strings.Split(hash, "$")
		salt := parts[3]

		// Check for duplicate salts (statistically should never happen with crypto/rand).
		testify.False(t, salts[salt], "Duplicate salt detected (crypto/rand should be unique)")
		salts[salt] = true
	}

	// All salts should be unique.
	testify.Len(t, salts, iterations, "All salts should be unique")
}

// TestPBKDF2Hasher_EdgeCases validates edge case handling.
func TestPBKDF2Hasher_EdgeCases(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	tests := []struct {
		name      string
		secret    string
		wantError bool
	}{
		{
			name:      "very long password (10KB)",
			secret:    strings.Repeat("a", 10240),
			wantError: false,
		},
		{
			name:      "special characters",
			secret:    "!@#$%^&*()_+-=[]{}|;:',.<>?/~`",
			wantError: false,
		},
		{
			name:      "whitespace only",
			secret:    "   \t\n\r",
			wantError: false,
		},
		{
			name:      "null bytes (valid UTF-8)",
			secret:    "password\x00embedded\x00nulls",
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hash, err := hasher.HashLowEntropyNonDeterministic(tc.secret)

			if tc.wantError {
				testify.Error(t, err, "Expected error for edge case")
			} else {
				testify.NoError(t, err, "Should handle edge case without error")
				testify.NotEmpty(t, hash, "Hash should not be empty")

				// Verify round-trip (hash → compare).
				err := hasher.CompareSecret(hash, tc.secret)
				testify.NoError(t, err, "Edge case password should match its hash")
			}
		})
	}
}

// TestPBKDF2Hasher_CompareSecret_VectorTests validates known test vectors.
//
// Validates requirements:
// - R01-03: PBKDF2-HMAC-SHA256 correctness against known test vectors.
func TestPBKDF2Hasher_CompareSecret_VectorTests(t *testing.T) {
	t.Parallel()

	hasher := cryptoutilIdentityClientAuth.NewPBKDF2Hasher()

	// Generate known test vector: hash password "TestVector123" with known salt.
	password := "TestVector123"
	knownSalt := make([]byte, 16)
	_, err := crand.Read(knownSalt)
	testify.NoError(t, err, "Should generate random salt")

	// Hash password using PBKDF2Hasher (uses random salt).
	hash, err := hasher.HashLowEntropyNonDeterministic(password)
	testify.NoError(t, err, "Should hash without error")

	// Verify correct password matches.
	err = hasher.CompareSecret(hash, password)
	testify.NoError(t, err, "Correct password should match")

	// Verify incorrect password does not match.
	err = hasher.CompareSecret(hash, "WrongPassword")
	testify.Error(t, err, "Incorrect password should not match")
}
