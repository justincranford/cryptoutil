// Copyright (c) 2025-2026 Justin Cranford.
//
//

package server_test

import (
	"bytes"
	"testing"

	cryptoutilIdentitySPA "cryptoutil/internal/apps/identity-spa"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	testify "github.com/stretchr/testify/require"
)

func TestSpa_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "help flag", args: []string{cryptoutilSharedMagic.CLIHelpFlag}},
		{name: "h flag", args: []string{"-h"}},
		{name: "help command", args: []string{cryptoutilSharedMagic.CLIHelpCommand}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilIdentitySPA.Spa(tc.args, nil, &stdout, &stderr)
			testify.Equal(t, 0, exitCode)
		})
	}
}

func TestSpa_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentitySPA.Spa([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	testify.Equal(t, 1, exitCode)
}

func TestSpaClient_HelpFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentitySPA.Spa([]string{"client", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	testify.Equal(t, 0, exitCode)
}

func TestSpaClient_NotImplemented(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentitySPA.Spa([]string{"client"}, nil, &stdout, &stderr)
	testify.Equal(t, 1, exitCode)
	testify.Contains(t, stderr.String(), "not yet implemented")
}

func TestSpaServiceInit_HelpFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentitySPA.Spa([]string{"init", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	testify.Equal(t, 0, exitCode)
}
