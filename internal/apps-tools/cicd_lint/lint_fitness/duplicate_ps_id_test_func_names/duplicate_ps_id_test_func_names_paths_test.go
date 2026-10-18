// Copyright (c) 2025-2026 Justin Cranford.
package duplicate_ps_id_test_func_names

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheckAndFindDuplicatesErrors(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("duplicate-test-functions")

	require.NoError(t, Check(logger))

	_, err := FindDuplicates(filepath.Join(t.TempDir(), "missing"))
	require.ErrorContains(t, err, "walking")

	err = CheckInDir(logger, filepath.Join(t.TempDir(), "missing"))
	require.ErrorContains(t, err, "scan failed")
}

func TestFindDuplicatesCollectError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.CMD_SERVER, "server_test.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filePath, []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	info, err := os.Stat(filePath)
	require.NoError(t, err)

	_, err = findDuplicatesWith(
		root,
		func(_ string, walkFn fs.WalkDirFunc) error {
			return walkFn(filePath, fs.FileInfoToDirEntry(info), nil)
		},
		func(string) ([]string, error) {
			return nil, errors.New("injected collect error")
		},
	)

	require.ErrorContains(t, err, "injected collect error")
}

func TestFindDuplicatesSkipsVendorDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	testPath := filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirVendor, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.CMD_SERVER, "server_test.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(testPath), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(testPath, []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	results, err := FindDuplicates(root)

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestExtractPSIDFromServerTestPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
		ok   bool
	}{
		{name: "valid server test", path: "internal/apps/sm-kms/server/server_test.go", want: cryptoutilSharedMagic.OTLPServiceSMKMS, ok: true},
		{name: "prefix missing", path: "apps/sm-kms/server/server_test.go"},
		{name: "service component missing", path: "internal/apps/sm-kms"},
		{name: "framework excluded", path: "internal/apps/framework/server/server_test.go"},
		{name: "tools excluded", path: "internal/apps/tools/server/server_test.go"},
		{name: "template service excluded", path: "internal/apps/template/server/server_test.go"},
		{name: "skeleton-template service included", path: "internal/apps/skeleton-template/server/server_test.go", want: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, ok: true},
		{name: "not server package", path: "internal/apps/sm-kms/client/client_test.go"},
		{name: "nested server package excluded", path: "internal/apps/sm-kms/server/handler/handler_test.go"},
		{name: "integration excluded", path: "internal/apps/sm-kms/server/server_integration_test.go"},
		{name: "e2e excluded", path: "internal/apps/sm-kms/server/server_e2e_test.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			psID, ok := extractPSIDFromServerTestPath(tc.path)

			require.Equal(t, tc.want, psID)
			require.Equal(t, tc.ok, ok)
		})
	}
}

func TestCollectTestFuncNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "top level tests only",
			content: `package server
import "testing"
func TestIncluded(t *testing.T) {}
func helper(t *testing.T) {}
type suite struct{}
func (suite) TestMethod(t *testing.T) {}
`,
			want: []string{"TestIncluded"},
		},
		{name: "malformed source is ignored", content: "package server\nfunc TestBroken( {\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "server_test.go")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			names, err := collectTestFuncNames(path)

			require.NoError(t, err)
			require.Equal(t, tc.want, names)
		})
	}

	_, err := collectTestFuncNames(filepath.Join(t.TempDir(), "missing_test.go"))
	require.ErrorContains(t, err, "reading")
}
