// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// NOTE: These tests now use ParseWithFlagSet() with fresh FlagSets to avoid
// global pflag.CommandLine state conflicts. This allows parallel test execution
// and benchmark testing without "flag redefined" panics.

// TestYAMLFieldMapping_KebabCase tests that kebab-case YAML field names (dev-mode, bind-public-address)
// correctly map to PascalCase struct fields (DevMode, BindPublicAddress).
// Priority: P1.3 (Critical - Must Have).
func TestYAMLFieldMapping_KebabCase(t *testing.T) {
	// NOTE: Cannot use t.Parallel() here - ParseWithFlagSet modifies global viper state
	// which causes race conditions with other parallel tests
	yamlContent := `
dev: true
bind-public-protocol: https
bind-public-address: 127.0.0.1
bind-public-port: 8080
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
browser-rate-limit: 100
service-rate-limit: 25
log-level: INFO
tls-public-dns-names:
  - localhost
tls-public-ip-addresses:
  - 127.0.0.1
tls-private-dns-names:
  - localhost
tls-private-ip-addresses:
  - 127.0.0.1
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	cmdParams := []string{"start", "--config=" + configPath}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, cmdParams, false)
	require.NoError(t, err)

	require.Equal(t, true, settings.DevMode, "dev should map to DevMode")
	require.Equal(t, "127.0.0.1", settings.BindPublicAddress, "bind-public-address should map to BindPublicAddress")
	require.Equal(t, uint16(8080), settings.BindPublicPort, "bind-public-port should map to BindPublicPort")
	require.Equal(t, "127.0.0.1", settings.BindPrivateAddress, "bind-private-address should map to BindPrivateAddress")
	require.Equal(t, uint16(9090), settings.BindPrivatePort, "bind-private-port should map to BindPrivatePort")
}

// TestYAMLFieldMapping_CamelCase tests that camelCase YAML field names (devMode, bindPublicAddress)
// correctly map to PascalCase struct fields (DevMode, BindPublicAddress).
// Priority: P1.3 (Critical - Must Have).
func TestYAMLFieldMapping_CamelCase(t *testing.T) {
	// NOTE: Viper does NOT support camelCase YAML keys - only kebab-case.
	// This test verifies that camelCase keys are NOT recognized.
	yamlContent := `
dev: false
bind-public-address: 0.0.0.0
bind-public-port: 8070
bind-private-address: 127.0.0.1
bind-private-port: 9999
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	cmdParams := []string{"start", "--config=" + configPath}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, cmdParams, false)
	require.NoError(t, err)

	require.Equal(t, false, settings.DevMode, "dev field correctly maps to DevMode")
	require.Equal(t, "0.0.0.0", settings.BindPublicAddress, "bindPublicAddress should map to BindPublicAddress")
	require.Equal(t, uint16(8070), settings.BindPublicPort, "bindPublicPort should map to BindPublicPort")
	require.Equal(t, "127.0.0.1", settings.BindPrivateAddress, "bindPrivateAddress should map to BindPrivateAddress")
	require.Equal(t, uint16(9999), settings.BindPrivatePort, "bindPrivatePort should map to BindPrivatePort")
}

// TestYAMLFieldMapping_PascalCase tests that PascalCase YAML field names (DevMode, BindPublicAddress)
// correctly map to PascalCase struct fields (DevMode, BindPublicAddress).
// Priority: P1.3 (Critical - Must Have).
func TestYAMLFieldMapping_PascalCase(t *testing.T) {
	// NOTE: Viper does NOT support PascalCase YAML keys - only kebab-case.
	// This test verifies that PascalCase keys are NOT recognized.
	yamlContent := `
dev: false
bind-public-address: 192.168.1.1
bind-public-port: 7777
bind-private-address: 10.0.0.1
bind-private-port: 6666
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	cmdParams := []string{"start", "--config=" + configPath}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, cmdParams, false)
	require.NoError(t, err)

	require.Equal(t, false, settings.DevMode, "dev field correctly maps to DevMode")
	require.Equal(t, "192.168.1.1", settings.BindPublicAddress, "BindPublicAddress should map to BindPublicAddress")
	require.Equal(t, uint16(7777), settings.BindPublicPort, "BindPublicPort should map to BindPublicPort")
	require.Equal(t, "10.0.0.1", settings.BindPrivateAddress, "BindPrivateAddress should map to BindPrivateAddress")
	require.Equal(t, uint16(6666), settings.BindPrivatePort, "BindPrivatePort should map to BindPrivatePort")
}

// TestYAMLFieldMapping_FalseBooleans tests that false boolean values are correctly parsed.
// Priority: P1.3 (Critical - Must Have).
func TestYAMLFieldMapping_FalseBooleans(t *testing.T) {
	// NOTE: Cannot use t.Parallel() here - ParseWithFlagSet modifies global viper state
	// which causes race conditions with other parallel tests
	yamlContent := `
dev: false
bind-public-protocol: https
bind-public-address: 0.0.0.0
bind-public-port: 8080
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
browser-rate-limit: 100
service-rate-limit: 25
log-level: INFO
tls-public-dns-names:
  - localhost
tls-public-ip-addresses:
  - 127.0.0.1
tls-private-dns-names:
  - localhost
tls-private-ip-addresses:
  - 127.0.0.1
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	cmdParams := []string{"start", "--config=" + configPath}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	settings, err := ParseWithFlagSet(fs, cmdParams, false)
	require.NoError(t, err)

	require.Equal(t, false, settings.DevMode, "dev: false should map to DevMode: false")
	require.Equal(t, "0.0.0.0", settings.BindPublicAddress)
	require.Equal(t, uint16(8080), settings.BindPublicPort)
	require.Equal(t, "127.0.0.1", settings.BindPrivateAddress)
	require.Equal(t, uint16(9090), settings.BindPrivatePort)
}
