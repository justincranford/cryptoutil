// Copyright (c) 2025-2026 Justin Cranford.
package dockerfile_healthcheck

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestKubernetesSuiteImageRequiresHelmProbes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		valid   bool
	}{
		{name: "Helm selects route", content: "FROM scratch\n", valid: true},
		{name: "invalid default route", content: "FROM scratch\nHEALTHCHECK CMD [\"/app/cryptoutil\", \"livez\"]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			directory := filepath.Join(root, "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment)
			require.NoError(t, os.MkdirAll(directory, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(filepath.Join(directory, "Dockerfile"), []byte(tt.content), cryptoutilSharedMagic.FilePermissionsDefault))

			violations, err := FindViolationsInDir(root)
			require.NoError(t, err)
			require.Equal(t, tt.valid, len(violations) == 0, "%v", violations)
		})
	}
}
