// Copyright (c) 2025-2026 Justin Cranford.
package cicd_checkov

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// InternalMain runs the repository-safe scanner and preserves scanner exit codes.
func InternalMain(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return internalMainWithRunner(args, stdout, stderr, func(ctx context.Context, runner Runner) error {
		return runner.Run(ctx)
	})
}

func internalMainWithRunner(args []string, stdout, stderr io.Writer, run func(context.Context, Runner) error) int {
	runner := Runner{Root: cryptoutilSharedMagic.CurrentDirectory, Args: args, Stdout: stdout, Stderr: stderr}
	if err := run(context.Background(), runner); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}

		_, _ = fmt.Fprintf(stderr, "cicd-checkov: %v\n", err)

		return 1
	}

	return 0
}
