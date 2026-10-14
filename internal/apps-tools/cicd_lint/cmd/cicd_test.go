// Copyright (c) 2025-2026 Justin Cranford.
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

	exitCode := Cicd([]string{testCicd}, strings.NewReader(""), &stdout, &stderr)
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
		{name: valueShortHelpFlag, arg: valueShortHelpFlag},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Cicd([]string{testCicd, tc.arg}, strings.NewReader(""), &stdout, &stderr)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout.String(), "Usage:")
		})
	}
}

func TestCicd_UnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{testCicd, "nonexistent-cmd"}, strings.NewReader(""), &stdout, &stderr)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown command: nonexistent-cmd")
}

func TestCicd_LintDocs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{testCicd, "lint-docs"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_LintDeployments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{testCicd, "lint-deployments"}, strings.NewReader(""), &stdout, &stderr)
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
		{name: "only flags", args: []string{testQ, cryptoutilSharedMagic.FlagSummary}, want: ""},
		{name: "command first", args: []string{testLintText}, want: testLintText},
		{name: "flag then command", args: []string{testQ, testLintText}, want: testLintText},
		{name: "multiple flags then command", args: []string{testQ, cryptoutilSharedMagic.FlagSummary, testLintGo}, want: testLintGo},
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
		{name: "no flags", args: []string{testLintText}, want: false},
		{name: "short -h", args: []string{valueShortHelpFlag}, want: true},
		{name: "--help flag", args: []string{cryptoutilSharedMagic.CLIHelpFlag}, want: true},
		{name: "help command", args: []string{cryptoutilSharedMagic.CLIHelpCommand}, want: true},
		{name: "help mixed with other args", args: []string{testLintText, cryptoutilSharedMagic.CLIHelpFlag}, want: true},
		{name: cryptoutilSharedMagic.EMPTY, args: []string{}, want: false},
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
	exitCode := Cicd([]string{testCicd, testQ, "lint-workflow"}, strings.NewReader(""), &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Expected exit code 0 for -q lint-workflow")
}
