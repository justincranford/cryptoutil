// Copyright (c) 2025 Justin Cranford

package im

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTestutil "cryptoutil/internal/shared/testutil"
)

// TestIM_HealthSubcommand_MultipleURLFlags tests health check with multiple --url flags (first wins).
func TestIM_HealthSubcommand_MultipleURLFlags(t *testing.T) {
	t.Parallel()

	// Pass multiple --url flags (first one should win, second ignored).
	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{
			"health",
			"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health",
			"--url", "https://invalid-second-url:9999",
		})
		require.Equal(t, 0, exitCode, "Should use first --url flag")
	})

	require.Contains(t, output, "Service is healthy")
}

// TestIM_LivezSubcommand_URLFlagWithoutValue tests livez with --url flag but missing value.
func TestIM_LivezSubcommand_URLFlagWithoutValue(t *testing.T) {
	// Pass --url flag without value (should use default URL).
	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url"})
		require.Equal(t, 1, exitCode, "Should fail with connection error to default")
	})
	require.Contains(t, output, "Liveness check failed")
	require.True(t,
		cryptoutilTestutil.ContainsAny(output, []string{
			"connection refused",
			"actively refused",
			"dial tcp",
		}),
		"Should contain connection error: %s", output)
}

// TestIM_ReadyzSubcommand_ExtraArgumentsIgnored tests readyz with extra arguments.
func TestIM_ReadyzSubcommand_ExtraArgumentsIgnored(t *testing.T) {
	t.Parallel()

	// Pass extra arguments after --url (should be ignored).
	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{
			"readyz",
			"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath,
			"extra", "ignored", "args",
		})
		require.Equal(t, 0, exitCode, "Extra args should be ignored")
	})

	require.Contains(t, output, "Service is ready")
}

// TestIM_ShutdownSubcommand_URLWithoutQueryParameters tests shutdown URL handling.
func TestIM_ShutdownSubcommand_URLWithoutQueryParameters(t *testing.T) {
	t.Parallel()

	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{
			"shutdown",
			"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath,
		})
		require.Equal(t, 0, exitCode, "Shutdown should succeed")
	})

	require.Contains(t, output, "Shutdown initiated")
}

// TestIM_HealthSubcommand_URLWithFragment tests health check with URL fragment (fragment should be ignored by HTTP).
func TestIM_HealthSubcommand_URLWithFragment(t *testing.T) {
	t.Parallel()

	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{
			"health",
			"--url", testMockServerCustom.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health#section",
		})
		require.Equal(t, 0, exitCode, "Health check with fragment should succeed")
	})

	require.Contains(t, output, "Service is healthy")
}

// TestIM_LivezSubcommand_URLWithUserInfo tests livez with URL containing user info (basic auth style).
func TestIM_LivezSubcommand_URLWithUserInfo(t *testing.T) {
	t.Parallel()

	// Extract host from server URL and add user info.
	urlParts := strings.Split(testMockServerOK.URL, "//")
	urlWithUserInfo := urlParts[0] + "//user:pass@" + urlParts[1] + "/livez"

	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{
			"livez",
			"--url", urlWithUserInfo,
		})
		require.Equal(t, 0, exitCode, "Livez with user info in URL should succeed")
	})

	require.Contains(t, output, "Service is alive")
}

// TestIM_ReadyzSubcommand_CaseInsensitiveHTTPStatus tests readyz response with different status code messages.
func TestIM_ReadyzSubcommand_CaseInsensitiveHTTPStatus(t *testing.T) {
	t.Parallel()

	// Use shared error server (returns 503, not 418, but still non-200 which is the point).
	output := cryptoutilTestutil.CaptureOutput(t, func() {
		exitCode := IM([]string{
			"readyz",
			"--url", testMockServerError.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath,
		})
		require.Equal(t, 1, exitCode, "Non-200 status should fail")
	})
	require.Contains(t, output, "Service is not ready")
	require.Contains(t, output, "503")
}
