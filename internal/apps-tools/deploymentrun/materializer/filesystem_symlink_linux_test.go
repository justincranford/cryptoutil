//go:build linux

// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishFilesRejectsRealSymlinkParent(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	symlinkPath := filepath.Join(parent, "linked")
	require.NoError(t, os.Symlink(t.TempDir(), symlinkPath))

	output := filepath.Join(symlinkPath, "materialized")
	err := publishFiles(context.Background(), operatingSystemFileSystem{}, output, oneTestFile())
	require.ErrorContains(t, err, testSymlinkOrNonDirectoryError)
}

func TestPublishFilesRejectsRealSymlinkStagingDirectory(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	output := filepath.Join(parent, "materialized")
	fileSystem := &faultFileSystem{
		base:             operatingSystemFileSystem{},
		outputDirectory:  output,
		realStageSymlink: true,
	}

	err := publishFiles(context.Background(), fileSystem, output, oneTestFile())
	require.ErrorContains(t, err, "staging path is a symlink")

	entries, readErr := os.ReadDir(parent)
	require.NoError(t, readErr)
	require.Empty(t, entries)
}
