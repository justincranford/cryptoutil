// Copyright (c) 2025-2026 Justin Cranford.

package registry

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindRegistryYAMLPathWithDeps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		getwdFn     func() (string, error)
		statFn      func(string) (os.FileInfo, error)
		wantPath    string
		wantErrText string
	}{
		{
			name: "working directory error",
			getwdFn: func() (string, error) {
				return "", errors.New("injected working directory failure")
			},
			statFn: func(string) (os.FileInfo, error) {
				return nil, errors.New("unexpected stat call")
			},
			wantErrText: "getwd: injected working directory failure",
		},
		{
			name: "registry file found",
			getwdFn: func() (string, error) {
				return filepath.Join("workspace", "nested"), nil
			},
			statFn: func(candidate string) (os.FileInfo, error) {
				if candidate == filepath.Join("workspace", "nested", defaultRegistryRelPath) {
					return nil, nil
				}

				return nil, os.ErrNotExist
			},
			wantPath: filepath.Join("workspace", "nested", defaultRegistryRelPath),
		},
		{
			name: "filesystem root reached without registry",
			getwdFn: func() (string, error) {
				return string(filepath.Separator) + filepath.Join("workspace", "missing"), nil
			},
			statFn: func(string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			wantErrText: "registry YAML not found: walked up from working directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path, err := findRegistryYAMLPathWithDeps(test.getwdFn, test.statFn)
			if test.wantErrText != "" {
				require.ErrorContains(t, err, test.wantErrText)
				require.Empty(t, path)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantPath, path)
		})
	}
}
