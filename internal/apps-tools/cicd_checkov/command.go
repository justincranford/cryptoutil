// Copyright (c) 2025-2026 Justin Cranford.
package cicd_checkov

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

//go:embed checkov_bootstrap.py
var pinnedCheckovBootstrap string

// pinnedCommandContext runs the scanner normally after a fail-closed, process-local
// backport of its known Terraform exclusion typo. UTF-8 matches repository text.
func pinnedCommandContext(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	python, arguments, err := pinnedCheckovCommandArguments(runtime.GOOS, executable, args)
	if err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, python, arguments...), nil
}

func pinnedCheckovCommandArguments(goos, executable string, args []string) (string, []string, error) {
	pythonName := cryptoutilSharedMagic.CICDCheckovPythonUnix
	if goos == cryptoutilSharedMagic.OSNameWindows {
		pythonName = cryptoutilSharedMagic.CICDCheckovPythonWindows
	}

	python := filepath.Join(filepath.Dir(executable), pythonName)

	info, err := os.Stat(python)
	if err != nil {
		return "", nil, fmt.Errorf("inspect pinned Checkov Python runtime %s: %w", python, err)
	}

	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("pinned Checkov Python runtime is not a regular file: %s", python)
	}

	return python, append([]string{"-X", "utf8", "-c", pinnedCheckovBootstrap}, args...), nil
}

// CommandContext invokes the installed Python entry point directly on Windows.
// Checkov's batch launcher forwards %*, which makes cmd.exe interpret regex
// metacharacters instead of passing them as arguments to the scanner.
func CommandContext(ctx context.Context, executable string, args ...string) (*exec.Cmd, error) {
	path, arguments, err := checkovCommandArguments(runtime.GOOS, executable, args)
	if err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, path, arguments...), nil
}

func checkovCommandArguments(goos, executable string, args []string) (string, []string, error) {
	if goos != cryptoutilSharedMagic.OSNameWindows || !strings.EqualFold(filepath.Ext(executable), ".cmd") {
		return executable, append([]string(nil), args...), nil
	}

	directory := filepath.Dir(executable)
	python := filepath.Join(directory, cryptoutilSharedMagic.CICDCheckovPythonWindows)

	script := strings.TrimSuffix(executable, filepath.Ext(executable))
	for _, path := range []string{python, script} {
		info, err := os.Stat(path)
		if err != nil {
			return "", nil, fmt.Errorf("inspect Checkov Python entry point %s: %w", path, err)
		}

		if !info.Mode().IsRegular() {
			return "", nil, fmt.Errorf("checkov Python entry point is not a regular file: %s", path)
		}
	}

	return python, append([]string{script}, args...), nil
}
