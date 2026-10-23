//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("injected output failure")
}

func TestGeneratorCLI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		args       []string
		readFault  string
		badModel   bool
		badYAML    bool
		badBounds  bool
		stale      bool
		writeFail  bool
		stdoutFail bool
		code       int
		message    string
		writes     int
	}{
		{name: "preview", message: "# Kubernetes deployment tree"},
		{name: cryptoutilSharedMagic.CLIHelpCommand, args: []string{"-h"}, message: "Usage of"},
		{name: "json subset", args: []string{"-format", "json", "-select", "s:sm-kms:pg"}, message: `"namespace": "cu-review-s-sm-kms-pg"`},
		{name: "alternate run", args: []string{"-run", "trial", "-select", "p:sm:mem"}, message: "cu-trial-p-sm-mem"},
		{name: "check", args: []string{"-check"}, message: "PASS:"},
		{name: cryptoutilSharedMagic.ScopeWrite, args: []string{"-write"}, message: "Generated", writes: 1},
		{name: "unknown flag", args: []string{"-unknown"}, code: 2, message: "flag provided"},
		{name: "invalid format", args: []string{"-format", "xml"}, code: 2, message: "invalid flags"},
		{name: "positional", args: []string{"extra"}, code: 2, message: "invalid flags"},
		{name: "check and write", args: []string{"-check", "-write"}, code: 2, message: "invalid flags"},
		{name: "subset write prohibited", args: []string{"-write", "-select", "s:sm-kms:pg"}, code: 2, message: "invalid flags"},
		{name: "missing model", readFault: "k8s-model.json", code: 1, message: "read model"},
		{name: "invalid model", badModel: true, code: 1, message: "model:"},
		{name: "invalid bounds", badBounds: true, code: 1, message: "replicas"},
		{name: "missing registry", readFault: "registry.yaml", code: 1, message: "read registry"},
		{name: "invalid registry", badYAML: true, code: 1, message: "registry:"},
		{name: "unknown selector", args: []string{"-select", "p:pki:pg"}, code: 1, message: "matched no release"},
		{name: "missing tree", args: []string{"-check"}, readFault: "k8s-deploy-tree.md", code: 1, message: "missing or stale"},
		{name: "stale tree", args: []string{"-check"}, stale: true, code: 1, message: "missing or stale"},
		{name: "write failure", args: []string{"-write"}, writeFail: true, code: 1, message: "write tree", writes: 1},
		{name: "stdout failure", stdoutFail: true, code: 1, message: "write output"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			inv, err := build(m, r, "")
			require.NoError(t, err)
			tree := render(m, r, inv)
			if tc.stale {
				tree = append(tree, '\n')
			}
			if tc.badBounds {
				m.Replicas.Min = 0
			}
			mBytes, err := json.Marshal(m)
			require.NoError(t, err)
			rBytes, err := yaml.Marshal(r)
			require.NoError(t, err)
			if tc.badModel {
				mBytes = []byte("{")
			}
			if tc.badYAML {
				rBytes = []byte("suites: [")
			}
			read := func(path string) ([]byte, error) {
				name := filepath.Base(path)
				if name == tc.readFault {
					return nil, fmt.Errorf("injected read failure")
				}
				switch name {
				case "k8s-model.json":
					return mBytes, nil
				case "registry.yaml":
					return rBytes, nil
				case "k8s-deploy-tree.md":
					return tree, nil
				default:
					return nil, fmt.Errorf("unexpected path: %s", path)
				}
			}
			writes := 0
			write := func(path string, data []byte, _ os.FileMode) error {
				writes++
				require.Equal(t, "k8s-deploy-tree.md", filepath.Base(path))
				require.Equal(t, tree, data)
				if tc.writeFail {
					return fmt.Errorf("injected write failure")
				}
				return nil
			}
			var stdout, stderr bytes.Buffer
			code := 0
			if tc.stdoutFail {
				code = internalMain(tc.args, failingWriter{}, &stderr, read, write)
			} else {
				code = internalMain(tc.args, &stdout, &stderr, read, write)
			}
			require.Equal(t, tc.code, code)
			require.Equal(t, tc.writes, writes)
			require.Contains(t, stdout.String()+stderr.String(), tc.message)
			if tc.name == "json subset" {
				require.True(t, json.Valid(stdout.Bytes()))
			}
		})
	}
}
