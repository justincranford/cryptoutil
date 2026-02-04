// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGetTLSPEMBytes tests getTLSPEMBytes with various inputs.
// NOTE: Cannot use t.Parallel() - this test accesses global viper state.
func TestGetTLSPEMBytes(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		key     string
		wantNil bool
	}{
		{
			name:    "nil value for non-existent key",
			setup:   func() {},
			key:     "non-existent-key",
			wantNil: true,
		},
		{
			name:    "nil for non-bytes value",
			setup:   resetFlags,
			key:     "log-level",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			result := getTLSPEMBytes(tc.key)
			if tc.wantNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}

// TestNewForServer tests NewForJOSEServer and NewForCAServer factory functions.
// NOTE: These tests cannot run in parallel due to global flag state.
func TestNewForServer(t *testing.T) {
	tests := []struct {
		name        string
		factory     func(address string, port uint16, devMode bool) *ServiceTemplateServerSettings
		address     string
		port        uint16
		devMode     bool
		wantService string
	}{
		{name: "JOSE dev mode", factory: NewForJOSEServer, address: "127.0.0.1", port: 8060, devMode: true, wantService: "jose-server"},
		{name: "JOSE production", factory: NewForJOSEServer, address: "127.0.0.1", port: 8061, devMode: false, wantService: "jose-server"},
		{name: "CA dev mode", factory: NewForCAServer, address: "127.0.0.1", port: 8050, devMode: true, wantService: "ca-server"},
		{name: "CA production", factory: NewForCAServer, address: "127.0.0.1", port: 8051, devMode: false, wantService: "ca-server"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetFlags()

			settings := tc.factory(tc.address, tc.port, tc.devMode)
			require.NotNil(t, settings)
			require.Equal(t, tc.address, settings.BindPublicAddress)
			require.Equal(t, tc.port, settings.BindPublicPort)
			require.Equal(t, tc.wantService, settings.OTLPService)
			require.Equal(t, tc.devMode, settings.DevMode)
		})
	}
}

// TestRegisterAsSettings tests all RegisterAs* setting functions.
func TestRegisterAsSettings(t *testing.T) {
	t.Parallel()

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: true}
		require.True(t, RegisterAsBoolSetting(&setting))
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: "test-value"}
		require.Equal(t, "test-value", RegisterAsStringSetting(&setting))
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: uint16(8080)}
		require.Equal(t, uint16(8080), RegisterAsUint16Setting(&setting))
	})

	t.Run("string slice", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: []string{"item1", "item2"}}
		require.Equal(t, []string{"item1", "item2"}, RegisterAsStringSliceSetting(&setting))
	})

	t.Run("string array", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: []string{"item1", "item2"}}
		require.Equal(t, []string{"item1", "item2"}, RegisterAsStringArraySetting(&setting))
	})

	t.Run("duration", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: 5 * time.Minute}
		require.Equal(t, 5*time.Minute, RegisterAsDurationSetting(&setting))
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		setting := Setting{Value: 100}
		require.Equal(t, 100, RegisterAsIntSetting(&setting))
	})
}
