// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	cryptoutilDeploymentGen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilLifecycleRun "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
)

const lifecycleExecuteCommand = "lifecycle-execute"

type lifecycleExecutionIO struct {
	read          func(string) ([]byte, error)
	load          func(string, string, string, func(string) ([]byte, error)) (cryptoutilDeploymentGen.BootstrapPreflightInput, error)
	admit         func(cryptoutilDeploymentGen.BootstrapPreflightInput) (cryptoutilDeploymentGen.CapacityEstimate, error)
	importSecrets func(cryptoutilDeploymentGen.Model, cryptoutilDeploymentGen.Inventory, cryptoutilDeploymentGen.LifecycleInput, string) ([]cryptoutilDeploymentGen.LifecycleSecret, error)
	execute       func(context.Context, cryptoutilLifecycleRun.Request, cryptoutilLifecycleRun.KubectlConfig, cryptoutilLifecycleRun.Policy) (cryptoutilLifecycleRun.Result, error)
	publish       func(context.Context, cryptoutilDeploymentGen.BootstrapPreflightInput, []cryptoutilDeploymentGen.LifecycleSecret, string, cryptoutilLifecycleRun.KubectlConfig, cryptoutilLifecycleRun.Policy) (cryptoutilLifecycleRun.Result, error)
	verifyPKI     func(cryptoutilDeploymentGen.BootstrapPreflightInput, cryptoutilLifecycleRun.Request, string) (cryptoutilLifecycleRun.Request, error)
}

func lifecycleExecuteMain(args []string, stdout, stderr io.Writer) int {
	return lifecycleExecuteMainWithIO(args, stdout, stderr, lifecycleExecutionProcess(executeLifecycleRequest))
}

func lifecycleExecutionProcess(
	execute func(context.Context, cryptoutilLifecycleRun.Request, cryptoutilLifecycleRun.KubectlConfig, cryptoutilLifecycleRun.Policy) (cryptoutilLifecycleRun.Result, error),
) lifecycleExecutionIO {
	return lifecycleExecutionIO{
		read: os.ReadFile, load: loadLifecycleExecutionInput,
		admit:         cryptoutilDeploymentGen.EstimateLifecycleCapacity,
		importSecrets: cryptoutilDeploymentGen.ImportPhysicalLifecycleSecrets,
		execute:       execute,
		verifyPKI: func(input cryptoutilDeploymentGen.BootstrapPreflightInput, request cryptoutilLifecycleRun.Request, file string) (cryptoutilLifecycleRun.Request, error) {
			manifest, err := readPKIPublicationInput(file)
			if err != nil {
				return request, errors.New("publication identity input unavailable")
			}

			return preparePKIVerification(input, request, manifest)
		},
	}
}

func lifecycleExecuteMainWithIO(args []string, stdout, stderr io.Writer, process lifecycleExecutionIO) int {
	return lifecycleCommandWithIO(args, stdout, stderr, process, lifecycleExecuteCommand)
}

func lifecycleCommandWithIO(args []string, stdout, stderr io.Writer, process lifecycleExecutionIO, command string) int {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)

	var publicationPath string
	flags.StringVar(&publicationPath, "publication-input", "", "explicit non-secret per-pod PKI trust/DNS publication and postJob verification contract")

	inputPath := flags.String("input", "", "explicit non-secret lifecycle operator input")
	runtimeInput := flags.String("runtime-input", "", "explicit non-secret runtime inputs for current quota/budget validation")
	selector := flags.String("select", "", "exact release selector; empty selects full matrix")
	root := flags.String("repository-root", "", "absolute certified canonical repository root")
	secretRoot := flags.String("secret-root", "", "absolute owner-protected material directory")
	kubectl := flags.String("kubectl", "", "absolute certified kubectl executable")
	helm := flags.String("helm", "", "absolute certified Helm executable")
	kubeconfig := flags.String("kubeconfig", "", "absolute explicit kubeconfig file")
	kubecontext := flags.String("context", "", "explicit Kubernetes context; no ambient default")
	admissionCertified := flags.Bool("admission-certified", false, "operator attestation: prior fresh lifecycle-aware host admission passed for this exact run/input; not a live probe")
	timeout := flags.Duration("timeout", 0, "explicit positive overall execution deadline")
	operationTimeout := flags.Duration("operation-timeout", 0, "explicit positive per-operation deadline")

	poll := flags.Duration("poll-interval", 0, "explicit positive polling interval")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	validPaths := true
	for _, name := range []string{*root, *secretRoot, *kubectl, *helm, *kubeconfig} {
		validPaths = validPaths && filepath.IsAbs(name) && filepath.Clean(name) == name
	}

	if flags.NArg() != 0 || *inputPath == "" || *runtimeInput == "" || !validPaths ||
		publicationPath == "" ||
		strings.TrimSpace(*kubecontext) == "" || !*admissionCertified ||
		*timeout <= 0 || *timeout > 24*time.Hour || *operationTimeout <= 0 ||
		*operationTimeout > time.Hour || *operationTimeout > *timeout || *poll <= 0 || *poll > time.Minute {
		return executionDiagnostic(stderr, 2, "lifecycle-execute requires explicit clean absolute target/material paths, input/runtime-input/publication-input, context, prior admission certification, and bounded positive policy durations")
	}

	input, err := process.load(*root, *selector, *runtimeInput, process.read)
	if err != nil {
		return executionDiagnostic(stderr, 1, "execution non-secret projections rejected")
	}

	raw, err := process.read(*inputPath)
	if err != nil {
		return executionDiagnostic(stderr, 1, "read execution lifecycle input failed")
	}

	lifecycle, err := cryptoutilDeploymentGen.DecodeLifecycleInput(raw)
	if err != nil {
		return executionDiagnostic(stderr, 1, "decode execution lifecycle input failed")
	}

	input.Lifecycle = &lifecycle
	if _, err := process.admit(input); err != nil {
		return executionDiagnostic(stderr, 1, "execution lifecycle ownership, revision, resources or quota budget rejected")
	}

	// No material is read until the explicit current contract and budgets pass.
	secrets, err := process.importSecrets(input.Model, input.Inventory, lifecycle, *secretRoot)
	if err != nil {
		return executionDiagnostic(stderr, 1, "protected execution material import rejected")
	}

	if command != pkiPublishCommand {
		if err := cryptoutilDeploymentGen.ValidateDatabaseJobMaterial(input.Inventory, lifecycle, secrets); err != nil {
			return executionDiagnostic(stderr, 1, "protected database Job material contract rejected")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	config := cryptoutilLifecycleRun.KubectlConfig{
		Kubectl: *kubectl, Helm: *helm, RepositoryRoot: *root, Kubeconfig: *kubeconfig, Context: *kubecontext, PollInterval: *poll,
	}
	policy := cryptoutilLifecycleRun.Policy{MaxDuration: *timeout, OperationTimeout: *operationTimeout}

	var result cryptoutilLifecycleRun.Result
	if command == pkiPublishCommand {
		result, err = process.publish(ctx, input, secrets, publicationPath, config, policy)
	} else {
		request := cryptoutilLifecycleRun.Request{
			Model: input.Model, Inventory: input.Inventory, Input: lifecycle, Secrets: secrets,
		}

		if process.verifyPKI == nil {
			return executionDiagnostic(stderr, 1, "host-prepublished verify-only PKI verification adapter required")
		}

		request, err = process.verifyPKI(input, request, publicationPath)
		if err != nil {
			return executionDiagnostic(stderr, 1, "PKI generation/output cannot be attested; protected host-prepublished verify-only contract required")
		}

		result, err = process.execute(ctx, request, config, policy)
	}

	if err != nil {
		return executionDiagnostic(stderr, 1, lifecycleExecutionFailureCategory(err))
	}

	if _, err := fmt.Fprintf(stdout, "PASS: guarded %s; secrets=%d jobs=%d; prior host admission operator-certified, not reprobed; runtime acceptance remains outstanding\n", command, result.Secrets, result.Jobs); err != nil {
		return executionDiagnostic(stderr, 1, "write execution result failed")
	}

	return 0
}

func executionDiagnostic(stderr io.Writer, code int, message string) int {
	if _, err := fmt.Fprintln(stderr, message); err != nil {
		return 1
	}

	return code
}

func lifecycleExecutionFailureCategory(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "lifecycle execution deadline exceeded"
	case errors.Is(err, context.Canceled):
		return "lifecycle execution canceled"
	case errors.Is(err, cryptoutilLifecycleRun.ErrCollision):
		return "lifecycle execution rejected an existing resource ownership/data/spec collision"
	case errors.Is(err, cryptoutilLifecycleRun.ErrPrerequisite):
		return "lifecycle execution prerequisites are not satisfied"
	case errors.Is(err, cryptoutilLifecycleRun.ErrJobFailed):
		return "lifecycle execution Job failed"
	case errors.Is(err, cryptoutilLifecycleRun.ErrReplaced):
		return "lifecycle execution resource identity was replaced"
	default:
		return "lifecycle execution transport rejected; private diagnostics suppressed"
	}
}

func loadLifecycleExecutionInput(
	root, selector, inputPath string, read func(string) ([]byte, error),
) (cryptoutilDeploymentGen.BootstrapPreflightInput, error) {
	return loadLifecycleExecutionInputWith(root, selector, inputPath, read, cryptoutilDeploymentGen.LoadBootstrapPreflightInput)
}

func loadLifecycleExecutionInputWith(
	root, selector, inputPath string, read func(string) ([]byte, error),
	load func(string, string, string, func(string) ([]byte, error)) (cryptoutilDeploymentGen.BootstrapPreflightInput, error),
) (cryptoutilDeploymentGen.BootstrapPreflightInput, error) {
	rootedRead := func(name string) ([]byte, error) {
		if !filepath.IsAbs(name) {
			name = filepath.Join(root, name)
		}

		return read(name)
	}

	return load(selector, "", inputPath, rootedRead)
}

func executeLifecycleRequest(
	ctx context.Context, request cryptoutilLifecycleRun.Request,
	config cryptoutilLifecycleRun.KubectlConfig, policy cryptoutilLifecycleRun.Policy,
) (cryptoutilLifecycleRun.Result, error) {
	client, err := cryptoutilLifecycleRun.NewKubectlClient(config, cryptoutilLifecycleRun.ExecRunner{})
	if err != nil {
		return cryptoutilLifecycleRun.Result{}, fmt.Errorf("configure lifecycle client: %w", err)
	}

	result, err := cryptoutilLifecycleRun.Execute(ctx, request, client, policy)
	if err != nil {
		return cryptoutilLifecycleRun.Result{}, fmt.Errorf("execute lifecycle request: %w", err)
	}

	return result, nil
}
