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

func main() {
	os.Exit(internalMain(os.Args[1:], os.Stdout, os.Stderr, os.ReadFile, os.WriteFile))
}

func marshalCanonicalJSON(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var canonical any
	if err := decoder.Decode(&canonical); err != nil {
		return nil, err
	}

	output, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		return nil, err
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
	format := flags.String("format", "tree", "tree or json")
	selector := flags.String("select", "", "one exact tier:scope:mode selector")
	run := flags.String(cryptoutilSharedMagic.WorkflowSubCmdRun, "", "override run identity for read-only previews")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || (*check && *write) ||
		((*check || *write) && (*selector != "" || *run != "" || *format != "tree")) ||
		(*format != "tree" && *format != "json") {
		fmt.Fprintln(stderr, "invalid flags: check/write require the full default inventory; format must be tree or json")
		return 2
	}

	mBytes, err := readFile(filepath.Join(cryptoutilSharedMagic.CICDExcludeDirDocs, "framework-v26", "k8s-model.json"))
	if err != nil {
		fmt.Fprintf(stderr, "read model: %v\n", err)
		return 1
	}
	m, err := decodeModel(mBytes)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *run != "" {
		m.Run = *run
	}
	rBytes, err := readFile(filepath.Join("api", "cryptosuite-registry", "registry.yaml"))
	if err != nil {
		fmt.Fprintf(stderr, "read registry: %v\n", err)
		return 1
	}
	r, err := decodeRegistry(rBytes)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	inv, err := build(m, r, *selector)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	output := render(m, r, inv)
	var inventoryJSON []byte
	if *format == "json" || *check || *write {
		inventoryJSON, err = marshalCanonicalJSON(inv)
		if err != nil {
			fmt.Fprintf(stderr, "encode inventory JSON: %v\n", err)
			return 1
		}
	}
	if *format == "json" {
		output = inventoryJSON
	}
	directory := filepath.Join(cryptoutilSharedMagic.CICDExcludeDirDocs, "framework-v26")
	treePath := filepath.Join(directory, "k8s-deploy-tree.md")
	inventoryPath := filepath.Join(directory, "k8s-expected-inventory.json")
	switch {
	case *check:
		for _, artifact := range []struct {
			path string
			data []byte
		}{{treePath, output}, {inventoryPath, inventoryJSON}} {
			current, readErr := readFile(artifact.path)
			if readErr != nil || !bytes.Equal(current, artifact.data) {
				fmt.Fprintf(stderr, "generated artifact missing or stale: %s (read error: %v)\n", artifact.path, readErr)
				return 1
			}
		}
		if _, err := fmt.Fprintln(stdout, "PASS: generated Kubernetes planning tree and expected inventory match inputs"); err != nil {
			fmt.Fprintf(stderr, "write status: %v\n", err)
			return 1
		}
	case *write:
		for _, artifact := range []struct {
			path string
			data []byte
		}{{treePath, output}, {inventoryPath, inventoryJSON}} {
			if err := writeFile(artifact.path, artifact.data, cryptoutilSharedMagic.CICDOutputFilePermissions); err != nil {
				fmt.Fprintf(stderr, "write artifact %s: %v\n", artifact.path, err)
				return 1
			}
		}
		if _, err := fmt.Fprintf(stdout, "Generated %s\nGenerated %s\n", treePath, inventoryPath); err != nil {
			fmt.Fprintf(stderr, "write status: %v\n", err)
			return 1
		}
	default:
		if _, err := stdout.Write(output); err != nil {
			fmt.Fprintf(stderr, "write output: %v\n", err)
			return 1
		}
	}
	return 0
}
