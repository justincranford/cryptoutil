// Copyright (c) 2025-2026 Justin Cranford.
package lint_catalog_files

import (
	"fmt"
	"io"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

const testMismatch = "mismatch"

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		exitCode int
		stdout   string
		stderr   string
		wantErr  string
	}{
		{name: "success logs output", stdout: "checked"},
		{name: "failure includes stderr", exitCode: 1, stderr: testMismatch, wantErr: testMismatch},
		{name: "failure without stderr", exitCode: 1, wantErr: "catalog file violations found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err := check(logger, func(stdout, stderr io.Writer) int {
				_, writeErr := fmt.Fprint(stdout, tc.stdout)
				require.NoError(t, writeErr)
				_, writeErr = fmt.Fprint(stderr, tc.stderr)
				require.NoError(t, writeErr)

				return tc.exitCode
			})

			if tc.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestCheckRunsCatalogFilesCommand(t *testing.T) {
	t.Parallel()

	require.NoError(t, Check(cryptoutilCmdCicdCommon.NewLogger("test")))
}
