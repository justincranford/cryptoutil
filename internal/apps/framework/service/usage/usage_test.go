// Copyright (c) 2025 Justin Cranford

package usage_test

import (
	"strings"
	"testing"

	cryptoutilUsage "cryptoutil/internal/apps/framework/service/usage"

	"github.com/stretchr/testify/require"
)

func TestBuildUsageMain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		productName        string
		serviceName        string
		serviceDisplayName string
		wantContains       []string
	}{
		{
			name:               "sm kms",
			productName:        "sm",
			serviceName:        "kms",
			serviceDisplayName: "Key Management Service",
			wantContains: []string{
				"Usage: sm kms <subcommand> [options]",
				"Key Management Service",
				"Use \"sm kms <subcommand> help\"",
			},
		},
		{
			name:               "jose ja",
			productName:        "jose",
			serviceName:        "ja",
			serviceDisplayName: "JWK Authority",
			wantContains: []string{
				"Usage: jose ja <subcommand> [options]",
				"JWK Authority",
				"Use \"jose ja <subcommand> help\"",
			},
		},
		{
			name:               "pki ca",
			productName:        "pki",
			serviceName:        "ca",
			serviceDisplayName: "Certificate Authority",
			wantContains: []string{
				"Usage: pki ca <subcommand> [options]",
				"Certificate Authority",
				"Use \"pki ca <subcommand> help\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageMain(tt.productName, tt.serviceName, tt.serviceDisplayName)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		productName        string
		serviceName        string
		serviceDisplayName string
		configFilePath     string
		wantContains       []string
	}{
		{
			name:               "sm kms server",
			productName:        "sm",
			serviceName:        "kms",
			serviceDisplayName: "Key Management Service",
			configFilePath:     "configs/sm-kms/sm-kms-framework.yml",
			wantContains: []string{
				"Usage: sm kms server [options]",
				"Key Management Service",
				"configs/sm-kms/sm-kms-framework.yml",
				"SQLite",
				"PostgreSQL",
			},
		},
		{
			name:               "jose ja server",
			productName:        "jose",
			serviceName:        "ja",
			serviceDisplayName: "JWK Authority",
			configFilePath:     "configs/jose-ja/jose-ja-framework.yml",
			wantContains: []string{
				"Usage: jose ja server [options]",
				"JWK Authority",
				"configs/jose-ja/jose-ja-framework.yml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageServer(tt.productName, tt.serviceName, tt.serviceDisplayName, tt.configFilePath)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		productName        string
		serviceName        string
		serviceDisplayName string
		wantContains       []string
	}{
		{
			name:               "sm kms client",
			productName:        "sm",
			serviceName:        "kms",
			serviceDisplayName: "Key Management Service",
			wantContains: []string{
				"Usage: sm kms client [options]",
				"Key Management Service",
				"sm kms client",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageClient(tt.productName, tt.serviceName, tt.serviceDisplayName)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		productName        string
		serviceName        string
		serviceDisplayName string
		configFilePath     string
		wantContains       []string
	}{
		{
			name:               "sm kms init",
			productName:        "sm",
			serviceName:        "kms",
			serviceDisplayName: "Key Management Service",
			configFilePath:     "configs/sm-kms/sm-kms-framework.yml",
			wantContains: []string{
				"Usage: sm kms init [options]",
				"Key Management Service",
				"configs/sm-kms/sm-kms-framework.yml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageInit(tt.productName, tt.serviceName, tt.serviceDisplayName, tt.configFilePath)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		productName       string
		serviceName       string
		defaultPublicPort string
		wantContains      []string
	}{
		{
			name:              "sm kms health port 8000",
			productName:       "sm",
			serviceName:       "kms",
			defaultPublicPort: "8000",
			wantContains: []string{
				"Usage: sm kms health [options]",
				"https://127.0.0.1:8000",
				"https://localhost:8000",
				"/browser/api/v1/health",
			},
		},
		{
			name:              "jose ja health port 8200",
			productName:       "jose",
			serviceName:       "ja",
			defaultPublicPort: "8200",
			wantContains: []string{
				"Usage: jose ja health [options]",
				"https://127.0.0.1:8200",
				"https://localhost:8200",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageHealth(tt.productName, tt.serviceName, tt.defaultPublicPort)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageLivez(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		productName  string
		serviceName  string
		wantContains []string
	}{
		{
			name:        "sm kms livez",
			productName: "sm",
			serviceName: "kms",
			wantContains: []string{
				"Usage: sm kms livez [options]",
				"https://127.0.0.1:9090",
				"/admin/api/v1/livez",
			},
		},
		{
			name:        "pki ca livez",
			productName: "pki",
			serviceName: "ca",
			wantContains: []string{
				"Usage: pki ca livez [options]",
				"https://127.0.0.1:9090",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageLivez(tt.productName, tt.serviceName)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageReadyz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		productName  string
		serviceName  string
		wantContains []string
	}{
		{
			name:        "sm kms readyz",
			productName: "sm",
			serviceName: "kms",
			wantContains: []string{
				"Usage: sm kms readyz [options]",
				"https://127.0.0.1:9090",
				"/admin/api/v1/readyz",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageReadyz(tt.productName, tt.serviceName)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}

func TestBuildUsageShutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		productName  string
		serviceName  string
		wantContains []string
	}{
		{
			name:        "sm kms shutdown",
			productName: "sm",
			serviceName: "kms",
			wantContains: []string{
				"Usage: sm kms shutdown [options]",
				"https://127.0.0.1:9090",
				"/admin/api/v1/shutdown",
				"--force",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cryptoutilUsage.BuildUsageShutdown(tt.productName, tt.serviceName)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(result, want), "expected %q to contain %q", result, want)
			}
		})
	}
}
