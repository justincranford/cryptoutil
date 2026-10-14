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
			name:               "skeleton template",
			productName:        cryptoutilSharedMagic.SkeletonProductName,
			serviceName:        cryptoutilSharedMagic.SkeletonTemplateServiceName,
			serviceDisplayName: cryptoutilSharedMagic.SkeletonTemplateDisplayName,
			wantContains: []string{
				"Usage: skeleton template <subcommand> [options]",
				cryptoutilSharedMagic.SkeletonTemplateDisplayName,
				"Use \"skeleton template <subcommand> help\"",
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
				testConfigsSmKMSSmKMSFrameworkYml,
				"SQLite",
				"PostgreSQL",
			},
		},
		{
			name:               "skeleton template server",
			productName:        cryptoutilSharedMagic.SkeletonProductName,
			serviceName:        cryptoutilSharedMagic.SkeletonTemplateServiceName,
			serviceDisplayName: cryptoutilSharedMagic.SkeletonTemplateDisplayName,
			configFilePath:     fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.SkeletonTemplateServiceID, cryptoutilSharedMagic.SkeletonTemplateServiceID),
			wantContains: []string{
				"Usage: skeleton template server [options]",
				cryptoutilSharedMagic.SkeletonTemplateDisplayName,
				"configs/skeleton-template/skeleton-template-framework.yml",
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
			name:               testSmKMSClient,
			productName:        cryptoutilSharedMagic.SMProductName,
			serviceName:        cryptoutilSharedMagic.KMSServiceName,
			serviceDisplayName: cryptoutilSharedMagic.KMSDisplayName,
			wantContains: []string{
				"Usage: sm kms client [options]",
				cryptoutilSharedMagic.KMSDisplayName,
				testSmKMSClient,
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
				testConfigsSmKMSSmKMSFrameworkYml,
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
				cryptoutilSharedMagic.DefaultBrowserHealthEndpoint,
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
				testHTTPS1270019090,
				testAdminAPIV1Livez,
			},
		},
		{
			name:        "skeleton template livez",
			productName: cryptoutilSharedMagic.SkeletonProductName,
			serviceName: cryptoutilSharedMagic.SkeletonTemplateServiceName,
			wantContains: []string{
				"Usage: skeleton template livez [options]",
				testHTTPS1270019090,
				testAdminAPIV1Livez,
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
				testHTTPS1270019090,
				testAdminAPIV1Readyz,
			},
		},
		{
			name:        "skeleton template readyz",
			productName: cryptoutilSharedMagic.SkeletonProductName,
			serviceName: cryptoutilSharedMagic.SkeletonTemplateServiceName,
			wantContains: []string{
				"Usage: skeleton template readyz [options]",
				testHTTPS1270019090,
				testAdminAPIV1Readyz,
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
				testHTTPS1270019090,
				testAdminAPIV1Shutdown,
				testForce,
			},
		},
		{
			name:        "skeleton template shutdown",
			productName: cryptoutilSharedMagic.SkeletonProductName,
			serviceName: cryptoutilSharedMagic.SkeletonTemplateServiceName,
			wantContains: []string{
				"Usage: skeleton template shutdown [options]",
				testHTTPS1270019090,
				testAdminAPIV1Shutdown,
				testForce,
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
