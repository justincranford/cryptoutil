// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGetTLSPEMBytes_NilValue tests getTLSPEMBytes with nil value.
func TestGetTLSPEMBytes_NilValue(t *testing.T) {
	result := getTLSPEMBytes("non-existent-key")
	require.Nil(t, result, "Expected nil for non-existent key")
}

// TestGetTLSPEMBytes_NonBytesValue tests getTLSPEMBytes with non-[]byte value.
func TestGetTLSPEMBytes_NonBytesValue(t *testing.T) {
	resetFlags()

	// Viper will convert string to string, not []byte for most keys
	// This tests the type assertion failure path
	result := getTLSPEMBytes("log-level") // string value
	require.Nil(t, result, "Expected nil for non-[]byte value")
}

// TestNewForJOSEServer_DevMode tests NewForJOSEServer with dev mode enabled.
func TestNewForJOSEServer_DevMode(t *testing.T) {
	resetFlags()

	settings := NewForJOSEServer("127.0.0.1", 9443, true)
	require.NotNil(t, settings)
	require.Equal(t, "127.0.0.1", settings.BindPublicAddress)
	require.Equal(t, uint16(9443), settings.BindPublicPort)
	require.Equal(t, "jose-server", settings.OTLPService)
	require.True(t, settings.DevMode)
}

// TestNewForJOSEServer_ProductionMode tests NewForJOSEServer with dev mode disabled.
func TestNewForJOSEServer_ProductionMode(t *testing.T) {
	resetFlags()

	settings := NewForJOSEServer("127.0.0.1", 8443, false)
	require.NotNil(t, settings)
	require.Equal(t, "127.0.0.1", settings.BindPublicAddress)
	require.Equal(t, uint16(8443), settings.BindPublicPort)
	require.Equal(t, "jose-server", settings.OTLPService)
	require.False(t, settings.DevMode)
}

// TestNewForCAServer_DevMode tests NewForCAServer with dev mode enabled.
func TestNewForCAServer_DevMode(t *testing.T) {
	resetFlags()

	settings := NewForCAServer("127.0.0.1", 8380, true)
	require.NotNil(t, settings)
	require.Equal(t, "127.0.0.1", settings.BindPublicAddress)
	require.Equal(t, uint16(8380), settings.BindPublicPort)
	require.Equal(t, "ca-server", settings.OTLPService)
	require.True(t, settings.DevMode)
}

// TestNewForCAServer_ProductionMode tests NewForCAServer with dev mode disabled.
func TestNewForCAServer_ProductionMode(t *testing.T) {
	resetFlags()

	settings := NewForCAServer("127.0.0.1", 9380, false)
	require.NotNil(t, settings)
	require.Equal(t, "127.0.0.1", settings.BindPublicAddress)
	require.Equal(t, uint16(9380), settings.BindPublicPort)
	require.Equal(t, "ca-server", settings.OTLPService)
	require.False(t, settings.DevMode)
}

// TestRegisterAsBoolSetting tests registerAsBoolSetting function.
func TestRegisterAsBoolSetting(t *testing.T) {
	setting := Setting{value: true}
	result := registerAsBoolSetting(&setting)
	require.True(t, result)
}

// TestRegisterAsStringSetting tests registerAsStringSetting function.
func TestRegisterAsStringSetting(t *testing.T) {
	setting := Setting{value: "test-value"}
	result := registerAsStringSetting(&setting)
	require.Equal(t, "test-value", result)
}

// TestRegisterAsUint16Setting tests registerAsUint16Setting function.
func TestRegisterAsUint16Setting(t *testing.T) {
	setting := Setting{value: uint16(8080)}
	result := registerAsUint16Setting(&setting)
	require.Equal(t, uint16(8080), result)
}

// TestRegisterAsStringSliceSetting tests registerAsStringSliceSetting function.
func TestRegisterAsStringSliceSetting(t *testing.T) {
	setting := Setting{value: []string{"item1", "item2"}}
	result := registerAsStringSliceSetting(&setting)
	require.Equal(t, []string{"item1", "item2"}, result)
}

// TestRegisterAsStringArraySetting tests registerAsStringArraySetting function.
func TestRegisterAsStringArraySetting(t *testing.T) {
	setting := Setting{value: []string{"item1", "item2"}}
	result := registerAsStringArraySetting(&setting)
	require.Equal(t, []string{"item1", "item2"}, result)
}

// TestRegisterAsDurationSetting tests registerAsDurationSetting function.
func TestRegisterAsDurationSetting(t *testing.T) {
	// Use time.Duration type, not string
	setting := Setting{value: time.Duration(5 * time.Minute)}
	result := registerAsDurationSetting(&setting)
	require.Equal(t, time.Duration(5*time.Minute), result)
}

// TestRegisterAsIntSetting tests registerAsIntSetting function.
func TestRegisterAsIntSetting(t *testing.T) {
	setting := Setting{value: 100}
	result := registerAsIntSetting(&setting)
	require.Equal(t, 100, result)
}
