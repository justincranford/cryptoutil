// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testEncryptedData = "encrypted-data"

// TestBarrierRootKey_GettersSetters tests all getter and setter methods.
func TestBarrierRootKey_GettersSetters(t *testing.T) {
	entity := &BarrierRootKey{}

	// Test UUID.
	testUUID := googleUuid.New()
	entity.SetUUID(testUUID)
	require.Equal(t, testUUID, entity.GetUUID())

	// Test Encrypted.
	entity.SetEncrypted(testEncryptedData)
	require.Equal(t, testEncryptedData, entity.GetEncrypted())

	// Test KEKUUID.
	testKEKUUID := googleUuid.New()
	entity.SetKEKUUID(testKEKUUID)
	require.Equal(t, testKEKUUID, entity.GetKEKUUID())
}

// TestBarrierIntermediateKey_GettersSetters tests all getter and setter methods.
func TestBarrierIntermediateKey_GettersSetters(t *testing.T) {
	entity := &BarrierIntermediateKey{}

	// Test UUID.
	testUUID := googleUuid.New()
	entity.SetUUID(testUUID)
	require.Equal(t, testUUID, entity.GetUUID())

	// Test Encrypted.
	entity.SetEncrypted(testEncryptedData)
	require.Equal(t, testEncryptedData, entity.GetEncrypted())

	// Test KEKUUID.
	testKEKUUID := googleUuid.New()
	entity.SetKEKUUID(testKEKUUID)
	require.Equal(t, testKEKUUID, entity.GetKEKUUID())
}

// TestBarrierContentKey_GettersSetters tests all getter and setter methods.
func TestBarrierContentKey_GettersSetters(t *testing.T) {
	entity := &BarrierContentKey{}

	// Test UUID.
	testUUID := googleUuid.New()
	entity.SetUUID(testUUID)
	require.Equal(t, testUUID, entity.GetUUID())

	// Test Encrypted.
	entity.SetEncrypted(testEncryptedData)
	require.Equal(t, testEncryptedData, entity.GetEncrypted())

	// Test KEKUUID.
	testKEKUUID := googleUuid.New()
	entity.SetKEKUUID(testKEKUUID)
	require.Equal(t, testKEKUUID, entity.GetKEKUUID())
}

// TestOrmRepository_Shutdown tests the shutdown functionality.
func TestOrmRepository_Shutdown(t *testing.T) {
	// Note: Cannot easily test shutdown without breaking other tests
	// since testGORMRepository is shared. This would require creating
	// a separate instance just for this test.
	// The shutdown functionality is already tested in TestMain's defer.
	t.Skip("Shutdown is tested in TestMain defer; testing here would break shared testGORMRepository")
}
