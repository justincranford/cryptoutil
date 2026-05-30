// Copyright (c) 2025-2026 Justin Cranford.
package usage_test

import (
	"fmt"
	"strings"
	"testing"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
			productName:        cryptoutilSharedMagic.SMProductName,
			serviceName:        cryptoutilSharedMagic.KMSServiceName,
			serviceDisplayName: cryptoutilSharedMagic.KMSDisplayName,
			wantContains: []string{
				"Usage: sm kms <subcommand> [options]",
				cryptoutilSharedMagic.KMSDisplayName,
				"Use \"sm kms <subcommand> help\"",
			},
		},
		{
			name:               "jose ja",
			productName:        cryptoutilSharedMagic.JoseProductName,
			serviceName:        cryptoutilSharedMagic.JoseJAServiceName,
			serviceDisplayName: cryptoutilSharedMagic.JoseJADisplayName,
			wantContains: []string{
				"Usage: jose ja <subcommand> [options]",
				cryptoutilSharedMagic.JoseJADisplayName,
				"Use \"jose ja <subcommand> help\"",
			},
		},
		{
			name:               "pki ca",
			productName:        cryptoutilSharedMagic.PKIProductName,
			serviceName:        cryptoutilSharedMagic.PKICAServiceName,
			serviceDisplayName: cryptoutilSharedMagic.PKICADisplayName,
			wantContains: []string{
				"Usage: pki ca <subcommand> [options]",
				cryptoutilSharedMagic.PKICADisplayName,
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
			productName:        cryptoutilSharedMagic.SMProductName,
			serviceName:        cryptoutilSharedMagic.KMSServiceName,
			serviceDisplayName: cryptoutilSharedMagic.KMSDisplayName,
			configFilePath:     fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.KMSServiceID),
			wantContains: []string{
				"Usage: sm kms server [options]",
				cryptoutilSharedMagic.KMSDisplayName,
				"configs/sm-kms/sm-kms-framework.yml",
				"SQLite",
				"PostgreSQL",
			},
		},
		{
			name:               "jose ja server",
			productName:        cryptoutilSharedMagic.JoseProductName,
			serviceName:        cryptoutilSharedMagic.JoseJAServiceName,
			serviceDisplayName: cryptoutilSharedMagic.JoseJADisplayName,
			configFilePath:     fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.JoseJAServiceID, cryptoutilSharedMagic.JoseJAServiceID),
			wantContains: []string{
				"Usage: jose ja server [options]",
				cryptoutilSharedMagic.JoseJADisplayName,
				"configs/sm-kms/sm-kms-framework.yml",
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
			productName:        cryptoutilSharedMagic.SMProductName,
			serviceName:        cryptoutilSharedMagic.KMSServiceName,
			serviceDisplayName: cryptoutilSharedMagic.KMSDisplayName,
			wantContains: []string{
				"Usage: sm kms client [options]",
				cryptoutilSharedMagic.KMSDisplayName,
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
			productName:        cryptoutilSharedMagic.SMProductName,
			serviceName:        cryptoutilSharedMagic.KMSServiceName,
			serviceDisplayName: cryptoutilSharedMagic.KMSDisplayName,
			configFilePath:     fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.KMSServiceID),
			wantContains: []string{
				"Usage: sm kms init [options]",
				cryptoutilSharedMagic.KMSDisplayName,
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
			name:              "sm kms health",
			productName:       cryptoutilSharedMagic.SMProductName,
			serviceName:       cryptoutilSharedMagic.KMSServiceName,
			defaultPublicPort: fmt.Sprintf("%d", cryptoutilSharedMagic.KMSServicePort),
			wantContains: []string{
				"Usage: sm kms health [options]",
				"https://127.0.0.1:8000",
				"https://localhost:8000",
				"/browser/api/v1/health",
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
			productName: cryptoutilSharedMagic.SMProductName,
			serviceName: cryptoutilSharedMagic.KMSServiceName,
			wantContains: []string{
				"Usage: sm kms livez [options]",
				"https://127.0.0.1:9090",
				"/admin/api/v1/livez",
			},
		},
		{
			name:        "pki ca livez",
			productName: cryptoutilSharedMagic.PKIProductName,
			serviceName: cryptoutilSharedMagic.PKICAServiceName,
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
			productName: cryptoutilSharedMagic.SMProductName,
			serviceName: cryptoutilSharedMagic.KMSServiceName,
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
			productName: cryptoutilSharedMagic.SMProductName,
			serviceName: cryptoutilSharedMagic.KMSServiceName,
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
