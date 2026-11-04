// Copyright (c) 2025-2026 Justin Cranford.
package cicd_checkov

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestCheckovCommandArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		goos            string
		batch           bool
		missing         bool
		scriptDirectory bool
	}{
		{name: "Windows batch entry point", goos: cryptoutilSharedMagic.OSNameWindows, batch: true},
		{name: "Windows native executable", goos: cryptoutilSharedMagic.OSNameWindows},
		{name: "Unix executable", goos: cryptoutilSharedMagic.MockRuntimeGoOS},
		{name: "missing Python runtime", goos: cryptoutilSharedMagic.OSNameWindows, batch: true, missing: true},
		{name: "script must be a regular file", goos: cryptoutilSharedMagic.OSNameWindows, batch: true, scriptDirectory: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()

			executable := filepath.Join(root, "checkov")

			if tc.batch {
				if tc.scriptDirectory {
					require.NoError(t, os.Mkdir(executable, 0o700))
				} else {
					require.NoError(t, os.WriteFile(executable, []byte("# Python entry point\n"), 0o600))
				}

				executable += ".cmd"

				if !tc.missing {
					require.NoError(t, os.WriteFile(filepath.Join(root, cryptoutilSharedMagic.CICDCheckovPythonWindows), nil, 0o600))
				}
			}

			args := []string{`--skip-path=^(?:\.[\\/])?nested(?:[\\/]|$)`, "--output", "json"}

			path, got, err := checkovCommandArguments(tc.goos, executable, args)
			if tc.missing {
				require.ErrorContains(t, err, "inspect Checkov Python entry point")

				return
			}

			if tc.scriptDirectory {
				require.ErrorContains(t, err, "not a regular file")

				return
			}

			require.NoError(t, err)

			if tc.batch {
				require.Equal(t, filepath.Join(root, cryptoutilSharedMagic.CICDCheckovPythonWindows), path)
				require.Equal(t, append([]string{filepath.Join(root, "checkov")}, args...), got)
			} else {
				require.Equal(t, executable, path)
				require.Equal(t, args, got)
			}
		})
	}
}

func TestCommandContext(t *testing.T) {
	t.Parallel()

	command, err := CommandContext(context.Background(), "checkov", cryptoutilSharedMagic.CLIVersionFlag)
	require.NoError(t, err)
	require.Equal(t, []string{"checkov", cryptoutilSharedMagic.CLIVersionFlag}, command.Args)

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		_, err = CommandContext(context.Background(), filepath.Join(t.TempDir(), "checkov.cmd"))
		require.ErrorContains(t, err, "inspect Checkov Python entry point")
	}
}
