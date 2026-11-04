// Copyright (c) 2025-2026 Justin Cranford.

package cryptoutil

import (
	"bytes"
	"testing"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestSuiteMaterializerExecutable(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		args []string
		code int
	}{
		{cryptoutilSharedMagic.CLIHelpCommand, []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, pki.Command, "-help"}, 0},
		{"explicit paths required", []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, pki.Command}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			require.Equal(t, test.code, Suite(test.args, nil, &stdout, &stderr))

			if test.code == 0 {
				require.Contains(t, stdout.String(), pki.Command)
				require.Contains(t, stdout.String(), pki.InputPath)
			}
		})
	}
}
