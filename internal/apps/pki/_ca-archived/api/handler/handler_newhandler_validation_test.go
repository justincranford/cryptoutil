// Copyright (c) 2025 Justin Cranford

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
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
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS server profile",
			Category:    "server",
		},
		"tls-client": {
			ID:          "tls-client",
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
	require.Contains(t, handler.profiles, "tls-server")
	require.Contains(t, handler.profiles, "tls-client")
	require.Equal(t, "TLS Server", handler.profiles["tls-server"].Name)
	require.Equal(t, "TLS Client", handler.profiles["tls-client"].Name)
}
