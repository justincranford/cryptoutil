// Copyright (c) 2025-2026 Justin Cranford.
//
//

package server_test

import (
	"bytes"
	"testing"

	cryptoutilIdentityRP "cryptoutil/internal/apps/identity-rp"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	testify "github.com/stretchr/testify/require"
)

func TestRp_HelpFlag(t *testing.T) {
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

			exitCode := cryptoutilIdentityRP.Rp(tc.args, nil, &stdout, &stderr)
			testify.Equal(t, 0, exitCode)
		})
	}
}

func TestRp_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentityRP.Rp([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	testify.Equal(t, 1, exitCode)
}

func TestRpClient_HelpFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentityRP.Rp([]string{"client", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	testify.Equal(t, 0, exitCode)
}

func TestRpClient_NotImplemented(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentityRP.Rp([]string{"client"}, nil, &stdout, &stderr)
	testify.Equal(t, 1, exitCode)
	testify.Contains(t, stderr.String(), "not yet implemented")
}

func TestRpServiceInit_HelpFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilIdentityRP.Rp([]string{"init", cryptoutilSharedMagic.CLIHelpFlag}, nil, &stdout, &stderr)
	testify.Equal(t, 0, exitCode)
}
