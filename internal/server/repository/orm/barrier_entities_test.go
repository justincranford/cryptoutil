package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestBarrierRootKeyGettersSetters tests BarrierRootKey getter and setter methods.
func TestBarrierRootKeyGettersSetters(t *testing.T) {
	t.Parallel()

	key := &BarrierRootKey{}

	// Test UUID.
	uuid := googleUuid.New()
	key.SetUUID(uuid)
	require.Equal(t, uuid, key.GetUUID())

	// Test Encrypted.
	encrypted := "encrypted-root-key-data"
	key.SetEncrypted(encrypted)
	require.Equal(t, encrypted, key.GetEncrypted())

	// Test KEK UUID.
	kekUUID := googleUuid.New()
	key.SetKEKUUID(kekUUID)
	require.Equal(t, kekUUID, key.GetKEKUUID())
}

// TestBarrierIntermediateKeyGettersSetters tests BarrierIntermediateKey getter and setter methods.
func TestBarrierIntermediateKeyGettersSetters(t *testing.T) {
	t.Parallel()

	key := &BarrierIntermediateKey{}

	// Test UUID.
	uuid := googleUuid.New()
	key.SetUUID(uuid)
	require.Equal(t, uuid, key.GetUUID())

	// Test Encrypted.
	encrypted := "encrypted-intermediate-key-data"
	key.SetEncrypted(encrypted)
	require.Equal(t, encrypted, key.GetEncrypted())

	// Test KEK UUID.
	kekUUID := googleUuid.New()
	key.SetKEKUUID(kekUUID)
	require.Equal(t, kekUUID, key.GetKEKUUID())
}

// TestBarrierContentKeyGettersSetters tests BarrierContentKey getter and setter methods.
func TestBarrierContentKeyGettersSetters(t *testing.T) {
	t.Parallel()

	key := &BarrierContentKey{}

	// Test UUID.
	uuid := googleUuid.New()
	key.SetUUID(uuid)
	require.Equal(t, uuid, key.GetUUID())

	// Test Encrypted.
	encrypted := "encrypted-content-key-data"
	key.SetEncrypted(encrypted)
	require.Equal(t, encrypted, key.GetEncrypted())

	// Test KEK UUID.
	kekUUID := googleUuid.New()
	key.SetKEKUUID(kekUUID)
	require.Equal(t, kekUUID, key.GetKEKUUID())
}
