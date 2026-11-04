// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlan(t *testing.T) {
	t.Parallel()

	modelJSON, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, k8sModelFileName))
	require.NoError(t, err)

	tests := []struct {
		name         string
		selector     string
		wantReleases int
		wantError    string
	}{
		{name: "full matrix", wantReleases: 55},
		{name: "selected product", selector: testProductIdentityMixedDiskSelector, wantReleases: 1},
		{name: "unknown scope", selector: testUnknownScopeSelector, wantError: testNoMatchedReleaseError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			planned, err := Plan(modelJSON, test.selector)
			if test.wantError != "" {
				require.ErrorContains(t, err, test.wantError)

				return
			}

			require.NoError(t, err)
			require.NotContains(t, string(planned), "-----BEGIN")
			require.Equal(t, byte('\n'), planned[len(planned)-1])

			repeated, err := Plan(modelJSON, test.selector)
			require.NoError(t, err)
			require.Equal(t, planned, repeated)
		})
	}

	_, err = Plan([]byte("{"), "")
	require.ErrorContains(t, err, testModelContext)
}
