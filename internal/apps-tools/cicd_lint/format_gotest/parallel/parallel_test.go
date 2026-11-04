package parallel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestFix_AddsParallelAndHonorsSequentialMarker(t *testing.T) {
	t.Parallel()

	content := "package example\n\nimport \"testing\"\n\nfunc TestNeedsParallel(t *testing.T) {}\n\n// Sequential: mutates global state.\nfunc TestSequential(t *testing.T) {}\n\nfunc TestAlreadyParallel(t *testing.T) { t.Parallel() }\n"
	filePath := filepath.Join(t.TempDir(), "test_test.go")
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	changed, fixes, err := fixFile(filePath)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, 1, fixes)

	formatted, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, 2, strings.Count(string(formatted), "t.Parallel()"), string(formatted))
	require.Contains(t, string(formatted), "// Sequential: mutates global state.")

	changed, fixes, err = fixFile(filePath)
	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, fixes)
}

func TestFix_LeavesNonTestAndNonTestingFunctionsUnchanged(t *testing.T) {
	t.Parallel()

	content := "package example\n\nimport \"testing\"\n\nfunc Helper(t *testing.T) {}\n\nfunc TestMain(m *testing.M) {}\n"
	filePath := filepath.Join(t.TempDir(), "test_test.go")
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	changed, fixes, err := fixFile(filePath)
	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, fixes)
}

func TestFix_AddsParallelWhenOnlySubtestIsParallel(t *testing.T) {
	t.Parallel()

	content := "package example\n\nimport \"testing\"\n\nfunc TestParent(t *testing.T) {\n\tt.Run(\"child\", func(t *testing.T) { t.Parallel() })\n}\n"
	filePath := filepath.Join(t.TempDir(), "test_test.go")
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	changed, fixes, err := fixFile(filePath)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, 1, fixes)

	formatted, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(formatted), "func TestParent(t *testing.T) {\n\tt.Parallel()")
}

func TestFix_DoesNotMutateNestedCheckout(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	nestedCheckoutDir := filepath.Join(rootDir, "ordinary-name")
	require.NoError(t, os.MkdirAll(nestedCheckoutDir, 0o700))

	rootFile := filepath.Join(rootDir, "root_test.go")
	rootContent := "package example\n\nimport \"testing\"\n\nfunc TestRoot(t *testing.T) {}\n"
	require.NoError(t, os.WriteFile(rootFile, []byte(rootContent), cryptoutilSharedMagic.CacheFilePermissions))

	nestedFile := filepath.Join(nestedCheckoutDir, "nested_test.go")
	nestedContent := "package example\n\nimport \"testing\"\n\nfunc TestNested(t *testing.T) {}\n"
	require.NoError(t, os.WriteFile(nestedFile, []byte(nestedContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(
		filepath.Join(nestedCheckoutDir, cryptoutilSharedMagic.CICDExcludeDirGit),
		[]byte("gitdir: ../.git/worktrees/ordinary-name"),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	processed, modified, fixed, err := Fix(cryptoutilCmdCicdCommon.NewLogger("test"), rootDir)

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, fixed)

	formattedRoot, err := os.ReadFile(rootFile)
	require.NoError(t, err)
	require.NotEqual(t, rootContent, string(formattedRoot))
	require.Contains(t, string(formattedRoot), "t.Parallel()")

	unmodifiedNested, err := os.ReadFile(nestedFile)
	require.NoError(t, err)
	require.Equal(t, nestedContent, string(unmodifiedNested))
}
