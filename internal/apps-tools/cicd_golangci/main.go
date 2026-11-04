// Copyright (c) 2025-2026 Justin Cranford.
// Package cicd_golangci provides the repository-safe golangci command entry point.
package cicd_golangci

import (
	"context"
	"fmt"
	"io"

	cryptoutilAppsToolsCicdLintGolangciRunner "cryptoutil/internal/apps-tools/cicd_lint/golangci_runner"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// InternalMain delegates supported arguments to repository-safe package enumeration.
func InternalMain(args []string, _ io.Reader, _ io.Writer, stderr io.Writer) int {
	return internalMainWithRunner(args, stderr, cryptoutilAppsToolsCicdLintGolangciRunner.Run)
}

func internalMainWithRunner(args []string, stderr io.Writer, run func(context.Context, string, []string) error) int {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if err := run(context.Background(), cryptoutilSharedMagic.CurrentDirectory, args); err != nil {
		_, _ = fmt.Fprintln(stderr, err)

		return 1
	}

	return 0
}
