// Copyright (c) 2025-2026 Justin Cranford.
package files

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	goSourceExtension = "go"
	goModuleExtension = "mod"
	goModuleFilename  = "go.mod"
)

// RepositoryGoPackageListing contains explicit Go package patterns and
// root-relative nested checkout roots skipped during repository enumeration.
type RepositoryGoPackageListing struct {
	PackagePatterns    []string
	SkippedDirectories []string
}

// ListRepositoryGoPackages discovers Go packages under startDirectory while
// respecting repository exclusions, nested Git checkout boundaries, and
// nested go.mod module boundaries. Returned patterns are safe explicit package
// paths such as "." and "./internal/example"; no recursive wildcard is used.
func ListRepositoryGoPackages(startDirectory string) (RepositoryGoPackageListing, error) {
	return ListRepositoryGoPackagesWithLstat(startDirectory, os.Lstat)
}

// ListRepositoryGoPackagesWithLstat is the Lstat-injectable form of ListRepositoryGoPackages.
func ListRepositoryGoPackagesWithLstat(
	startDirectory string,
	lstatFn func(string) (os.FileInfo, error),
) (RepositoryGoPackageListing, error) {
	if lstatFn == nil {
		return RepositoryGoPackageListing{}, fmt.Errorf("missing Lstat function")
	}

	listing, err := ListRepositoryFilesWithOptionsAndLstat(startDirectory, []string{goSourceExtension, goModuleExtension}, nil, lstatFn)
	if err != nil {
		return RepositoryGoPackageListing{}, err
	}

	rootDir := filepath.Clean(startDirectory)
	moduleDirectories := make(map[string]struct{})

	for _, moduleFile := range listing.FilesByExtension[goModuleExtension] {
		if filepath.Base(moduleFile) == goModuleFilename {
			moduleDirectories[filepath.Clean(filepath.Dir(moduleFile))] = struct{}{}
		}
	}

	packages := make(map[string]struct{})

	for _, goFile := range listing.FilesByExtension[goSourceExtension] {
		packageDir := filepath.Clean(filepath.Dir(goFile))

		nestedModule, moduleErr := isNestedOSModulePackage(rootDir, packageDir, moduleDirectories)
		if moduleErr != nil {
			return RepositoryGoPackageListing{}, moduleErr
		}

		if nestedModule {
			continue
		}

		relativeDir, relErr := filepath.Rel(rootDir, packageDir)
		if relErr != nil {
			return RepositoryGoPackageListing{}, fmt.Errorf("failed to determine package path for %s: %w", packageDir, relErr)
		}

		if relativeDir == cryptoutilSharedMagic.CurrentDirectory {
			packages[cryptoutilSharedMagic.CurrentDirectory] = struct{}{}

			continue
		}

		packages["./"+filepath.ToSlash(relativeDir)] = struct{}{}
	}

	patterns := make([]string, 0, len(packages))
	for pattern := range packages {
		patterns = append(patterns, pattern)
	}

	sort.Strings(patterns)

	return RepositoryGoPackageListing{PackagePatterns: patterns, SkippedDirectories: listing.SkippedDirectories}, nil
}

// ListRepositoryGoPackagesFromFS discovers Go packages using fsys for both traversal and nested .git marker inspection.
func ListRepositoryGoPackagesFromFS(fsys fs.FS, startDirectory string) (RepositoryGoPackageListing, error) {
	listing, err := ListRepositoryFilesFromFSWithOptions(fsys, startDirectory, []string{goSourceExtension, goModuleExtension}, nil)
	if err != nil {
		return RepositoryGoPackageListing{}, err
	}

	rootDir := path.Clean(startDirectory)
	moduleDirectories := make(map[string]struct{})

	for _, moduleFile := range listing.FilesByExtension[goModuleExtension] {
		if path.Base(moduleFile) == goModuleFilename {
			moduleDirectories[fsRelativePath(rootDir, path.Clean(path.Dir(moduleFile)))] = struct{}{}
		}
	}

	packages := make(map[string]struct{})

	for _, goFile := range listing.FilesByExtension[goSourceExtension] {
		packageDir := fsRelativePath(rootDir, path.Clean(path.Dir(goFile)))
		if isNestedFSModulePackage(packageDir, moduleDirectories) {
			continue
		}

		if packageDir == cryptoutilSharedMagic.CurrentDirectory {
			packages[cryptoutilSharedMagic.CurrentDirectory] = struct{}{}

			continue
		}

		packages["./"+packageDir] = struct{}{}
	}

	patterns := make([]string, 0, len(packages))
	for pattern := range packages {
		patterns = append(patterns, pattern)
	}

	sort.Strings(patterns)

	return RepositoryGoPackageListing{PackagePatterns: patterns, SkippedDirectories: listing.SkippedDirectories}, nil
}

func isNestedOSModulePackage(rootDir, packageDir string, moduleDirectories map[string]struct{}) (bool, error) {
	relativeDir, err := filepath.Rel(rootDir, packageDir)
	if err != nil {
		return false, fmt.Errorf("failed to determine module-relative package path for %s: %w", packageDir, err)
	}

	currentDir := packageDir
	for relativeDir != cryptoutilSharedMagic.CurrentDirectory {
		if _, exists := moduleDirectories[currentDir]; exists {
			return true, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
		relativeDir = filepath.Dir(relativeDir)
	}

	return false, nil
}

func isNestedFSModulePackage(packageDir string, moduleDirectories map[string]struct{}) bool {
	for currentDir := packageDir; currentDir != cryptoutilSharedMagic.CurrentDirectory && currentDir != ""; currentDir = path.Dir(currentDir) {
		if _, exists := moduleDirectories[currentDir]; exists {
			return true
		}
	}

	return false
}

func fsRelativePath(rootDir, candidatePath string) string {
	if rootDir == cryptoutilSharedMagic.CurrentDirectory {
		return path.Clean(candidatePath)
	}

	if candidatePath == rootDir {
		return cryptoutilSharedMagic.CurrentDirectory
	}

	return strings.TrimPrefix(candidatePath, rootDir+"/")
}
