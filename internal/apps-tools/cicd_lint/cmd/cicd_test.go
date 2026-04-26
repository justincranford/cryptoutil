// Copyright (c) 2025 Justin Cranford

package cmd

import (
	"bytes"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCicd_NoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd"}, strings.NewReader(""), &stdout, &stderr)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "Usage:")
}

func TestCicd_Help(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: cryptoutilSharedMagic.CLIHelpCommand, arg: cryptoutilSharedMagic.CLIHelpCommand},
		{name: cryptoutilSharedMagic.CLIHelpFlag, arg: cryptoutilSharedMagic.CLIHelpFlag},
		{name: "-h", arg: "-h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Cicd([]string{"cicd", tc.arg}, strings.NewReader(""), &stdout, &stderr)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout.String(), "Usage:")
		})
	}
}

func TestCicd_UnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "nonexistent-cmd"}, strings.NewReader(""), &stdout, &stderr)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown command: nonexistent-cmd")
}

func TestCicd_LintDocs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "lint-docs"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_LintDeployments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "lint-deployments"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	printUsage(&buf)
	output := buf.String()
	require.Contains(t, output, "cicd-lint - Cryptoutil CI/CD linter and formatter tools")
	require.Contains(t, output, "Usage:")
	require.Contains(t, output, "Commands:")
	require.Contains(t, output, "lint-deployments")
	require.Contains(t, output, "lint-docs")
	require.Contains(t, output, "github-cleanup")
}

func TestFirstNonFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "no args", args: []string{}, want: ""},
		{name: "only flags", args: []string{"-q", cryptoutilSharedMagic.FlagSummary}, want: ""},
		{name: "command first", args: []string{"lint-text"}, want: "lint-text"},
		{name: "flag then command", args: []string{"-q", "lint-text"}, want: "lint-text"},
		{name: "multiple flags then command", args: []string{"-q", cryptoutilSharedMagic.FlagSummary, "lint-go"}, want: "lint-go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, firstNonFlag(tc.args))
		})
	}
}

func TestHasHelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no flags", args: []string{"lint-text"}, want: false},
		{name: "short -h", args: []string{"-h"}, want: true},
		{name: "--help flag", args: []string{cryptoutilSharedMagic.CLIHelpFlag}, want: true},
		{name: "help command", args: []string{cryptoutilSharedMagic.CLIHelpCommand}, want: true},
		{name: "help mixed with other args", args: []string{"lint-text", cryptoutilSharedMagic.CLIHelpFlag}, want: true},
		{name: "empty", args: []string{}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, hasHelpFlag(tc.args))
		})
	}
}

func TestCicd_QuietFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	// lint-workflow is fast and always passes in the project root.
	exitCode := Cicd([]string{"cicd", "-q", "lint-workflow"}, strings.NewReader(""), &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Expected exit code 0 for -q lint-workflow")
}
