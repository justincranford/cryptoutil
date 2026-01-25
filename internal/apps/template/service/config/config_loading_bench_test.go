// Copyright (c) 2025 Justin Cranford
//
//

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
)

// BenchmarkYAMLFileLoading benchmarks the performance of loading and parsing YAML configuration files.
// Measures the time to read and parse a YAML file into ServiceTemplateServerSettings.
func BenchmarkYAMLFileLoading(b *testing.B) {
	// Create temp directory for test YAML files.
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yml")

	// Create realistic YAML config file.
	configContent := `log-level: INFO
dev-mode: false
bind-public-protocol: https
bind-public-address: 192.168.1.100
bind-public-port: 8443
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
database-url: "postgres://user:pass@localhost:5432/testdb"
browser-ip-rate-limit: 100
service-ip-rate-limit: 100
otlp-endpoint: "http://otel-collector:4317"
tls-public-dns-names: ["api.example.com"]
tls-private-dns-names: ["localhost"]
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(b, err)

	b.ResetTimer() // Reset timer after setup.

	for i := 0; i < b.N; i++ {
		// Benchmark ParseWithFlagSet with fresh FlagSet per iteration (prevents "flag redefined" panic).
		fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)

		_, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(
			fs,
			[]string{"start", "--config", configPath},
			false, // Don't exit on help
		)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkConfigValidation benchmarks the performance of validateConfiguration.
// Tests validation overhead on a realistic production configuration.
func BenchmarkConfigValidation(b *testing.B) {
	// Create realistic production configuration.
	// Note: We can't directly call validateConfiguration (it's not exported),
	// so we benchmark Parse which includes validation.
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "validation-bench.yml")

	configContent := `log-level: INFO
dev-mode: false
bind-public-protocol: https
bind-public-address: 192.168.1.100
bind-public-port: 8443
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
database-url: "postgres://user:pass@localhost:5432/testdb"
browser-ip-rate-limit: 100
service-ip-rate-limit: 100
otlp-endpoint: "http://otel-collector:4317"
tls-public-dns-names: ["api.example.com", "www.example.com"]
tls-private-dns-names: ["localhost"]
tls-public-ip-addresses: ["192.168.1.100"]
tls-private-ip-addresses: ["127.0.0.1"]
cors-allowed-origins: ["http://localhost:3000", "https://app.example.com"]
allowed-ips: ["192.168.1.0/24", "10.0.0.0/8"]
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// ParseWithFlagSet includes validation, so this benchmarks both parsing + validation.
		// Use fresh FlagSet per iteration to prevent "flag redefined" panic.
		fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)

		_, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(
			fs,
			[]string{"start", "--config", configPath},
			false,
		)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkConfigMerging benchmarks the performance of merging configuration from multiple sources.
// Tests config file + CLI parameters + environment variables merging.
func BenchmarkConfigMerging(b *testing.B) {
	// Create temp directory and config file.
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "base-config.yml")

	// Base config from file.
	configContent := `log-level: INFO
dev-mode: false
bind-public-protocol: https
bind-public-address: 192.168.1.100
bind-public-port: 8443
database-url: "postgres://user:pass@localhost:5432/testdb"
browser-ip-rate-limit: 100
service-ip-rate-limit: 100
otlp-endpoint: "http://otel-collector:4317"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(b, err)

	// Set environment variables to merge.
	require.NoError(b, os.Setenv("CRYPTOUTIL_BIND_PRIVATE_PORT", "9191"))
	require.NoError(b, os.Setenv("CRYPTOUTIL_LOG_LEVEL", "DEBUG"))

	defer func() {
		_ = os.Unsetenv("CRYPTOUTIL_BIND_PRIVATE_PORT")
		_ = os.Unsetenv("CRYPTOUTIL_LOG_LEVEL")
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Benchmark ParseWithFlagSet with config file + CLI overrides + env vars.
		// This tests the full merging logic: file → env vars → CLI flags.
		// Use fresh FlagSet per iteration to prevent "flag redefined" panic.
		fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)

		_, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(
			fs,
			[]string{
				"start",
				"--config", configPath,
				"--bind-private-address", "127.0.0.1", // CLI override.
			},
			false,
		)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}
