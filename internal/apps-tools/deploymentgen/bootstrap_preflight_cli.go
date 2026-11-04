// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"time"
)

const bootstrapPreflightCommand = "bootstrap-preflight"

// LoadBootstrapPreflightInput loads schema-validated non-secret projections.
// Callers must attach explicit Lifecycle input before admitting or executing.
// The injected reader controls the repository-root boundary; this function
// neither probes a host nor mutates a cluster.
func LoadBootstrapPreflightInput(
	selector, runIdentity, inputPath string, readFile func(string) ([]byte, error),
) (BootstrapPreflightInput, error) {
	return loadBootstrapPreflightInput(selector, runIdentity, inputPath, readFile)
}

func bootstrapPreflightMain(args []string, stdout, stderr io.Writer, readFile func(string) ([]byte, error)) int {
	return bootstrapPreflightMainWithIO(args, stdout, stderr, readFile, runBootstrapDockerCommand, net.Listen)
}

func runBootstrapDockerCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	process := exec.CommandContext(ctx, command, args...)
	process.WaitDelay = time.Second

	output, err := process.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("execute bootstrap probe %s: %w", command, err)
	}

	return output, nil
}

func bootstrapPreflightMainWithIO(
	args []string,
	stdout, stderr io.Writer,
	readFile func(string) ([]byte, error),
	run DockerCommandRunner,
	listen func(string, string) (net.Listener, error),
) int {
	flags := flag.NewFlagSet(bootstrapPreflightCommand, flag.ContinueOnError)
	flags.SetOutput(stderr)
	selector := flags.String("select", "", "one exact tier:scope:mode selector; empty selects the complete matrix")
	runIdentity := flags.String(cryptoutilSharedMagic.WorkflowSubCmdRun, "", "override bootstrap run identity")
	inputPath := flags.String("runtime-input", "", "explicit non-secret runtime chart input JSON covering selected releases")
	lifecyclePath := flags.String("lifecycle-input", "", "required explicit non-secret lifecycle input including Job requests")
	reserveCPU := flags.Int64("reserve-cpu-milli", 0, "explicit positive physical-host/K3s CPU reserve in millicores")
	reserveMemory := flags.Int64("reserve-memory-bytes", 0, "explicit positive physical-host/K3s memory reserve in bytes")
	timeout := flags.Duration("timeout", time.Minute, "positive deadline for read-only Docker admission probes")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	if flags.NArg() != 0 || *inputPath == "" || *lifecyclePath == "" || *reserveCPU <= 0 || *reserveMemory <= 0 || *timeout <= 0 {
		return failWithDiagnostic(stderr, 2, "bootstrap-preflight requires -runtime-input and -lifecycle-input, positive -reserve-cpu-milli/-reserve-memory-bytes, and positive -timeout\n")
	}

	input, err := loadBootstrapPreflightInput(*selector, *runIdentity, *inputPath, readFile)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "bootstrap preflight input: %v\n", err)
	}

	rawLifecycle, err := readFile(*lifecyclePath)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read admission lifecycle input failed\n")
	}

	lifecycle, err := DecodeLifecycleInput(rawLifecycle)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "decode admission lifecycle input: %v\n", err)
	}

	input.Lifecycle = &lifecycle
	input.Reserve = ResourceCapacity{CPUMilli: *reserveCPU, MemoryBytes: *reserveMemory}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	estimate, err := CheckLocalK3sPreflight(ctx, input, run, listen)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "bootstrap admission rejected: %v\n", err)
	}

	if _, err := fmt.Fprintf(stdout,
		"PASS: fresh Compose bootstrap host-total admission; run=%s app-pods=%d infrastructure-pods=%d lifecycle-pods=%d requests=%dm/%d bytes; no cluster changes; node placement/readiness not verified\n",
		input.Model.Run, estimate.ApplicationPods, estimate.InfrastructurePods, estimate.LifecyclePods,
		estimate.TotalRequests.CPUMilli, estimate.TotalRequests.MemoryBytes,
	); err != nil {
		return failWithDiagnostic(stderr, 1, "write bootstrap admission: %v\n", err)
	}

	return 0
}

func loadBootstrapPreflightInput(
	selector, runIdentity, inputPath string,
	readFile func(string) ([]byte, error),
) (BootstrapPreflightInput, error) {
	if readFile == nil {
		return BootstrapPreflightInput{}, fmt.Errorf("bootstrap preflight file reader is required")
	}

	modelData, err := readFile(defaultPlannerPaths().model)
	if err != nil {
		return BootstrapPreflightInput{}, fmt.Errorf("read bootstrap model: %w", err)
	}

	model, err := decodeModel(modelData)
	if err != nil {
		return BootstrapPreflightInput{}, err
	}

	if runIdentity != "" {
		model.Run = runIdentity
	}

	lockData, err := readFile(filepath.Join(filepath.Dir(defaultPlannerPaths().model), k8sToolchainLockFileName))
	if err != nil {
		return BootstrapPreflightInput{}, fmt.Errorf("read bootstrap toolchain lock: %w", err)
	}

	lock, err := DecodeToolchainLock(lockData)
	if err != nil {
		return BootstrapPreflightInput{}, err
	}

	inventory, err := build(model, RegisteredScopes(), selector)
	if err != nil {
		return BootstrapPreflightInput{}, err
	}

	inputData, err := readFile(inputPath)
	if err != nil {
		return BootstrapPreflightInput{}, fmt.Errorf("read bootstrap runtime input: %w", err)
	}

	inputSchema, err := readFile(chartValuesInputSchemaPath)
	if err != nil {
		return BootstrapPreflightInput{}, fmt.Errorf("%s: %w", chartValuesInputSchemaReadError, err)
	}

	if err := ValidateChartValuesInputSchema(inputSchema, inputData); err != nil {
		return BootstrapPreflightInput{}, err
	}

	chartInput, err := DecodeChartValuesInput(inputData)
	if err != nil {
		return BootstrapPreflightInput{}, err
	}

	runtimeValues := make([]RuntimeChartValues, 0, len(inventory.Releases))

	for _, release := range inventory.Releases {
		selected := inventory
		selected.Releases = []Release{release}

		values, err := BuildRuntimeChartValues(
			model, selected, release.Key,
			chartInput.Images, chartInput.PKIMaterializer, chartInput.Storage, chartInput.Runtime,
		)
		if err != nil {
			return BootstrapPreflightInput{}, err
		}

		runtimeValues = append(runtimeValues, values)
	}

	defaultsData, err := readFile(infrastructureChartDefaultsPath)
	if err != nil {
		return BootstrapPreflightInput{}, fmt.Errorf("read bootstrap infrastructure defaults: %w", err)
	}

	defaults, err := DecodeInfrastructureChartValues(defaultsData)
	if err != nil {
		return BootstrapPreflightInput{}, err
	}

	infrastructureValues, err := BuildInfrastructureChartValues(model, inventory, defaults)
	if err != nil {
		return BootstrapPreflightInput{}, err
	}

	return BootstrapPreflightInput{
		Model: model, Lock: lock, Inventory: inventory, RuntimeValues: runtimeValues,
		InfrastructureValues: infrastructureValues,
	}, nil
}
