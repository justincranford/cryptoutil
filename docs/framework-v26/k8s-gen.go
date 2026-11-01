//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

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

const (
	k8sModelFileName             = "k8s-model.json"
	k8sRegistryFileName          = "registry.yaml"
	k8sDeployTreeFileName        = "k8s-deploy-tree.md"
	k8sExpectedInventoryFileName = "k8s-expected-inventory.json"
	jsonOutputFormat             = "json"
	treeOutputFormat             = "tree"
)

func main() {
	os.Exit(internalMain(os.Args[1:], os.Stdout, os.Stderr, os.ReadFile, os.WriteFile))
}

func failWithDiagnostic(stderr io.Writer, exitCode int, format string, args ...any) int {
	if _, err := fmt.Fprintf(stderr, format, args...); err != nil {
		return 1
	}

	return exitCode
}

func marshalCanonicalJSON(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical JSON input: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var canonical any
	if err := decoder.Decode(&canonical); err != nil {
		return nil, fmt.Errorf("decode canonical JSON input: %w", err)
	}

	output, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("indent canonical JSON output: %w", err)
	}

	return append(output, '\n'), nil
}

func internalMain(args []string, stdout, stderr io.Writer,
	readFile func(string) ([]byte, error), writeFile func(string, []byte, os.FileMode) error,
) int {
	flags := flag.NewFlagSet("k8s-plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	check := flags.Bool("check", false, "compare full generated artifacts to checked-in files")
	write := flags.Bool(cryptoutilSharedMagic.ScopeWrite, false, "write the full generated artifacts (no cluster changes)")
	format := flags.String("format", treeOutputFormat, "tree or json")
	selector := flags.String("select", "", "one exact tier:scope:mode selector")
	run := flags.String(cryptoutilSharedMagic.WorkflowSubCmdRun, "", "override run identity for read-only previews")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	if flags.NArg() != 0 || (*check && *write) ||
		((*check || *write) && (*selector != "" || *run != "" || *format != treeOutputFormat)) ||
		(*format != treeOutputFormat && *format != jsonOutputFormat) {
		return failWithDiagnostic(stderr, 2, "invalid flags: check/write require the full default inventory; format must be tree or json\n")
	}

	mBytes, err := readFile(filepath.Join(cryptoutilSharedMagic.CICDExcludeDirDocs, "framework-v26", k8sModelFileName))
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

	rBytes, err := readFile(filepath.Join("api", "cryptosuite-registry", k8sRegistryFileName))
	if err != nil {
		return failWithDiagnostic(stderr, 1, "read registry: %v\n", err)
	}

	r, err := decodeRegistry(rBytes)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	inv, err := build(m, r, *selector)
	if err != nil {
		return failWithDiagnostic(stderr, 1, "%v\n", err)
	}

	output := render(m, r, inv)

	var inventoryJSON []byte
	if *format == jsonOutputFormat || *check || *write {
		inventoryJSON, err = marshalCanonicalJSON(inv)
		if err != nil {
			return failWithDiagnostic(stderr, 1, "encode inventory JSON: %v\n", err)
		}
	}

	if *format == jsonOutputFormat {
		output = inventoryJSON
	}

	directory := filepath.Join(cryptoutilSharedMagic.CICDExcludeDirDocs, "framework-v26")
	treePath := filepath.Join(directory, k8sDeployTreeFileName)
	inventoryPath := filepath.Join(directory, k8sExpectedInventoryFileName)

	switch {
	case *check:
		for _, artifact := range []struct {
			path string
			data []byte
		}{{treePath, output}, {inventoryPath, inventoryJSON}} {
			current, readErr := readFile(artifact.path)
			if readErr != nil || !bytes.Equal(current, artifact.data) {
				return failWithDiagnostic(stderr, 1, "generated artifact missing or stale: %s (read error: %v)\n", artifact.path, readErr)
			}
		}

		if _, err := fmt.Fprintln(stdout, "PASS: generated Kubernetes planning tree and expected inventory match inputs"); err != nil {
			return failWithDiagnostic(stderr, 1, "write status: %v\n", err)
		}
	case *write:
		for _, artifact := range []struct {
			path string
			data []byte
		}{{treePath, output}, {inventoryPath, inventoryJSON}} {
			if err := writeFile(artifact.path, artifact.data, cryptoutilSharedMagic.CICDOutputFilePermissions); err != nil {
				return failWithDiagnostic(stderr, 1, "write artifact %s: %v\n", artifact.path, err)
			}
		}

		if _, err := fmt.Fprintf(stdout, "Generated %s\nGenerated %s\n", treePath, inventoryPath); err != nil {
			return failWithDiagnostic(stderr, 1, "write status: %v\n", err)
		}
	default:
		if _, err := stdout.Write(output); err != nil {
			return failWithDiagnostic(stderr, 1, "write output: %v\n", err)
		}
	}

	return 0
}
