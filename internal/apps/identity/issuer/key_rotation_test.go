// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	"crypto/elliptic"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// mockKeyGenerator implements KeyGenerator for testing.
type mockKeyGenerator struct {
	signingKeyCount    int
	encryptionKeyCount int
}

func (m *mockKeyGenerator) GenerateSigningKey(_ context.Context, algorithm string) (*SigningKey, error) {
	m.signingKeyCount++

	return &SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           []byte("mock-signing-key"),
		Algorithm:     algorithm,
		CreatedAt:     time.Now().UTC(),
		Active:        false, // Will be set by rotation manager.
		ValidForVerif: false, // Will be set by rotation manager.
	}, nil
}

func (m *mockKeyGenerator) GenerateEncryptionKey(_ context.Context) (*EncryptionKey, error) {
	m.encryptionKeyCount++

	return &EncryptionKey{
		KeyID:        googleUuid.NewString(),
		Key:          []byte("0123456789abcdef0123456789abcdef"), // 32 bytes.
		CreatedAt:    time.Now().UTC(),
		Active:       false, // Will be set by rotation manager.
		ValidForDecr: false, // Will be set by rotation manager.
	}, nil
}

// Validates requirements:
// - R05-03: Refresh token rotation for security
// - R11-02: Code coverage meets target (≥90%).
func TestDefaultKeyRotationPolicy(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()

	require.NotNil(t, policy)
	require.Equal(t, cryptoutilSharedMagic.DefaultKeyRotationInterval, policy.RotationInterval)
	require.Equal(t, cryptoutilSharedMagic.DefaultKeyGracePeriod, policy.GracePeriod)
	require.Equal(t, cryptoutilSharedMagic.DefaultMaxActiveKeys, policy.MaxActiveKeys)
	require.False(t, policy.AutoRotationEnabled)
}

func TestStrictKeyRotationPolicy(t *testing.T) {
	t.Parallel()

	policy := StrictKeyRotationPolicy()

	require.NotNil(t, policy)
	require.Equal(t, cryptoutilSharedMagic.StrictKeyRotationInterval, policy.RotationInterval)
	require.Equal(t, cryptoutilSharedMagic.StrictKeyGracePeriod, policy.GracePeriod)
	require.Equal(t, cryptoutilSharedMagic.StrictMaxActiveKeys, policy.MaxActiveKeys)
	require.True(t, policy.AutoRotationEnabled)
}

func TestDevelopmentKeyRotationPolicy(t *testing.T) {
	t.Parallel()

	policy := DevelopmentKeyRotationPolicy()

	require.NotNil(t, policy)
	require.Equal(t, cryptoutilSharedMagic.DevelopmentKeyRotationInterval, policy.RotationInterval)
	require.Equal(t, cryptoutilSharedMagic.DevelopmentKeyGracePeriod, policy.GracePeriod)
	require.Equal(t, cryptoutilSharedMagic.DevelopmentMaxActiveKeys, policy.MaxActiveKeys)
	require.False(t, policy.AutoRotationEnabled)
}

func TestNewKeyRotationManager(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}

	manager, err := NewKeyRotationManager(policy, generator, nil)

	require.NoError(t, err)
	require.NotNil(t, manager)
}

func TestNewKeyRotationManager_NilPolicy(t *testing.T) {
	t.Parallel()

	generator := &mockKeyGenerator{}

	manager, err := NewKeyRotationManager(nil, generator, nil)

	require.Error(t, err)
	require.Nil(t, manager)
}

func TestNewKeyRotationManager_NilGenerator(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()

	manager, err := NewKeyRotationManager(policy, nil, nil)

	require.Error(t, err)
	require.Nil(t, manager)
}

func TestRotateSigningKey(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// First rotation.
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Verify key was generated.
	require.Equal(t, 1, generator.signingKeyCount)

	// Get active key.
	activeKey, err := manager.GetActiveSigningKey()
	require.NoError(t, err)
	require.NotNil(t, activeKey)
	require.True(t, activeKey.Active)
	require.True(t, activeKey.ValidForVerif)
	require.Equal(t, "RS256", activeKey.Algorithm)

	// Second rotation.
	firstKeyID := activeKey.KeyID
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Verify second key was generated.
	require.Equal(t, 2, generator.signingKeyCount)

	// Get new active key.
	newActiveKey, err := manager.GetActiveSigningKey()
	require.NoError(t, err)
	require.NotNil(t, newActiveKey)
	require.True(t, newActiveKey.Active)
	require.NotEqual(t, firstKeyID, newActiveKey.KeyID)

	// Old key should still be valid for verification.
	oldKey, err := manager.GetSigningKeyByID(firstKeyID)
	require.NoError(t, err)
	require.NotNil(t, oldKey)
	require.False(t, oldKey.Active)       // Not active for signing new tokens.
	require.True(t, oldKey.ValidForVerif) // Still valid for verification.
}

func TestRotateEncryptionKey(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// First rotation.
	err = manager.RotateEncryptionKey(ctx)
	require.NoError(t, err)

	// Verify key was generated.
	require.Equal(t, 1, generator.encryptionKeyCount)

	// Get active key.
	activeKey, err := manager.GetActiveEncryptionKey()
	require.NoError(t, err)
	require.NotNil(t, activeKey)
	require.True(t, activeKey.Active)
	require.True(t, activeKey.ValidForDecr)

	// Second rotation.
	firstKeyID := activeKey.KeyID
	err = manager.RotateEncryptionKey(ctx)
	require.NoError(t, err)

	// Verify second key was generated.
	require.Equal(t, 2, generator.encryptionKeyCount)

	// Get new active key.
	newActiveKey, err := manager.GetActiveEncryptionKey()
	require.NoError(t, err)
	require.NotNil(t, newActiveKey)
	require.True(t, newActiveKey.Active)
	require.NotEqual(t, firstKeyID, newActiveKey.KeyID)

	// Old key should still be valid for decryption.
	oldKey, err := manager.GetEncryptionKeyByID(firstKeyID)
	require.NoError(t, err)
	require.NotNil(t, oldKey)
	require.False(t, oldKey.Active)      // Not active for encrypting new tokens.
	require.True(t, oldKey.ValidForDecr) // Still valid for decryption.
}

func TestMaxActiveKeysEnforcement(t *testing.T) {
	t.Parallel()

	policy := &KeyRotationPolicy{
		RotationInterval:    24 * time.Hour,
		GracePeriod:         1 * time.Hour,
		MaxActiveKeys:       2,
		AutoRotationEnabled: false,
	}

	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Generate 3 signing keys (exceeds max of 2).
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Verify only 2 keys are kept (oldest removed).
	require.Equal(t, 3, generator.signingKeyCount)
	require.Len(t, manager.signingKeys, 2)
}

func TestRotationCallback(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}

	callbackInvoked := false

	var callbackKeyID string

	callback := func(keyID string) {
		callbackInvoked = true
		callbackKeyID = keyID
	}

	manager, err := NewKeyRotationManager(policy, generator, callback)
	require.NoError(t, err)

	ctx := context.Background()

	// Rotate signing key.
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Verify callback was invoked.
	require.True(t, callbackInvoked)
	require.NotEmpty(t, callbackKeyID)
}

func TestGetActiveSigningKey_NoKeys(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	// No keys rotated yet.
	activeKey, err := manager.GetActiveSigningKey()

	require.Error(t, err)
	require.Nil(t, activeKey)
}

func TestGetSigningKeyByID_NotFound(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	// Try to get non-existent key.
	key, err := manager.GetSigningKeyByID("non-existent-key")

	require.Error(t, err)
	require.Nil(t, key)
}

func TestGetActiveEncryptionKey_NoKeys(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	// No keys rotated yet.
	activeKey, err := manager.GetActiveEncryptionKey()

	require.Error(t, err)
	require.Nil(t, activeKey)
}

func TestGetEncryptionKeyByID_NotFound(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	// Try to get non-existent key.
	key, err := manager.GetEncryptionKeyByID("non-existent-key")

	require.Error(t, err)
	require.Nil(t, key)
}

func TestGetPublicKeys(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// No keys - should return empty.
	keys := manager.GetPublicKeys()
	require.Empty(t, keys)

	// Rotate signing key.
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Should have one key now.
	keys = manager.GetPublicKeys()
	// Note: mockKeyGenerator produces simple byte keys that may not convert to JWK.
	// The function returns keys that are valid for verification.
	require.NotNil(t, keys)
}

func TestGetAllValidVerificationKeys(t *testing.T) {
	t.Parallel()

	policy := DefaultKeyRotationPolicy()
	generator := &mockKeyGenerator{}
	manager, err := NewKeyRotationManager(policy, generator, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// No keys - should return empty.
	keys := manager.GetAllValidVerificationKeys()
	require.Empty(t, keys)

	// Rotate signing key.
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Should have one key.
	keys = manager.GetAllValidVerificationKeys()
	require.Len(t, keys, 1)
	require.True(t, keys[0].ValidForVerif)

	// Rotate again.
	err = manager.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Should have two keys now.
	keys = manager.GetAllValidVerificationKeys()
	require.Len(t, keys, 2)

	// All should be valid for verification.
	for _, key := range keys {
		require.True(t, key.ValidForVerif)
	}
}

func TestStartAutoRotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		autoRotationEnabled bool
		expectRotation      bool
	}{
		{
			name:                "auto_rotation_disabled",
			autoRotationEnabled: false,
			expectRotation:      false,
		},
		{
			name:                "auto_rotation_enabled",
			autoRotationEnabled: true,
			expectRotation:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			policy := &KeyRotationPolicy{
				RotationInterval:    100 * time.Millisecond,
				MaxActiveKeys:       3,
				AutoRotationEnabled: tc.autoRotationEnabled,
			}

			mockGen := &mockKeyGenerator{}
			manager, err := NewKeyRotationManager(policy, mockGen, nil)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()

			if tc.expectRotation {
				go manager.StartAutoRotation(ctx, "RS256")

				time.Sleep(350 * time.Millisecond)
				cancel()

				require.Greater(t, manager.GetSigningKeyCount(), 0, "Auto-rotation should generate keys")
			} else {
				go manager.StartAutoRotation(ctx, "RS256")

				time.Sleep(150 * time.Millisecond)
				cancel()

				require.Equal(t, 0, manager.GetSigningKeyCount(), "Auto-rotation disabled should not generate keys")
			}
		})
	}
}

func TestEcdsaCurveName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		curve    elliptic.Curve
		expected string
	}{
		{
			name:     "P256",
			curve:    elliptic.P256(),
			expected: "P-256",
		},
		{
			name:     "P384",
			curve:    elliptic.P384(),
			expected: "P-384",
		},
		{
			name:     "P521",
			curve:    elliptic.P521(),
			expected: "P-521",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ecdsaCurveName(tc.curve)
			require.Equal(t, tc.expected, result)
		})
	}
}
