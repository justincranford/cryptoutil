// Copyright (c) 2025 Justin Cranford

package lint_gotest_test_sleep

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheck_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{})

	require.NoError(t, err)
}

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileName    string
		fileContent string
		wantErr     bool
	}{
		{
			name:     "no_violation_channel_sync",
			fileName: "handler_test.go",
			fileContent: `package test

import (
	"testing"
)

func TestSync(t *testing.T) {
	t.Parallel()
	done := make(chan struct{})
	close(done)
	<-done
}
`,
			wantErr: false,
		},
		{
			name:     "violation_time_sleep",
			fileName: "handler_test.go",
			fileContent: `package test

import (
	"testing"
	"time"
)

func TestBad(t *testing.T) {
	t.Parallel()
	time.Sleep(100 * time.Millisecond)
}
`,
			wantErr: true,
		},
		{
			name:     "exempt_rate_limiter",
			fileName: "rate_limiter_test.go",
			fileContent: `package test

import "time"

func TestRate(t *testing.T) {
	time.Sleep(time.Second)
}
`,
			wantErr: false,
		},
		{
			name:     "exempt_cleanup_suffix",
			fileName: "session_cleanup_test.go",
			fileContent: `package test

import "time"

func TestCleanup(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
}
`,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, tc.fileName)
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = Check(logger, []string{testFile})

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckTestSleepUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name: "time_sleep_violation",
			fileContent: `time.Sleep(100 * time.Millisecond)
`,
			wantIssues: true,
		},
		{
			name: "time_now_no_violation",
			fileContent: `now := time.Now()
`,
			wantIssues: false,
		},
		{
			name: "time_after_no_violation",
			fileContent: `<-time.After(time.Second)
`,
			wantIssues: false,
		},
		{
			name: "multiple_violations",
			fileContent: `time.Sleep(100 * time.Millisecond)
doSomething()
time.Sleep(200 * time.Millisecond)
`,
			wantIssues: true,
		},
		{
			name:        "empty_file",
			fileContent: ``,
			wantIssues:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "example_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			issues := CheckTestSleepUsage(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckTestSleepUsage_ReadError(t *testing.T) {
	t.Parallel()

	issues := CheckTestSleepUsage("/nonexistent/path/test_test.go")

	require.NotEmpty(t, issues)
	require.Contains(t, issues[0], "Error reading file")
}

func TestIsExemptedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{
			name:     "exempt_rate_limiter",
			filePath: "/internal/apps/sm-kms/rate_limiter_test.go",
			want:     true,
		},
		{
			name:     "exempt_session_manager",
			filePath: "/internal/apps/framework/session_manager_test.go",
			want:     true,
		},
		{
			name:     "exempt_cleanup_suffix",
			filePath: "/internal/apps/sm-kms/session_cleanup_test.go",
			want:     true,
		},
		{
			name:     "exempt_cleanup_integration_suffix",
			filePath: "/internal/apps/sm-kms/session_cleanup_integration_test.go",
			want:     true,
		},
		{
			name:     "exempt_rotation_suffix",
			filePath: "/internal/apps/sm-kms/key_rotation_test.go",
			want:     true,
		},
		{
			name:     "exempt_pool_dir",
			filePath: "/internal/shared/pool/keygen_pool_test.go",
			want:     true,
		},
		{
			name:     "exempt_telemetry_dir",
			filePath: "/internal/shared/telemetry/sidecar_test.go",
			want:     true,
		},
		{
			name:     "exempt_listener_suffix",
			filePath: "/internal/apps/sm-im/server_listener_test.go",
			want:     true,
		},
		{
			name:     "exempt_listener_db_suffix",
			filePath: "/internal/apps/sm-im/server_listener_db_test.go",
			want:     true,
		},
		{
			name:     "exempt_listener_send_suffix",
			filePath: "/internal/apps/sm-im/server_listener_send_test.go",
			want:     true,
		},
		{
			name:     "exempt_application_test",
			filePath: "/internal/apps/sm-kms/application_test.go",
			want:     true,
		},
		{
			name:     "exempt_concurrent_suffix",
			filePath: "/internal/apps/sm-kms/handler_concurrent_test.go",
			want:     true,
		},
		{
			name:     "exempt_shutdown_suffix",
			filePath: "/internal/apps/sm-kms/server_shutdown_test.go",
			want:     true,
		},
		{
			name:     "exempt_health_shutdown_suffix",
			filePath: "/internal/apps/sm-kms/server_health_shutdown_test.go",
			want:     true,
		},
		{
			name:     "exempt_coverage2_suffix",
			filePath: "/internal/apps/sm-kms/service_coverage2_test.go",
			want:     true,
		},
		{
			name:     "exempt_public_suffix",
			filePath: "/internal/apps/sm-kms/repo/public_test.go",
			want:     true,
		},
		{
			name:     "exempt_table_suffix",
			filePath: "/internal/apps/sm-kms/repo/keys_table_test.go",
			want:     true,
		},
		{
			name:     "exempt_integration_suffix",
			filePath: "/internal/apps/sm-kms/service/keys_integration_test.go",
			want:     true,
		},
		{
			name:     "exempt_repository_suffix",
			filePath: "/internal/apps/sm-kms/repo/keys_repository_test.go",
			want:     true,
		},
		{
			name:     "exempt_testserver_dir",
			filePath: "/internal/apps/framework/service/testing/testserver/helper_test.go",
			want:     true,
		},
		{
			name:     "exempt_authorization_suffix",
			filePath: "/internal/apps/identity-authz/device_authorization_test.go",
			want:     true,
		},
		{
			name:     "exempt_authenticator_suffix",
			filePath: "/internal/apps/identity-idp/webauthn_authenticator_test.go",
			want:     true,
		},
		{
			name:     "exempt_cache_suffix",
			filePath: "/internal/apps/identity-authz/policy_cache_test.go",
			want:     true,
		},
		{
			name:     "exempt_revocation_suffix",
			filePath: "/internal/apps/identity-authz/token_revocation_test.go",
			want:     true,
		},
		{
			name:     "exempt_highcov_suffix",
			filePath: "/internal/apps/sm-kms/service/keys_highcov_test.go",
			want:     true,
		},
		{
			name:     "exempt_cert_ops_suffix",
			filePath: "/internal/apps/pki-ca/handler_cert_ops_test.go",
			want:     true,
		},
		{
			name:     "exempt_observability",
			filePath: "/internal/apps/pki-ca/observability/observability_test.go",
			want:     true,
		},
		{
			name:     "exempt_cancel_suffix",
			filePath: "/internal/apps/pki-ca/ra_cancel_test.go",
			want:     true,
		},
		{
			name:     "exempt_http_errors",
			filePath: "/internal/apps/sm-im/http_errors_test.go",
			want:     true,
		},
		{
			name:     "exempt_util_network",
			filePath: "/internal/shared/util/network/http_test.go",
			want:     true,
		},
		{
			name:     "exempt_util_thread",
			filePath: "/internal/shared/util/thread/pool_test.go",
			want:     true,
		},
		{
			name:     "exempt_logger_suffix",
			filePath: "/internal/apps/tools/cicd_lint/logger_test.go",
			want:     true,
		},
		{
			name:     "exempt_outdated_deps",
			filePath: "/internal/apps/tools/cicd_lint/lint_go/outdated_deps/check_test.go",
			want:     true,
		},
		{
			name:     "exempt_error_validation_suffix",
			filePath: "/internal/apps/pki-ca/hardware_error_validation_test.go",
			want:     true,
		},
		{
			name:     "exempt_certificate_dir",
			filePath: "/internal/shared/crypto/certificate/cert_test.go",
			want:     true,
		},
		{
			name:     "exempt_clientauth",
			filePath: "/internal/apps/identity-authz/clientauth/revocation_test.go",
			want:     true,
		},
		{
			name:     "exempt_identity_authz_handlers",
			filePath: "/internal/apps/identity-authz/handlers/token_test.go",
			want:     true,
		},
		{
			name:     "not_exempt_handler",
			filePath: "/internal/apps/sm-kms/handler/keys_test.go",
			want:     false,
		},
		{
			name:     "not_exempt_repository",
			filePath: "/internal/apps/sm-kms/repo/some_test.go",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isExemptedFile(tc.filePath)

			require.Equal(t, tc.want, got)
		})
	}
}
