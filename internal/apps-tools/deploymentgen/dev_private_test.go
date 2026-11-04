// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentPrivateOutputNeverAdoptsOrOverwrites(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testDevProtectedRoot, testDevRelativeBase, testDevAbsentBase, testDevExistingOutput} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			base := t.TempDir()

			switch test {
			case testDevRelativeBase:
				base = lifecycleRootPath
			case testDevAbsentBase:
				base = filepath.Join(base, "missing")
			case testDevExistingOutput:
				require.NoError(t, os.Mkdir(filepath.Join(base, "phase2-runtime-p2sm1103"), 0o700))
			}

			directory, err := CreateDevelopmentPrivateDirectory(base)
			if test != testDevProtectedRoot {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NoError(t, directory.Write("secrets/p2sm1103/k3s-join-token", []byte("test-only")))
			require.Error(t, directory.Write("secrets/p2sm1103/k3s-join-token", []byte("overwrite")))
			require.Error(t, directory.Write("../escape", []byte("escape")))
			require.Error(t, directory.Write(cryptoutilSharedMagic.EMPTY, nil))
			require.NoError(t, directory.Close())
		})
	}
}
