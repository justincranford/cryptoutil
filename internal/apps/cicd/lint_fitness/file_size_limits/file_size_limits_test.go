// Copyright (c) 2025 Justin Cranford

package file_size_limits

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func makeLines(n int) string {
	var sb strings.Builder
	sb.WriteString("package foo\n\n")

	for range n {
		sb.WriteString("// line ")
		sb.WriteString(strings.Repeat("x", 1))
		sb.WriteString("\n")
	}

	return sb.String()
}

func TestCheckInDir_SmallFile_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	f := filepath.Join(tmp, "small.go")
	require.NoError(t, os.WriteFile(f, []byte("package foo\n\nfunc Foo() {}\n"), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_MediumFile_WarnOnly(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// 350 lines - exceeds soft limit (300) but below hard limit (500), should produce WARN but no error.
	f := filepath.Join(tmp, "medium.go")
	require.NoError(t, os.WriteFile(f, []byte(makeLines(350)), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err) // warning only, no error
}

func TestCheckInDir_LargeFile_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// 510 lines - above hard limit (500), should fail.
	f := filepath.Join(tmp, "large.go")
	require.NoError(t, os.WriteFile(f, []byte(makeLines(510)), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
}

func TestCheckInDir_GeneratedFile_Excluded(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Generated files (*_gen.go) are excluded from size checks.
	f := filepath.Join(tmp, "server.gen.go")
	require.NoError(t, os.WriteFile(f, []byte(makeLines(600)), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_APIDir_Excluded(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	apiDir := filepath.Join(tmp, "api")
	require.NoError(t, os.MkdirAll(apiDir, cryptoutilSharedMagic.DirPermissions))
	f := filepath.Join(apiDir, "models.go")
	require.NoError(t, os.WriteFile(f, []byte(makeLines(600)), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_MagicDir_Excluded(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	magicDir := filepath.Join(tmp, "internal", "shared", "magic")
	require.NoError(t, os.MkdirAll(magicDir, cryptoutilSharedMagic.DirPermissions))
	f := filepath.Join(magicDir, "constants.go")
	require.NoError(t, os.WriteFile(f, []byte(makeLines(600)), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_TestFile_Excluded(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	f := filepath.Join(tmp, "foo_test.go")
	require.NoError(t, os.WriteFile(f, []byte(makeLines(510)), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestShouldExclude_Various(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantExcl bool
	}{
		{"normal go file", "internal/foo/bar.go", false},
		{"generated file", "api/server.gen.go", true},
		{"test file", "foo_test.go", true},
		{"magic dir", "internal/shared/magic/constants.go", true},
		// vendor and api dirs are skipped at directory-walk level, not in shouldExclude.
		{"api dir file (not excluded by shouldExclude)", "api/model/types.go", false},
		{"vendor dir file (not excluded by shouldExclude)", "vendor/github.com/foo.go", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := shouldExclude(tc.path)
			require.Equal(t, tc.wantExcl, got)
		})
	}
}

func TestCountLines_ValidFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.go")
	require.NoError(t, os.WriteFile(f, []byte("line1\nline2\nline3\n"), cryptoutilSharedMagic.CacheFilePermissions))
	n, err := countLines(f)
	require.NoError(t, err)
	require.Equal(t, 3, n)
}

func TestCountLines_NonexistentFile_Error(t *testing.T) {
	t.Parallel()

	_, err := countLines("/nonexistent/file.go")
	require.Error(t, err)
}


func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-file-size-limits")

	err = Check(logger)
	require.NoError(t, err)
}
