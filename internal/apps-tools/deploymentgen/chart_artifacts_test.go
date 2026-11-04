// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestCanonicalChartArtifactsCreateDeterministicMirror(t *testing.T) {
	t.Parallel()

	source := os.DirFS(filepath.Join("..", "..", ".."))
	first, err := canonicalChartArtifacts(source)
	require.NoError(t, err)
	require.Len(t, first, 15)

	second, err := canonicalChartArtifacts(source)
	require.NoError(t, err)
	require.Equal(t, first, second)

	foundRuntimeQuotaTemplate := false

	for _, artifact := range first {
		require.True(t, strings.HasPrefix(artifact.path, "deployments/k8s/charts/"))
		require.NotContains(t, artifact.path, `\`)
		require.NotContains(t, artifact.path, "fixtures")
		require.NotEmpty(t, artifact.data)

		if artifact.path == "deployments/k8s/charts/cryptoutil-runtime/templates/quota.yaml" {
			foundRuntimeQuotaTemplate = true
		}
	}

	require.True(t, foundRuntimeQuotaTemplate)
}

func TestCanonicalChartArtifactsRejectMissingSource(t *testing.T) {
	t.Parallel()

	_, err := canonicalChartArtifacts(os.DirFS(t.TempDir()))
	require.ErrorContains(t, err, "walk canonical chart source")
}

func TestValidateCanonicalChartMirror(t *testing.T) {
	t.Parallel()

	artifacts := []generatedArtifact{
		{path: "deployments/k8s/charts/cryptoutil-infra/Chart.yaml"},
		{path: "deployments/k8s/charts/cryptoutil-runtime/Chart.yaml"},
	}
	validFiles := fstest.MapFS{
		artifacts[0].path: &fstest.MapFile{Data: []byte("name: cryptoutil-infra\n")},
		artifacts[1].path: &fstest.MapFile{Data: []byte("name: cryptoutil-runtime\n")},
	}

	tests := []struct {
		name      string
		files     fstest.MapFS
		wantError string
	}{
		{
			name:  "exact mirror",
			files: validFiles,
		},
		{
			name: "missing expected chart",
			files: fstest.MapFS{
				artifacts[0].path: &fstest.MapFile{Data: []byte("name: cryptoutil-infra\n")},
			},
			wantError: "missing expected files",
		},
		{
			name: "unexpected extra file",
			files: fstest.MapFS{
				artifacts[0].path: &fstest.MapFile{Data: []byte("name: cryptoutil-infra\n")},
				artifacts[1].path: &fstest.MapFile{Data: []byte("name: cryptoutil-runtime\n")},
				"deployments/k8s/charts/cryptoutil-runtime/templates/stale.yaml": &fstest.MapFile{Data: []byte("stale\n")},
			},
			wantError: unexpectedChartMirrorFileMessage,
		},
		{
			name: "symlink is rejected",
			files: fstest.MapFS{
				artifacts[0].path: &fstest.MapFile{Data: []byte(cryptoutilSharedMagic.CICDExcludeDirTarget), Mode: fs.ModeSymlink},
				artifacts[1].path: &fstest.MapFile{Data: []byte("name: cryptoutil-runtime\n")},
			},
			wantError: "only regular files",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateCanonicalChartMirror(test.files, artifacts)
			if test.wantError != "" {
				require.ErrorContains(t, err, test.wantError)

				return
			}

			require.NoError(t, err)
		})
	}
}
