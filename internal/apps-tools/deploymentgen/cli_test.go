// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	cliFormatFlag           = "-format"
	cliRuntimeInputFlag     = "-runtime-input"
	cliRuntimeInputPath     = "inputs.json"
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
	cliUnknownFlag          = "-unknown"
	cliGeneratedMessage     = "Generated"
	cliPassMessage          = "PASS:"
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

	_, err = decodeCanonicalJSON([]byte("{"))
	require.Error(t, err)
}

func TestMarshalCanonicalJSONInjectedErrors(t *testing.T) {
	t.Parallel()

	injected := errors.New("injected canonical JSON failure")

	tests := []struct {
		name      string
		marshalFn func(any) ([]byte, error)
		decodeFn  func([]byte) (any, error)
		indentFn  func(any) ([]byte, error)
		wantError string
	}{
		{
			name: "marshal",
			marshalFn: func(any) ([]byte, error) {
				return nil, injected
			},
			decodeFn: func([]byte) (any, error) {
				return nil, nil
			},
			indentFn: func(any) ([]byte, error) {
				return nil, nil
			},
			wantError: "marshal canonical JSON input",
		},
		{
			name: "decode",
			marshalFn: func(any) ([]byte, error) {
				return []byte("{}"), nil
			},
			decodeFn: func([]byte) (any, error) {
				return nil, injected
			},
			indentFn: func(any) ([]byte, error) {
				return nil, nil
			},
			wantError: "decode canonical JSON input",
		},
		{
			name: "indent",
			marshalFn: func(any) ([]byte, error) {
				return []byte("{}"), nil
			},
			decodeFn: func([]byte) (any, error) {
				return map[string]any{}, nil
			},
			indentFn: func(any) ([]byte, error) {
				return nil, injected
			},
			wantError: "indent canonical JSON output",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := marshalCanonicalJSONWith(struct{}{}, tc.marshalFn, tc.decodeFn, tc.indentFn)
			require.ErrorContains(t, err, tc.wantError)
			require.ErrorIs(t, err, injected)
		})
	}
}

func TestMainInvalidArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := Main([]string{cliUnknownFlag}, &stdout, &stderr, nil, nil)

	require.Equal(t, 2, code)
	require.Contains(t, stderr.String(), "flag provided")
}

func TestInternalMainCanonicalJSONFailure(t *testing.T) {
	t.Parallel()

	m, _ := fixture(t)
	modelJSON, err := json.Marshal(m)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer

	code := internalMainWithJSON(
		[]string{cliFormatFlag, jsonOutputFormat},
		&stdout,
		&stderr,
		func(string) ([]byte, error) {
			return modelJSON, nil
		},
		func(string, []byte, os.FileMode) error {
			return nil
		},
		func(any) ([]byte, error) {
			return nil, errors.New("injected JSON encoding failure")
		},
	)

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "encode inventory JSON")
	require.Contains(t, stderr.String(), "injected JSON encoding failure")
}

func TestGeneratorUsesDurablePaths(t *testing.T) {
	t.Parallel()

	expectedPaths := plannerPaths{
		model:     filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, k8sModelFileName),
		tree:      filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, k8sDeployTreeFileName),
		inventory: filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, k8sExpectedInventoryFileName),
	}
	require.Equal(t, expectedPaths, defaultPlannerPaths())

	m, scopes := fixture(t)
	mBytes, err := json.Marshal(m)
	require.NoError(t, err)

	inv, err := build(m, scopes, "")
	require.NoError(t, err)

	tree := render(m, scopes, inv)
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
		badBounds         bool
		staleArtifact     string
		writeFailArtifact string
		stdoutFail        bool
		stderrFail        bool
		code              int
		message           string
		writes            int
	}{
		{name: testPreviewCaseName, message: "# Kubernetes deployment tree"},
		{name: cryptoutilSharedMagic.CLIHelpCommand, args: []string{"-h"}, message: "Usage of"},
		{name: cliJSONSubsetTest, args: []string{cliFormatFlag, jsonOutputFormat, cliSelectFlag, testServiceSMKMSPGSelector}, message: `"namespace": "cu-review-s-sm-kms-pg"`},
		{name: "alternate run", args: []string{cliRunFlag, cliTrialRun, cliSelectFlag, testProductSMMemSelector}, message: "cu-trial-p-sm-mem"},
		{name: testCheckCaseName, args: []string{cliCheckFlag}, message: cliPassMessage},
		{name: cryptoutilSharedMagic.ScopeWrite, args: []string{cliWriteFlag}, message: cliGeneratedMessage, writes: 2},
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
		{name: "unknown selector", args: []string{cliSelectFlag, testUnknownScopeSelector}, code: 1, message: testNoMatchedReleaseError},
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
			m, scopes := fixture(t)
			inv, err := build(m, scopes, "")
			require.NoError(t, err)

			tree := render(m, scopes, inv)
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

			if tc.badModel {
				mBytes = []byte("{")
			}

			read := func(path string) ([]byte, error) {
				name := filepath.Base(path)
				if name == tc.readFault {
					return nil, fmt.Errorf("injected read failure")
				}

				switch name {
				case k8sModelFileName:
					return mBytes, nil
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
