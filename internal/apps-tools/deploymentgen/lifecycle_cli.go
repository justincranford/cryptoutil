// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// lifecycleMain produces non-secret projections or validates a protected import.
// Secret material is never written to stdout, artifacts, or Kubernetes.
func lifecycleMain(
	args []string, stdout, stderr io.Writer,
	readFile func(string) ([]byte, error), writeFile func(string, []byte, os.FileMode) error,
) int {
	return lifecycleMainWithImporter(args, stdout, stderr, readFile, writeFile, ImportPhysicalLifecycleSecrets)
}

func lifecycleMainWithImporter(
	args []string, stdout, stderr io.Writer,
	readFile func(string) ([]byte, error), writeFile func(string, []byte, os.FileMode) error,
	importSecrets func(Model, Inventory, LifecycleInput, string) ([]LifecycleSecret, error),
) int {
	flags := flag.NewFlagSet("lifecycle", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "required non-secret lifecycle operator input JSON")
	runtimeInputPath := flags.String("runtime-input", "", "explicit runtime inputs for the runtime-values format")
	secretRoot := flags.String("secret-root", "", "absolute owner-protected physical directory for import-check")
	selector := flags.String("select", "", "exact release selector, or empty for the complete union")
	format := flags.String("format", lifecycleValuesFormat, "values, secret-plan, revision, import-check, runtime-values, or infrastructure-values (never Secret data)")
	write := flags.Bool(cryptoutilSharedMagic.ScopeWrite, false, "write non-secret lifecycle chart values")
	check := flags.Bool("check", false, "check non-secret lifecycle chart values")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	if flags.NArg() != 0 || *inputPath == "" || (*write && *check) ||
		(*format != lifecycleValuesFormat && *format != lifecycleSecretPlanFormat && *format != lifecycleRevisionFormat &&
			*format != runtimeValuesFormat && *format != infrastructureValuesFormat && *format != lifecycleImportCheckFormat) ||
		((*write || *check) && (*format == lifecycleSecretPlanFormat || *format == lifecycleRevisionFormat || *format == lifecycleImportCheckFormat)) ||
		(*format == lifecycleImportCheckFormat && *secretRoot == "") ||
		(*format != lifecycleImportCheckFormat && *secretRoot != "") ||
		(*format == runtimeValuesFormat && (*selector == "" || *runtimeInputPath == "")) ||
		(*format != runtimeValuesFormat && *runtimeInputPath != "") {
		return failWithDiagnostic(stderr, 2, "lifecycle requires -input and an explicit format; runtime-values also requires -select and -runtime-input\n")
	}

	raw, err := readFile(*inputPath)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read lifecycle input: %v\n", err)
	}

	input, err := DecodeLifecycleInput(raw)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	if *format == lifecycleRevisionFormat {
		revision, err := LifecycleRevision(input)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "%v\n", err)
		}

		if _, err := fmt.Fprintln(stdout, revision); err != nil {
			return failWithDiagnostic(stderr, 1, "write lifecycle revision: %v\n", err)
		}

		return 0
	}

	modelJSON, err := readFile(defaultPlannerPaths().model)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read lifecycle model: %v\n", err)
	}

	model, err := DecodeModel(modelJSON)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	inventory, err := build(model, RegisteredScopes(), *selector)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	values, err := BuildLifecycleChartValues(model, inventory, input)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	if *format == lifecycleImportCheckFormat {
		secrets, err := importSecrets(model, inventory, input, *secretRoot)
		if err != nil {
			// An injected filesystem/transport error may contain private material.
			// The physical adapter exposes classified errors; never print arbitrary
			// material-bearing downstream errors at the public CLI boundary.
			var classified protectedLifecycleImportError
			if errors.As(err, &classified) {
				return failWithDiagnostic(stderr, 1, "protected lifecycle import rejected: %s\n", classified.Error())
			}

			return failWithDiagnostic(stderr, 1, "protected lifecycle import rejected; verify source ownership, paths, ACL/modes, and file sizes\n")
		}

		if _, err := fmt.Fprintf(stdout, "PASS: validated %d protected lifecycle Secrets in memory; no material output or cluster changes\n", len(secrets)); err != nil {
			return failWithDiagnostic(stderr, 1, "write protected import result failed\n")
		}

		return 0
	}

	outputPath := filepath.Join("deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "generated", model.Run, "lifecycle", model.Revision+".json")
	if *selector != "" {
		outputPath = filepath.Join("deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "generated", model.Run, inventory.Releases[0].Helm, "lifecycle-"+model.Revision+".json")
	}

	output, err := lifecycleOutput(*format, *runtimeInputPath, model, inventory, values, readFile)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "encode lifecycle output: %v\n", err)
	}

	if *format == runtimeValuesFormat {
		outputPath = runtimeValuesOutputPath(model.Run, inventory.Releases[0].Helm)
	}

	if *format == infrastructureValuesFormat {
		outputPath = infrastructureValuesOutputPath(model.Run)
	}

	switch {
	case *check:
		current, readErr := readFile(outputPath)
		if readErr != nil || !bytes.Equal(current, output) {
			return failWithDiagnostic(stderr, 1, "lifecycle artifact missing or stale: %s\n", outputPath)
		}
	case *write:
		if err := writeFile(outputPath, output, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly); err != nil {
			return failWithDiagnostic(stderr, 1, "write lifecycle artifact: %v\n", err)
		}

	default:
		if _, err := stdout.Write(output); err != nil {
			return failWithDiagnostic(stderr, 1, "write lifecycle output: %v\n", err)
		}
	}

	return 0
}

func lifecycleOutput(
	format, runtimeInputPath string, model Model, inventory Inventory, lifecycle LifecycleChartValues,
	readFile func(string) ([]byte, error),
) ([]byte, error) {
	var value any

	schemaPath := lifecycleValuesSchemaPath

	switch format {
	case lifecycleSecretPlanFormat:
		return marshalCanonicalJSON(lifecycle.Secrets)
	case runtimeValuesFormat:
		raw, err := readFile(runtimeInputPath)
		if err != nil {
			return nil, fmt.Errorf("read lifecycle runtime input: %w", err)
		}

		raw, err = buildRuntimeChartValuesFromInput(model, inventory, inventory.Releases[0].Key, raw, readFile)
		if err != nil {
			return nil, err
		}

		var runtime RuntimeChartValues

		if err := json.Unmarshal(raw, &runtime); err != nil {
			return nil, fmt.Errorf("decode projected runtime values: %w", err)
		}

		runtime, err = AddLifecycleRuntimeQuota(runtime, lifecycle)
		if err != nil {
			return nil, err
		}

		value, schemaPath = runtime, runtimeChartValuesSchemaPath
	case infrastructureValuesFormat:
		raw, err := buildInfrastructureChartValuesFromInventory(model, inventory, readFile)
		if err != nil {
			return nil, err
		}

		var infra InfrastructureChartValues

		if err := json.Unmarshal(raw, &infra); err != nil {
			return nil, fmt.Errorf("decode projected infrastructure values: %w", err)
		}

		infra, err = AddLifecycleInfrastructureQuota(infra, lifecycle)
		if err != nil {
			return nil, err
		}

		value, schemaPath = infra, infrastructureChartValuesSchemaPath
	default:
		value = lifecycle
	}

	raw, err := marshalCanonicalJSON(value)
	if err != nil {
		return nil, fmt.Errorf("encode lifecycle chart values: %w", err)
	}

	schema, err := readFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("read lifecycle output schema: %w", err)
	}

	if err := validateJSONSchemaDocument("https://cryptoutil.dev/schemas/lifecycle-output.json", "lifecycle output schema", "lifecycle output", schema, raw); err != nil {
		return nil, err
	}

	return raw, nil
}
