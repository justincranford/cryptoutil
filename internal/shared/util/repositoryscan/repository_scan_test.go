// Copyright (c) 2025-2026 Justin Cranford.
package repositoryscan

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testGitMarkerFile      = "git marker file"
	testGitMarkerDirectory = "git marker directory"
	testDanglingGitMarker  = "dangling git marker symlink"
	testExcludedDirectory  = "excluded"
	testNestedSourceFile   = "repo/nested/source.go"
)

func TestShouldSkipDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		markerIsDir bool
		markerIsSym bool
		marked      bool
		excluded    bool
		wantSkip    bool
	}{
		{name: "unmarked directory remains scannable"},
		{name: testGitMarkerFile, marked: true, wantSkip: true},
		{name: testGitMarkerDirectory, markerIsDir: true, marked: true, wantSkip: true},
		{name: testDanglingGitMarker, markerIsSym: true, marked: true, wantSkip: true},
		{name: "shared excluded basename", excluded: true, wantSkip: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()

			candidateName := "ordinary-name"
			if tc.excluded {
				candidateName = testExcludedDirectory
			}

			candidateDir := filepath.Join(rootDir, candidateName)
			require.NoError(t, os.MkdirAll(candidateDir, 0o700))

			if tc.marked {
				gitMarkerPath := filepath.Join(candidateDir, cryptoutilSharedMagic.CICDExcludeDirGit)
				if tc.markerIsSym {
					err := os.Symlink("missing-marker-target", gitMarkerPath)
					if err != nil {
						t.Skipf("symlink creation is unavailable on this platform: %v", err)
					}
				} else if tc.markerIsDir {
					require.NoError(t, os.Mkdir(gitMarkerPath, 0o700))
				} else {
					require.NoError(t, os.WriteFile(gitMarkerPath, []byte("gitdir: ../.git/worktrees/ordinary-name"), 0o600))
				}
			}

			shouldSkip, err := ShouldSkipDir(rootDir, candidateDir, func(name string) bool {
				return name == testExcludedDirectory
			})

			require.NoError(t, err)
			require.Equal(t, tc.wantSkip, shouldSkip)
		})
	}
}

func TestShouldSkipDirFS_UsesSuppliedFilesystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		marked      bool
		markerIsDir bool
		markerIsSym bool
	}{
		{name: "unmarked directory"},
		{name: testGitMarkerFile, marked: true},
		{name: testGitMarkerDirectory, markerIsDir: true, marked: true},
		{name: testDanglingGitMarker, markerIsSym: true, marked: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			filesystem := fstest.MapFS{testNestedSourceFile: {Data: []byte("package nested")}}
			if test.markerIsSym {
				filesystem["repo/nested/.git"] = &fstest.MapFile{
					Data: []byte("missing-marker-target"),
					Mode: fs.ModeSymlink,
				}
			} else if test.markerIsDir {
				filesystem["repo/nested/.git/HEAD"] = &fstest.MapFile{Data: []byte("ref: refs/heads/main")}
			} else if test.marked {
				filesystem["repo/nested/.git"] = &fstest.MapFile{Data: []byte("gitdir: ../.git/worktrees/nested")}
			}

			shouldSkip, err := ShouldSkipDirFS(filesystem, "repo", "repo/nested", nil)

			require.NoError(t, err)
			require.Equal(t, test.marked, shouldSkip)
		})
	}
}

func TestShouldSkipDirFS_PropagatesInspectionErrors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected supplied filesystem error")
	filesystem := failingReadDirFS{
		FS:        fstest.MapFS{testNestedSourceFile: {Data: []byte("package nested")}},
		errorPath: "repo/nested",
		err:       injectedErr,
	}

	shouldSkip, err := ShouldSkipDirFS(filesystem, "repo", "repo/nested", nil)

	require.ErrorIs(t, err, injectedErr)
	require.False(t, shouldSkip)
}

func TestHasGitMarkerFS_RejectsMissingFilesystem(t *testing.T) {
	t.Parallel()

	hasMarker, err := HasGitMarkerFS(nil, "repo")

	require.ErrorContains(t, err, "missing filesystem")
	require.False(t, hasMarker)
}

func TestShouldSkipDir_RootIsRetainedBeforeExclusionsAndMarkerInspection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: cryptoutilSharedMagic.CICDExcludeDirDotWorktree},
		{name: cryptoutilSharedMagic.CICDExcludeDirWorktree},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := filepath.Join(t.TempDir(), tc.name)
			require.NoError(t, os.MkdirAll(rootDir, 0o700))

			shouldSkip, err := ShouldSkipDirWithLstat(
				rootDir,
				rootDir,
				func(string) bool { return true },
				func(string) (os.FileInfo, error) {
					return nil, errors.New("root marker inspection must not run")
				},
			)

			require.NoError(t, err)
			require.False(t, shouldSkip)
		})
	}
}

func TestShouldSkipDir_PropagatesMarkerInspectionErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	candidateDir := filepath.Join(rootDir, "ordinary-name")
	injectedErr := errors.New("injected marker inspection error")

	shouldSkip, err := ShouldSkipDirWithLstat(
		rootDir,
		candidateDir,
		func(string) bool { return false },
		func(string) (os.FileInfo, error) {
			return nil, injectedErr
		},
	)

	require.ErrorIs(t, err, injectedErr)
	require.False(t, shouldSkip)
}

type failingReadDirFS struct {
	fs.FS
	errorPath string
	err       error
}

func (filesystem failingReadDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == filesystem.errorPath {
		return nil, filesystem.err
	}

	return fs.ReadDir(filesystem.FS, name)
}
