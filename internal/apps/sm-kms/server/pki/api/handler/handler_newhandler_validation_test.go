// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestNewHandler_NilStorage tests NewHandler with nil storage.
func TestNewHandler_NilStorage(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	handler, err := NewHandler(testSetup.Issuer, nil, nil)
	require.Error(t, err)
	require.Nil(t, handler)
	require.Contains(t, err.Error(), "storage is required")
}

// TestNewHandler_NilProfiles tests NewHandler with nil profiles map creates empty map.
func TestNewHandler_NilProfiles(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	storage := cryptoutilCAStorage.NewMemoryStore()

	handler, err := NewHandler(testSetup.Issuer, storage, nil)
	require.NoError(t, err)
	require.NotNil(t, handler)
	require.NotNil(t, handler.profiles)
	require.Len(t, handler.profiles, 0)
}

// TestNewHandler_WithProfiles tests NewHandler with provided profiles.
func TestNewHandler_WithProfiles(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	storage := cryptoutilCAStorage.NewMemoryStore()

	profiles := map[string]*ProfileConfig{
		cryptoutilSharedMagic.PKIProfileIDTLSServer: {
			ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
			Name:        testProfileNameTLSServer,
			Description: "TLS server profile",
			Category:    cryptoutilSharedMagic.CMD_SERVER,
		},
		cryptoutilSharedMagic.PKIProfileIDTLSClient: {
			ID:          cryptoutilSharedMagic.PKIProfileIDTLSClient,
			Name:        "TLS Client",
			Description: "TLS client profile",
			Category:    "client",
		},
	}

	handler, err := NewHandler(testSetup.Issuer, storage, profiles)
	require.NoError(t, err)
	require.NotNil(t, handler)
	require.NotNil(t, handler.profiles)
	require.Len(t, handler.profiles, 2)
	require.Contains(t, handler.profiles, cryptoutilSharedMagic.PKIProfileIDTLSServer)
	require.Contains(t, handler.profiles, cryptoutilSharedMagic.PKIProfileIDTLSClient)
	require.Equal(t, testProfileNameTLSServer, handler.profiles[cryptoutilSharedMagic.PKIProfileIDTLSServer].Name)
	require.Equal(t, "TLS Client", handler.profiles[cryptoutilSharedMagic.PKIProfileIDTLSClient].Name)
}
