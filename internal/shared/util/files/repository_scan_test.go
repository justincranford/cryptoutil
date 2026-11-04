// Copyright (c) 2025-2026 Justin Cranford.
package files_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

func TestListRepositoryFilesFromFSWithOptions_SkipsNestedCheckouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		markerIsDir bool
		markerIsSym bool
	}{
		{name: testGitMarkerFile},
		{name: testGitMarkerDirectory, markerIsDir: true},
		{name: testDanglingGitMarker, markerIsSym: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filesystem := fstest.MapFS{
				testRepositoryRootFile:    {Data: []byte("package root")},
				"repo/ordinary/nested.go": {Data: []byte("package nested")},
			}
			if test.markerIsSym {
				filesystem["repo/ordinary/.git"] = &fstest.MapFile{
					Data: []byte("missing-marker-target"),
					Mode: fs.ModeSymlink,
				}
			} else if test.markerIsDir {
				filesystem["repo/ordinary/.git/HEAD"] = &fstest.MapFile{Data: []byte("ref: refs/heads/main")}
			} else {
				filesystem["repo/ordinary/.git"] = &fstest.MapFile{Data: []byte("gitdir: ../.git/worktrees/ordinary")}
			}

			listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesFromFSWithOptions(
				filesystem,
				"repo",
				[]string{testGo},
				nil,
			)

			require.NoError(t, err)
			require.Equal(t, []string{testRepositoryRootFile}, listing.FilesByExtension[testGo])
			require.Equal(t, []string{"ordinary"}, listing.SkippedDirectories)
		})
	}
}

func TestListRepositoryFilesWithOptionsAndLstat_SkipsNestedCheckoutMarkers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		markerIsDir bool
		markerIsSym bool
	}{
		{name: testGitMarkerFile},
		{name: testGitMarkerDirectory, markerIsDir: true},
		{name: testDanglingGitMarker, markerIsSym: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			nestedDir := filepath.Join(rootDir, "ordinary-name")
			require.NoError(t, os.MkdirAll(nestedDir, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(rootDir, "root.go"), []byte("package root"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "nested.go"), []byte("package nested"), 0o600))

			markerPath := filepath.Join(nestedDir, cryptoutilSharedMagic.CICDExcludeDirGit)

			switch {
			case test.markerIsSym:
				if err := os.Symlink("missing-marker-target", markerPath); err != nil {
					t.Skipf("symlink creation is unavailable on this platform: %v", err)
				}
			case test.markerIsDir:
				require.NoError(t, os.Mkdir(markerPath, 0o700))
			default:
				require.NoError(t, os.WriteFile(markerPath, []byte("gitdir: ../.git/worktrees/ordinary-name"), 0o600))
			}

			listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesWithOptionsAndLstat(
				rootDir,
				[]string{testGo},
				nil,
				os.Lstat,
			)

			require.NoError(t, err)
			require.Equal(t, []string{filepath.ToSlash(filepath.Join(rootDir, "root.go"))}, listing.FilesByExtension[testGo])
			require.Equal(t, []string{"ordinary-name"}, listing.SkippedDirectories)
		})
	}
}

func TestListRepositoryFilesWithOptions_SkipsNestedWorktreeContainers(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "root.go"), []byte("package root"), 0o600))

	for _, directoryName := range []string{
		cryptoutilSharedMagic.CICDExcludeDirDotWorktree,
		cryptoutilSharedMagic.CICDExcludeDirWorktree,
	} {
		nestedDir := filepath.Join(rootDir, directoryName)
		require.NoError(t, os.MkdirAll(nestedDir, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "nested.go"), []byte("package nested"), 0o600))
	}

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesWithOptions(
		rootDir,
		[]string{testGo},
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, []string{filepath.ToSlash(filepath.Join(rootDir, "root.go"))}, listing.FilesByExtension[testGo])
	require.Empty(t, listing.SkippedDirectories)
}

func TestListRepositoryFilesWithOptionsAndLstat_ExplicitWorktreeRootIsScanned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: cryptoutilSharedMagic.CICDExcludeDirDotWorktree},
		{name: cryptoutilSharedMagic.CICDExcludeDirWorktree},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rootDir := filepath.Join(t.TempDir(), test.name)
			nestedDir := filepath.Join(rootDir, testNestedDirectory)
			require.NoError(t, os.MkdirAll(nestedDir, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGit), []byte("root marker"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(rootDir, "root.go"), []byte("package selected"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(nestedDir, cryptoutilSharedMagic.CICDExcludeDirGit), []byte("nested marker"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "nested.go"), []byte("package nested"), 0o600))

			listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesWithOptionsAndLstat(
				rootDir,
				[]string{testGo},
				nil,
				os.Lstat,
			)

			require.NoError(t, err)
			require.Equal(t, []string{filepath.ToSlash(filepath.Join(rootDir, "root.go"))}, listing.FilesByExtension[testGo])
			require.Equal(t, []string{testNestedDirectory}, listing.SkippedDirectories)
		})
	}
}

func TestListRepositoryFilesWithOptionsAndLstat_PropagatesMarkerInspectionErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	nestedDir := filepath.Join(rootDir, testNestedDirectory)
	require.NoError(t, os.MkdirAll(nestedDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "nested.go"), []byte("package nested"), 0o600))

	injectedErr := errors.New("injected Lstat failure")

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesWithOptionsAndLstat(
		rootDir,
		[]string{testGo},
		nil,
		func(path string) (os.FileInfo, error) {
			if filepath.Clean(path) == filepath.Join(nestedDir, cryptoutilSharedMagic.CICDExcludeDirGit) {
				return nil, injectedErr
			}

			return os.Lstat(path)
		},
	)

	require.ErrorIs(t, err, injectedErr)
	require.Nil(t, listing.FilesByExtension)
}

func TestListRepositoryFilesFromFSWithOptions_SkipsNestedWorktreeContainers(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		testRepositoryRootFile:     {Data: []byte("package root")},
		"repo/.worktree/nested.go": {Data: []byte("package nested")},
		"repo/worktree/nested.go":  {Data: []byte("package nested")},
	}

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesFromFSWithOptions(
		filesystem,
		"repo",
		[]string{testGo},
		nil,
	)

	require.NoError(t, err)
	require.Equal(t, []string{testRepositoryRootFile}, listing.FilesByExtension[testGo])
	require.Empty(t, listing.SkippedDirectories)
}

func TestListRepositoryFilesFromFSWithOptions_ExplicitWorktreeRootIsScanned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: cryptoutilSharedMagic.CICDExcludeDirDotWorktree},
		{name: cryptoutilSharedMagic.CICDExcludeDirWorktree},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filesystem := fstest.MapFS{
				test.name + "/.git":        {Data: []byte("gitdir: ../.git/worktrees/selected")},
				test.name + "/root.go":     {Data: []byte("package selected")},
				test.name + "/nested/a.go": {Data: []byte("package nested")},
				test.name + "/nested/.git": {Data: []byte("gitdir: ../.git/worktrees/nested")},
			}

			listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesFromFSWithOptions(
				filesystem,
				test.name,
				[]string{testGo},
				nil,
			)

			require.NoError(t, err)
			require.Equal(t, []string{test.name + "/root.go"}, listing.FilesByExtension[testGo])
			require.Equal(t, []string{testNestedDirectory}, listing.SkippedDirectories)
		})
	}
}

func TestListRepositoryFilesFromFSWithOptions_PropagatesMarkerInspectionErrors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected directory inspection failure")
	filesystem := readDirErrorFS{
		FS:        fstest.MapFS{"repo/nested/file.go": {Data: []byte("package nested")}},
		errorPath: "repo/nested",
		err:       injectedErr,
	}

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesFromFSWithOptions(
		filesystem,
		"repo",
		[]string{testGo},
		nil,
	)

	require.ErrorIs(t, err, injectedErr)
	require.Nil(t, listing.FilesByExtension)
}

func TestListRepositoryFilesFromFSWithOptions_PreservesRelativeExclusions(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		testRepositorySourceFile:      {Data: []byte("package source")},
		"repo/generated/cache.go":     {Data: []byte("package generated")},
		"repo/generated/sub/cache.go": {Data: []byte("package generated")},
	}

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryFilesFromFSWithOptions(
		filesystem,
		"repo",
		[]string{testGo},
		[]string{testGeneratedDirectory},
	)

	require.NoError(t, err)
	require.Equal(t, []string{testRepositorySourceFile}, listing.FilesByExtension[testGo])
}

type readDirErrorFS struct {
	fs.FS
	errorPath string
	err       error
}

func (filesystem readDirErrorFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == filesystem.errorPath {
		return nil, filesystem.err
	}

	return fs.ReadDir(filesystem.FS, name)
}
