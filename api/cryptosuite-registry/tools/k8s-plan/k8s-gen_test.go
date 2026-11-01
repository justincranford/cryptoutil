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

const (
	cliFormatFlag           = "-format"
	cliRunFlag              = "-run"
	cliSelectFlag           = "-select"
	cliCheckFlag            = "-check"
	cliWriteFlag            = "-write"
	cliTrialRun             = "trial"
	cliUnsupportedFormat    = "xml"
	cliJSONSubsetTest       = "json subset"
	cliInvalidFlagsMessage  = "invalid flags"
	cliStaleArtifactMessage = "missing or stale"
	cliWriteArtifactMessage = "write artifact"
	cliWriteStatusMessage   = "write status"
)

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("injected output failure")
}

func TestMarshalCanonicalJSON(t *testing.T) {
	t.Parallel()

	output, err := marshalCanonicalJSON(struct {
		Zulu  int `json:"zulu"`
		Alpha int `json:"alpha"`
	}{Zulu: 2, Alpha: 1})
	require.NoError(t, err)
	require.Equal(t, "{\n  \"alpha\": 1,\n  \"zulu\": 2\n}\n", string(output))

	_, err = marshalCanonicalJSON(make(chan int))
	require.Error(t, err)
}

func TestGeneratorUsesDurablePaths(t *testing.T) {
	t.Parallel()

	expectedPaths := plannerPaths{
		model:     filepath.Join("api", "cryptosuite-registry", "k8s", k8sModelFileName),
		registry:  filepath.Join("api", "cryptosuite-registry", k8sRegistryFileName),
		tree:      filepath.Join("api", "cryptosuite-registry", "k8s", k8sDeployTreeFileName),
		inventory: filepath.Join("api", "cryptosuite-registry", "k8s", k8sExpectedInventoryFileName),
	}
	require.Equal(t, expectedPaths, defaultPlannerPaths())

	m, r := fixture(t)
	mBytes, err := json.Marshal(m)
	require.NoError(t, err)

	rBytes, err := yaml.Marshal(r)
	require.NoError(t, err)

	inv, err := build(m, r, "")
	require.NoError(t, err)

	tree := render(m, r, inv)
	inventoryJSON, err := marshalCanonicalJSON(inv)
	require.NoError(t, err)

	artifacts := map[string][]byte{
		expectedPaths.tree:      tree,
		expectedPaths.inventory: inventoryJSON,
	}

	var reads []string

	read := func(path string) ([]byte, error) {
		reads = append(reads, path)

		switch path {
		case expectedPaths.model:
			return mBytes, nil
		case expectedPaths.registry:
			return rBytes, nil
		case expectedPaths.tree, expectedPaths.inventory:
			return artifacts[path], nil
		default:
			return nil, fmt.Errorf("unexpected path: %s", path)
		}
	}

	var stdout, stderr bytes.Buffer

	code := internalMain([]string{cliCheckFlag}, &stdout, &stderr, read,
		func(string, []byte, os.FileMode) error {
			return fmt.Errorf("unexpected write")
		},
	)

	require.Equal(t, 0, code, stderr.String())
	require.Equal(t, []string{
		expectedPaths.model,
		expectedPaths.registry,
		expectedPaths.tree,
		expectedPaths.inventory,
	}, reads)
	require.Contains(t, stdout.String(), "PASS:")
}

func TestGeneratorCLI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		args              []string
		readFault         string
		badModel          bool
		badYAML           bool
		badBounds         bool
		staleArtifact     string
		writeFailArtifact string
		stdoutFail        bool
		stderrFail        bool
		code              int
		message           string
		writes            int
	}{
		{name: "preview", message: "# Kubernetes deployment tree"},
		{name: cryptoutilSharedMagic.CLIHelpCommand, args: []string{"-h"}, message: "Usage of"},
		{name: cliJSONSubsetTest, args: []string{cliFormatFlag, jsonOutputFormat, cliSelectFlag, testServiceSMKMSPGSelector}, message: `"namespace": "cu-review-s-sm-kms-pg"`},
		{name: "alternate run", args: []string{cliRunFlag, cliTrialRun, cliSelectFlag, testProductSMMemSelector}, message: "cu-trial-p-sm-mem"},
		{name: "check", args: []string{cliCheckFlag}, message: "PASS:"},
		{name: cryptoutilSharedMagic.ScopeWrite, args: []string{cliWriteFlag}, message: "Generated", writes: 2},
		{name: "unknown flag", args: []string{"-unknown"}, code: 2, message: "flag provided"},
		{name: "invalid format", args: []string{cliFormatFlag, cliUnsupportedFormat}, code: 2, message: cliInvalidFlagsMessage},
		{name: "invalid flags diagnostic write failure", args: []string{cliFormatFlag, cliUnsupportedFormat}, stderrFail: true, code: 1},
		{name: "positional", args: []string{testExtraIdentity}, code: 2, message: cliInvalidFlagsMessage},
		{name: "check and write", args: []string{cliCheckFlag, cliWriteFlag}, code: 2, message: cliInvalidFlagsMessage},
		{name: "subset write prohibited", args: []string{cliWriteFlag, cliSelectFlag, testServiceSMKMSPGSelector}, code: 2, message: cliInvalidFlagsMessage},
		{name: "alternate run check prohibited", args: []string{cliCheckFlag, cliRunFlag, cliTrialRun}, code: 2, message: cliInvalidFlagsMessage},
		{name: "json format write prohibited", args: []string{cliWriteFlag, cliFormatFlag, jsonOutputFormat}, code: 2, message: cliInvalidFlagsMessage},
		{name: "missing model", readFault: k8sModelFileName, code: 1, message: "read model"},
		{name: "invalid model", badModel: true, code: 1, message: testModelContext},
		{name: "invalid bounds", badBounds: true, code: 1, message: testReplicasKey},
		{name: "missing registry", readFault: k8sRegistryFileName, code: 1, message: "read registry"},
		{name: "invalid registry", badYAML: true, code: 1, message: testRegistryContext},
		{name: "unknown selector", args: []string{cliSelectFlag, "p:pki:pg"}, code: 1, message: "matched no release"},
		{name: "missing tree", args: []string{cliCheckFlag}, readFault: k8sDeployTreeFileName, code: 1, message: cliStaleArtifactMessage},
		{name: "stale tree", args: []string{cliCheckFlag}, staleArtifact: k8sDeployTreeFileName, code: 1, message: cliStaleArtifactMessage},
		{name: "missing expected inventory", args: []string{cliCheckFlag}, readFault: k8sExpectedInventoryFileName, code: 1, message: cliStaleArtifactMessage},
		{name: "stale expected inventory", args: []string{cliCheckFlag}, staleArtifact: k8sExpectedInventoryFileName, code: 1, message: cliStaleArtifactMessage},
		{name: "tree write failure", args: []string{cliWriteFlag}, writeFailArtifact: k8sDeployTreeFileName, code: 1, message: cliWriteArtifactMessage, writes: 1},
		{name: "expected inventory write failure", args: []string{cliWriteFlag}, writeFailArtifact: k8sExpectedInventoryFileName, code: 1, message: cliWriteArtifactMessage, writes: 2},
		{name: "stdout failure", stdoutFail: true, code: 1, message: "write output"},
		{name: "check status stdout failure", args: []string{cliCheckFlag}, stdoutFail: true, code: 1, message: cliWriteStatusMessage},
		{name: "write status stdout failure", args: []string{cliWriteFlag}, stdoutFail: true, code: 1, message: cliWriteStatusMessage, writes: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			inv, err := build(m, r, "")
			require.NoError(t, err)

			tree := render(m, r, inv)
			if tc.badBounds {
				m.Replicas.Min = 0
			}

			inventoryJSON, err := marshalCanonicalJSON(inv)
			require.NoError(t, err)

			artifacts := map[string][]byte{
				k8sDeployTreeFileName:        tree,
				k8sExpectedInventoryFileName: inventoryJSON,
			}
			if tc.staleArtifact != "" {
				artifacts[tc.staleArtifact] = append(append([]byte(nil), artifacts[tc.staleArtifact]...), '\n')
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
				case k8sModelFileName:
					return mBytes, nil
				case k8sRegistryFileName:
					return rBytes, nil
				case k8sDeployTreeFileName, k8sExpectedInventoryFileName:
					return artifacts[name], nil
				default:
					return nil, fmt.Errorf("unexpected path: %s", path)
				}
			}
			writes := 0
			write := func(path string, data []byte, _ os.FileMode) error {
				writes++
				name := filepath.Base(path)
				want, ok := artifacts[name]
				require.True(t, ok, "unexpected artifact path: %s", path)
				require.Equal(t, want, data)

				if tc.writeFailArtifact == name {
					return fmt.Errorf("injected write failure")
				}

				return nil
			}

			var stdout, stderr bytes.Buffer

			code := 0
			if tc.stdoutFail {
				code = internalMain(tc.args, failingWriter{}, &stderr, read, write)
			} else if tc.stderrFail {
				code = internalMain(tc.args, &stdout, failingWriter{}, read, write)
			} else {
				code = internalMain(tc.args, &stdout, &stderr, read, write)
			}

			require.Equal(t, tc.code, code)
			require.Equal(t, tc.writes, writes)

			if tc.message != "" {
				require.Contains(t, stdout.String()+stderr.String(), tc.message)
			}

			if tc.name == cliJSONSubsetTest {
				require.True(t, json.Valid(stdout.Bytes()))
				require.Equal(t, byte('\n'), stdout.Bytes()[len(stdout.Bytes())-1])
			}
		})
	}
}
