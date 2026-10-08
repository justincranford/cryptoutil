package failures

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestFixFile_ConvertsTestingFailureCalls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		want     []string
		wantPath []string
	}{
		{
			name: "adds default assertion imports and preserves formatted args",
			content: `package example

import "testing"

func TestExample(t *testing.T) {
	t.Errorf("value %d", 1)
	t.Fatalf("missing %s", "key")
}
`,
			want: []string{
				"assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, \"value %d\", 1)",
				"require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, \"missing %s\", \"key\")",
			},
			wantPath: []string{
				"cryptoutil/internal/shared/magic",
				"github.com/stretchr/testify/assert",
				"github.com/stretchr/testify/require",
			},
		},
		{
			name: "preserves existing aliases",
			content: `package example

import (
	"testing"
	testAssert "github.com/stretchr/testify/assert"
	testRequire "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	t.Errorf("failure")
	t.Fatalf("fatal")
}
`,
			want: []string{
				"testAssert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, \"failure\")",
				"testRequire.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, \"fatal\")",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample_test.go")
			require.NoError(t, os.WriteFile(filePath, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			changed, fixes, err := fixFile(filePath)
			require.NoError(t, err)
			require.True(t, changed)
			require.Equal(t, 2, fixes)

			formatted, err := os.ReadFile(filePath)
			require.NoError(t, err)

			for _, expected := range tc.want {
				require.Contains(t, string(formatted), expected)
			}

			for _, expected := range tc.wantPath {
				require.Contains(t, string(formatted), expected)
			}

			_, err = parser.ParseFile(token.NewFileSet(), filePath, formatted, parser.AllErrors)
			require.NoError(t, err)

			changed, fixes, err = fixFile(filePath)
			require.NoError(t, err)
			require.False(t, changed)
			require.Zero(t, fixes)
		})
	}
}

func TestFixFile_AddsImportWithoutExistingImportBlock(t *testing.T) {
	t.Parallel()

	content := `package example

type testT struct{}

func (*testT) Errorf(string, ...any) {}

func TestExample(t *testT) {
	t.Errorf("failure")
}
`
	filePath := filepath.Join(t.TempDir(), "sample_test.go")
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	changed, fixes, err := fixFile(filePath)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, 1, fixes)

	formatted, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(formatted), "github.com/stretchr/testify/assert")
	require.Contains(t, string(formatted), "assert.Failf(t,")
}

func TestFixFile_LeavesOtherCallsUnchanged(t *testing.T) {
	t.Parallel()

	content := `package example

import "testing"

func TestExample(t *testing.T) {
	t.Error("failure")
}
`
	filePath := filepath.Join(t.TempDir(), "sample_test.go")
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	changed, fixes, err := fixFile(filePath)
	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, fixes)
}

func TestFixFile_ReplacesLegacyFailureMessage(t *testing.T) {
	t.Parallel()

	content := fmt.Sprintf(`package example

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	assert.Failf(t, %q, "failure")
}
`, cryptoutilSharedMagic.CICDTestErrorfFailureMessage)
	filePath := filepath.Join(t.TempDir(), "sample_test.go")
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	changed, fixes, err := fixFile(filePath)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, 1, fixes)

	formatted, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(formatted), "assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, \"failure\")")

	changed, fixes, err = fixFile(filePath)
	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, fixes)
}

func TestFix_DiscoversAndFormatsFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	filePath := filepath.Join(rootDir, "sample_test.go")
	content := `package example

import "testing"

func TestExample(t *testing.T) {
	t.Fatalf("fatal")
}
`
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	processed, modified, fixed, err := Fix(cryptoutilCmdCicdCommon.NewLogger("test"), rootDir)
	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, 1, modified)
	require.Equal(t, 1, fixed)
}
