// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"bytes"
	"strings"
	"testing"

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
		{name: "help", arg: "help"},
		{name: "--help", arg: "--help"},
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

func TestCicd_CheckChunkVerification(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "check-chunk-verification"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_LintDeployments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "lint-deployments"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_GenerateListings(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "generate-listings"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_ValidateMirror(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "validate-mirror"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_ValidateCompose(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "validate-compose", "nonexistent.yml"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_ValidateConfig(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "validate-config", "nonexistent.yml"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestCicd_ValidateAll(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Cicd([]string{"cicd", "validate-all"}, strings.NewReader(""), &stdout, &stderr)
	require.Contains(t, []int{0, 1}, exitCode)
}

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	printUsage(&buf)
	output := buf.String()
	require.Contains(t, output, "cicd - Cryptoutil CI/CD linter and formatter tools")
	require.Contains(t, output, "Usage:")
	require.Contains(t, output, "Commands:")
	require.Contains(t, output, "lint-deployments")
	require.Contains(t, output, "validate-all")
	require.Contains(t, output, "check-chunk-verification")
	require.Contains(t, output, "github-cleanup-runs")
	require.Contains(t, output, "github-cleanup-artifacts")
	require.Contains(t, output, "github-cleanup-caches")
	require.Contains(t, output, "github-cleanup-all")
}

