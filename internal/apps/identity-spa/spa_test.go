// Copyright (c) 2025 Justin Cranford
//
//

package spa

import (
	"bytes"
	"testing"

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

			exitCode := Spa(tc.args, nil, &stdout, &stderr)
			testify.Equal(t, 0, exitCode)
		})
	}
}

func TestSpa_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Spa([]string{"unknown-subcommand"}, nil, &stdout, &stderr)
	testify.Equal(t, 1, exitCode)
}

func TestSpaClient_HelpFlag(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	exitCode := spaClient([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stderr)
	testify.Equal(t, 0, exitCode)
}

func TestSpaClient_NotImplemented(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	exitCode := spaClient([]string{}, nil, &stderr)
	testify.Equal(t, 1, exitCode)
	testify.Contains(t, stderr.String(), "not yet implemented")
}

func TestSpaServiceInit_HelpFlag(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	exitCode := spaServiceInit([]string{cryptoutilSharedMagic.CLIHelpFlag}, nil, &stderr)
	testify.Equal(t, 0, exitCode)
}
