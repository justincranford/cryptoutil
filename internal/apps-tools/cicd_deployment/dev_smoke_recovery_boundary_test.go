// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	json "encoding/json"
	"path/filepath"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"

	"github.com/stretchr/testify/require"
)

const testRecoverySystemPod = "coredns-owned-system-pod"

func TestDevelopmentNativeDiagnosticMissingExecutable(t *testing.T) {
	t.Parallel()

	cluster := gen.LocalK3sPlan{
		Run: gen.DevelopmentSmokeRun, Project: testDevCleanupProject, Image: testDevCleanupImage,
		Server: gen.LocalK3sNode{Container: "native-diagnostic-server", Volume: "native-diagnostic-volume"},
	}
	runner := &developmentDiagnosticRunner{
		owner: &developmentCleanupRunner{t: t, cluster: cluster}, state: testDiagnosticRunning,
	}
	snapshot := developmentComputeSnapshot{Nodes: map[string]string{cluster.Server.Container: "owned-id-" + cluster.Server.Container}}
	require.Equal(t, "bootstrap running; startup log observation unavailable",
		nativeDevelopmentBootstrapObservation(filepath.Join(t.TempDir(), "absent-native-tool"), cluster, runner, snapshot))
	require.Zero(t, runner.owner.deletes)
}

func TestDevelopmentBootstrapInventorySystemWorkloadBoundary(t *testing.T) {
	t.Parallel()

	for _, name := range []string{testRecoverySystemPod, "unowned-system-namespace-app"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			objects := make([]map[string]any, 0, developmentSystemNSCount+1)
			for _, namespace := range []string{developmentDefaultContext, developmentKubeSystem, developmentKubePublic, developmentKubeNodeLease} {
				objects = append(objects, map[string]any{developmentKind: developmentNamespaceKind, developmentMetadata: map[string]string{developmentName: namespace}})
			}

			objects = append(objects, map[string]any{
				developmentKind:     developmentPodKind,
				developmentMetadata: map[string]string{developmentName: name, testImageNamespace: developmentKubeSystem},
			})
			raw, err := json.Marshal(map[string]any{developmentKind: developmentListKind, developmentItems: objects})
			require.NoError(t, err)

			err = validateDevelopmentBootstrapInventory(raw)
			if name == testRecoverySystemPod {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "non-baseline")
			}
		})
	}
}

func TestDevelopmentProtectedWritePreservesRedactedFailure(t *testing.T) {
	t.Parallel()

	private, err := gen.CreateDevelopmentPrivateDirectory(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, writeDevelopmentSmokeProtected(private, "already-published", []byte(testDiagnosticPrivate)))
	require.Error(t, writeDevelopmentSmokeProtected(private, "already-published", []byte("replacement")))
	require.NoError(t, private.Close())
	require.Equal(t, "bootstrap observation unavailable", developmentBootstrapObservation(context.Background(), "unused", gen.LocalK3sPlan{}, nil, developmentComputeSnapshot{}))
}
