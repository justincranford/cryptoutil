// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DevelopmentPrivateDirectory is a fresh owner-only disposable root lease.
// It never removes or adopts an existing directory.
type DevelopmentPrivateDirectory struct{ root *os.Root }

func CreateDevelopmentPrivateDirectory(base string) (*DevelopmentPrivateDirectory, error) {
	return createDevelopmentPrivateDirectoryWith(base, secureDevelopmentDirectory, os.OpenRoot)
}

func createDevelopmentPrivateDirectoryWith(base string, protect func(string) error, open func(string) (*os.Root, error)) (*DevelopmentPrivateDirectory, error) {
	if !filepath.IsAbs(base) || filepath.Clean(base) != base {
		return nil, errors.New("explicit absolute development output base required")
	}

	for current := base; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("development output base has unsafe components")
		}

		if current == filepath.Dir(current) {
			break
		}
	}

	directory := filepath.Join(base, "phase2-runtime-p2sm1103")
	if err := os.Mkdir(directory, developmentDirectoryMode); err != nil {
		return nil, errors.New("fresh development output root unavailable")
	}

	if err := protect(directory); err != nil {
		return nil, errors.New("development output protection rejected")
	}

	root, err := open(directory)
	if err != nil {
		return nil, errors.New("development output lease unavailable")
	}

	return &DevelopmentPrivateDirectory{root: root}, nil
}

func (directory *DevelopmentPrivateDirectory) Close() error {
	if err := directory.root.Close(); err != nil {
		return fmt.Errorf("development directory closure: %w", err)
	}

	return nil
}

// Write publishes only fresh protected files; repeated writes are collisions.
func (directory *DevelopmentPrivateDirectory) Write(name string, raw []byte) error {
	if !fs.ValidPath(name) || len(raw) == 0 || len(raw) > 1<<20 {
		return errors.New("protected development output rejected")
	}

	parent := filepath.ToSlash(filepath.Dir(name))
	if parent != "." {
		current := ""
		for _, part := range strings.Split(parent, "/") {
			current = filepath.Join(current, part)
			if err := directory.root.Mkdir(current, developmentDirectoryMode); err != nil && !errors.Is(err, fs.ErrExist) {
				return errors.New("protected development parent creation failed")
			}
		}
	}

	return writeDevelopmentProtectedFile(name, raw, func(name string) (*os.File, error) {
		return directory.root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
	})
}

func writeDevelopmentProtectedFile(name string, raw []byte, open func(string) (*os.File, error)) error {
	file, err := open(name)
	if err != nil {
		return errors.New("fresh protected development file unavailable")
	}

	info, statErr := file.Stat()
	if statErr == nil {
		statErr = validatePhysicalLifecycleHandle(file, info, false)
	}

	if statErr != nil {
		_ = file.Close()

		return errors.New("protected development file ownership rejected")
	}

	count, writeErr := file.Write(raw)

	closeErr := file.Close()
	if writeErr != nil || closeErr != nil || count != len(raw) {
		return errors.New("protected development file publication failed")
	}

	return nil
}
