// Copyright (c) 2025-2026 Justin Cranford.
// Package cicd_deployment exposes the deployment command entry point.
package cicd_deployment

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	cryptoutilDeploymentGen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Main runs the deployment command with process file I/O.
func Main(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return mainWithLifecycle(args, stdout, stderr, lifecycleExecuteMain)
}

func mainWithLifecycle(args []string, stdout, stderr io.Writer, lifecycle func([]string, io.Writer, io.Writer) int) int {
	if len(args) > 0 && args[0] == devSmokeCommand {
		return developmentSmokeMain(args[1:], stdout, stderr)
	}

	if len(args) > 0 && args[0] == pkiPublishCommand {
		return pkiPublishMain(args[1:], stdout, stderr)
	}

	if len(args) > 0 && args[0] == lifecycleExecuteCommand {
		return lifecycle(args[1:], stdout, stderr)
	}

	return cryptoutilDeploymentGen.Main(args, stdout, stderr, os.ReadFile, writeDeploymentFile)
}

func writeDeploymentFile(path string, data []byte, mode os.FileMode) error {
	return writeDeploymentFileWithIO(path, data, mode, os.MkdirAll, os.WriteFile)
}

func writeDeploymentFileWithIO(
	path string,
	data []byte,
	mode os.FileMode,
	mkdirAll func(string, os.FileMode) error,
	writeFile func(string, []byte, os.FileMode) error,
) error {
	if err := mkdirAll(filepath.Dir(path), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute); err != nil {
		return fmt.Errorf("create generated output directory: %w", err)
	}

	if err := writeFile(path, data, mode); err != nil {
		return fmt.Errorf("write generated output: %w", err)
	}

	return nil
}
