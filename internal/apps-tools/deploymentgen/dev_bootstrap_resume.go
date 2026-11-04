// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"io"
	"net"
	"path"
	"reflect"
	"strings"
	"time"
)

const developmentEngineFormat = "{{.ID}}"

const developmentLocalVolumeDriver = cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault

// DevelopmentRetainedVolumes freezes an explicit same-engine volume identity
// snapshot. It is not an adoption grant. The dev-smoke CLI requires its separate
// protected incomplete-bootstrap lease before obtaining or using this snapshot.
// Existing compute, incomplete volume sets, and renamed run/project resources
// are always rejected. No resource is filtered out of fresh admission.
type DevelopmentRetainedVolumes struct {
	Engine  string
	Volumes []DevelopmentRetainedVolume
}

type DevelopmentRetainedVolume struct {
	Name, Driver, Scope, Mountpoint, CreatedAt string
	Labels                                     map[string]string
	Options                                    map[string]string
}

func CaptureDevelopmentRetainedVolumes(ctx context.Context, plan LocalK3sPlan, run DockerCommandRunner) (DevelopmentRetainedVolumes, error) {
	result := DevelopmentRetainedVolumes{}
	if ctx == nil || run == nil || plan.Run != DevelopmentSmokeRun || ctx.Err() != nil {
		return result, errors.New("explicit owned development bootstrap recovery required")
	}

	if _, bounded := ctx.Deadline(); !bounded {
		return result, errors.New("bootstrap recovery requires a context deadline")
	}

	raw, err := run(ctx, "docker", dockerInfoCommand, "--format", developmentEngineFormat)
	if err != nil || len(bytes.TrimSpace(raw)) == 0 {
		return result, errors.New("bootstrap recovery Docker identity unavailable")
	}

	result.Engine = strings.TrimSpace(string(raw))
	for _, node := range append([]LocalK3sNode{plan.Server}, plan.Agents...) {
		raw, err := run(ctx, "docker", bootstrapVolumeKind, "inspect", "--format", bootstrapJSONFormat, node.Volume)
		if err != nil {
			return DevelopmentRetainedVolumes{}, errors.New("bootstrap recovery volume unavailable")
		}

		var volume DevelopmentRetainedVolume
		if err := json.Unmarshal(raw, &volume); err != nil || !validDevelopmentRetainedVolume(volume, node, plan) {
			return DevelopmentRetainedVolumes{}, errors.New("bootstrap recovery volume identity rejected")
		}

		result.Volumes = append(result.Volumes, volume)
	}

	return result, nil
}

func validDevelopmentRetainedVolume(volume DevelopmentRetainedVolume, node LocalK3sNode, plan LocalK3sPlan) bool {
	_, err := time.Parse(time.RFC3339Nano, volume.CreatedAt)

	return err == nil && volume.Name == node.Volume && volume.Driver == developmentLocalVolumeDriver && volume.Scope == developmentLocalVolumeDriver &&
		len(volume.Options) == 0 && path.IsAbs(volume.Mountpoint) &&
		path.Base(volume.Mountpoint) == "_data" && path.Base(path.Dir(volume.Mountpoint)) == node.Volume &&
		volume.Labels[cryptoutilSharedMagic.KubernetesLabelRun] == plan.Run && volume.Labels["cu.io/node-id"] == node.Container &&
		volume.Labels["com.docker.compose.project"] == plan.Project &&
		strings.TrimSpace(volume.Labels["com.docker.compose.config-hash"]) != ""
}

func CheckDevelopmentBootstrapResume(ctx context.Context, input BootstrapPreflightInput, snapshot DevelopmentRetainedVolumes,
	run DockerCommandRunner, listen func(string, string) (net.Listener, error),
) (CapacityEstimate, error) {
	return checkLocalK3sPreflightWithOwnership(ctx, input, run, listen,
		func(ctx context.Context, plan LocalK3sPlan, run DockerCommandRunner) error {
			if plan.Run != DevelopmentSmokeRun {
				return errors.New("bootstrap recovery is restricted to the explicit disposable run")
			}

			current, err := CaptureDevelopmentRetainedVolumes(ctx, plan, run)
			if err != nil || !reflect.DeepEqual(snapshot, current) {
				return errors.New("bootstrap recovery retained identity changed")
			}

			expected := map[string]bool{}
			for _, volume := range current.Volumes {
				if expected[volume.Name] {
					return errors.New("bootstrap recovery duplicate volume binding")
				}

				expected[volume.Name] = true
			}

			seen := map[string]bool{}

			for _, kind := range []string{bootstrapContainerKind, bootstrapNetworkKind, bootstrapVolumeKind} {
				args := []string{kind, bootstrapListCommand}
				if kind == bootstrapContainerKind {
					args = append(args, bootstrapAllContainersFlag)
				}

				args = append(args, "--format", bootstrapJSONFormat)

				raw, err := run(ctx, "docker", args...)
				if err != nil {
					return errors.New("bootstrap recovery complete ownership inventory unavailable")
				}

				decoder := json.NewDecoder(bytes.NewReader(raw))

				for {
					var resource struct{ Name, Names, Labels string }
					if err := decoder.Decode(&resource); err != nil {
						if errors.Is(err, io.EOF) {
							break
						}

						return errors.New("bootstrap recovery complete ownership inventory rejected")
					}

					name := resource.Name
					if kind == bootstrapContainerKind {
						name = resource.Names
					}

					if strings.TrimSpace(name) == "" {
						return errors.New("bootstrap recovery unnamed resource rejected")
					}

					owned := false
					for _, label := range strings.Split(resource.Labels, ",") {
						owned = owned || label == "com.docker.compose.project="+plan.Project || label == "cu.io/run="+plan.Run
					}

					reserved := name == plan.Network || expected[name]
					for _, node := range append([]LocalK3sNode{plan.Server}, plan.Agents...) {
						reserved = reserved || name == node.Container
					}

					if !owned && !reserved {
						continue
					}

					if kind != bootstrapVolumeKind || !expected[name] || !owned || seen[name] {
						return errors.New("bootstrap recovery compute, renamed, foreign or duplicate collision")
					}

					seen[name] = true
				}
			}

			if len(seen) != len(expected) {
				return errors.New("bootstrap recovery incomplete retained inventory")
			}

			return nil
		})
}
