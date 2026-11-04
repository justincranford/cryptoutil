// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testReopenRelative        = "relative-base"
	testReopenMissingRoot     = "missing-root"
	testReopenEmptyTopology   = "empty-topology"
	testReopenMissingTopology = "missing-topology"
	testReopenMissingToken    = "missing-token"
	testReopenOversized       = "oversized-file"
)

func TestDevelopmentBootstrapReopenInvalidRoots(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{testReopenRelative, testReopenMissingRoot, testReopenEmptyTopology, testReopenMissingTopology, testReopenMissingToken} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()

			base := t.TempDir()
			if failure != testReopenMissingRoot && failure != testReopenRelative {
				private, err := CreateDevelopmentPrivateDirectory(base)
				require.NoError(t, err)

				if failure != testReopenMissingTopology {
					require.NoError(t, private.Write(cryptoutilSharedMagic.COMPOSE_YML, []byte("topology")))
				}

				if failure != testReopenMissingToken {
					require.NoError(t, private.Write("secrets/p2sm1103/k3s-join-token", []byte(strings.Repeat("a", 64))))
				}

				require.NoError(t, private.Close())
			}

			expected := []byte("topology")
			if failure == testReopenEmptyTopology {
				expected = nil
			}

			if failure == testReopenRelative {
				base = "relative-recovery-base"
			}

			directory, token, err := OpenDevelopmentBootstrapDirectory(base, expected)
			require.Error(t, err)
			require.Nil(t, directory)
			require.Nil(t, token)
		})
	}
}

func TestDevelopmentBootstrapReopenFileSizeAndType(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{testReopenOversized, "directory-instead-of-file"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()

			base := t.TempDir()
			private, err := CreateDevelopmentPrivateDirectory(base)
			require.NoError(t, err)
			require.NoError(t, private.Write(cryptoutilSharedMagic.COMPOSE_YML, []byte("topology")))
			require.NoError(t, private.Write("secrets/p2sm1103/k3s-join-token", []byte(strings.Repeat("a", 64))))
			require.NoError(t, private.Close())

			target := filepath.Join(base, "phase2-runtime-p2sm1103", cryptoutilSharedMagic.COMPOSE_YML)
			if failure == testReopenOversized {
				require.NoError(t, os.WriteFile(target, []byte(strings.Repeat("x", developmentRecoveryReadLimit+1)), 0o600))
			} else {
				require.NoError(t, os.Remove(target))
				require.NoError(t, os.Mkdir(target, 0o700))
			}

			directory, token, err := OpenDevelopmentBootstrapDirectory(base, []byte("topology"))
			require.Error(t, err)
			require.Nil(t, directory)
			require.Nil(t, token)
		})
	}
}
