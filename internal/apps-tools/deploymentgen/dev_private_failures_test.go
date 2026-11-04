// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentPrivateOutputRejectsNonDirectoriesAndBlockedParents(t *testing.T) {
	t.Parallel()

	for mode := range 3 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			base := t.TempDir()
			if mode == 0 {
				base = filepath.Join(base, "not-directory")
				require.NoError(t, os.WriteFile(base, []byte("test-only"), 0o600))
				_, err := CreateDevelopmentPrivateDirectory(base)
				require.Error(t, err)

				return
			}

			directory, err := CreateDevelopmentPrivateDirectory(base)
			require.NoError(t, err)

			if mode == 1 {
				require.NoError(t, directory.Write("blocked", []byte("test-only")))
				require.Error(t, directory.Write("blocked/child", []byte("test-only")))
				require.NoError(t, directory.Close())

				return
			}

			require.NoError(t, directory.Close())
			require.NoError(t, directory.Close())
			require.Error(t, directory.Write("after-close", []byte("test-only")))
		})
	}
}
