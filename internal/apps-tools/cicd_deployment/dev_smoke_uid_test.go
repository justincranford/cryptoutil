// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"path/filepath"
	"strconv"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentCleanupRejectsReplacedUIDAndPreservesUnrecordedCompute(t *testing.T) {
	t.Parallel()

	for mode := range testDevThirdCase {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			cluster := gen.LocalK3sPlan{
				Run: gen.DevelopmentSmokeRun, Project: testDevCleanupProject, Network: testDevCleanupNetwork, Image: testDevCleanupImage,
				Server: gen.LocalK3sNode{Container: testDevCleanupServer, Volume: testDevCleanupServerVolume},
				Agents: []gen.LocalK3sNode{{Container: testDevCleanupAgent, Volume: testDevCleanupAgentVolume}},
			}
			runner := &developmentCleanupRunner{t: t, cluster: cluster}
			snapshot, err := captureDevelopmentCompute(t.Context(), filepath.Join(t.TempDir(), "docker.exe"), cluster, runner)
			require.NoError(t, err)

			switch mode {
			case 0:
				snapshot.Nodes[cluster.Server.Container] = "previous-owned-id"
			case 1:
				delete(snapshot.Nodes, cluster.Server.Container)
				snapshot.Complete = false
			case 2:
				snapshot.Network = "previous-owned-network-id"
			}

			err = cleanupDevelopmentCompute(t.Context(), filepath.Join(t.TempDir(), "docker.exe"), cluster, runner, snapshot)
			if mode == 1 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			require.NotContains(t, runner.deletedIDs, testDevCleanupNetworkUID)

			if mode != 2 {
				require.Equal(t, []string{testDevCleanupAgentUID, testDevCleanupAgentUID}, runner.deletedIDs)
			}
		})
	}
}
