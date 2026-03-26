// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

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
				"--url", testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + "/health",
				"--url", "https://invalid-second-url:9999",
			},
			wantExitCode: 0,
			wantContains: []string{"Service is healthy"},
		},
		{
			name: "HealthSubcommand_URLWithFragment",
			args: []string{
				"health",
				"--url", testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + "/health#section",
			},
			wantExitCode: 0,
			wantContains: []string{"Service is healthy"},
		},
		{
			name: "ReadyzSubcommand_ExtraArgumentsIgnored",
			args: []string{
				"readyz",
				"--url", testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
				"extra", "ignored", "args",
			},
			wantExitCode: 0,
			wantContains: []string{"Service is ready"},
		},
		{
			name: "ShutdownSubcommand_URLWithoutQueryParameters",
			args: []string{
				"shutdown",
				"--url", testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath,
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

					return urlParts[0] + "//user:pass@" + urlParts[1] + cryptoutilSharedMagic.PrivateAdminLivezRequestPath
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
				"EOF", // Can happen when nothing is listening on default port.
			},
		},
		{
			name: "ReadyzSubcommand_CaseInsensitiveHTTPStatus",
			args: []string{
				"readyz",
				"--url", testMockServerError.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			},
			wantExitCode: 1,
			wantContains: []string{"Service is not ready", "503"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Im(tt.args, nil, &stdout, &stderr)
			require.Equal(t, tt.wantExitCode, exitCode)

			output := stdout.String() + stderr.String()

			for _, want := range tt.wantContains {
				require.Contains(t, output, want)
			}

			if len(tt.wantAny) > 0 {
				require.True(t,
					cryptoutilSharedTestutil.ContainsAny(output, tt.wantAny),
					"Should contain one of: %v, got: %s", tt.wantAny, output)
			}
		})
	}
}
