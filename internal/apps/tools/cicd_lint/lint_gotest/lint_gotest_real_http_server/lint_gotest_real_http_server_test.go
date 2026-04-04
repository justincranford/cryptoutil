// Copyright (c) 2025 Justin Cranford

package lint_gotest_real_http_server

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
			name:     "no_violation_app_test",
			fileName: "handler_test.go",
			fileContent: `package test

import (
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/path", nil)
	_ = req
}
`,
			wantErr: false,
		},
		{
			name:     "violation_new_server",
			fileName: "handler_test.go",
			fileContent: `package test

import (
	"net/http/httptest"
	"testing"
)

func TestBadHandler(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(nil)
	defer srv.Close()
}
`,
			wantErr: true,
		},
		{
			name:     "exempt_client_path",
			fileName: "client/oauth_test.go",
			fileContent: `package test

import "net/http/httptest"

func TestOAuth(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()
}
`,
			wantErr: false,
		},
		{
			name:     "exempt_realm_path",
			fileName: "realm/federation_test.go",
			fileContent: `package test

import "net/http/httptest"

func TestFed(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()
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
			err := os.MkdirAll(filepath.Dir(testFile), cryptoutilSharedMagic.CICDTempDirPermissions)
			require.NoError(t, err)
			err = os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
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

func TestCheckRealHTTPServerUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name: "new_server_violation",
			fileContent: `srv := httptest.NewServer(handler)
`,
			wantIssues: true,
		},
		{
			name: "new_recorder_no_violation",
			fileContent: `rec := httptest.NewRecorder()
`,
			wantIssues: false,
		},
		{
			name: "new_request_no_violation",
			fileContent: `req := httptest.NewRequest("GET", "/", nil)
`,
			wantIssues: false,
		},
		{
			name: "multiple_violations",
			fileContent: `srv1 := httptest.NewServer(h1)
srv2 := httptest.NewServer(h2)
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

			issues := CheckRealHTTPServerUsage(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckRealHTTPServerUsage_ReadError(t *testing.T) {
	t.Parallel()

	issues := CheckRealHTTPServerUsage("/nonexistent/path/test_test.go")

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
			name:     "exempt_client_dir",
			filePath: "/internal/apps/identity-authz/client/oauth_test.go",
			want:     true,
		},
		{
			name:     "exempt_realm_dir",
			filePath: "/internal/apps/identity-idp/realm/federation_test.go",
			want:     true,
		},
		{
			name:     "exempt_clientauth_dir",
			filePath: "/internal/apps/identity-authz/clientauth/revocation_test.go",
			want:     true,
		},
		{
			name:     "exempt_testing_dir",
			filePath: "/internal/apps/framework/service/testing/healthclient/healthclient_test.go",
			want:     true,
		},
		{
			name:     "exempt_util_network_dir",
			filePath: "/internal/shared/util/network/http_test.go",
			want:     true,
		},
		{
			name:     "exempt_backchannel",
			filePath: "/internal/apps/identity-idp/backchannel_logout_test.go",
			want:     true,
		},
		{
			name:     "not_exempt_handler",
			filePath: "/internal/apps/sm-kms/handler/keys_test.go",
			want:     false,
		},
		{
			name:     "not_exempt_service",
			filePath: "/internal/apps/sm-kms/service/keys_test.go",
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
