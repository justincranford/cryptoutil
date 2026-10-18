// Copyright (c) 2025-2026 Justin Cranford.
package format_text

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
)

func TestFormatWithFileOperations_ErrorPaths(t *testing.T) {
	t.Parallel()

	injectedError := errors.New("injected file operation error")
	tests := []struct {
		name      string
		statError error
		writeErr  error
		wantError string
	}{
		{name: "stat error", statError: injectedError, wantError: "failed to stat"},
		{name: "write error", writeErr: injectedError, wantError: "failed to write"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "sample.txt")
			require.NoError(t, os.WriteFile(filePath, []byte("one\r\ntwo\r\n"), 0o600))

			operations := formatFileOperations{
				readFile: os.ReadFile,
				stat: func(path string) (os.FileInfo, error) {
					if tc.statError != nil {
						return nil, tc.statError
					}

					return os.Stat(path)
				},
				writeFile: func(string, []byte, os.FileMode) error {
					return tc.writeErr
				},
			}

			err := formatWithFileOperations(
				cryptoutilCmdCicdCommon.NewLogger(t.Name()),
				map[string][]string{testTextExtension: {filePath}},
				operations,
			)

			require.ErrorContains(t, err, tc.wantError)
			require.ErrorIs(t, err, injectedError)
		})
	}
}
