// Copyright (c) 2025 Justin Cranford
//
//

package hsm

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrNotImplemented(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		expected  string
	}{
		{
			name:      "sign operation",
			operation: "Sign",
			expected:  "HSM operation not implemented: Sign",
		},
		{
			name:      "decrypt operation",
			operation: "Decrypt",
			expected:  "HSM operation not implemented: Decrypt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &ErrNotImplemented{Operation: tc.operation}
			require.Equal(t, tc.expected, err.Error())
		})
	}
}

func TestErrHSMNotAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		reason   string
		expected string
	}{
		{
			name:     "device disconnected",
			reason:   "device disconnected",
			expected: "HSM not available: device disconnected",
		},
		{
			name:     "library not found",
			reason:   "PKCS#11 library not found",
			expected: "HSM not available: PKCS#11 library not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &ErrHSMNotAvailable{Reason: tc.reason}
			require.Equal(t, tc.expected, err.Error())
		})
	}
}

func TestErrKeyNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "by label",
			identifier: "my-signing-key",
			expected:   "HSM key not found: my-signing-key",
		},
		{
			name:       "by id",
			identifier: "01:02:03:04",
			expected:   "HSM key not found: 01:02:03:04",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &ErrKeyNotFound{Identifier: tc.identifier}
			require.Equal(t, tc.expected, err.Error())
		})
	}
}

func TestProviderConfigFields(t *testing.T) {
	t.Parallel()

	config := &ProviderConfig{
		Type:        "pkcs11",
		LibraryPath: "/usr/lib/softhsm/libsofthsm2.so",
		PIN:         "1234",
		SlotID:      0,
		TokenLabel:  "test-token",
	}

	require.Equal(t, "pkcs11", config.Type)
	require.Equal(t, "/usr/lib/softhsm/libsofthsm2.so", config.LibraryPath)
	require.Equal(t, "1234", config.PIN)
	require.Equal(t, uint(0), config.SlotID)
	require.Equal(t, "test-token", config.TokenLabel)
}

func TestKeyInfoFields(t *testing.T) {
	t.Parallel()

	info := KeyInfo{
		ID:         "key-123",
		Label:      "My RSA Key",
		Type:       cryptoutilSharedMagic.KeyTypeRSA,
		Size:       cryptoutilSharedMagic.DefaultMetricsBatchSize,
		CanSign:    true,
		CanEncrypt: true,
		CanDecrypt: true,
		CanWrap:    false,
		CanUnwrap:  false,
	}

	require.Equal(t, "key-123", info.ID)
	require.Equal(t, "My RSA Key", info.Label)
	require.Equal(t, cryptoutilSharedMagic.KeyTypeRSA, info.Type)
	require.Equal(t, cryptoutilSharedMagic.DefaultMetricsBatchSize, info.Size)
	require.True(t, info.CanSign)
	require.True(t, info.CanEncrypt)
	require.True(t, info.CanDecrypt)
	require.False(t, info.CanWrap)
	require.False(t, info.CanUnwrap)
}

func TestKeySpecFields(t *testing.T) {
	t.Parallel()

	spec := &KeySpec{
		Label:       "my-ec-key",
		Type:        "EC",
		Size:        cryptoutilSharedMagic.MaxUnsealSharedSecrets,
		Curve:       "P-256",
		Extractable: false,
		Usage: KeyUsage{
			Sign:    true,
			Verify:  true,
			Encrypt: false,
			Decrypt: false,
			Wrap:    false,
			Unwrap:  false,
		},
	}

	require.Equal(t, "my-ec-key", spec.Label)
	require.Equal(t, "EC", spec.Type)
	require.Equal(t, cryptoutilSharedMagic.MaxUnsealSharedSecrets, spec.Size)
	require.Equal(t, "P-256", spec.Curve)
	require.False(t, spec.Extractable)
	require.True(t, spec.Usage.Sign)
	require.True(t, spec.Usage.Verify)
	require.False(t, spec.Usage.Encrypt)
}
