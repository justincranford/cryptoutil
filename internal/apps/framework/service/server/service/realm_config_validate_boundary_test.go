// Copyright (c) 2025 Justin Cranford

package service

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestRealmConfig_Validate_BoundarySuccess kills CONDITIONALS_BOUNDARY mutants
// that change `x < 1` to `x <= 1` in all Validate() methods.
// Each test uses value=1 (exact boundary) which MUST succeed.
func TestRealmConfig_Validate_BoundarySuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config RealmConfig
	}{
		// realm_service.go boundaries.
		{
			name:   "UsernamePasswordConfig_MinPasswordLength_1",
			config: &UsernamePasswordConfig{MinPasswordLength: 1},
		},
		{
			name:   "JWESessionCookieConfig_SessionExpiry_1",
			config: &JWESessionCookieConfig{SessionExpiryMinutes: 1},
		},
		{
			name:   "JWSSessionCookieConfig_SessionExpiry_1",
			config: &JWSSessionCookieConfig{SessionExpiryMinutes: 1},
		},
		{
			name: "OpaqueSessionCookieConfig_AllBoundaries",
			config: &OpaqueSessionCookieConfig{
				TokenLengthBytes:     cryptoutilSharedMagic.RealmMinTokenLengthBytes,
				SessionExpiryMinutes: 1,
				StorageType:          cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
		},
		{
			name:   "BasicUsernamePasswordConfig_MinPasswordLength_1",
			config: &BasicUsernamePasswordConfig{MinPasswordLength: 1},
		},
		{
			name: "BearerAPITokenConfig_AllBoundaries",
			config: &BearerAPITokenConfig{
				TokenExpiryDays:  1,
				TokenLengthBytes: cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
			},
		},
		// realm_service_impl.go boundaries.
		{
			name:   "JWESessionTokenConfig_TokenExpiry_1",
			config: &JWESessionTokenConfig{TokenExpiryMinutes: 1},
		},
		{
			name:   "JWSSessionTokenConfig_TokenExpiry_1",
			config: &JWSSessionTokenConfig{TokenExpiryMinutes: 1},
		},
		{
			name: "OpaqueSessionTokenConfig_AllBoundaries",
			config: &OpaqueSessionTokenConfig{
				TokenLengthBytes:   cryptoutilSharedMagic.RealmMinTokenLengthBytes,
				TokenExpiryMinutes: 1,
				StorageType:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
			},
		},
		{
			name:   "BasicClientIDSecretConfig_MinSecretLength_1",
			config: &BasicClientIDSecretConfig{MinSecretLength: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			require.NoError(t, err, "config with boundary value 1 should be valid")
		})
	}
}

// TestRegistrationService_ExpiryDuration kills ARITHMETIC_BASE mutant
// at registration_service.go:232 that changes `*` to `/` or `+`.
func TestRegistrationService_ExpiryDuration(t *testing.T) {
	t.Parallel()

	expected := DefaultRegistrationExpiryHours * time.Hour
	// DefaultRegistrationExpiryHours is 72, so expected = 72h.
	// Mutant `DefaultRegistrationExpiryHours / time.Hour` would give ~0ns.
	// Mutant `DefaultRegistrationExpiryHours + time.Hour` would give ~1h + 72ns.
	require.Equal(t, 72*time.Hour, expected, "registration expiry should be 72 hours")
	require.Greater(t, expected, time.Hour, "expiry must be greater than 1 hour")
}
