// Copyright (c) 2025-2026 Justin Cranford.
package magic

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetProjectRootWithInjectedFilesystemFunctions(t *testing.T) {
	t.Parallel()

	volumeRoot := filepath.VolumeName(os.TempDir()) + string(os.PathSeparator)
	projectRoot := filepath.Join(volumeRoot, "project")
	nestedDirectory := filepath.Join(projectRoot, "internal")
	tests := []struct {
		name    string
		getwdFn func() (string, error)
		statFn  func(string) (os.FileInfo, error)
		want    string
	}{
		{
			name: "get working directory failure",
			getwdFn: func() (string, error) {
				return "", errors.New("injected working directory failure")
			},
			statFn: func(string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			want: CurrentDirectory,
		},
		{
			name: "finds project root",
			getwdFn: func() (string, error) {
				return nestedDirectory, nil
			},
			statFn: func(path string) (os.FileInfo, error) {
				if path == filepath.Join(projectRoot, ".git") {
					return nil, nil
				}

				return nil, os.ErrNotExist
			},
			want: projectRoot,
		},
		{
			name: "falls back at filesystem root",
			getwdFn: func() (string, error) {
				return volumeRoot, nil
			},
			statFn: func(string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			want: CurrentDirectory,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, getProjectRootWithFns(tc.getwdFn, tc.statFn))
		})
	}

	require.NotEmpty(t, getProjectRoot())
}
