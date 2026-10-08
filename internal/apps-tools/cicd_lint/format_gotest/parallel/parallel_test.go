package parallel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

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
