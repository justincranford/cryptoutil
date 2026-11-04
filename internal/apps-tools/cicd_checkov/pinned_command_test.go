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

func TestPinnedCommandContext(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	scanner := filepath.Join(directory, "checkov")
	_, err := pinnedCommandContext(t.Context(), scanner)
	require.Error(t, err)

	pythonName := cryptoutilSharedMagic.CICDCheckovPythonUnix
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		pythonName = cryptoutilSharedMagic.CICDCheckovPythonWindows
	}

	require.NoError(t, os.WriteFile(filepath.Join(directory, pythonName), nil, cryptoutilSharedMagic.FilePermissionsDefault))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	command, err := pinnedCommandContext(ctx, scanner, cryptoutilSharedMagic.CLIVersionFlag)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.CLIVersionFlag, command.Args[len(command.Args)-1])
	require.ErrorIs(t, command.Run(), context.Canceled)
}

func TestPinnedCheckovCommandPreservesArgumentsAndUsesUTF8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		goos   string
		python string
	}{
		{name: "Windows", goos: cryptoutilSharedMagic.OSNameWindows, python: cryptoutilSharedMagic.CICDCheckovPythonWindows},
		{name: "Unix", goos: cryptoutilSharedMagic.MockRuntimeGoOS, python: cryptoutilSharedMagic.CICDCheckovPythonUnix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			directory := t.TempDir()
			python := filepath.Join(directory, tt.python)
			require.NoError(t, os.WriteFile(python, nil, cryptoutilSharedMagic.FilePermissionsDefault))

			args := []string{"--skip-path", "(^|/)ignored(/|$)", testCommandDirectoryArgument, testCurrentDirectory}
			executable, actual, err := pinnedCheckovCommandArguments(tt.goos, filepath.Join(directory, "checkov"), args)
			require.NoError(t, err)
			require.Equal(t, python, executable)
			require.Equal(t, []string{"-X", "utf8", "-c"}, actual[:3])
			require.Equal(t, pinnedCheckovBootstrap, actual[3])
			require.Equal(t, args, actual[4:])
			require.Equal(t, "--skip-path", args[0])
		})
	}
}

func TestPinnedCheckovCommandFailsClosedWithoutRegularRuntime(t *testing.T) {
	t.Parallel()

	for _, directoryRuntime := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "directory"}[directoryRuntime], func(t *testing.T) {
			t.Parallel()

			directory := t.TempDir()
			if directoryRuntime {
				require.NoError(t, os.Mkdir(filepath.Join(directory, cryptoutilSharedMagic.DevSetupToolTypePython), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			}

			_, _, err := pinnedCheckovCommandArguments(cryptoutilSharedMagic.MockRuntimeGoOS, filepath.Join(directory, "checkov"), nil)
			require.Error(t, err)
		})
	}
}
