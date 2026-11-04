// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package files provides file system utility functions.
package files

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedRepositoryScan "cryptoutil/internal/shared/util/repositoryscan"
)

// WriteFile writes content to a file with the specified permissions.
// If permissions is 0, it defaults to CacheFilePermissions.
// The content can be a string or []byte.
func WriteFile(filePath string, bytesOrString any, permissions os.FileMode) error {
	if permissions == 0 {
		return fmt.Errorf("missing file permissions")
	}

	var data []byte

	switch v := bytesOrString.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("content must be string or []byte, got %T", bytesOrString)
	}

	err := os.WriteFile(filePath, data, permissions)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ListAllFiles walks the directory tree starting from startDirectory and returns files grouped by extension.
// It uses the extension inclusions and directory exclusions from magic constants.
// Returns a map where keys are file extensions (without dot) and values are slices of relative file paths.
// Paths are normalized to use forward slashes for cross-platform compatibility.
// Excluded directories are skipped entirely.
func ListAllFiles(startDirectory string) (map[string][]string, error) {
	listing, err := ListRepositoryFilesWithOptions(startDirectory, cryptoutilSharedMagic.TextFilenameExtensionInclusions, nil)
	if err != nil {
		return nil, err
	}

	return listing.FilesByExtension, nil
}

// ListAllFilesWithOptions walks the directory tree starting from startDirectory and returns files grouped by extension.
// It uses the provided extension inclusions and directory exclusions.
// Returns a map where keys are file extensions (without dot) and values are slices of relative file paths.
// Paths are normalized to use forward slashes for cross-platform compatibility.
// Only the caller-provided exclusions are applied; nested Git checkout detection is intentionally opt-in through
// ListRepositoryFilesWithOptions.
func ListAllFilesWithOptions(startDirectory string, inclusions []string, exclusions []string) (map[string][]string, error) {
	listing, err := listFilesWithOptions(startDirectory, inclusions, exclusions, false, os.Lstat)
	if err != nil {
		return nil, err
	}

	return listing.FilesByExtension, nil
}

// RepositoryFileListing contains repository-scoped files and root-relative nested checkout roots skipped by
// repository-safe enumeration.
type RepositoryFileListing struct {
	FilesByExtension   map[string][]string
	SkippedDirectories []string
}

// ListRepositoryFilesWithOptions walks the directory tree starting from startDirectory and returns files grouped by
// extension. It applies the repository directory exclusions, caller exclusions, and nested Git checkout boundary.
// The explicitly selected root is always scanned, even when its basename is normally excluded or it contains .git.
func ListRepositoryFilesWithOptions(
	startDirectory string,
	inclusions []string,
	exclusions []string,
) (RepositoryFileListing, error) {
	return ListRepositoryFilesWithOptionsAndLstat(startDirectory, inclusions, exclusions, os.Lstat)
}

// ListRepositoryFilesWithOptionsAndLstat is the Lstat-injectable form of
// ListRepositoryFilesWithOptions. The supplied Lstat function is used for
// nested marker inspection while traversal remains OS-backed.
func ListRepositoryFilesWithOptionsAndLstat(
	startDirectory string,
	inclusions []string,
	exclusions []string,
	lstatFn func(string) (os.FileInfo, error),
) (RepositoryFileListing, error) {
	if lstatFn == nil {
		return RepositoryFileListing{}, fmt.Errorf("missing Lstat function")
	}

	exclusions = repositoryFileExclusions(exclusions)

	return listFilesWithOptions(startDirectory, inclusions, exclusions, true, lstatFn)
}

// ListRepositoryFilesFromFSWithOptions performs repository-safe enumeration using fsys for both walking and nested
// .git marker inspection. It applies the repository directory exclusions and caller exclusions. Returned paths use
// slash separators and are rooted at startDirectory; skipped checkout roots are relative to startDirectory.
func ListRepositoryFilesFromFSWithOptions(
	fsys fs.FS,
	startDirectory string,
	inclusions []string,
	exclusions []string,
) (RepositoryFileListing, error) {
	if fsys == nil {
		return RepositoryFileListing{}, fmt.Errorf("missing filesystem")
	}

	rootDir := path.Clean(startDirectory)
	if !fs.ValidPath(rootDir) {
		return RepositoryFileListing{}, fmt.Errorf("invalid filesystem root %q", rootDir)
	}

	exclusions = repositoryFileExclusions(exclusions)

	includedExtensions := make(map[string]bool, len(inclusions))
	for _, extension := range inclusions {
		includedExtensions[extension] = true
	}

	listing := RepositoryFileListing{FilesByExtension: make(map[string][]string)}

	err := fs.WalkDir(fsys, rootDir, func(walkPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if path.Clean(walkPath) == rootDir {
				return nil
			}

			if isExcludedFSDirectory(rootDir, walkPath, exclusions) {
				return fs.SkipDir
			}

			shouldSkip, skipErr := cryptoutilSharedRepositoryScan.ShouldSkipDirFS(
				fsys,
				rootDir,
				walkPath,
				func(string) bool { return false },
			)
			if skipErr != nil {
				return fmt.Errorf("inspect filesystem repository boundary %s: %w", walkPath, skipErr)
			}

			if shouldSkip {
				relativePath := strings.TrimPrefix(path.Clean(walkPath), rootDir)
				relativePath = strings.TrimPrefix(relativePath, "/")
				listing.SkippedDirectories = append(listing.SkippedDirectories, relativePath)

				return fs.SkipDir
			}

			return nil
		}

		extension := strings.TrimPrefix(path.Ext(walkPath), ".")
		if extension == "" {
			baseName := path.Base(walkPath)
			if strings.HasPrefix(baseName, ".") {
				extension = strings.TrimPrefix(baseName, ".")
			}
		}

		if includedExtensions[extension] {
			listing.FilesByExtension[extension] = append(listing.FilesByExtension[extension], path.Clean(walkPath))
		}

		return nil
	})
	if err != nil {
		return RepositoryFileListing{}, fmt.Errorf("failed to walk filesystem directory: %w", err)
	}

	return listing, nil
}

func repositoryFileExclusions(exclusions []string) []string {
	repositoryExclusions := make([]string, 0, len(cryptoutilSharedMagic.DirectoryNameExclusions)+len(exclusions)+1)
	repositoryExclusions = append(repositoryExclusions, cryptoutilSharedMagic.DirectoryNameExclusions...)
	repositoryExclusions = append(repositoryExclusions, cryptoutilSharedMagic.DirectoryNameExclusionTestLoadTarget)
	repositoryExclusions = append(repositoryExclusions, exclusions...)

	return repositoryExclusions
}

func listFilesWithOptions(
	startDirectory string,
	inclusions []string,
	exclusions []string,
	skipNestedCheckouts bool,
	lstatFn func(string) (os.FileInfo, error),
) (RepositoryFileListing, error) {
	matches := make(map[string][]string)
	rootDir := filepath.Clean(startDirectory)

	// Build a set of included extensions for fast lookup.
	includedExtensions := make(map[string]bool, len(inclusions))
	for _, ext := range inclusions {
		includedExtensions[ext] = true
	}

	listing := RepositoryFileListing{FilesByExtension: matches}

	err := filepath.Walk(startDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Normalize path to forward slashes for consistency.
		normalizedPath := filepath.ToSlash(path)

		if info.IsDir() {
			shouldSkip, nestedCheckout, skipErr := shouldSkipDirectoryForFileListing(
				rootDir,
				path,
				normalizedPath,
				exclusions,
				skipNestedCheckouts,
				lstatFn,
			)
			if skipErr != nil {
				return skipErr
			}

			if shouldSkip {
				if nestedCheckout {
					relativePath, relErr := filepath.Rel(rootDir, path)
					if relErr != nil {
						return fmt.Errorf("failed to determine relative path for %s: %w", path, relErr)
					}

					listing.SkippedDirectories = append(listing.SkippedDirectories, filepath.ToSlash(relativePath))
				}

				return filepath.SkipDir
			}

			return nil
		}

		// Extract the extension (without the dot).
		ext := strings.TrimPrefix(filepath.Ext(path), ".")

		// Handle files without extension (like .gitignore, .dockerignore).
		if ext == "" {
			// Check if the base name itself is in inclusions (e.g., "gitignore" for ".gitignore").
			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, ".") {
				ext = strings.TrimPrefix(baseName, ".")
			}
		}

		// Only include files with matching extensions.
		if includedExtensions[ext] {
			listing.FilesByExtension[ext] = append(listing.FilesByExtension[ext], normalizedPath)
		}

		return nil
	})
	if err != nil {
		return RepositoryFileListing{}, fmt.Errorf("failed to walk directory: %w", err)
	}

	return listing, nil
}

func shouldSkipDirectoryForFileListing(
	rootDir string,
	candidateDir string,
	normalizedPath string,
	exclusions []string,
	skipNestedCheckouts bool,
	lstatFn func(string) (os.FileInfo, error),
) (bool, bool, error) {
	if filepath.Clean(candidateDir) == rootDir {
		return false, false, nil
	}

	isExcludedDirectory := func(name string) bool {
		for _, exclusion := range exclusions {
			if !filepath.IsAbs(exclusion) && filepath.Base(exclusion) == exclusion && name == exclusion {
				return true
			}
		}

		return false
	}

	if isExcludedDirectory(filepath.Base(candidateDir)) {
		return true, false, nil
	}

	relPath, relErr := filepath.Rel(rootDir, candidateDir)
	if relErr != nil {
		return false, false, fmt.Errorf("failed to determine relative path for %s: %w", candidateDir, relErr)
	}

	relSlash := filepath.ToSlash(relPath)

	for _, exclusion := range exclusions {
		if relSlash == exclusion || strings.HasPrefix(relSlash, exclusion+"/") ||
			normalizedPath == exclusion || strings.HasPrefix(normalizedPath, exclusion+"/") {
			return true, false, nil
		}
	}

	if skipNestedCheckouts {
		shouldSkip, err := cryptoutilSharedRepositoryScan.ShouldSkipDirWithLstat(
			rootDir,
			candidateDir,
			nil,
			lstatFn,
		)
		if err != nil {
			return false, false, fmt.Errorf("inspect repository boundary %s: %w", candidateDir, err)
		}

		return shouldSkip, shouldSkip, nil
	}

	return false, false, nil
}

func isExcludedFSDirectory(rootDir, candidateDir string, exclusions []string) bool {
	if path.Clean(rootDir) == path.Clean(candidateDir) {
		return false
	}

	for _, exclusion := range exclusions {
		if !path.IsAbs(exclusion) && path.Base(exclusion) == exclusion && path.Base(candidateDir) == exclusion {
			return true
		}

		relativePath := fsRelativePath(path.Clean(rootDir), path.Clean(candidateDir))

		if relativePath == exclusion || strings.HasPrefix(relativePath, exclusion+"/") {
			return true
		}
	}

	return false
}

// ReadFilesBytes reads the contents of multiple files and returns them as byte slices.
func ReadFilesBytes(filePaths []string) ([][]byte, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files specified")
	}

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return nil, fmt.Errorf("empty file path %d of %d in list", i+1, len(filePaths))
		}
	}

	filesContents := make([][]byte, 0, len(filePaths))

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)

		fileContents, err := ReadFileBytes(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %d of %d (%s): %w", i+1, len(filePaths), filePath, err)
		}

		filesContents = append(filesContents, fileContents)
	}

	return filesContents, nil
}

// ReadFileBytes reads the contents of a single file and returns it as a byte slice.
func ReadFileBytes(filePath string) ([]byte, error) {
	fileData, err := os.ReadFile(filePath) // #nosec G304 -- General purpose file utility function
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return fileData, nil
}

// ReadFilesBytesLimit reads multiple files with size and count limits.
func ReadFilesBytesLimit(filePaths []string, maxFiles, maxBytesPerFile int64) ([][]byte, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files specified")
	} else if len(filePaths) > int(maxFiles) {
		return nil, fmt.Errorf("too many files specified: maximum is %d, got %d", maxFiles, len(filePaths))
	}

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return nil, fmt.Errorf("empty file path %d of %d in list", i+1, len(filePaths))
		}
	}

	filesContents := make([][]byte, 0, len(filePaths))

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)

		fileContents, err := ReadFileBytesLimit(filePath, maxBytesPerFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %d of %d (%s): %w", i+1, len(filePaths), filePath, err)
		}

		filesContents = append(filesContents, fileContents)
	}

	return filesContents, nil
}

// ReadFileBytesLimit reads a single file with a size limit.
func ReadFileBytesLimit(filePath string, maxBytes int64) ([]byte, error) {
	return readFileBytesLimit(filePath, maxBytes, os.Open, (*os.File).Stat, (*os.File).Read, (*os.File).Close)
}

func readFileBytesLimit(
	filePath string,
	maxBytes int64,
	openFileFn func(string) (*os.File, error),
	statFn func(*os.File) (os.FileInfo, error),
	readFn func(*os.File, []byte) (int, error),
	closeFn func(*os.File) error,
) ([]byte, error) {
	if maxBytes <= 0 {
		// If no limit or negative value, read the entire file
		return ReadFileBytes(filePath)
	}

	// Open file
	file, err := openFileFn(filePath) // #nosec G304 -- General purpose file utility function
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func() {
		if closeErr := closeFn(file); closeErr != nil {
			// Log error but don't override the main return error
			fmt.Printf("Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	// Get file info to determine file size
	fileInfo, err := statFn(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats for %s: %w", filePath, err)
	}

	// Determine how many bytes to read
	bytesToRead := maxBytes
	if fileInfo.Size() < maxBytes {
		bytesToRead = fileInfo.Size()
	}

	// Read the limited bytes
	buffer := make([]byte, bytesToRead)

	n, err := readFn(file, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes from file %s: %w", filePath, err)
	}

	return buffer[:n], nil
}
