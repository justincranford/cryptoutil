// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentPrivateProtectionAndLeaseFaultsNeverReturnWriter(t *testing.T) {
	t.Parallel()

	for mode := range 2 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			protect := secureDevelopmentDirectory
			open := os.OpenRoot

			if mode == 0 {
				protect = func(string) error { return errors.New("PRIVATE") }
			} else {
				open = func(string) (*os.Root, error) { return nil, errors.New("PRIVATE") }
			}

			directory, err := createDevelopmentPrivateDirectoryWith(t.TempDir(), protect, open)
			require.Error(t, err)
			require.Nil(t, directory)
			require.NotContains(t, err.Error(), "PRIVATE")
		})
	}
}

func TestDevelopmentProtectedFileFaultsNeverReportPublication(t *testing.T) {
	t.Parallel()

	for mode := range 2 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()
			directory, err := CreateDevelopmentPrivateDirectory(t.TempDir())

			require.NoError(t, err)
			defer func() { require.NoError(t, directory.Close()) }()

			require.NoError(t, directory.Write("owned-test-input", []byte("test-only")))
			file, err := directory.root.OpenFile("owned-test-input", os.O_RDONLY, 0)
			require.NoError(t, err)

			if mode == 1 {
				require.NoError(t, file.Close())
			}

			err = writeDevelopmentProtectedFile("owned-test-input", []byte("PRIVATE"), func(string) (*os.File, error) { return file, nil })
			require.Error(t, err)
			require.NotContains(t, err.Error(), "PRIVATE")
			actual, err := directory.root.ReadFile("owned-test-input")
			require.NoError(t, err)
			require.Equal(t, []byte("test-only"), actual)
		})
	}
}
