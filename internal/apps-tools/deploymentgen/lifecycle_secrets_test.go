// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleOwnerOnlyError = "owner-only"
	testLifecycleSymlinkError   = "symlink"
)

func lifecycleSecretFiles(input LifecycleInput) fstest.MapFS {
	files := fstest.MapFS{}

	for _, secret := range input.Secrets {
		for _, file := range secret.Files {
			files[file] = &fstest.MapFile{Data: []byte("fixture-material\n"), Mode: 0o600}
		}
	}

	return files
}

func TestLifecycleImportProtectedFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(fstest.MapFS, string)
		want string
	}{
		{"owner-only material", func(fstest.MapFS, string) {}, ""},
		{"missing material", func(files fstest.MapFS, file string) { delete(files, file) }, testLifecycleMissingFileError},
		{"group-readable", func(files fstest.MapFS, file string) { files[file].Mode = 0o640 }, testLifecycleOwnerOnlyError},
		{"world-readable", func(files fstest.MapFS, file string) { files[file].Mode = 0o644 }, testLifecycleOwnerOnlyError},
		{"empty material", func(files fstest.MapFS, file string) { files[file].Data = nil }, testNonemptyError},
		{"oversized material", func(files fstest.MapFS, file string) { files[file].Data = []byte(strings.Repeat("x", (1<<20)+1)) }, "1 MiB"},
		{"symlink file", func(files fstest.MapFS, file string) { files[file].Mode = fs.ModeSymlink | 0o600 }, testLifecycleSymlinkError},
		{"symlink parent", func(files fstest.MapFS, file string) {
			files[strings.Split(file, "/")[0]] = &fstest.MapFile{Mode: fs.ModeSymlink}
		}, testLifecycleSymlinkError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			files := lifecycleSecretFiles(input)
			file := input.Secrets[0].Files["value"]
			test.edit(files, file)
			secrets, err := ImportLifecycleSecrets(model, inventory, input, files)

			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, secrets, "failed imports must not expose a partial deployment")

				return
			}

			require.NoError(t, err)
			require.Len(t, secrets, len(input.Secrets))

			for _, secret := range secrets {
				require.Equal(t, "v1", secret.APIVersion)
				require.Equal(t, cryptoutilSharedMagic.KubernetesSecretKind, secret.Kind)
				require.Equal(t, []byte("fixture-material\n"), secret.Data["value"])
				require.Equal(t, model.Run, secret.Metadata.Labels[cryptoutilSharedMagic.KubernetesLabelRun])
			}
		})
	}
}

func TestLifecycleImportRejectsInvalidContractBeforeReading(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	input.Jobs = nil
	_, err := ImportLifecycleSecrets(model, inventory, input, nil)
	require.Error(t, err)
}
