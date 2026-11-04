// Copyright (c) 2025-2026 Justin Cranford.

//go:build !windows

package deploymentgen

import "os"

func secureDevelopmentDirectory(path string) error {
	return os.Chmod(path, 0o700)
}
