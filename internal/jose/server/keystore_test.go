// Copyright (c) 2025 Justin Cranford

package server

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestKeyStoreStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     *StoredKey
		wantErr bool
	}{
		{
			name: "valid key",
			key: &StoredKey{
				KID:       googleUuid.New(),
				Algorithm: "EC/P256",
				Use:       "sig",
			},
			wantErr: false,
		},
		{
			name:    "nil key",
			key:     nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ks := NewKeyStore()
			err := ks.Store(tc.key)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyStoreDuplicateKey(t *testing.T) {
	t.Parallel()

	ks := NewKeyStore()
	kid := googleUuid.New()

	key := &StoredKey{
		KID:       kid,
		Algorithm: "EC/P256",
		Use:       "sig",
	}

	// First store should succeed.
	err := ks.Store(key)
	require.NoError(t, err)

	// Second store with same KID should fail.
	err = ks.Store(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestKeyStoreGet(t *testing.T) {
	t.Parallel()

	ks := NewKeyStore()
	kid := googleUuid.New()

	key := &StoredKey{
		KID:       kid,
		Algorithm: "EC/P256",
		Use:       "sig",
	}

	_ = ks.Store(key)

	tests := []struct {
		name       string
		searchKID  string
		wantExists bool
	}{
		{
			name:       "existing key",
			searchKID:  kid.String(),
			wantExists: true,
		},
		{
			name:       "non-existing key",
			searchKID:  googleUuid.New().String(),
			wantExists: false,
		},
		{
			name:       "invalid uuid",
			searchKID:  "invalid-uuid",
			wantExists: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotKey, exists := ks.Get(tc.searchKID)

			require.Equal(t, tc.wantExists, exists)

			if tc.wantExists {
				require.NotNil(t, gotKey)
				require.Equal(t, kid, gotKey.KID)
			}
		})
	}
}

func TestKeyStoreDelete(t *testing.T) {
	t.Parallel()

	ks := NewKeyStore()
	kid := googleUuid.New()

	key := &StoredKey{
		KID:       kid,
		Algorithm: "EC/P256",
		Use:       "sig",
	}

	_ = ks.Store(key)

	// Delete existing key.
	deleted := ks.Delete(kid.String())
	require.True(t, deleted)

	// Verify key is gone.
	_, exists := ks.Get(kid.String())
	require.False(t, exists)

	// Delete non-existing key.
	deleted = ks.Delete(kid.String())
	require.False(t, deleted)
}

func TestKeyStoreList(t *testing.T) {
	t.Parallel()

	ks := NewKeyStore()

	// Empty store.
	keys := ks.List()
	require.Empty(t, keys)

	// Add keys.
	for i := 0; i < 3; i++ {
		key := &StoredKey{
			KID:       googleUuid.New(),
			Algorithm: "EC/P256",
			Use:       "sig",
		}
		_ = ks.Store(key)
	}

	// List should have 3 keys.
	keys = ks.List()
	require.Len(t, keys, 3)
}

func TestKeyStoreCount(t *testing.T) {
	t.Parallel()

	ks := NewKeyStore()
	require.Equal(t, 0, ks.Count())

	for i := 0; i < 5; i++ {
		key := &StoredKey{
			KID:       googleUuid.New(),
			Algorithm: "EC/P256",
			Use:       "sig",
		}
		_ = ks.Store(key)
	}

	require.Equal(t, 5, ks.Count())
}

func TestKeyStoreGetJWKS(t *testing.T) {
	t.Parallel()

	ks := NewKeyStore()

	// Empty JWKS.
	jwks := ks.GetJWKS()
	require.Equal(t, 0, jwks.Len())

	// JWKS with nil public keys (symmetric keys).
	key := &StoredKey{
		KID:       googleUuid.New(),
		Algorithm: "oct/256",
		Use:       "enc",
		PublicJWK: nil, // Symmetric keys don't have public component.
	}
	_ = ks.Store(key)

	// JWKS should still be empty since PublicJWK is nil.
	jwks = ks.GetJWKS()
	require.Equal(t, 0, jwks.Len())
}
