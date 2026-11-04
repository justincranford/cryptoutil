// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"time"
)

const developmentRuntimePlanCommand = "dev-runtime-plan"

func developmentRuntimePlanMain(args []string, stdout, stderr io.Writer, read func(string) ([]byte, error), run DockerCommandRunner, listen func(string, string) (net.Listener, error)) int {
	flags := flag.NewFlagSet(developmentRuntimePlanCommand, flag.ContinueOnError)
	flags.SetOutput(stderr)
	runID := flags.String(cryptoutilSharedMagic.WorkflowSubCmdRun, "", "explicit new disposable run (1-8 lowercase alphanumeric)")
	selector := flags.String("select", "s:sm-kms:mem", "documented initial dev smoke selector; empty admits complete matrix")
	image := flags.String("image", "", "actual local-built owned suite image tag; inspected, never fabricated")
	reserveCPU := flags.Int64("reserve-cpu-milli", 0, "explicit positive physical-host/K3s CPU reserve")
	reserveMemory := flags.Int64("reserve-memory-bytes", 0, "explicit positive physical-host/K3s memory reserve")

	timeout := flags.Duration("timeout", time.Minute, "bounded read-only preparation and host probes")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	if flags.NArg() != 0 || *runID == "" || *image == "" || *reserveCPU <= 0 || *reserveMemory <= 0 || *timeout <= 0 || *timeout > time.Hour {
		return failWithDiagnostic(stderr, 2, "dev-runtime-plan requires explicit owned run/image, positive reserves and bounded timeout\n")
	}

	modelData, err := read(defaultPlannerPaths().model)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read development model failed\n")
	}

	model, err := DecodeModel(modelData)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "decode development model failed\n")
	}

	rawLock, err := read(filepath.Join(filepath.Dir(defaultPlannerPaths().model), k8sToolchainLockFileName))
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read locked toolchain failed\n")
	}

	lock, err := DecodeToolchainLock(rawLock)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "decode locked toolchain failed\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	// Output includes only explicit image provenance, never image environment,
	// command arguments or protected values inherited from unrelated images.
	format := `{"digests":{{json .RepoDigests}},"os":{{json .Os}},"architecture":{{json .Architecture}},` +
		`"run":{{json (index .Config.Labels "cu.io/run")}},"owner":{{json (index .Config.Labels "cu.io/owner")}}}`

	rawImage, err := run(ctx, "docker", "image", "inspect", "--format", format, *image)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read owned suite image failed; private diagnostics suppressed\n")
	}

	var inspected struct {
		Digests                      []string
		OS, Architecture, Run, Owner string
	}
	if err := json.Unmarshal(rawImage, &inspected); err != nil || len(inspected.Digests) != 1 || inspected.OS != "linux" ||
		inspected.Architecture != "amd64" || inspected.Run != *runID || inspected.Owner != "phase2-runtime" {
		return failWithDiagnostic(stderr, 1, "suite image provenance/platform/run rejected\n")
	}

	plan, err := BuildDevelopmentRuntimePlan(model, lock, *runID, *selector, inspected.Digests[0])
	if err != nil {
		return failWithDiagnostic(stderr, 1, "development graph/argv/identity planning failed: %v\n", err)
	}

	inventory := plan.inventory

	rawDefaults, err := read(infrastructureChartDefaultsPath)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read canonical infrastructure defaults failed\n")
	}

	defaults, err := DecodeInfrastructureChartValues(rawDefaults)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "decode canonical infrastructure defaults failed\n")
	}

	defaults.Storage.ClassName = plan.RuntimeInput.Storage.ClassName

	infra, err := BuildInfrastructureChartValues(plan.Model, inventory, defaults)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "development infrastructure projection failed: %v\n", err)
	}

	input := BootstrapPreflightInput{
		Model: plan.Model, Inventory: inventory, Lock: lock, InfrastructureValues: infra, Lifecycle: &plan.Lifecycle,
		Reserve: ResourceCapacity{CPUMilli: *reserveCPU, MemoryBytes: *reserveMemory},
	}
	input.RuntimeValues = plan.runtimeValues

	estimate, err := CheckLocalK3sPreflight(ctx, input, run, listen)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "development bootstrap preflight rejected: %v\n", err)
	}

	const reportFormat = "PASS: read-only ownership/ports/host request preflight; NOT runtime/material certification\n" +
		"run=%s selector=%q revision=%s project=%s network=%s\nsuite=%s\n" +
		"requests cpuMilli=%d memoryBytes=%d reserveCPU=%d reserveMemory=%d\n" +
		"releases=%d protectedSecretRequirements=%d jobs=%d\n"

	if _, err := fmt.Fprintf(stdout, reportFormat, *runID, *selector, plan.Model.Revision, plan.Cluster.Project, plan.Cluster.Network,
		inspected.Digests[0], estimate.TotalRequests.CPUMilli, estimate.TotalRequests.MemoryBytes, *reserveCPU, *reserveMemory,
		len(inventory.Releases), len(plan.Lifecycle.Secrets), len(plan.Lifecycle.Jobs)); err != nil {
		return 1
	}

	return 0
}
