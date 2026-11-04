// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	privateDirectoryMode = 0o750
	privateFileMode      = 0o440
)

type operatingSystemFileSystem struct{}

func (operatingSystemFileSystem) Lstat(name string) (os.FileInfo, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("lstat %q: %w", name, err)
	}

	return info, nil
}

func (operatingSystemFileSystem) Mkdir(name string, mode os.FileMode) error {
	if err := os.Mkdir(name, mode); err != nil {
		return fmt.Errorf("create directory %q: %w", name, err)
	}

	return nil
}

func (operatingSystemFileSystem) MkdirTemp(directory string, pattern string) (string, error) {
	name, err := os.MkdirTemp(directory, pattern)
	if err != nil {
		return "", fmt.Errorf("create temporary directory in %q: %w", directory, err)
	}

	return name, nil
}

func (operatingSystemFileSystem) OpenFile(name string, flag int, mode os.FileMode) (FileHandle, error) {
	file, err := os.OpenFile(name, flag, mode)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", name, err)
	}

	return file, nil
}

func (operatingSystemFileSystem) Chmod(name string, mode os.FileMode) error {
	if err := os.Chmod(name, mode); err != nil {
		return fmt.Errorf("change permissions for %q: %w", name, err)
	}

	return nil
}

func (operatingSystemFileSystem) Rename(oldPath string, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename %q to %q: %w", oldPath, newPath, err)
	}

	return nil
}

func (operatingSystemFileSystem) RemoveAll(name string) error {
	if err := os.RemoveAll(name); err != nil {
		return fmt.Errorf("remove path %q: %w", name, err)
	}

	return nil
}

func osReadFile(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", name, err)
	}

	return data, nil
}

func publishFiles(ctx context.Context, fileSystem FileSystem, outputDirectory string, files map[string]materializedFile) error {
	if err := contextError(ctx); err != nil {
		return fmt.Errorf("materializer canceled before publication: %w", err)
	}

	if err := preflightOutputDirectory(fileSystem, outputDirectory); err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("materializer has no validated output files")
	}

	parentDirectory := filepath.Dir(outputDirectory)

	stageDirectory, err := fileSystem.MkdirTemp(parentDirectory, ".pki-materializer-*")
	if err != nil {
		return fmt.Errorf("create materializer staging directory: %w", err)
	}

	cleanup := func(primary error) error {
		if cleanupErr := fileSystem.RemoveAll(stageDirectory); cleanupErr != nil {
			return errors.Join(primary, fmt.Errorf("remove materializer staging directory: %w", cleanupErr))
		}

		return primary
	}

	if filepath.Dir(stageDirectory) != parentDirectory {
		return fmt.Errorf("materializer staging directory escaped its output parent")
	}

	stageInfo, err := fileSystem.Lstat(stageDirectory)
	if err != nil {
		return cleanup(fmt.Errorf("inspect materializer staging directory: %w", err))
	}

	if stageInfo.Mode()&os.ModeSymlink != 0 || !stageInfo.IsDir() {
		return cleanup(fmt.Errorf("materializer staging path is a symlink or non-directory"))
	}

	if err := fileSystem.Chmod(stageDirectory, privateDirectoryMode); err != nil {
		return cleanup(fmt.Errorf("secure materializer staging directory: %w", err))
	}

	fileNames := make([]string, 0, len(files))
	for relativeName := range files {
		if !safeRelativeOutputPath(relativeName) {
			return cleanup(fmt.Errorf("materializer output file path is invalid"))
		}

		fileNames = append(fileNames, relativeName)
	}

	sort.Strings(fileNames)

	directories := make(map[string]struct{})

	for _, fileName := range fileNames {
		if err := contextError(ctx); err != nil {
			return cleanup(fmt.Errorf("materializer canceled while staging output: %w", err))
		}

		relativeDirectory := path.Dir(fileName)
		if err := createOutputDirectories(fileSystem, stageDirectory, relativeDirectory, directories); err != nil {
			return cleanup(err)
		}

		if err := writePrivateFile(fileSystem, filepath.Join(stageDirectory, filepath.FromSlash(fileName)), files[fileName].data); err != nil {
			return cleanup(err)
		}
	}

	if err := contextError(ctx); err != nil {
		return cleanup(fmt.Errorf("materializer canceled before atomic publication: %w", err))
	}

	if _, err := fileSystem.Lstat(outputDirectory); err == nil {
		return cleanup(fmt.Errorf("materializer output directory was reused during publication"))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return cleanup(fmt.Errorf("recheck materializer output directory: %w", err))
	}

	if err := fileSystem.Rename(stageDirectory, outputDirectory); err != nil {
		return cleanup(fmt.Errorf("atomically publish materialized PKI files: %w", err))
	}

	return nil
}

func preflightOutputDirectory(fileSystem FileSystem, outputDirectory string) error {
	if !filepath.IsAbs(outputDirectory) || filepath.Clean(outputDirectory) != outputDirectory {
		return fmt.Errorf("materializer output directory is not a clean absolute path")
	}

	parentDirectory := filepath.Dir(outputDirectory)
	if err := validateDirectoryChain(fileSystem, parentDirectory); err != nil {
		return err
	}

	if _, err := fileSystem.Lstat(outputDirectory); err == nil {
		return fmt.Errorf("materializer output directory already exists")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("inspect materializer output directory: %w", err)
	}

	return nil
}

func validateDirectoryChain(fileSystem FileSystem, directory string) error {
	return validateDirectoryChainWithAbs(fileSystem, directory, filepath.Abs)
}

func validateDirectoryChainWithAbs(
	fileSystem FileSystem,
	directory string,
	resolveAbsPath func(string) (string, error),
) error {
	absoluteDirectory, err := resolveAbsPath(directory)
	if err != nil {
		return fmt.Errorf("resolve materializer output parent: %w", err)
	}

	cleanDirectory := filepath.Clean(absoluteDirectory)
	volume := filepath.VolumeName(cleanDirectory)
	root := volume + string(filepath.Separator)
	relative := strings.TrimPrefix(cleanDirectory, root)
	current := root

	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" {
			continue
		}

		current = filepath.Join(current, component)

		info, statErr := fileSystem.Lstat(current)
		if statErr != nil {
			return fmt.Errorf("inspect materializer output parent: %w", statErr)
		}

		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("materializer output parent contains a symlink or non-directory")
		}
	}

	return nil
}

func createOutputDirectories(fileSystem FileSystem, stageDirectory string, relativeDirectory string, created map[string]struct{}) error {
	if relativeDirectory == cryptoutilSharedMagic.CurrentDirectory {
		return nil
	}

	current := stageDirectory
	for _, component := range strings.Split(relativeDirectory, "/") {
		current = filepath.Join(current, component)
		if _, exists := created[current]; exists {
			continue
		}

		if err := fileSystem.Mkdir(current, privateDirectoryMode); err != nil {
			return fmt.Errorf("create materializer role directory: %w", err)
		}

		if err := fileSystem.Chmod(current, privateDirectoryMode); err != nil {
			return fmt.Errorf("secure materializer role directory: %w", err)
		}

		created[current] = struct{}{}
	}

	return nil
}

func writePrivateFile(fileSystem FileSystem, name string, data []byte) error {
	file, err := fileSystem.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, privateFileMode)
	if err != nil {
		return fmt.Errorf("create materializer output file: %w", err)
	}

	written, writeErr := file.Write(data)
	if writeErr == nil && written != len(data) {
		writeErr = io.ErrShortWrite
	}

	if writeErr == nil {
		writeErr = file.Sync()
	}

	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.Join(
			wrapOptionalError("write materializer output file", writeErr),
			wrapOptionalError("close materializer output file", closeErr),
		)
	}

	if err := fileSystem.Chmod(name, privateFileMode); err != nil {
		return fmt.Errorf("secure materializer output file: %w", err)
	}

	return nil
}

func wrapOptionalError(message string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", message, err)
}

func safeRelativeOutputPath(name string) bool {
	if name == "" || path.IsAbs(name) || path.Clean(name) != name || strings.Contains(name, `\`) {
		return false
	}

	for _, component := range strings.Split(name, "/") {
		if component == "" || component == cryptoutilSharedMagic.CurrentDirectory || component == cryptoutilSharedMagic.DeploymentMaterializerPathParent {
			return false
		}
	}

	return true
}
