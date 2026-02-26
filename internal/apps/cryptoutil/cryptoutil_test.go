// Copyright (c) 2025 Justin Cranford
//
//

package cryptoutil

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuite_NoArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Suite([]string{}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stderr.String()
	require.Contains(t, output, "Usage: cryptoutil")
	require.Contains(t, output, "Available products:")
}

func TestSuite_OneArgument(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Suite([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stderr.String()
	require.Contains(t, output, "Usage: cryptoutil")
}

func TestSuite_HelpCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "help command", args: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "help"}},
		{name: "help flag long", args: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "--help"}},
		{name: "help flag short", args: []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Suite(tt.args, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			output := stderr.String()
			require.Contains(t, output, "Usage: cryptoutil")
			require.Contains(t, output, "Available products:")
			require.Contains(t, output, cryptoutilSharedMagic.IdentityProductName)
			require.Contains(t, output, cryptoutilSharedMagic.JoseProductName)
			require.Contains(t, output, cryptoutilSharedMagic.PKIProductName)
			require.Contains(t, output, "sm")
		})
	}
}

func TestSuite_UnknownProduct(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Suite([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "nonexistent"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode)

	output := stderr.String()
	require.Contains(t, output, "Unknown product: nonexistent")
	require.Contains(t, output, "Usage: cryptoutil")
}

func TestSuite_ProductRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		product     string
		expectedStr string
	}{
		{name: "identity help", product: cryptoutilSharedMagic.IdentityProductName, expectedStr: "Usage: identity"},
		{name: "jose help", product: cryptoutilSharedMagic.JoseProductName, expectedStr: "Usage: jose"},
		{name: "pki help", product: cryptoutilSharedMagic.PKIProductName, expectedStr: "Usage: pki"},
		{name: "sm help", product: "sm", expectedStr: "Usage: sm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			// Route to product with help flag to verify routing works.
			exitCode := Suite([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, tt.product, "help"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, tt.expectedStr)
		})
	}
}

func TestSuite_ProductVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		product     string
		expectedStr string
	}{
		{name: "identity version", product: cryptoutilSharedMagic.IdentityProductName, expectedStr: "identity product"},
		{name: "jose version", product: cryptoutilSharedMagic.JoseProductName, expectedStr: "jose product"},
		{name: "pki version", product: cryptoutilSharedMagic.PKIProductName, expectedStr: "pki product"},
		{name: "sm version", product: "sm", expectedStr: "sm product"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Suite([]string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, tt.product, "version"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode)

			combinedOutput := stdout.String() + stderr.String()
			require.Contains(t, combinedOutput, tt.expectedStr)
		})
	}
}

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	require.NotPanics(t, func() {
		printUsage(&stderr)
	})

	output := stderr.String()
	require.Contains(t, output, "Usage: cryptoutil")
	require.Contains(t, output, "Available products:")
}
