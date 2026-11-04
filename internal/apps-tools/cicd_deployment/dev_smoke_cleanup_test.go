// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	json "encoding/json"
	"errors"
	"path/filepath"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type developmentCleanupRunner struct {
	t          *testing.T
	cluster    gen.LocalK3sPlan
	failure    string
	deletes    int
	deletedIDs []string
}

const (
	testDevNetworkLabels  = "cleanup-network-labels"
	testDevNetworkRemoval = "cleanup-network-removal"
)

func (r *developmentCleanupRunner) Run(_ context.Context, _ string, args []string, _ []byte) ([]byte, error) {
	require.NotContains(r.t, args, "--volumes")
	require.NotContains(r.t, args, "--force")
	require.NotContains(r.t, args, "prune")

	if args[1] == developmentRemove || args[1] == testDevStop {
		require.Contains(r.t, args[len(args)-1], "owned-id-")

		if r.failure == args[1] || (r.failure == testDevNetworkRemoval && args[0] == developmentNetwork) {
			return nil, errors.New("PRIVATE")
		}

		r.deletes++
		r.deletedIDs = append(r.deletedIDs, args[len(args)-1])

		return nil, nil
	}

	if r.failure == developmentInspect {
		return nil, errors.New("PRIVATE")
	}

	if args[0] == developmentNetwork {
		network := map[string]any{
			testDevID:         testDevCleanupNetworkUID,
			developmentLabels: map[string]string{testPKIRunLabel: r.cluster.Run, "cu.io/role": "k3s-node-network"}, testDevContainersField: map[string]any{},
		}
		if r.failure == testDevOccupiedNetwork {
			network[testDevContainersField] = map[string]any{"unowned-container": true}
		}

		if r.failure == testDevNetworkLabels {
			network[developmentLabels] = map[string]string{testPKIRunLabel: "different-owner-run"}
		}

		return json.Marshal(network)
	}

	var selected gen.LocalK3sNode

	for _, node := range append([]gen.LocalK3sNode{r.cluster.Server}, r.cluster.Agents...) {
		if node.Container == args[len(args)-1] {
			selected = node
		}
	}

	require.NotEmpty(r.t, selected.Container)

	object := map[string]any{
		testDevID: "owned-id-" + selected.Container, developmentImage: r.cluster.Image,
		developmentLabels: map[string]string{testPKIRunLabel: r.cluster.Run, "cu.io/node-id": selected.Container, "com.docker.compose.project": r.cluster.Project},
		"mounts":          []any{map[string]string{testDevTypeField: developmentVolume, cryptoutilSharedMagic.ClaimName: selected.Volume, "destination": "/var/lib/rancher/k3s"}},
	}
	switch r.failure {
	case testDevMissingID:
		object[testDevID] = ""
	case testDevWrongImage:
		object[developmentImage] = "shared-image"
	case testDevWrongRun:
		object[developmentLabels] = map[string]string{testPKIRunLabel: "unowned-other-run"}
	case testDevMissingVolume:
		object["mounts"] = []any{}
	}

	return json.Marshal(object)
}

func TestDevelopmentCleanupOnlyDeletesVerifiedOwnedCompute(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{"", developmentInspect, testDevMissingID, testDevWrongImage, testDevWrongRun, testDevMissingVolume, testDevStop, developmentRemove, testDevOccupiedNetwork, testDevNetworkLabels, testDevNetworkRemoval} {
		t.Run("cleanup="+failure, func(t *testing.T) {
			t.Parallel()

			cluster := gen.LocalK3sPlan{
				Run: gen.DevelopmentSmokeRun, Project: testDevCleanupProject, Network: testDevCleanupNetwork, Image: testDevCleanupImage,
				Server: gen.LocalK3sNode{Container: testDevCleanupServer, Volume: testDevCleanupServerVolume},
				Agents: []gen.LocalK3sNode{{Container: testDevCleanupAgent, Volume: testDevCleanupAgentVolume}},
			}
			runner := &developmentCleanupRunner{t: t, cluster: cluster, failure: failure}

			err := cleanupDevelopmentNodes(t.Context(), filepath.Join(t.TempDir(), "docker.exe"), cluster, runner)
			if failure == "" {
				require.NoError(t, err)
				require.Equal(t, 5, runner.deletes)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")
			}
		})
	}
}
