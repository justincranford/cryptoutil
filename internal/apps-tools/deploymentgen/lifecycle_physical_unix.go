// Copyright (c) 2025-2026 Justin Cranford.

//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package deploymentgen

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func validatePhysicalLifecycleHandle(_ *os.File, info fs.FileInfo, directory bool) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return fmt.Errorf("protected source must be owned by the current process user")
	}

	if info.Mode().Perm()&0o077 != 0 || info.Mode()&(fs.ModeSetuid|fs.ModeSetgid|fs.ModeSticky) != 0 {
		return fmt.Errorf("protected source requires owner-only permissions without special mode bits")
	}

	if !directory && stat.Nlink != 1 {
		return fmt.Errorf("protected source must not be a hard-linked file")
	}

	return nil
}
