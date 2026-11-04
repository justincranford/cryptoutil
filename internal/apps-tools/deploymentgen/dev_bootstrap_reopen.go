// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const developmentRecoveryReadLimit = 1 << 20

// OpenDevelopmentBootstrapDirectory reopens only an incomplete bootstrap:
// exactly the canonical topology and original join token, with no kubeconfig,
// material, or execution output. This is not Kubernetes/resource adoption.
// The caller must separately admit every retained Docker volume and prove
// absence of application deployment after obtaining live Kubernetes access.
func OpenDevelopmentBootstrapDirectory(base string, compose []byte) (directory *DevelopmentPrivateDirectory, token []byte, result error) {
	if !filepath.IsAbs(base) || filepath.Clean(base) != base || len(compose) == 0 {
		return nil, nil, errors.New("explicit protected bootstrap recovery input required")
	}

	target := filepath.Join(base, "phase2-runtime-p2sm1103")
	for current := target; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
			return nil, nil, errors.New("protected bootstrap recovery path rejected")
		}

		if current == filepath.Dir(current) {
			break
		}
	}

	root, err := os.OpenRoot(target)
	if err != nil {
		return nil, nil, errors.New("protected bootstrap recovery root unavailable")
	}

	defer func() {
		if result != nil {
			_ = root.Close()
		}
	}()

	source := protectedLifecycleFS{root: root}
	expected := map[string]bool{
		lifecycleRootPath: true, "compose.yml": false, "secrets": true,
		"secrets/p2sm1103": true, "secrets/p2sm1103/k3s-join-token": false,
	}

	seen := 0
	if err := fs.WalkDir(root.FS(), lifecycleRootPath, func(name string, entry fs.DirEntry, walkErr error) error {
		directory, ok := expected[name]
		if walkErr != nil || !ok || entry.IsDir() != directory {
			return errors.New("bootstrap recovery requires exact incomplete protected output")
		}

		if directory {
			if err := source.checkDirectory(name); err != nil {
				return errors.New("bootstrap recovery directory ownership rejected")
			}
		}

		seen++

		return nil
	}); err != nil || seen != len(expected) {
		return nil, nil, errors.New("bootstrap recovery protected inventory rejected")
	}

	read := func(name string) ([]byte, error) {
		file, err := source.Open(name)
		if err != nil {
			return nil, errors.New("bootstrap recovery file ownership rejected")
		}

		raw, readErr := io.ReadAll(io.LimitReader(file, developmentRecoveryReadLimit+1))

		closeErr := file.Close()
		if readErr != nil || closeErr != nil || len(raw) > developmentRecoveryReadLimit {
			return nil, errors.New("bootstrap recovery file read rejected")
		}

		return raw, nil
	}

	actual, err := read("compose.yml")
	if err != nil || !bytes.Equal(actual, compose) {
		return nil, nil, errors.New("bootstrap recovery canonical topology mismatch")
	}

	token, err = read("secrets/p2sm1103/k3s-join-token")
	if err != nil || len(token) != 64 {
		return nil, nil, errors.New("bootstrap recovery original token rejected")
	}

	decoded, err := hex.DecodeString(string(token))
	if err != nil || hex.EncodeToString(decoded) != string(token) {
		return nil, nil, errors.New("bootstrap recovery original token encoding rejected")
	}

	return &DevelopmentPrivateDirectory{root: root}, token, nil
}
