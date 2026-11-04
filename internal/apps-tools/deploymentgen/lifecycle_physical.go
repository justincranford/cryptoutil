// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const lifecyclePrivateDirectoryMode = 0o700

// ImportPhysicalLifecycleSecrets reads a caller-owned protected directory through
// an os.Root confinement boundary. It never changes permissions, logs material,
// writes artifacts, or contacts Kubernetes. The owner must not mutate the source
// concurrently; every opened handle is checked before any bytes are read.
func ImportPhysicalLifecycleSecrets(
	model Model, inventory Inventory, input LifecycleInput, directory string,
) (secrets []LifecycleSecret, returnErr error) {
	defer func() {
		if returnErr != nil {
			returnErr = protectedLifecycleImportError{cause: returnErr}
		}
	}()
	// Validate non-secret ownership before touching the physical filesystem.
	if _, err := BuildLifecycleChartValues(model, inventory, input); err != nil {
		return nil, err
	}

	if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory {
		return nil, fmt.Errorf("protected source requires a clean absolute directory")
	}

	// Reject links in the absolute root path as well as beneath the root.
	for current := directory; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			return nil, fmt.Errorf("inspect protected source directory: unavailable")
		}

		if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("protected source directory has a symlink or non-directory component")
		}

		if parent := filepath.Dir(current); parent == current {
			break
		}
	}

	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, fmt.Errorf("open protected source directory: unavailable")
	}

	defer func() {
		if err := root.Close(); err != nil && returnErr == nil {
			secrets = nil
			returnErr = fmt.Errorf("close protected source directory: unavailable")
		}
	}()

	source := protectedLifecycleFS{root: root}
	if err := source.checkDirectory(lifecycleRootPath); err != nil {
		return nil, err
	}

	return ImportLifecycleSecrets(model, inventory, input, source)
}

type protectedLifecycleImportError struct{ cause error }

func (err protectedLifecycleImportError) Error() string { return err.cause.Error() }
func (err protectedLifecycleImportError) Unwrap() error { return err.cause }

type lifecyclePhysicalRoot interface {
	Lstat(string) (fs.FileInfo, error)
	Open(string) (*os.File, error)
	FS() fs.FS
}

type protectedLifecycleFS struct{ root lifecyclePhysicalRoot }

func (source protectedLifecycleFS) checkDirectory(name string) error {
	handle, err := source.openChecked(name, true)
	if err != nil {
		return err
	}

	if err := handle.Close(); err != nil {
		return fmt.Errorf("close protected directory: unavailable")
	}

	return nil
}

func (source protectedLifecycleFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) || strings.ContainsAny(name, `:\`) {
		return nil, fmt.Errorf("protected directory path is invalid")
	}

	if err := source.checkDirectory(name); err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(source.root.FS(), name)
	if err != nil {
		return nil, fmt.Errorf("read protected directory: unavailable")
	}

	return entries, nil
}

func (source protectedLifecycleFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) || name == lifecycleRootPath || strings.ContainsAny(name, `:\`) {
		return nil, fmt.Errorf("protected file path is invalid")
	}

	parent := path.Dir(name)
	for {
		if err := source.checkDirectory(parent); err != nil {
			return nil, err
		}

		if parent == lifecycleRootPath {
			break
		}

		parent = path.Dir(parent)
	}

	return source.openChecked(name, false)
}

func (source protectedLifecycleFS) openChecked(name string, directory bool) (fs.File, error) {
	before, err := source.root.Lstat(name)
	if err != nil || before.Mode()&fs.ModeSymlink != 0 {
		return nil, fmt.Errorf("protected source has a missing or symlink component")
	}

	if (directory && !before.IsDir()) || (!directory && !before.Mode().IsRegular()) {
		return nil, fmt.Errorf("protected source requires an expected regular file or directory type")
	}

	handle, err := source.root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open protected source handle: unavailable")
	}

	info, err := handle.Stat()
	if err != nil {
		_ = handle.Close()

		return nil, fmt.Errorf("inspect protected source handle: unavailable")
	}

	after, err := source.root.Lstat(name)
	if err != nil || after.Mode()&fs.ModeSymlink != 0 || !os.SameFile(before, info) || !os.SameFile(info, after) ||
		(directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		_ = handle.Close()

		return nil, fmt.Errorf("protected source identity or type changed")
	}

	if err := validatePhysicalLifecycleHandle(handle, info, directory); err != nil {
		_ = handle.Close()

		return nil, err
	}

	return protectedLifecycleFile{File: handle}, nil
}

// ACL validation on Windows is authoritative: portable mode bits do not encode
// Windows DACLs. Expose owner-only modes only AFTER checking the opened handle.
type protectedLifecycleFile struct{ *os.File }

func (file protectedLifecycleFile) Stat() (fs.FileInfo, error) {
	info, err := file.File.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect protected file: unavailable")
	}

	return protectedLifecycleInfo{FileInfo: info}, nil
}

type protectedLifecycleInfo struct{ fs.FileInfo }

func (info protectedLifecycleInfo) Mode() fs.FileMode {
	if info.IsDir() {
		return fs.ModeDir | lifecyclePrivateDirectoryMode
	}

	return cryptoutilSharedMagic.FilePermOwnerReadWriteOnly
}
