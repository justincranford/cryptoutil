// Copyright (c) 2025-2026 Justin Cranford.

package cryptoutil

import (
	"bytes"
	"testing"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestSuiteDatabaseLifecycleRoute(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		args []string
		code int
	}{
		{"suite database help", []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, db.Command, cryptoutilSharedMagic.CLIHelpFlag}, 0},
		{"explicit input required", []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, db.Command}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			require.Equal(t, test.code, Suite(test.args, nil, &stdout, &stderr))

			if test.code == 0 {
				require.Contains(t, stdout.String(), db.Command)
			}
		})
	}
}
