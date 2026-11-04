// Copyright (c) 2025-2026 Justin Cranford.
package common

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestIsRepositoryScanExcludedDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{name: "git metadata", dir: cryptoutilSharedMagic.CICDExcludeDirGit, want: true},
		{name: "hidden worktree container", dir: cryptoutilSharedMagic.CICDExcludeDirDotWorktree, want: true},
		{name: "worktree container", dir: cryptoutilSharedMagic.CICDExcludeDirWorktree, want: true},
		{name: "DAST reports", dir: cryptoutilSharedMagic.CICDExcludeDirDastReports, want: true},
		{name: "E2E reports", dir: cryptoutilSharedMagic.CICDExcludeDirE2EReports, want: true},
		{name: "load reports", dir: cryptoutilSharedMagic.CICDExcludeDirLoadReports, want: true},
		{name: "test results", dir: cryptoutilSharedMagic.CICDExcludeDirTestResults, want: true},
		{name: "temporary output", dir: cryptoutilSharedMagic.CICDExcludeDirTemp, want: true},
		{name: "legacy CI cache", dir: cryptoutilSharedMagic.CICDExcludeDirDotCICD, want: true},
		{name: "CI lint cache", dir: cryptoutilSharedMagic.CICDExcludeDirDotCICDLint, want: true},
		{name: "build output", dir: cryptoutilSharedMagic.CICDExcludeDirBuild, want: true},
		{name: "distribution output", dir: cryptoutilSharedMagic.CICDExcludeDirDist, want: true},
		{name: "generated output", dir: cryptoutilSharedMagic.CICDExcludeDirOutput, want: true},
		{name: "legacy Python virtual environment", dir: cryptoutilSharedMagic.CICDExcludeDirVenvLegacy, want: true},
		{name: "Python environment", dir: cryptoutilSharedMagic.CICDExcludeDirEnv, want: true},
		{name: "hidden environment", dir: cryptoutilSharedMagic.CICDExcludeDirDotEnv, want: true},
		{name: "cache", dir: cryptoutilSharedMagic.CICDExcludeDirCache, want: true},
		{name: "generic target source remains scannable", dir: cryptoutilSharedMagic.CICDExcludeDirTarget, want: false},
		{name: "Python bytecode cache", dir: cryptoutilSharedMagic.CICDExcludeDirPycache, want: true},
		{name: "pytest cache", dir: cryptoutilSharedMagic.CICDExcludeDirPytestCache, want: true},
		{name: "tox environment", dir: cryptoutilSharedMagic.CICDExcludeDirTox, want: true},
		{name: "nox environment", dir: cryptoutilSharedMagic.CICDExcludeDirNox, want: true},
		{name: "Hypothesis cache", dir: cryptoutilSharedMagic.CICDExcludeDirHypothesis, want: true},
		{name: "HTML coverage output", dir: cryptoutilSharedMagic.CICDExcludeDirHTMLCov, want: true},
		{name: "Ruff cache", dir: cryptoutilSharedMagic.CICDExcludeDirRuffCache, want: true},
		{name: "mypy cache", dir: cryptoutilSharedMagic.CICDExcludeDirMypyCache, want: true},
		{name: "mutation test output", dir: cryptoutilSharedMagic.CICDExcludeDirGremlins, want: true},
		{name: "vendored dependencies", dir: cryptoutilSharedMagic.CICDExcludeDirVendor, want: true},
		{name: "generated test output", dir: cryptoutilSharedMagic.CICDExcludeDirTestOutput, want: true},
		{name: "Python virtual environment", dir: cryptoutilSharedMagic.CICDExcludeDirVenv, want: true},
		{name: "Node.js dependencies", dir: cryptoutilSharedMagic.CICDExcludeDirNodeModules, want: true},
		{name: "generated workflow reports", dir: cryptoutilSharedMagic.CICDExcludeDirWorkflowReports, want: true},
		{name: "tracked source directory", dir: "internal", want: false},
		{name: "API source root", dir: "api", want: false},
		{name: "generated API component remains scannable", dir: "client", want: false},
		{name: "generated API model component remains scannable", dir: "model", want: false},
		{name: "generated API server component remains scannable", dir: cryptoutilSharedMagic.CMD_SERVER, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, IsRepositoryScanExcludedDir(test.dir))
		})
	}
}

func TestShouldSkipRepositoryScanDir_NestedCheckoutMarkers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		markerIsDir bool
	}{
		{name: "git marker file"},
		{name: "git marker directory", markerIsDir: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			candidateDir := filepath.Join(rootDir, "ordinary-name")
			require.NoError(t, os.MkdirAll(candidateDir, 0o700))

			gitMarkerPath := filepath.Join(candidateDir, cryptoutilSharedMagic.CICDExcludeDirGit)
			if tc.markerIsDir {
				require.NoError(t, os.Mkdir(gitMarkerPath, 0o700))
			} else {
				require.NoError(t, os.WriteFile(gitMarkerPath, []byte("gitdir: ../.git/worktrees/ordinary-name"), 0o600))
			}

			shouldSkip, err := ShouldSkipRepositoryScanDir(rootDir, candidateDir)

			require.NoError(t, err)
			require.True(t, shouldSkip)
		})
	}
}

func TestShouldSkipRepositoryScanDir_UnmarkedDirectoryRemainsScannable(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	candidateDir := filepath.Join(rootDir, "ordinary-source")
	require.NoError(t, os.MkdirAll(candidateDir, 0o700))

	shouldSkip, err := ShouldSkipRepositoryScanDir(rootDir, candidateDir)

	require.NoError(t, err)
	require.False(t, shouldSkip)
}

func TestShouldSkipRepositoryScanDir_SelectedWorktreeRootRemainsScannable(t *testing.T) {
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
			require.NoError(t, os.WriteFile(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGit), []byte("gitdir: ../.git/worktrees/selected"), 0o600))

			shouldSkip, err := ShouldSkipRepositoryScanDir(rootDir, rootDir)

			require.NoError(t, err)
			require.False(t, shouldSkip)
		})
	}
}

func TestShouldSkipRepositoryScanDir_PropagatesMarkerInspectionErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	candidateDir := filepath.Join(rootDir, "ordinary-name")
	require.NoError(t, os.MkdirAll(candidateDir, 0o700))

	injectedErr := errors.New("injected marker inspection error")

	var inspectedPath string

	shouldSkip, err := shouldSkipRepositoryScanDir(rootDir, candidateDir, func(path string) (os.FileInfo, error) {
		inspectedPath = path

		return nil, injectedErr
	})

	require.ErrorIs(t, err, injectedErr)
	require.Equal(t, filepath.Join(candidateDir, cryptoutilSharedMagic.CICDExcludeDirGit), inspectedPath)
	require.False(t, shouldSkip)
}
