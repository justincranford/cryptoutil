// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestHelmInfrastructurePostgresDataDirectoryUsesPersistentMount(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	charts := []struct {
		name string
		path string
	}{
		{
			name: "canonical chart",
			path: filepath.Join(
				"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-infra",
			),
		},
		{
			name: "generated mirror",
			path: filepath.Join("..", "..", "..", "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", "cryptoutil-infra"),
		},
	}
	tests := []struct {
		name              string
		leaderMountPath   string
		followerMountPath string
	}{
		{
			name:              "canonical retained mounts",
			leaderMountPath:   defaults.Components.PostgresLeader.Storage.MountPath,
			followerMountPath: defaults.Components.PostgresFollower.Storage.MountPath,
		},
		{
			name:              "operator changed retained mounts",
			leaderMountPath:   "/mnt/postgres-leader-data",
			followerMountPath: "/mnt/postgres-follower-data",
		},
	}

	for chartIndex, chart := range charts {
		t.Run(chart.name, func(t *testing.T) {
			t.Parallel()

			for index, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					t.Parallel()

					values, projectionErr := BuildInfrastructureChartValues(model, inventory, defaults)
					require.NoError(t, projectionErr)

					values.Components.PostgresLeader.Storage.MountPath = test.leaderMountPath
					values.Components.PostgresFollower.Storage.MountPath = test.followerMountPath

					valuesJSON, marshalErr := MarshalInfrastructureChartValues(values)
					require.NoError(t, marshalErr)
					valuesPath := filepath.Join(t.TempDir(), "infrastructure-values.json")
					require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

					outputDirectory := t.TempDir()
					rendered, renderErr := renderHelmDocuments(
						t, helmBinary, chart.path, values.HelmRelease, values.Namespace,
						valuesPath, outputDirectory, chartIndex*len(tests)+index,
					)
					require.NoError(t, renderErr)

					assertPostgresDataDirectory(
						t, rendered, cryptoutilSharedMagic.KMSPostgresLeaderContainer, test.leaderMountPath,
					)
					assertPostgresDataDirectory(
						t, rendered, cryptoutilSharedMagic.PKIInitPostgresFollowerService, test.followerMountPath,
					)
				})
			}
		})
	}
}

func assertPostgresDataDirectory(
	t *testing.T,
	documents []helmRenderedDocument,
	statefulSetName, storageMountPath string,
) {
	t.Helper()

	var statefulSet helmRenderedDocument

	found := false

	for _, document := range documents {
		if document.Kind == testStatefulSetKind && document.Metadata.Name == statefulSetName {
			statefulSet = document
			found = true

			break
		}
	}

	require.True(t, found, "rendered PostgreSQL StatefulSet %q", statefulSetName)
	require.Len(t, statefulSet.Spec.Template.Spec.Containers, 1)

	container := statefulSet.Spec.Template.Spec.Containers[0]
	pgData := ""

	for _, variable := range container.Env {
		if variable.Name == cryptoutilSharedMagic.PostgresDataDirectoryEnvironment {
			pgData = variable.Value

			break
		}
	}

	normalizedMountPath := strings.TrimSuffix(storageMountPath, "/")
	expectedPGData := fmt.Sprintf("%s/pgdata", normalizedMountPath)
	require.Equal(t, expectedPGData, pgData)
	require.NotEqual(t, cryptoutilSharedMagic.Postgres18ImageDefaultDataDirectory, pgData)

	mountFound := false

	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.MountPath == storageMountPath {
			mountFound = true

			break
		}
	}

	require.True(t, mountFound, "persistent volume must be mounted at %q", storageMountPath)
	require.True(t, strings.HasPrefix(pgData, normalizedMountPath+"/"),
		"PostgreSQL data directory must stay under its retained PVC mount")
}
