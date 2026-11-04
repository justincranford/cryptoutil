// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	k8sModelFileName             = "k8s-model.json"
	k8sDeployTreeFileName        = "k8s-deploy-tree.md"
	k8sExpectedInventoryFileName = "k8s-expected-inventory.json"
	jsonOutputFormat             = "json"
	treeOutputFormat             = "tree"
	lifecycleIdentitiesFormat    = "lifecycle-identities"
)

type plannerPaths struct {
	model     string
	tree      string
	inventory string
}

type generatedArtifact struct {
	path string
	data []byte
}

func defaultPlannerPaths() plannerPaths {
	modelDir := filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment)

	return plannerPaths{
		model:     filepath.Join(modelDir, k8sModelFileName),
		tree:      filepath.Join(modelDir, k8sDeployTreeFileName),
		inventory: filepath.Join(modelDir, k8sExpectedInventoryFileName),
	}
}

// Main runs the deployment generator with injected process I/O.
func Main(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	readFile func(string) ([]byte, error),
	writeFile func(string, []byte, os.FileMode) error,
) int {
	if len(args) > 0 && args[0] == developmentRuntimePlanCommand {
		return developmentRuntimePlanMain(args[1:], stdout, stderr, readFile, runBootstrapDockerCommand, net.Listen)
	}

	if len(args) > 0 && args[0] == bootstrapPreflightCommand {
		return bootstrapPreflightMain(args[1:], stdout, stderr, readFile)
	}

	if len(args) > 0 && args[0] == "lifecycle" {
		return lifecycleMain(args[1:], stdout, stderr, readFile, writeFile)
	}

	return internalMain(args, stdout, stderr, readFile, writeFile)
}

func failWithDiagnostic(stderr io.Writer, exitCode int, format string, args ...any) int {
	if _, err := fmt.Fprintf(stderr, format, args...); err != nil {
		return 1
	}

	return exitCode
}

func marshalCanonicalJSON(value any) ([]byte, error) {
	return marshalCanonicalJSONWith(
		value,
		json.Marshal,
		decodeCanonicalJSON,
		func(value any) ([]byte, error) {
			return json.MarshalIndent(value, "", "  ")
		},
	)
}

func marshalCanonicalJSONWith(
	value any,
	marshalFn func(any) ([]byte, error),
	decodeFn func([]byte) (any, error),
	indentFn func(any) ([]byte, error),
) ([]byte, error) {
	raw, err := marshalFn(value)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical JSON input: %w", err)
	}

	canonical, err := decodeFn(raw)
	if err != nil {
		return nil, fmt.Errorf("decode canonical JSON input: %w", err)
	}

	output, err := indentFn(canonical)
	if err != nil {
		return nil, fmt.Errorf("indent canonical JSON output: %w", err)
	}

	return append(output, '\n'), nil
}

func decodeCanonicalJSON(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var canonical any
	if err := decoder.Decode(&canonical); err != nil {
		return nil, fmt.Errorf("decode canonical JSON value: %w", err)
	}

	return canonical, nil
}

func internalMain(args []string, stdout, stderr io.Writer,
	readFile func(string) ([]byte, error), writeFile func(string, []byte, os.FileMode) error,
) int {
	return internalMainWithJSON(args, stdout, stderr, readFile, writeFile, marshalCanonicalJSON)
}

func internalMainWithJSON(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	readFile func(string) ([]byte, error),
	writeFile func(string, []byte, os.FileMode) error,
	marshalJSON func(any) ([]byte, error),
) int {
	return internalMainWithJSONAndFS(args, stdout, stderr, readFile, writeFile, marshalJSON, os.DirFS("."))
}

func internalMainWithJSONAndFS(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	readFile func(string) ([]byte, error),
	writeFile func(string, []byte, os.FileMode) error,
	marshalJSON func(any) ([]byte, error),
	source fs.FS,
) int {
	flags := flag.NewFlagSet("k8s-plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	check := flags.Bool("check", false, "compare generated artifacts to expected bytes")
	write := flags.Bool(cryptoutilSharedMagic.ScopeWrite, false, "write generated artifacts (no cluster changes)")
	format := flags.String(
		"format", treeOutputFormat,
		"tree, json, charts, runtime-values, infrastructure-values, or lifecycle-identities",
	)
	selector := flags.String("select", "", "one exact tier:scope:mode selector")
	run := flags.String(cryptoutilSharedMagic.WorkflowSubCmdRun, "", "override run identity for read-only previews")
	runtimeInputPath := flags.String("runtime-input", "", "explicit non-secret runtime chart input JSON")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	formatSupported := *format == treeOutputFormat || *format == jsonOutputFormat ||
		*format == runtimeValuesFormat || *format == infrastructureValuesFormat ||
		*format == lifecycleIdentitiesFormat || *format == chartTemplatesFormat
	invalidValueMode := (*format == runtimeValuesFormat && (*selector == "" || *runtimeInputPath == "")) ||
		(*format != runtimeValuesFormat && *runtimeInputPath != "")

	invalidArtifactMode := (*check || *write) && ((*format == treeOutputFormat && (*selector != "" || *run != "")) ||
		(*format == jsonOutputFormat) ||
		(*format == lifecycleIdentitiesFormat) ||
		(*format == infrastructureValuesFormat && (*selector != "" || *run != "")) ||
		(*format == chartTemplatesFormat && (*selector != "" || *run != "")))
	if flags.NArg() != 0 || (*check && *write) || !formatSupported || invalidValueMode || invalidArtifactMode {
		return failWithDiagnostic(
			stderr,
			2,
			"invalid flags: use tree/json previews, runtime-values with -select and -runtime-input, or infrastructure-values; check/write require a complete generated artifact\n",
		)
	}

	paths := defaultPlannerPaths()

	mBytes, err := readFile(paths.model)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read model: %v\n", err)
	}

	m, err := decodeModel(mBytes)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	if *run != "" {
		m.Run = *run
	}

	scopes := RegisteredScopes()

	inv, err := build(m, scopes, *selector)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	if *format == runtimeValuesFormat && len(inv.Releases) != 1 {
		return failWithDiagnostic(stderr, 1, "runtime-values requires exactly one selected release\n")
	}

	var inventoryJSON []byte
	if *format == jsonOutputFormat || ((*check || *write) && *format == treeOutputFormat) {
		inventoryJSON, err = marshalJSON(inv)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "encode inventory JSON: %v\n", err)
		}
	}

	var (
		output    []byte
		artifacts []generatedArtifact
	)

	switch *format {
	case treeOutputFormat:
		output = render(m, scopes, inv)
	case jsonOutputFormat:
		output = inventoryJSON
	case lifecycleIdentitiesFormat:
		lifecycle, lifecycleErr := BuildLifecycleInventory(inv)
		if lifecycleErr != nil {
			return failWithDiagnostic(stderr, 1, "generate lifecycle identities: %v\n", lifecycleErr)
		}

		output, err = marshalJSON(lifecycle)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "encode lifecycle identities: %v\n", err)
		}
	case runtimeValuesFormat:
		inputJSON, readErr := readFile(*runtimeInputPath)
		if readErr != nil {
			return failWithDiagnostic(stderr, 1, "read runtime chart input: %v\n", readErr)
		}

		output, err = buildRuntimeChartValuesFromInput(m, inv, inv.Releases[0].Key, inputJSON, readFile)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "generate runtime chart values: %v\n", err)
		}
	case infrastructureValuesFormat:
		output, err = buildInfrastructureChartValuesFromInventory(m, inv, readFile)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "generate infrastructure chart values: %v\n", err)
		}
	case chartTemplatesFormat:
		artifacts, err = canonicalChartArtifacts(source)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "collect canonical chart artifacts: %v\n", err)
		}

		artifactPaths := make([]string, len(artifacts))
		for index, artifact := range artifacts {
			artifactPaths[index] = artifact.path
		}

		output, err = marshalJSON(artifactPaths)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "encode chart artifact paths: %v\n", err)
		}
	}

	if *format != chartTemplatesFormat {
		artifacts = generatedArtifacts(*format, m, inv, paths, output, inventoryJSON)
	}

	switch {
	case *check:
		for _, artifact := range artifacts {
			current, readErr := readFile(artifact.path)
			if readErr != nil || !bytes.Equal(current, artifact.data) {
				return failWithDiagnostic(stderr, 1, "generated artifact missing or stale: %s (read error: %v)\n", artifact.path, readErr)
			}
		}

		if *format == chartTemplatesFormat {
			if err := validateCanonicalChartMirror(source, artifacts); err != nil {
				return failWithDiagnostic(stderr, 1, "validate generated chart mirror: %v\n", err)
			}
		}

		if _, err := fmt.Fprintln(stdout, "PASS: generated Kubernetes artifact matches inputs"); err != nil {
			return failWithDiagnostic(stderr, 1, "write status: %v\n", err)
		}
	case *write:
		for _, artifact := range artifacts {
			if err := writeFile(artifact.path, artifact.data, cryptoutilSharedMagic.CICDOutputFilePermissions); err != nil {
				return failWithDiagnostic(stderr, 1, "write artifact %s: %v\n", artifact.path, err)
			}
		}

		for _, artifact := range artifacts {
			if _, err := fmt.Fprintf(stdout, "Generated %s\n", artifact.path); err != nil {
				return failWithDiagnostic(stderr, 1, "write status: %v\n", err)
			}
		}
	default:
		if _, err := stdout.Write(output); err != nil {
			return failWithDiagnostic(stderr, 1, "write output: %v\n", err)
		}
	}

	return 0
}

func generatedArtifacts(
	format string,
	model Model,
	inventory Inventory,
	paths plannerPaths,
	output, inventoryJSON []byte,
) []generatedArtifact {
	switch format {
	case treeOutputFormat:
		return []generatedArtifact{{path: paths.tree, data: output}, {path: paths.inventory, data: inventoryJSON}}
	case runtimeValuesFormat:
		if len(inventory.Releases) == 1 {
			return []generatedArtifact{{
				path: runtimeValuesOutputPath(model.Run, inventory.Releases[0].Helm),
				data: output,
			}}
		}
	case infrastructureValuesFormat:
		return []generatedArtifact{{path: infrastructureValuesOutputPath(model.Run), data: output}}
	}

	return nil
}
