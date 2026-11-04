// Copyright (c) 2025-2026 Justin Cranford.
package cicd_checkov

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternalMainWithRunner(t *testing.T) {
	t.Parallel()

	executable, err := os.Executable()
	require.NoError(t, err)

	exitErr := exec.Command(executable, "-cryptoutil-invalid-test-flag").Run()

	var processErr *exec.ExitError
	require.ErrorAs(t, exitErr, &processErr)

	tests := []struct {
		name string
		err  error
		code int
	}{
		{name: "success"},
		{name: "failure", err: errors.New("scanner failure"), code: 1},
		{name: "wrapped scanner exit", err: fmt.Errorf("scan: %w", exitErr), code: processErr.ExitCode()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			args := []string{testCommandDirectoryArgument, testCurrentDirectory}
			code := internalMainWithRunner(args, &stdout, &stderr, func(ctx context.Context, runner Runner) error {
				require.NotNil(t, ctx)
				require.Equal(t, ".", runner.Root)
				require.Equal(t, args, runner.Args)
				require.Same(t, &stdout, runner.Stdout)
				require.Same(t, &stderr, runner.Stderr)

				return tt.err
			})
			require.Equal(t, tt.code, code)

			if tt.code == 1 {
				require.Contains(t, stderr.String(), tt.err.Error())
			} else {
				require.Empty(t, stderr.String())
			}
		})
	}
}

func TestInternalMainRejectsUnsafeArguments(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	require.Equal(t, 1, InternalMain([]string{"--bc-api-key", "forbidden"}, nil, nil, &stderr))
	require.NotEmpty(t, stderr.String())
}
