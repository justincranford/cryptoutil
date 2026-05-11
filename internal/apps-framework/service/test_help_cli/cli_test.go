// Copyright (c) 2025-2026 Justin Cranford.

package test_help_cli

import (
	"io"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestRunCLITests_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entryFn EntryFunc
	}{
		{
			name: "standard entry behavior",
			entryFn: func(args []string, _ io.Reader, _, _ io.Writer) int {
				switch {
				case len(args) == 1 && args[0] == cryptoutilSharedMagic.CLIHelpFlag:
					return 0
				case len(args) == 1 && args[0] == cryptoutilSharedMagic.CLIVersionCommand:
					return 0
				default:
					return 1
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			RunCLITests(t, tc.entryFn)
			require.True(t, true)
		})
	}
}
