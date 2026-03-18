// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	crand "crypto/rand"
	"encoding/base64"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/pbkdf2"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAuthenticator_VerifyPasswordErrors(t *testing.T) {
	t.Parallel()

	initialConfig := &Config{
		Realms: []RealmConfig{
			{
				ID:      testRealmID1,
				Name:    "test-realm",
				Type:    RealmTypeFile,
				Enabled: true,
			},
		},
		Defaults: RealmDefaults{
			PasswordPolicy: DefaultPasswordPolicy(),
		},
	}

	auth, err := NewAuthenticator(initialConfig)
	require.NoError(t, err)

	policy := &PasswordPolicyConfig{
		Algorithm:  cryptoutilSharedMagic.PBKDF2DefaultAlgorithm,
		Iterations: cryptoutilSharedMagic.IMPBKDF2Iterations,
		SaltBytes:  cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
		HashBytes:  cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
	}

	tests := []struct {
		name     string
		hash     string
		password string
	}{
		{name: "invalid hash format", hash: "invalid", password: googleUuid.Must(googleUuid.NewV7()).String()},
		{name: "wrong algorithm", hash: "$bcrypt$10$salt$hash", password: googleUuid.Must(googleUuid.NewV7()).String()},
		{name: "invalid iterations", hash: "$pbkdf2-sha256$abc$salt$hash", password: googleUuid.Must(googleUuid.NewV7()).String()},
		{name: "invalid salt encoding", hash: "$pbkdf2-sha256$10000$!!!invalid!!!$hash", password: googleUuid.Must(googleUuid.NewV7()).String()},
		{name: "too few hash parts", hash: "$pbkdf2-sha256$10000", password: googleUuid.Must(googleUuid.NewV7()).String()},
		{name: "empty hash", hash: "", password: googleUuid.Must(googleUuid.NewV7()).String()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := auth.verifyPassword(tc.password, tc.hash, policy)
			require.Error(t, err)
		})
	}
}

// createTestPasswordHash creates a PBKDF2-SHA256 password hash for testing.
func createTestPasswordHash(t *testing.T, password string) string {
	t.Helper()

	salt := make([]byte, cryptoutilSharedMagic.PBKDF2DefaultSaltBytes)
	_, err := crand.Read(salt)
	require.NoError(t, err)

	hashFunc := cryptoutilSharedMagic.PBKDF2HashFunction(cryptoutilSharedMagic.PBKDF2DefaultAlgorithm)
	derivedKey := pbkdf2.Key(
		[]byte(password),
		salt,
		cryptoutilSharedMagic.PBKDF2DefaultIterations,
		cryptoutilSharedMagic.PBKDF2DefaultHashBytes,
		hashFunc,
	)

	return "$" + cryptoutilSharedMagic.PBKDF2DefaultHashName + "$" +
		"600000$" +
		base64.StdEncoding.EncodeToString(salt) + "$" +
		base64.StdEncoding.EncodeToString(derivedKey)
}
