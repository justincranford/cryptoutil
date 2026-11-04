// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleMaterialFile     = "material"
	testLifecycleMissingFileError = "missing"
)

func TestLifecyclePhysicalRejectsUnprotectedSource(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	for _, directory := range []string{"relative", filepath.Join(t.TempDir(), "absent")} {
		t.Run(directory, func(t *testing.T) {
			t.Parallel()

			secrets, err := ImportPhysicalLifecycleSecrets(model, inventory, input, directory)
			require.Error(t, err)
			require.Nil(t, secrets)
		})
	}

	directory := t.TempDir()
	root, err := os.OpenRoot(directory)
	require.NoError(t, err)

	defer func() { require.NoError(t, root.Close()) }()

	source := protectedLifecycleFS{root: root}

	for _, name := range []string{".", "../escape", `drive:stream`, `dir\file`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handle, err := source.Open(name)
			require.Error(t, err)
			require.Nil(t, handle)
		})
	}
}
