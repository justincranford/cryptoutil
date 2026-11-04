// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testLifecycleNestedMaterial = "nested/material"

func TestLifecyclePhysicalWindowsPathAndSizeBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, binding, want string
		edit                func(*testing.T, string)
	}{
		{"nested protected source", testLifecycleNestedMaterial, "", func(t *testing.T, root string) {
			require.NoError(t, os.Mkdir(filepath.Join(root, "nested"), 0o700))
			protectLifecycleTestPath(t, filepath.Join(root, "nested"), "")
			require.NoError(t, os.WriteFile(filepath.Join(root, "nested", "material"), []byte(testLifecyclePhysicalMaterial), 0o600))
			protectLifecycleTestPath(t, filepath.Join(root, "nested", "material"), "")
		}},
		{"nested unprotected directory", testLifecycleNestedMaterial, testLifecycleOutsideOwnerError, func(t *testing.T, root string) {
			require.NoError(t, os.Mkdir(filepath.Join(root, "nested"), 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(root, "nested", "material"), []byte(testLifecyclePhysicalMaterial), 0o600))
			protectLifecycleTestPath(t, filepath.Join(root, "nested"), "(A;OICI;FR;;;WD)")
			protectLifecycleTestPath(t, filepath.Join(root, "nested", "material"), "")
		}},
		{"junction cannot escape", "alias/material", "symlink", func(t *testing.T, root string) {
			target := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(target, "material"), []byte(testLifecyclePhysicalMaterial), 0o600))
			// Junction creation needs no symlink privilege or developer-mode change.
			output, err := exec.Command("cmd", "/c", "mklink", "/J", filepath.Join(root, "alias"), target).CombinedOutput()
			require.NoError(t, err, string(output))
		}},
		{"directory instead of file", testLifecycleMaterialFile, "regular", func(t *testing.T, root string) {
			require.NoError(t, os.Mkdir(filepath.Join(root, "material"), 0o700))
			protectLifecycleTestPath(t, filepath.Join(root, "material"), "")
		}},
		{"oversized file", testLifecycleMaterialFile, "at most 1 MiB", func(t *testing.T, root string) {
			require.NoError(t, os.WriteFile(filepath.Join(root, "material"), []byte(strings.Repeat("a", lifecycleMaxSecretBytes+1)), 0o600))
			protectLifecycleTestPath(t, filepath.Join(root, "material"), "")
		}},
		{"aggregate Secret limit", testLifecycleMaterialFile, "exceeds 1 MiB", func(t *testing.T, root string) {
			require.NoError(t, os.WriteFile(filepath.Join(root, "material"), []byte(strings.Repeat("a", lifecycleMaxSecretBytes/2)), 0o600))
			protectLifecycleTestPath(t, filepath.Join(root, "material"), "")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			root := t.TempDir()
			protectLifecycleTestPath(t, root, "")

			for index := range input.Secrets {
				for key := range input.Secrets[index].Files {
					input.Secrets[index].Files[key] = test.binding
				}
			}

			var err error

			model.Revision, err = LifecycleRevision(input)
			require.NoError(t, err)
			inventory, err := build(model, RegisteredScopes(), testServiceSMKMSMemSelector)
			require.NoError(t, err)
			test.edit(t, root)

			secrets, err := ImportPhysicalLifecycleSecrets(model, inventory, input, root)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, secrets)
				require.NotContains(t, err.Error(), testLifecyclePhysicalMaterial)

				return
			}

			require.NoError(t, err)
			require.Len(t, secrets, len(input.Secrets))
		})
	}
}

func TestLifecyclePhysicalRejectsRootJunctionAndRegularFile(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	directory := t.TempDir()
	target := t.TempDir()
	alias := filepath.Join(directory, "alias")
	output, err := exec.Command("cmd", "/c", "mklink", "/J", alias, target).CombinedOutput()
	require.NoError(t, err, string(output))
	_, err = ImportPhysicalLifecycleSecrets(model, inventory, input, alias)
	require.ErrorContains(t, err, "symlink or non-directory")

	file := filepath.Join(directory, testLifecycleMaterialFile)
	require.NoError(t, os.WriteFile(file, []byte(testLifecyclePhysicalMaterial), 0o600))
	_, err = ImportPhysicalLifecycleSecrets(model, inventory, input, file)
	require.ErrorContains(t, err, "symlink or non-directory")
}
