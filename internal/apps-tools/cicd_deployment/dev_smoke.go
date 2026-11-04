// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"crypto/elliptic"
	sha256 "crypto/sha256"
	"encoding/hex"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const devSmokeCommand = "dev-smoke"

type developmentSmokeInput struct {
	root, image, docker, kubectl, helm string
	timeout                            time.Duration
	compute                            developmentComputeSnapshot
	resumeBootstrap                    bool
	preflightOnly                      bool
}

// This route is restricted to user-authorized new disposable p2sm1103/mem.
// No cluster certification flag or ambient target/context is accepted.
func developmentSmokeMain(args []string, stdout, stderr io.Writer) int {
	return developmentSmokeMainWithRuntime(args, stdout, stderr, exec.LookPath, executeDevelopmentSmoke)
}

func developmentSmokeMainWithRuntime(args []string, stdout, stderr io.Writer, lookup func(string) (string, error),
	execute func(context.Context, developmentSmokeInput, execution.Runner, io.Writer) error,
) int {
	flags := flag.NewFlagSet(devSmokeCommand, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	input := developmentSmokeInput{}
	flags.StringVar(&input.root, "repository-root", "", "absolute owned source root")
	flags.StringVar(&input.image, "image", "", "actual run-owned locally built suite image")
	flags.StringVar(&input.kubectl, "kubectl", "", "absolute lock-verified kubectl")
	flags.StringVar(&input.helm, "helm", "", "absolute lock-verified Helm")
	flags.DurationVar(&input.timeout, "timeout", 0, "bounded complete smoke deadline")
	flags.BoolVar(&input.resumeBootstrap, "resume-bootstrap", false, "explicit same-run incomplete bootstrap recovery; retain original token/data")
	flags.BoolVar(&input.preflightOnly, "preflight-only", false, "read-only admission only; no generated material or node launch")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(stdout, "dev-smoke requires explicit root/image/kubectl/helm/timeout; disposable p2sm1103 only")
			if err == nil {
				return 0
			}
		}

		return executionDiagnostic(stderr, 2, "disposable smoke flags rejected")
	}

	if flags.NArg() != 0 || input.image == "" || input.timeout <= 0 || input.timeout > time.Hour {
		return executionDiagnostic(stderr, 2, "explicit bounded disposable smoke inputs required")
	}

	for _, value := range []string{input.root, input.kubectl, input.helm} {
		if !filepath.IsAbs(value) || filepath.Clean(value) != value {
			return executionDiagnostic(stderr, 2, "explicit clean absolute disposable target paths required")
		}
	}

	docker, err := lookup("docker")
	if err != nil {
		return executionDiagnostic(stderr, 1, "native Docker unavailable")
	}

	input.docker, err = filepath.Abs(docker)
	if err != nil {
		return executionDiagnostic(stderr, 1, "native Docker path unavailable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), input.timeout)
	defer cancel()

	if err := execute(ctx, input, execution.ExecRunner{}, stdout); err != nil {
		return executionDiagnostic(stderr, 1, err.Error())
	}

	return 0
}

func executeDevelopmentSmoke(ctx context.Context, input developmentSmokeInput, runner execution.Runner, stdout io.Writer) (result error) {
	return executeDevelopmentSmokeWithOperations(ctx, input, runner, stdout, developmentSmokeOperations{
		listen: net.Listen, deploy: deployDevelopmentSmoke, verifyImages: verifyNativeDevelopmentImages, writeProtected: writeDevelopmentSmokeProtected,
		observation: nativeDevelopmentBootstrapObservation,
	})
}

type developmentSmokeOperations struct {
	listen         func(string, string) (net.Listener, error)
	verifyImages   func(context.Context, developmentSmokeInput, []string) error
	writeProtected func(*gen.DevelopmentPrivateDirectory, string, []byte) error
	observation    func(string, gen.LocalK3sPlan, execution.Runner, developmentComputeSnapshot) string
	deploy         func(context.Context, developmentSmokeInput, gen.BootstrapPreflightInput, gen.LocalK3sPlan, execution.Request,
		execution.KubectlConfig, execution.Runner, io.Writer) error
}

func executeDevelopmentSmokeWithOperations(ctx context.Context, input developmentSmokeInput, runner execution.Runner,
	stdout io.Writer, operations developmentSmokeOperations,
) (result error) {
	if operations.listen == nil || operations.deploy == nil || operations.verifyImages == nil || operations.writeProtected == nil {
		return errors.New("explicit disposable smoke transports required")
	}

	call := func(binary string, args []string, stdin []byte) ([]byte, error) {
		out, err := runner.Run(ctx, binary, args, stdin)
		if err != nil {
			return nil, errors.New("disposable smoke transport rejected; private diagnostics suppressed")
		}

		return out, nil
	}
	read := func(relative string) ([]byte, error) {
		raw, err := os.ReadFile(filepath.Join(input.root, filepath.FromSlash(relative)))
		if err != nil {
			return nil, errors.New("disposable checked-in input unavailable")
		}

		return raw, nil
	}

	raw, err := read("api/cryptosuite-registry/k8s/k8s-model.json")
	if err != nil {
		return err
	}

	model, err := gen.DecodeModel(raw)
	if err != nil {
		return errors.New("disposable model rejected")
	}

	raw, err = read("api/cryptosuite-registry/k8s/k8s-toolchain.lock.json")
	if err != nil {
		return err
	}

	lock, err := gen.DecodeToolchainLock(raw)
	if err != nil {
		return errors.New("disposable toolchain rejected")
	}

	binary, err := os.ReadFile(input.kubectl)

	digest := sha256.Sum256(binary)
	if err != nil || hex.EncodeToString(digest[:]) != lock.Kubernetes.Kubectl.WindowsAMD64SHA256 {
		return errors.New("disposable kubectl lock mismatch")
	}

	version, err := call(input.helm, []string{cryptoutilSharedMagic.CLIVersionCommand, "--template", "{{.Version}}"}, nil)
	if err != nil || strings.TrimSpace(string(version)) != lock.Kubernetes.Helm.Version {
		return errors.New("disposable Helm lock mismatch")
	}

	raw, err = call(input.docker, []string{
		developmentImage, developmentInspect, developmentFormatFlag,
		`{"digests":{{json .RepoDigests}},"os":{{json .Os}},"architecture":{{json .Architecture}},"labels":{{json .Config.Labels}}}`, input.image,
	}, nil)
	if err != nil {
		return err
	}

	var inspected struct {
		Digests          []string
		OS, Architecture string
		Labels           map[string]string
	}
	if err := json.Unmarshal(raw, &inspected); err != nil || len(inspected.Digests) != 1 ||
		inspected.OS != "linux" || inspected.Architecture != "amd64" ||
		inspected.Labels[cryptoutilSharedMagic.KubernetesLabelRun] != gen.DevelopmentSmokeRun || inspected.Labels["cu.io/owner"] != "phase2-runtime" {
		return errors.New("disposable suite image ownership/platform rejected")
	}

	plan, err := gen.BuildDevelopmentRuntimePlan(model, lock, gen.DevelopmentSmokeRun, gen.DevelopmentSmokeSelector, inspected.Digests[0])
	if err != nil {
		return errors.New("disposable graph planning rejected")
	}
	// Subsequent reads and saves must not follow a tag that can be replaced
	// after this owner/platform inspection.
	input.image = inspected.Digests[0]

	raw, err = read("configs/sm-kms/sm-kms-domain.yml")
	if err != nil {
		return err
	}

	plan, err = gen.ConfigureDevelopmentSmoke(plan, raw)
	if err != nil {
		return fmt.Errorf("development configuration: %w", err)
	}

	preflight, err := developmentSmokePreflight(plan, lock, read)
	if err != nil {
		return err
	}

	raw, err = read("api/cryptosuite-registry/templates/kubernetes/local-k3s/compose.yml.tmpl")
	if err != nil {
		return err
	}

	compose, err := gen.RenderLocalK3sCompose(raw, plan.Cluster)
	if err != nil {
		return errors.New("disposable canonical node topology rejected")
	}

	dockerRun := func(ctx context.Context, _ string, args ...string) ([]byte, error) {
		return runner.Run(ctx, input.docker, args, nil)
	}

	var private *gen.DevelopmentPrivateDirectory

	var originalToken []byte

	var retained gen.DevelopmentRetainedVolumes

	defer func() {
		if private != nil {
			result = errors.Join(result, private.Close())
		}
	}()

	if input.resumeBootstrap {
		private, originalToken, err = gen.OpenDevelopmentBootstrapDirectory(filepath.Join(input.root, cryptoutilSharedMagic.CICDExcludeDirTestOutput), compose)
		if err != nil {
			return errors.New("explicit incomplete protected bootstrap recovery rejected")
		}

		retained, err = gen.CaptureDevelopmentRetainedVolumes(ctx, plan.Cluster, dockerRun)
		if err != nil {
			return errors.New("explicit retained bootstrap volume identity rejected")
		}
	}

	admit := func() error {
		if input.resumeBootstrap {
			_, err := gen.CheckDevelopmentBootstrapResume(ctx, preflight, retained, dockerRun, operations.listen)
			if err != nil {
				return errors.New("retained bootstrap admission rejected")
			}

			return nil
		}

		_, err := gen.CheckLocalK3sPreflight(ctx, preflight, dockerRun, operations.listen)
		if err != nil {
			return errors.New("fresh bootstrap admission rejected")
		}

		return nil
	}
	if err := admit(); err != nil {
		return errors.New("disposable ownership/ports/reserves admission rejected")
	}

	// Read-only immutable archive interoperability is proved before generating
	// protected output or starting any node. No archive is persisted.
	if err := operations.verifyImages(ctx, input, []string{inspected.Digests[0], lock.Images.OTelCollector.Reference, lock.Images.LGTM.Reference}); err != nil {
		return errors.New("disposable pinned image stream admission rejected")
	}

	if input.preflightOnly {
		if _, err := fmt.Fprintln(stdout, "p2sm1103 ownership, protected recovery inputs, ports, reserves and immutable image streams admitted; no runtime acceptance"); err != nil {
			return errors.New("disposable admission report unavailable")
		}

		return nil
	}

	material, err := gen.GenerateDevelopmentMaterial(ctx, plan, gen.DevelopmentMaterialPolicy{Curve: elliptic.P384(), Lifetime: developmentMaterialLifetime})
	if err != nil {
		return fmt.Errorf("development material: %w", err)
	}

	request := execution.Request{Model: plan.Model, Inventory: preflight.Inventory, Input: plan.Lifecycle, Secrets: material.Secrets}

	manifest, err := developmentPublication(preflight.RuntimeValues)
	if err != nil {
		return err
	}

	request, err = preparePKIVerification(preflight, request, manifest)
	if err != nil {
		return err
	}

	if !input.resumeBootstrap {
		private, err = gen.CreateDevelopmentPrivateDirectory(filepath.Join(input.root, cryptoutilSharedMagic.CICDExcludeDirTestOutput))
		if err != nil {
			return fmt.Errorf("development private output: %w", err)
		}

		if err := operations.writeProtected(private, "secrets/p2sm1103/k3s-join-token", material.Token); err != nil {
			return errors.New("development protected token rejected")
		}

		if err := operations.writeProtected(private, "compose.yml", compose); err != nil {
			return errors.New("development protected topology rejected")
		}
	} else {
		// The retained datastore and agents remain bound to the original token.
		// Newly generated material is memory-only and cannot replace this file.
		material.Token = originalToken
	}

	composePath := filepath.Join(input.root, cryptoutilSharedMagic.CICDExcludeDirTestOutput, "phase2-runtime-p2sm1103", "compose.yml")
	// The same admission repeats immediately before mutation, comparing every
	// retained identity without omitting collisions or changing resource budgets.
	if err := admit(); err != nil {
		return errors.New("final disposable launch admission rejected")
	}

	var compute developmentComputeSnapshot

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), developmentCleanupTimeout)
		defer cancel()

		if compute.Nodes == nil {
			var err error

			compute, err = captureDevelopmentCompute(cleanupCtx, input.docker, plan.Cluster, runner)
			result = errors.Join(result, err)
		}

		if result != nil {
			observe := operations.observation
			if observe == nil {
				observe = observeDevelopmentBootstrap
			}

			result = errors.Join(result, errors.New(observe(input.docker, plan.Cluster, runner, compute)))
		}

		result = errors.Join(result, cleanupDevelopmentCompute(cleanupCtx, input.docker, plan.Cluster, runner, compute))
	}()

	if _, err := call(input.docker, []string{"compose", "--project-name", plan.Cluster.Project, developmentFileFlag, composePath, "up", "-d", "--no-build", "--pull", "never"}, nil); err != nil {
		return err
	}

	compute, err = captureDevelopmentCompute(ctx, input.docker, plan.Cluster, runner)
	if err != nil {
		return err
	}

	input.compute = compute

	if err := waitDevelopment(ctx, func() error {
		_, err := call(input.docker, []string{developmentExec, compute.Nodes[plan.Cluster.Server.Container], developmentK3s, "kubectl", developmentGetCommand, "--raw=/readyz"}, nil)

		return err
	}); err != nil {
		return errors.New("owned K3s server readiness unavailable")
	}

	raw, err = call(input.docker, []string{developmentExec, compute.Nodes[plan.Cluster.Server.Container], "cat", "/etc/rancher/k3s/k3s.yaml"}, nil)
	if err != nil {
		return err
	}

	kubeconfig, err := developmentKubeconfig(raw, plan.Cluster)
	if err != nil {
		return err
	}

	if err := operations.writeProtected(private, "kubeconfig", kubeconfig); err != nil {
		return errors.New("development protected kubeconfig rejected")
	}

	config := execution.KubectlConfig{
		Kubectl: input.kubectl, Helm: input.helm, RepositoryRoot: input.root,
		Kubeconfig: filepath.Join(filepath.Dir(composePath), "kubeconfig"), Context: gen.DevelopmentSmokeRun, PollInterval: time.Second,
	}

	if input.resumeBootstrap {
		raw, err := call(input.kubectl, []string{
			"--kubeconfig", config.Kubeconfig, "--context", config.Context,
			developmentGetCommand, "namespaces,pods,jobs", "--all-namespaces", "-o", "json",
		}, nil)
		if err != nil || validateDevelopmentBootstrapInventory(raw) != nil {
			return errors.New("bootstrap recovery live application absence rejected; retained data untouched")
		}
	}

	return operations.deploy(ctx, input, preflight, plan.Cluster, request, config, runner, stdout)
}

func developmentSmokePreflight(plan gen.DevelopmentRuntimePlan, lock gen.ToolchainLock, read func(string) ([]byte, error)) (gen.BootstrapPreflightInput, error) {
	raw, err := read("api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra/values.yaml")
	if err != nil {
		return gen.BootstrapPreflightInput{}, fmt.Errorf("development runtime: %w", err)
	}

	defaults, err := gen.DecodeInfrastructureChartValues(raw)
	if err != nil {
		return gen.BootstrapPreflightInput{}, fmt.Errorf("development defaults: %w", err)
	}

	inventory, err := gen.Build(plan.Model, plan.Selector)
	if err != nil {
		return gen.BootstrapPreflightInput{}, fmt.Errorf("development inventory: %w", err)
	}

	defaults.Storage.ClassName = plan.RuntimeInput.Storage.ClassName

	infra, err := gen.BuildInfrastructureChartValues(plan.Model, inventory, defaults)
	if err != nil {
		return gen.BootstrapPreflightInput{}, fmt.Errorf("development infrastructure: %w", err)
	}

	input := gen.BootstrapPreflightInput{
		Model: plan.Model, Lock: lock, Inventory: inventory,
		InfrastructureValues: infra, Lifecycle: &plan.Lifecycle,
		Reserve: gen.ResourceCapacity{CPUMilli: developmentCPUReserve, MemoryBytes: developmentMemoryReserve},
	}
	for _, release := range inventory.Releases {
		values, err := gen.BuildRuntimeChartValues(plan.Model, inventory, release.Key, plan.RuntimeInput.Images,
			plan.RuntimeInput.PKIMaterializer, plan.RuntimeInput.Storage, plan.RuntimeInput.Runtime)
		if err != nil {
			return gen.BootstrapPreflightInput{}, fmt.Errorf("development runtime values: %w", err)
		}

		input.RuntimeValues = append(input.RuntimeValues, values)
	}

	return input, nil
}
