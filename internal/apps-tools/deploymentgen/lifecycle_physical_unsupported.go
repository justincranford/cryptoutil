// Copyright (c) 2025-2026 Justin Cranford.

//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package deploymentgen

import (
	"fmt"
	"io/fs"
	"os"
)

func validatePhysicalLifecycleHandle(_ *os.File, _ fs.FileInfo, _ bool) error {
	return fmt.Errorf("protected physical import is unsupported on this operating system")
}
