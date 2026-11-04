// Copyright (c) 2025-2026 Justin Cranford.
package files_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

func TestListRepositoryGoPackages_SkipsNestedCheckoutsAndModules(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writePackageFixtureFile(t, rootDir, "go.mod", "module example.test/root\ngo 1.22\n")
	writePackageFixtureFile(t, rootDir, "root.go", "package root\n")
	writePackageFixtureFile(t, rootDir, filepath.Join("pkg", "pkg.go"), "package pkg\n")
	writePackageFixtureFile(t, rootDir, filepath.Join(testForeignCheckout, cryptoutilSharedMagic.CICDExcludeDirGit), "gitdir: ../.git/worktrees/foreign\n")
	writePackageFixtureFile(t, rootDir, filepath.Join(testForeignCheckout, "foreign.go"), "package foreign\n")
	writePackageFixtureFile(t, rootDir, filepath.Join("nested-module", "go.mod"), "module example.test/nested\ngo 1.22\n")
	writePackageFixtureFile(t, rootDir, filepath.Join("nested-module", "nested.go"), "package nested\n")

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryGoPackages(rootDir)

	require.NoError(t, err)
	require.Equal(t, []string{testRootPackage, testSubPackage}, listing.PackagePatterns)
	require.Equal(t, []string{testForeignCheckout}, listing.SkippedDirectories)
}

func TestListRepositoryGoPackagesFromFS_UsesSuppliedFilesystem(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"repo/go.mod":                  {Data: []byte("module example.test/root\ngo 1.22\n")},
		testRepositoryRootFile:         {Data: []byte("package root\n")},
		"repo/pkg/pkg.go":              {Data: []byte("package pkg\n")},
		"repo/foreign/.git":            {Data: []byte("missing-marker-target"), Mode: 0o777 | os.ModeSymlink},
		"repo/foreign/foreign.go":      {Data: []byte("package foreign\n")},
		"repo/nested-module/go.mod":    {Data: []byte("module example.test/nested\ngo 1.22\n")},
		"repo/nested-module/nested.go": {Data: []byte("package nested\n")},
	}

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryGoPackagesFromFS(filesystem, "repo")

	require.NoError(t, err)
	require.Equal(t, []string{testRootPackage, testSubPackage}, listing.PackagePatterns)
	require.Equal(t, []string{testForeignCheckout}, listing.SkippedDirectories)
}

func TestListRepositoryGoPackages_EmptyRootHasNoPackages(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writePackageFixtureFile(t, rootDir, "go.mod", "module example.test/empty\ngo 1.22\n")

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryGoPackages(rootDir)

	require.NoError(t, err)
	require.Empty(t, listing.PackagePatterns)
}

func TestListRepositoryGoPackagesWithLstat_RejectsMissingLstat(t *testing.T) {
	t.Parallel()

	listing, err := cryptoutilSharedUtilFiles.ListRepositoryGoPackagesWithLstat(t.TempDir(), nil)

	require.ErrorContains(t, err, "missing Lstat function")
	require.Nil(t, listing.PackagePatterns)
}

func writePackageFixtureFile(t *testing.T, rootDir, relativePath, content string) {
	t.Helper()

	filePath := filepath.Join(rootDir, relativePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o700))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))
}
