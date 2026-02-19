//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
//

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testEncryptedData = "encrypted-data"

// TestRootKey_GettersSetters tests all getter and setter methods.
func TestRootKey_GettersSetters(t *testing.T) {
	t.Parallel()

	entity := &RootKey{}

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

// TestIntermediateKey_GettersSetters tests all getter and setter methods.
func TestIntermediateKey_GettersSetters(t *testing.T) {
	t.Parallel()

	entity := &IntermediateKey{}

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

// TestContentKey_GettersSetters tests all getter and setter methods.
func TestContentKey_GettersSetters(t *testing.T) {
	t.Parallel()

	entity := &ContentKey{}

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
	t.Parallel()
	// Note: Cannot easily test shutdown without breaking other tests
	// since testGORMRepository is shared. This would require creating
	// a separate instance just for this test.
	// The shutdown functionality is already tested in TestMain's defer.
	t.Skip("Shutdown is tested in TestMain defer; testing here would break shared testGORMRepository")
}
