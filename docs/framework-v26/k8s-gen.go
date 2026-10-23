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

func internalMain(args []string, stdout, stderr io.Writer,
	readFile func(string) ([]byte, error), writeFile func(string, []byte, os.FileMode) error,
) int {
	flags := flag.NewFlagSet("k8s-plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	check := flags.Bool("check", false, "compare the full generated tree to the checked-in document")
	write := flags.Bool(cryptoutilSharedMagic.ScopeWrite, false, "write the full generated tree (no cluster changes)")
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
		fmt.Fprintln(stderr, "invalid flags: check/write require the full default tree; format must be tree or json")
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
	if *format == "json" {
		output, err = json.MarshalIndent(inv, "", "    ")
		if err != nil {
			fmt.Fprintf(stderr, "encode inventory: %v\n", err)
			return 1
		}
		output = append(output, '\n')
	}
	path := filepath.Join(cryptoutilSharedMagic.CICDExcludeDirDocs, "framework-v26", "k8s-deploy-tree.md")
	switch {
	case *check:
		current, readErr := readFile(path)
		if readErr != nil || !bytes.Equal(current, output) {
			fmt.Fprintf(stderr, "generated tree missing or stale: %s (read error: %v)\n", path, readErr)
			return 1
		}
		fmt.Fprintln(stdout, "PASS: generated Kubernetes planning tree matches inputs")
	case *write:
		if err := writeFile(path, output, 0o644); err != nil {
			fmt.Fprintf(stderr, "write tree: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Generated "+path)
	default:
		if _, err := stdout.Write(output); err != nil {
			fmt.Fprintf(stderr, "write output: %v\n", err)
			return 1
		}
	}
	return 0
}
