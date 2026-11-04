// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func deployDevelopmentSmoke(ctx context.Context, input developmentSmokeInput, preflight gen.BootstrapPreflightInput,
	cluster gen.LocalK3sPlan, request execution.Request, config execution.KubectlConfig, runner execution.Runner, stdout io.Writer,
) error {
	return deployDevelopmentSmokeWithAPI(ctx, input, preflight, cluster, request, config, runner, stdout, verifyDevelopmentPublicAPI)
}

func deployDevelopmentSmokeWithAPI(ctx context.Context, input developmentSmokeInput, preflight gen.BootstrapPreflightInput,
	cluster gen.LocalK3sPlan, request execution.Request, config execution.KubectlConfig, runner execution.Runner, stdout io.Writer,
	verifyAPI func(context.Context, gen.BootstrapPreflightInput, []gen.LifecycleSecret, execution.KubectlConfig) error,
) error {
	return deployDevelopmentSmokeWithTransports(ctx, input, preflight, cluster, request, config, runner, stdout, verifyAPI, importDevelopmentImage)
}

func deployDevelopmentSmokeWithTransports(ctx context.Context, input developmentSmokeInput, preflight gen.BootstrapPreflightInput,
	cluster gen.LocalK3sPlan, request execution.Request, config execution.KubectlConfig, runner execution.Runner, stdout io.Writer,
	verifyAPI func(context.Context, gen.BootstrapPreflightInput, []gen.LifecycleSecret, execution.KubectlConfig) error,
	importImage func(context.Context, developmentSmokeInput, gen.LocalK3sPlan, string, execution.Runner) error,
) error {
	docker := func(args []string, stdin []byte) ([]byte, error) {
		raw, err := runner.Run(ctx, input.docker, args, stdin)
		if err != nil {
			return nil, errors.New("owned image/node transport unavailable")
		}

		return raw, nil
	}
	kubectl := func(args []string, stdin []byte) ([]byte, error) {
		args = append([]string{developmentKubeconfigFlag, config.Kubeconfig, developmentContextFlag, config.Context, "--request-timeout=30s"}, args...)

		raw, err := runner.Run(ctx, config.Kubectl, args, stdin)
		if err != nil {
			return nil, errors.New("owned smoke Kubernetes operation rejected")
		}

		return raw, nil
	}

	for _, image := range []string{input.image, preflight.Lock.Images.OTelCollector.Reference, preflight.Lock.Images.LGTM.Reference} {
		raw, err := docker([]string{developmentImage, developmentInspect, developmentFormatFlag, "{{json .RepoDigests}}", image}, nil)
		if err != nil {
			return err
		}

		var digests []string
		if err := json.Unmarshal(raw, &digests); err != nil || len(digests) != 1 {
			return errors.New("owned image digest transport rejected")
		}

		if err := importImage(ctx, input, cluster, digests[0], runner); err != nil {
			return errors.New("owned immutable image import rejected")
		}
	}

	if err := waitDevelopment(ctx, func() error {
		_, err := kubectl([]string{"wait", "nodes", "--all", "--for=condition=Ready", "--timeout=30s"}, nil)

		return err
	}); err != nil {
		return errors.New("owned K3s nodes not ready")
	}

	values, err := gen.BuildLifecycleChartValues(request.Model, request.Inventory, request.Input)
	if err != nil {
		return errors.New("owned smoke lifecycle revision rejected")
	}

	for _, namespace := range values.Namespaces {
		raw, err := json.Marshal(map[string]any{
			developmentAPIVersion: "v1", developmentKind: developmentNamespaceKind,
			developmentMetadata: map[string]any{
				developmentName: namespace.Name,
				developmentLabels: map[string]string{
					pki.SecretLabelRun: request.Model.Run, cryptoutilSharedMagic.KubernetesLabelRelease: namespace.Release,
					"pod-security.kubernetes.io/enforce": "restricted",
				},
			},
		})
		if err != nil {
			return errors.New("owned namespace encoding rejected")
		}

		if _, err := kubectl([]string{developmentCreate, developmentFileFlag, developmentStdin}, raw); err != nil {
			return err
		}
	}

	storage, err := json.Marshal(map[string]any{
		developmentAPIVersion: "storage.k8s.io/v1", developmentKind: "StorageClass",
		developmentMetadata: map[string]any{
			developmentName:   preflight.InfrastructureValues.Storage.ClassName,
			developmentLabels: map[string]string{pki.SecretLabelRun: cluster.Run},
		},
		"provisioner": "rancher.io/local-path", "reclaimPolicy": "Retain", "volumeBindingMode": "WaitForFirstConsumer",
	})
	if err != nil {
		return errors.New("owned retained storage encoding rejected")
	}

	if _, err := kubectl([]string{developmentCreate, developmentFileFlag, developmentStdin}, storage); err != nil {
		return err
	}

	client, err := execution.NewKubectlClient(config, runner)
	if err != nil {
		return fmt.Errorf("development guarded client: %w", err)
	}

	if err := client.Prepare(ctx, values); err != nil {
		return fmt.Errorf("development guarded preparation: %w", err)
	}

	for _, secret := range request.Secrets {
		if err := client.EnsureSecret(ctx, secret); err != nil {
			return fmt.Errorf("development protected Secret: %w", err)
		}
	}

	createChart := func(chart, release string, values any) error {
		raw, err := json.Marshal(values)
		if err != nil {
			return errors.New("owned chart values encoding rejected")
		}

		rendered, err := runner.Run(ctx, config.Helm, []string{
			cryptoutilSharedMagic.HelmCommandTemplate, release,
			filepath.Join(input.root, "api", "cryptosuite-registry", "templates", "kubernetes", chart),
			"--values", developmentStdin, "--kube-version", strings.Split(strings.TrimPrefix(preflight.Lock.Kubernetes.K3S.Version, "v"), "-")[0],
		}, raw)
		if err != nil {
			return errors.New("owned canonical chart rendering rejected")
		}

		_, err = kubectl([]string{developmentCreate, developmentFileFlag, developmentStdin}, rendered)

		return err
	}
	if err := createChart("cryptoutil-infra", preflight.InfrastructureValues.HelmRelease, preflight.InfrastructureValues); err != nil {
		return err
	}

	client, err = execution.NewKubectlClient(config, runner)
	if err != nil {
		return fmt.Errorf("development execution client: %w", err)
	}

	if _, err := execution.Execute(ctx, request, client, execution.Policy{MaxDuration: input.timeout, OperationTimeout: time.Minute}); err != nil {
		return fmt.Errorf("development guarded lifecycle: %w", err)
	}

	for _, values := range preflight.RuntimeValues {
		if err := createChart("cryptoutil-runtime", values.HelmRelease, values); err != nil {
			return err
		}

		for _, workload := range values.Workloads {
			if _, err := kubectl([]string{developmentNamespaceFlag, values.Namespace, "rollout", cryptoutilSharedMagic.StringStatus, "statefulset/" + workload.Name, "--timeout=300s"}, nil); err != nil {
				return err
			}

			for _, instance := range workload.InstanceServices {
				for _, command := range [][]string{workload.Image.Probes.Startup, workload.Image.Probes.Liveness, workload.Image.Probes.Readiness} {
					if _, err := kubectl(append([]string{developmentNamespaceFlag, values.Namespace, developmentExec, instance.PodName, "--"}, command...), nil); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := verifyAPI(ctx, preflight, request.Secrets, config); err != nil {
		return errors.New("owned authenticated API validation rejected")
	}

	_, err = fmt.Fprintln(stdout, "PASS disposable p2sm1103 smoke: generated PKI attested, live probes/admin mTLS and authenticated API verified; NOT full-matrix/production certification")
	if err != nil {
		return errors.New("owned smoke result output rejected")
	}

	return nil
}

func verifyDevelopmentImageArchive(raw []byte, reference string) error {
	return verifyDevelopmentImageStream(bytes.NewReader(raw), reference, io.Discard)
}

func importDevelopmentImage(ctx context.Context, input developmentSmokeInput, cluster gen.LocalK3sPlan, reference string, runner execution.Runner) error {
	return importDevelopmentImageWithFactory(ctx, input, cluster, reference, runner, nativeDevelopmentImageProcess)
}

func importDevelopmentImageWithFactory(ctx context.Context, input developmentSmokeInput, cluster gen.LocalK3sPlan,
	reference string, runner execution.Runner, factory developmentImageFactory,
) error {
	nodes := append([]gen.LocalK3sNode{cluster.Server}, cluster.Agents...)
	if !input.compute.Complete || len(input.compute.Nodes) != len(nodes) {
		return errors.New("owned image import requires complete captured compute")
	}

	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		id := input.compute.Nodes[node.Container]
		if id == "" {
			return errors.New("owned image import node UID absent")
		}

		ids = append(ids, id)
	}

	if err := streamDevelopmentImage(ctx, input.docker, reference, ids, factory); err != nil {
		return err
	}
	// CRI normalizes short Docker references. Require the immutable name that
	// the actual Pod will request, not merely any image with matching content.
	expected := reference

	first, _, qualified := strings.Cut(reference, "/")
	if !qualified {
		expected = "docker.io/library/" + reference
	} else if !strings.ContainsAny(first, ".:") && first != cryptoutilSharedMagic.DefaultOTLPHostnameDefault {
		expected = "docker.io/" + reference
	}

	for _, id := range ids {
		raw, err := runner.Run(ctx, input.docker, []string{
			developmentExec, id, developmentK3s, developmentContainerd, developmentNamespaceFlag,
			developmentContainerdNamespace, developmentContainerdImages, "ls", "--quiet",
		}, nil)
		if err != nil {
			return errors.New("owned immutable image availability unavailable")
		}

		found := false
		for _, name := range strings.Fields(string(raw)) {
			found = found || name == expected
		}

		if !found {
			return errors.New("owned immutable image name absent on node")
		}
	}

	return nil
}

func cleanupDevelopmentNodes(ctx context.Context, docker string, cluster gen.LocalK3sPlan, runner execution.Runner) error {
	snapshot, err := captureDevelopmentCompute(ctx, docker, cluster, runner)

	return errors.Join(err, cleanupDevelopmentCompute(ctx, docker, cluster, runner, snapshot))
}

func developmentOwnedNode(ctx context.Context, docker string, cluster gen.LocalK3sPlan, node gen.LocalK3sNode, runner execution.Runner) (string, error) {
	raw, err := runner.Run(ctx, docker, []string{
		developmentContainer, developmentInspect, developmentFormatFlag,
		`{"id":{{json .Id}},"image":{{json .Config.Image}},"labels":{{json .Config.Labels}},"mounts":{{json .Mounts}}}`, node.Container,
	}, nil)
	if err != nil {
		return "", errors.New("owned node identity unavailable; no unverified deletion")
	}

	var object struct {
		ID, Image string
		Labels    map[string]string
		Mounts    []struct{ Type, Name, Destination string }
	}
	if err := json.Unmarshal(raw, &object); err != nil || object.ID == "" || object.Image != cluster.Image ||
		object.Labels[pki.SecretLabelRun] != cluster.Run || object.Labels["cu.io/node-id"] != node.Container ||
		object.Labels["com.docker.compose.project"] != cluster.Project {
		return "", errors.New("owned node identity collision; no deletion")
	}

	for _, mount := range object.Mounts {
		if mount.Type == developmentVolume && mount.Name == node.Volume && mount.Destination == "/var/lib/rancher/k3s" {
			return object.ID, nil
		}
	}

	return "", errors.New("owned node retained-volume binding rejected")
}

type developmentComputeSnapshot struct {
	Nodes    map[string]string
	Network  string
	Complete bool
}

type developmentNetworkIdentity struct {
	ID         string
	Labels     map[string]string
	Containers map[string]any
}

func developmentOwnedNetwork(ctx context.Context, docker string, cluster gen.LocalK3sPlan, runner execution.Runner) (developmentNetworkIdentity, error) {
	raw, err := runner.Run(ctx, docker, []string{
		developmentNetwork, developmentInspect, developmentFormatFlag,
		`{"id":{{json .Id}},"labels":{{json .Labels}},"containers":{{json .Containers}}}`, cluster.Network,
	}, nil)

	var network developmentNetworkIdentity
	if err != nil {
		return network, errors.New("owned network identity unavailable")
	}

	if err := json.Unmarshal(raw, &network); err != nil || network.ID == "" || network.Labels[pki.SecretLabelRun] != cluster.Run ||
		network.Labels["cu.io/role"] != "k3s-node-network" {
		return developmentNetworkIdentity{}, errors.New("owned network identity collision")
	}

	return network, nil
}

func captureDevelopmentCompute(ctx context.Context, docker string, cluster gen.LocalK3sPlan, runner execution.Runner) (developmentComputeSnapshot, error) {
	snapshot := developmentComputeSnapshot{Nodes: map[string]string{}}

	var result error

	for _, node := range append(append([]gen.LocalK3sNode(nil), cluster.Agents...), cluster.Server) {
		id, err := developmentOwnedNode(ctx, docker, cluster, node, runner)
		if err != nil {
			result = errors.Join(result, err)
		} else {
			snapshot.Nodes[node.Container] = id
		}
	}

	network, err := developmentOwnedNetwork(ctx, docker, cluster, runner)
	if err != nil {
		result = errors.Join(result, err)
	} else {
		snapshot.Network = network.ID
	}

	snapshot.Complete = result == nil

	return snapshot, result
}

func cleanupDevelopmentCompute(ctx context.Context, docker string, cluster gen.LocalK3sPlan, runner execution.Runner, snapshot developmentComputeSnapshot) error {
	var result error

	for _, node := range append(append([]gen.LocalK3sNode(nil), cluster.Agents...), cluster.Server) {
		if snapshot.Nodes[node.Container] == "" {
			continue
		}

		id, err := developmentOwnedNode(ctx, docker, cluster, node, runner)
		if err != nil || id != snapshot.Nodes[node.Container] {
			result = errors.Join(result, errors.New("owned node cleanup UID changed or identity rejected; no deletion"))

			continue
		}

		if _, err := runner.Run(ctx, docker, []string{developmentContainer, "stop", "--time", "30", id}, nil); err != nil {
			result = errors.Join(result, errors.New("owned node stop rejected"))

			continue
		}

		if _, err := runner.Run(ctx, docker, []string{developmentContainer, developmentRemove, id}, nil); err != nil {
			result = errors.Join(result, errors.New("owned node removal rejected; retained volumes untouched"))
		}
	}

	if result != nil || snapshot.Network == "" || !snapshot.Complete {
		return result
	}

	network, err := developmentOwnedNetwork(ctx, docker, cluster, runner)
	if err != nil || network.ID != snapshot.Network || len(network.Containers) != 0 {
		return errors.New("owned network cleanup collision")
	}

	if _, err := runner.Run(ctx, docker, []string{developmentNetwork, developmentRemove, network.ID}, nil); err != nil {
		return errors.New("owned network cleanup rejected")
	}

	return nil
}
