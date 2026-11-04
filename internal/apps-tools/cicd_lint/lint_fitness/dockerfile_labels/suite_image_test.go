// Copyright (c) 2025-2026 Justin Cranford.
package dockerfile_labels

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestKubernetesSuiteImageLabelsAndEntrypoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		title      string
		entrypoint []string
		valid      bool
	}{
		{name: "suite route selected by Helm", title: cryptoutilSharedMagic.DefaultOTLPServiceDefault, entrypoint: []string{cryptoutilSharedMagic.KubernetesSuiteImageBinary}, valid: true},
		{name: "wrong title", title: "wrong", entrypoint: []string{cryptoutilSharedMagic.KubernetesSuiteImageBinary}},
		{name: "hardcoded service route", title: cryptoutilSharedMagic.DefaultOTLPServiceDefault, entrypoint: []string{cryptoutilSharedMagic.KubernetesSuiteImageBinary, cryptoutilSharedMagic.IdentityProductName, cryptoutilSharedMagic.SPAServiceName}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			directory := filepath.Join(root, "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment)
			require.NoError(t, os.MkdirAll(directory, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			content := dockerfileWithLabelsAndEntrypoint(tt.title, "Local Helm suite image", tt.entrypoint)
			require.NoError(t, os.WriteFile(filepath.Join(directory, "Dockerfile"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

			violations, err := FindViolationsInDir(root)
			require.NoError(t, err)
			require.Equal(t, tt.valid, len(violations) == 0, "%v", violations)
		})
	}
}
