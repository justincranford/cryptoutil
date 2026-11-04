// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"fmt"
	"net"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testResumeChangedSnapshot = "changed-snapshot"
	testResumeCapacity        = "capacity"
	testResumeEngine          = "engine"
	testResumeDuplicateVolume = "duplicate-volume"
	testResumeVolumeRead      = "volume-read"
	testResumeVolumeJSON      = "volume-json"
	testResumeVolumeForeign   = "volume-foreign"
	testResumeListRead        = "list-read"
	testResumeListJSON        = "list-json"
	testResumeListUnnamed     = "list-unnamed"
	testResumeMissingVolume   = "missing-volume"
	testResumeRenamedVolume   = "renamed-volume"
	testResumeForeignVolume   = "foreign-volume"
	testResumeExistingNode    = "existing-container"
	testResumeExistingNetwork = "existing-network"
)

func developmentResumeFixture(t *testing.T) (BootstrapPreflightInput, LocalK3sPlan, DevelopmentRetainedVolumes) {
	t.Helper()

	input, _ := bootstrapPreflightFixture(t, "s:sm-kms:pg")
	raw, err := json.Marshal(input)
	require.NoError(t, err)

	raw = []byte(strings.ReplaceAll(string(raw), input.Model.Run, DevelopmentSmokeRun))
	require.NoError(t, json.Unmarshal(raw, &input))

	input.Model.Revision, err = LifecycleRevision(*input.Lifecycle)
	require.NoError(t, err)
	input.Inventory, err = Build(input.Model, "s:sm-kms:pg")
	require.NoError(t, err)
	input.RuntimeValues, input.InfrastructureValues = capacityProjections(t, input.Model, RegisteredScopes(), input.Inventory)

	plan, err := BuildLocalK3sPlan(input.Model, input.Lock)
	require.NoError(t, err)

	snapshot := DevelopmentRetainedVolumes{Engine: "explicit-owned-engine"}
	for _, node := range append([]LocalK3sNode{plan.Server}, plan.Agents...) {
		snapshot.Volumes = append(snapshot.Volumes, DevelopmentRetainedVolume{
			Name: node.Volume, Driver: developmentLocalVolumeDriver, Scope: developmentLocalVolumeDriver,
			Mountpoint: path.Join("/var/lib/docker/volumes", node.Volume, "_data"), CreatedAt: "2026-11-03T10:00:00Z",
			Labels: map[string]string{
				cryptoutilSharedMagic.KubernetesLabelRun: plan.Run, "cu.io/node-id": node.Container,
				"com.docker.compose.project": plan.Project, "com.docker.compose.config-hash": strings.Repeat("a", 64),
			},
		})
	}

	return input, plan, snapshot
}

func TestDevelopmentBootstrapResumeAdmission(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{
		"", testResumeEngine, testResumeVolumeRead, testResumeVolumeJSON, testResumeVolumeForeign, testResumeChangedSnapshot, testResumeListRead,
		testResumeListJSON, testResumeListUnnamed, testResumeMissingVolume, testResumeDuplicateVolume, testResumeRenamedVolume, testResumeForeignVolume, testResumeExistingNode, testResumeExistingNetwork, testResumeCapacity,
	} {
		t.Run("recovery="+failure, func(t *testing.T) {
			t.Parallel()

			input, plan, snapshot := developmentResumeFixture(t)

			expected := snapshot
			if failure == testResumeChangedSnapshot {
				expected.Engine = "replaced-engine"
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			run := func(_ context.Context, binary string, args ...string) ([]byte, error) {
				require.Equal(t, "docker", binary)
				require.NotContains(t, args, "up")
				require.NotContains(t, args, "rm")

				if args[0] == dockerInfoCommand {
					if args[len(args)-1] == developmentEngineFormat {
						if failure == testResumeEngine {
							return nil, errors.New("PRIVATE")
						}

						return []byte(snapshot.Engine), nil
					}

					if failure == testResumeCapacity {
						return []byte("1 1024"), nil
					}

					return []byte("14 8326418432"), nil
				}

				if args[1] == "inspect" {
					if failure == testResumeVolumeRead {
						return nil, errors.New("PRIVATE")
					}

					if failure == testResumeVolumeJSON {
						return []byte("{"), nil
					}

					for _, volume := range snapshot.Volumes {
						if volume.Name == args[len(args)-1] {
							if failure == testResumeVolumeForeign {
								volume.Driver = "unowned-driver"
							}

							return json.Marshal(volume)
						}
					}

					require.FailNow(t, "unexpected volume")
				}

				if failure == testResumeListRead {
					return nil, errors.New("PRIVATE")
				}

				if failure == testResumeListJSON {
					return []byte("{"), nil
				}

				if failure == testResumeListUnnamed {
					return []byte("{}"), nil
				}

				if args[0] != cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume {
					if (failure == testResumeExistingNode && args[0] == bootstrapContainerKind) || (failure == testResumeExistingNetwork && args[0] == bootstrapNetworkKind) {
						return json.Marshal(map[string]string{testBootstrapNameKey: plan.Network, testBootstrapNamesKey: plan.Server.Container, testBootstrapLabelsKey: "cu.io/run=" + plan.Run})
					}

					return nil, nil
				}

				var records []string

				for index, volume := range snapshot.Volumes {
					if failure == testResumeMissingVolume && index == 0 {
						continue
					}

					name := volume.Name
					if failure == testResumeRenamedVolume && index == 0 {
						name += "-renamed"
					}

					label := "cu.io/run=" + plan.Run
					if failure == testResumeForeignVolume && index == 0 {
						label = "cu.io/run=foreign"
					}

					raw, err := json.Marshal(map[string]string{testBootstrapNameKey: name, testBootstrapLabelsKey: label})
					require.NoError(t, err)

					records = append(records, string(raw))
					if failure == testResumeDuplicateVolume && index == 0 {
						records = append(records, string(raw))
					}
				}

				return []byte(strings.Join(records, "\n")), nil
			}
			listen := func(_, _ string) (net.Listener, error) {
				return &bootstrapTestListener{close: func() error { return nil }}, nil
			}

			_, err := CheckDevelopmentBootstrapResume(ctx, input, expected, run, listen)
			if failure == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err, fmt.Sprintf("negative admission %s must fail", failure))
				require.NotContains(t, err.Error(), "PRIVATE")
			}
		})
	}
}
