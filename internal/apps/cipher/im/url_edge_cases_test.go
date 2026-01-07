// Copyright (c) 2025 Justin Cranford

package im

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTestutil "cryptoutil/internal/shared/testutil"
)

// TestIM_URLEdgeCases tests various URL edge cases using table-driven tests.
func TestIM_URLEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantContains []string
		wantAny      []string // for cases where we check ContainsAny
	}{
		{
			name: "HealthSubcommand_ExtraURLIgnored",
			args: []string{
				"health",
				"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health",
				"--url", "https://invalid-second-url:9999",
			},
			wantExitCode: 0,
			wantContains: []string{"Service is healthy"},
		},
		{
			name: "HealthSubcommand_URLWithFragment",
			args: []string{
				"health",
				"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health#section",
			},
			wantExitCode: 0,
			wantContains: []string{"Service is healthy"},
		},
		{
			name: "ReadyzSubcommand_ExtraArgumentsIgnored",
			args: []string{
				"readyz",
				"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath,
				"extra", "ignored", "args",
			},
			wantExitCode: 0,
			wantContains: []string{"Service is ready"},
		},
		{
			name: "ShutdownSubcommand_URLWithoutQueryParameters",
			args: []string{
				"shutdown",
				"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath,
			},
			wantExitCode: 0,
			wantContains: []string{"Shutdown initiated"},
		},
		{
			name: "LivezSubcommand_URLWithUserInfo",
			args: []string{
				"livez",
				"--url", func() string {
					// Extract host from server URL and add user info.
					urlParts := strings.Split(testMockServerOK.URL, "//")

					return urlParts[0] + "//user:pass@" + urlParts[1] + "/livez"
				}(),
			},
			wantExitCode: 0,
			wantContains: []string{"Service is alive"},
		},
		{
			name:         "LivezSubcommand_URLFlagWithoutValue",
			args:         []string{"livez", "--url"},
			wantExitCode: 1,
			wantContains: []string{"Liveness check failed"},
			wantAny: []string{
				"connection refused",
				"actively refused",
				"dial tcp",
			},
		},
		{
			name: "ReadyzSubcommand_CaseInsensitiveHTTPStatus",
			args: []string{
				"readyz",
				"--url", testMockServerError.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath,
			},
			wantExitCode: 1,
			wantContains: []string{"Service is not ready", "503"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			exitCode := internalIM(tt.args, &stdout, &stderr)
			require.Equal(t, tt.wantExitCode, exitCode)

			output := stdout.String() + stderr.String()

			for _, want := range tt.wantContains {
				require.Contains(t, output, want)
			}

			if len(tt.wantAny) > 0 {
				require.True(t,
					cryptoutilTestutil.ContainsAny(output, tt.wantAny),
					"Should contain one of: %v, got: %s", tt.wantAny, output)
			}
		})
	}
}
