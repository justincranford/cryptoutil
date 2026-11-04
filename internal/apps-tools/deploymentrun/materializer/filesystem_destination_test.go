// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishFilesFlatOutput(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	output := filepath.Join(parent, "materialized")
	err := publishFiles(context.Background(), operatingSystemFileSystem{}, output, map[string]materializedFile{
		"manifest": {data: []byte("complete")},
	})
	require.NoError(t, err)

	content, readErr := os.ReadFile(filepath.Join(output, "manifest"))
	require.NoError(t, readErr)
	require.Equal(t, []byte("complete"), content)
}

func TestPublishFilesRejectsUnsafeParentsAndReuse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		prepare             func(*testing.T, string, string)
		injectedSymlinkPath func(string) string
		outputPath          func(string) string
		wantErr             string
	}{
		{
			name: "existing destination is rejected",
			prepare: func(t *testing.T, _ string, output string) {
				require.NoError(t, os.Mkdir(output, privateDirectoryMode))
			},
			outputPath: func(parent string) string { return filepath.Join(parent, "materialized") },
			wantErr:    testExistingOutputError,
		},
		{
			name:       "missing parent is rejected",
			outputPath: func(parent string) string { return filepath.Join(parent, "missing", "materialized") },
			wantErr:    "inspect materializer output parent",
		},
		{
			name: "non-directory parent is rejected",
			prepare: func(t *testing.T, parent string, _ string) {
				require.NoError(t, os.WriteFile(filepath.Join(parent, "file"), []byte("x"), privateFileMode))
			},
			outputPath: func(parent string) string { return filepath.Join(parent, "file", "materialized") },
			wantErr:    testSymlinkOrNonDirectoryError,
		},
		{
			name:                "injected symlink parent is rejected",
			injectedSymlinkPath: func(parent string) string { return filepath.Join(parent, "linked") },
			outputPath:          func(parent string) string { return filepath.Join(parent, "linked", "materialized") },
			wantErr:             testSymlinkOrNonDirectoryError,
		},
		{
			name:       "relative destination is rejected",
			outputPath: func(string) string { return "relative/path" },
			wantErr:    testCleanAbsolutePathError,
		},
		{
			name: "non-clean destination is rejected",
			outputPath: func(parent string) string {
				return parent + string(filepath.Separator) + ".." + string(filepath.Separator) + "unsafe" +
					string(filepath.Separator) + cryptoutilSharedMagic.CICDExcludeDirOutput
			},
			wantErr: testCleanAbsolutePathError,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parent := t.TempDir()

			output := testCase.outputPath(parent)
			if testCase.prepare != nil {
				testCase.prepare(t, parent, output)
			}

			fileSystem := &faultFileSystem{
				base:            operatingSystemFileSystem{},
				outputDirectory: output,
			}
			if testCase.injectedSymlinkPath != nil {
				fileSystem.symlinkPath = testCase.injectedSymlinkPath(parent)
			}

			err := publishFiles(context.Background(), fileSystem, output, oneTestFile())
			require.ErrorContains(t, err, testCase.wantErr)
		})
	}
}
