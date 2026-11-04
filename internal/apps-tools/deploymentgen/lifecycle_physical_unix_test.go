// Copyright (c) 2025-2026 Justin Cranford.

//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package deploymentgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLifecyclePhysicalUnixPermissionAndLinkBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*testing.T, string)
		want string
	}{
		{"protected import", func(_ *testing.T, _ string) {}, ""},
		{"group readable file", func(t *testing.T, root string) {
			require.NoError(t, os.Chmod(filepath.Join(root, "material"), 0o640))
		}, "owner-only"},
		{"world writable directory", func(t *testing.T, root string) {
			require.NoError(t, os.Chmod(root, 0o702))
		}, "owner-only"},
		{"file symlink", func(t *testing.T, root string) {
			require.NoError(t, os.Rename(filepath.Join(root, "material"), filepath.Join(root, "actual")))
			require.NoError(t, os.Symlink("actual", filepath.Join(root, "material")))
		}, "symlink"},
		{"hardlink", func(t *testing.T, root string) {
			require.NoError(t, os.Link(filepath.Join(root, "material"), filepath.Join(root, "alias")))
		}, "hard-linked"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			root := t.TempDir()
			require.NoError(t, os.Chmod(root, 0o700))
			for index := range input.Secrets {
				for key := range input.Secrets[index].Files {
					input.Secrets[index].Files[key] = "material"
				}
			}

			var err error
			model.Revision, err = LifecycleRevision(input)
			require.NoError(t, err)
			inventory, err := build(model, RegisteredScopes(), testServiceSMKMSMemSelector)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(root, "material"), []byte("nonproduction fixture"), 0o600))
			test.edit(t, root)
			secrets, err := ImportPhysicalLifecycleSecrets(model, inventory, input, root)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, secrets)
				return
			}

			require.NoError(t, err)
			require.Len(t, secrets, len(input.Secrets))
		})
	}
}
