// Copyright (c) 2025-2026 Justin Cranford.
// Package testcli provides shared test helpers for PS-ID service CLI entry point tests.
// All 10 PS-ID service packages call RunCLITests to reduce their *_test.go files
// to a single function that delegates the standard three-case suite to this helper.
package testcli

import (
	"bytes"
	"io"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// EntryFunc is the signature of a PS-ID service entry function.
type EntryFunc func(args []string, stdin io.Reader, stdout, stderr io.Writer) int

// RunCLITests runs the standard three-case CLI test suite for a PS-ID entry function:
//
//   - help flag returns exit code 0
//   - version subcommand returns exit code 0
//   - unknown subcommand returns exit code 1
func RunCLITests(t *testing.T, entryFn EntryFunc) {
	t.Helper()

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{name: "help_flag", args: []string{cryptoutilSharedMagic.CLIHelpFlag}, wantCode: 0},
		{name: "version_cmd", args: []string{cryptoutilSharedMagic.CLIVersionCommand}, wantCode: 0},
		{name: "unknown_subcommand", args: []string{"unknown-subcommand"}, wantCode: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := entryFn(tt.args, nil, &stdout, &stderr)
			require.Equal(t, tt.wantCode, exitCode)
		})
	}
}
