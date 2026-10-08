package format_text

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testTextExtension   = "txt"
	testNormalizedLines = "one\ntwo\n"
)

func TestFormat_NormalizesLineEndings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "crlf", content: "one\r\ntwo\r\n", want: testNormalizedLines},
		{name: "mixed line endings", content: "one\r\ntwo\nthree\r\n", want: "one\ntwo\nthree\n"},
		{name: "lf unchanged", content: testNormalizedLines, want: testNormalizedLines},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample.txt")
			require.NoError(t, os.WriteFile(filePath, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			err := Format(cryptoutilCmdCicdCommon.NewLogger("test"), map[string][]string{testTextExtension: {filePath}})
			require.NoError(t, err)

			content, err := os.ReadFile(filePath)
			require.NoError(t, err)
			require.Equal(t, tc.want, string(content))
		})
	}
}

func TestFormat_ReadError(t *testing.T) {
	t.Parallel()

	filePath := filepath.Join(t.TempDir(), "missing.txt")
	err := Format(cryptoutilCmdCicdCommon.NewLogger("test"), map[string][]string{testTextExtension: {filePath}})
	require.ErrorContains(t, err, "failed to read")
}
