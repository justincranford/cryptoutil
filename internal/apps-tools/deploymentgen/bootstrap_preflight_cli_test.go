// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testBootstrapFreshRun           = "fresh"
	testBootstrapFreshAdmission     = "explicit run admission"
	testBootstrapInputPath          = "bootstrap-runtime-input.json"
	testBootstrapSelect             = "s:sm-kms:pg"
	testBootstrapHost               = "14 8326410240"
	testBootstrapCPUReserveFlag     = "-reserve-cpu-milli"
	testBootstrapMemoryReserveFlag  = "-reserve-memory-bytes"
	testBootstrapLifecycleFlag      = "-lifecycle-input"
	testBootstrapLifecycleFile      = "bootstrap-lifecycle.json"
	testBootstrapMalformedLifecycle = "malformed admission lifecycle"
)

func TestBootstrapPreflightCLIAdmission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		collision  bool
		failInput  bool
		failOutput bool
		wantCode   int
		wantErr    string
		wantCalls  int
		wantCloses int
	}{
		{name: "selected admission", args: []string{cliSelectFlag, testBootstrapSelect}, wantCalls: 4, wantCloses: 2},
		{name: testBootstrapFreshAdmission, args: []string{cliSelectFlag, testBootstrapSelect, "-run", testBootstrapFreshRun}, wantCalls: 4, wantCloses: 2},
		{name: "missing admission lifecycle", args: []string{cliSelectFlag, testBootstrapSelect, testBootstrapLifecycleFlag, "absent-lifecycle.json"}, wantCode: 1, wantErr: "read admission lifecycle"},
		{name: testBootstrapMalformedLifecycle, args: []string{cliSelectFlag, testBootstrapSelect}, wantCode: 1, wantErr: "decode admission lifecycle"},
		{name: "full matrix rejects measured host", wantCode: 1, wantErr: testBootstrapInsufficientHost, wantCalls: 4},
		{name: "stopped resource collision before ports", args: []string{cliSelectFlag, testBootstrapSelect}, collision: true, wantCode: 1, wantErr: testBootstrapContainerExists, wantCalls: 1},
		{name: "input rejected before Docker", failInput: true, wantCode: 1, wantErr: "bootstrap preflight input"},
		{name: "output error", args: []string{cliSelectFlag, testBootstrapSelect}, failOutput: true, wantCode: 1, wantErr: "write bootstrap admission", wantCalls: 4, wantCloses: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			selector := testBootstrapSelect
			if len(test.args) == 0 {
				selector = ""
			}

			runIdentity := ""
			if test.name == testBootstrapFreshAdmission {
				runIdentity = testBootstrapFreshRun
			}

			readFile := bootstrapCLIReadFixture(t, selector, runIdentity)
			input, err := LoadBootstrapPreflightInput(testBootstrapSelect, "", testBootstrapInputPath, readFile)
			require.NoError(t, err)
			plan, err := BuildLocalK3sPlan(input.Model, input.Lock)
			require.NoError(t, err)

			if test.failInput {
				readFile = func(string) ([]byte, error) { return nil, errors.New("injected input failure") }
			}

			if test.name == testBootstrapMalformedLifecycle {
				readValid := readFile
				readFile = func(name string) ([]byte, error) {
					if name == testBootstrapLifecycleFile {
						return []byte("{"), nil
					}

					return readValid(name)
				}
			}

			args := append([]string{
				cliRuntimeInputFlag, testBootstrapInputPath,
				testBootstrapLifecycleFlag, testBootstrapLifecycleFile,
				testBootstrapCPUReserveFlag, "1000", testBootstrapMemoryReserveFlag, "1073741824",
			}, test.args...)
			calls, closes := 0, 0
			runner := func(ctx context.Context, command string, args ...string) ([]byte, error) {
				calls++

				require.Equal(t, "docker", command)

				_, bounded := ctx.Deadline()
				require.True(t, bounded)

				if calls == 1 {
					require.Equal(t, []string{bootstrapContainerKind, bootstrapListCommand, bootstrapAllContainersFlag, dockerOutputFormatFlag, bootstrapJSONFormat}, args)

					if test.collision {
						return json.Marshal(map[string]string{testBootstrapNamesKey: plan.Server.Container})
					}
				}

				if calls == 4 {
					require.Equal(t, []string{dockerInfoCommand, dockerOutputFormatFlag, dockerHostFormat}, args)

					return []byte(testBootstrapHost), nil
				}

				return nil, nil
			}
			listen := func(string, string) (net.Listener, error) {
				return &bootstrapTestListener{close: func() error {
					closes++

					return nil
				}}, nil
			}

			var stdout, stderr bytes.Buffer

			var output io.Writer = &stdout
			if test.failOutput {
				output = bootstrapCLIErrorWriter{}
			}

			code := bootstrapPreflightMainWithIO(args, output, &stderr, readFile, runner, listen)
			require.Equal(t, test.wantCode, code, stderr.String())
			require.Equal(t, test.wantCalls, calls)
			require.Equal(t, test.wantCloses, closes)

			if test.wantErr != "" {
				require.Contains(t, stderr.String(), test.wantErr)
			} else {
				require.Empty(t, stderr.String())
				require.Contains(t, stdout.String(), "host-total admission")
				require.Contains(t, stdout.String(), "app-pods=2 infrastructure-pods=5")
				require.Contains(t, stdout.String(), "no cluster changes")
				require.Contains(t, stdout.String(), "node placement/readiness not verified")
			}
		})
	}
}

func TestBootstrapPreflightCLIFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{name: cryptoutilSharedMagic.CLIHelpCommand, args: []string{"-help"}},
		{name: "bootstrap mutation flag forbidden", args: []string{cliWriteFlag}, wantCode: 2},
		{name: "no input", wantCode: 2},
		{name: "no reserves", args: []string{cliRuntimeInputFlag, testBootstrapInputPath}, wantCode: 2},
		{name: "zero timeout", args: []string{cliRuntimeInputFlag, testBootstrapInputPath, testBootstrapCPUReserveFlag, "1000", testBootstrapMemoryReserveFlag, "1073741824", "-timeout", "0"}, wantCode: 2},
		{name: "positional argument", args: []string{"extra"}, wantCode: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := Main(append([]string{bootstrapPreflightCommand}, test.args...), &stdout, &stderr,
				func(string) ([]byte, error) {
					t.Fatal("invalid flags must not read inputs or execute Docker")

					return nil, nil
				}, func(string, []byte, os.FileMode) error {
					t.Fatal("preflight must never write artifacts")

					return nil
				})
			require.Equal(t, test.wantCode, code)
		})
	}
}

func TestLoadBootstrapPreflightInputFailures(t *testing.T) {
	t.Parallel()

	lockPath := filepath.Join(filepath.Dir(defaultPlannerPaths().model), k8sToolchainLockFileName)
	incompleteRuntime := chartValuesInputFixture(t)
	incompleteRuntime.Runtime.Workloads = nil
	incompleteRuntimeJSON, err := marshalChartValuesInput(incompleteRuntime)
	require.NoError(t, err)

	tests := []struct {
		name        string
		failPath    string
		editPath    string
		editData    []byte
		selector    string
		nilReader   bool
		broadSchema bool
		wantErr     string
	}{
		{name: "missing reader", nilReader: true, wantErr: "file reader is required"},
		{name: "model read", failPath: defaultPlannerPaths().model, wantErr: "read bootstrap model"},
		{name: "unparseable bootstrap model", editPath: defaultPlannerPaths().model, editData: []byte("{"), wantErr: testModelContext},
		{name: "lock read", failPath: lockPath, wantErr: "read bootstrap toolchain lock"},
		{name: "invalid lock", editPath: lockPath, editData: []byte("{}"), wantErr: "toolchain lock has invalid"},
		{name: "invalid selector", selector: "invalid-selector", wantErr: "selector"},
		{name: "runtime input read", failPath: testBootstrapInputPath, wantErr: "read bootstrap runtime input"},
		{name: "runtime schema read", failPath: chartValuesInputSchemaPath, wantErr: chartValuesInputSchemaReadError},
		{name: "runtime schema mismatch", editPath: testBootstrapInputPath, editData: []byte("{}"), wantErr: "schema"},
		{name: "runtime strict decode", editPath: testBootstrapInputPath, editData: []byte(`{"unknown":true}`), broadSchema: true, wantErr: "unknown field"},
		{name: "runtime projection incomplete", editPath: testBootstrapInputPath, editData: incompleteRuntimeJSON, broadSchema: true, wantErr: "runtime"},
		{name: "infrastructure read", failPath: infrastructureChartDefaultsPath, wantErr: "read bootstrap infrastructure defaults"},
		{name: "invalid infrastructure defaults", editPath: infrastructureChartDefaultsPath, editData: []byte("{"), wantErr: "infrastructure"},
		{name: "infrastructure projection invalid", editPath: infrastructureChartDefaultsPath, editData: []byte("{}"), wantErr: "storage class"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			baseReader := bootstrapCLIReadFixture(t)

			reader := func(path string) ([]byte, error) {
				if test.broadSchema && path == chartValuesInputSchemaPath {
					return []byte("{}"), nil
				}

				if path == test.failPath {
					return nil, errors.New("injected bootstrap read")
				}

				if path == test.editPath {
					return test.editData, nil
				}

				return baseReader(path)
			}
			if test.nilReader {
				reader = nil
			}

			selector := testBootstrapSelect
			if test.selector != "" {
				selector = test.selector
			}

			_, err := loadBootstrapPreflightInput(selector, "", testBootstrapInputPath, reader)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func bootstrapCLIReadFixture(t *testing.T, selectors ...string) func(string) ([]byte, error) {
	t.Helper()

	selector := testBootstrapSelect
	if len(selectors) > 0 {
		selector = selectors[0]
	}

	model, _, lifecycle := lifecycleDeployFixture(t, selector, func(model *Model) {
		if len(selectors) > 1 && selectors[1] != "" {
			model.Run = selectors[1]
		}
	})
	lifecycleJSON, err := json.Marshal(lifecycle)
	require.NoError(t, err)
	modelJSON, err := json.Marshal(model)
	require.NoError(t, err)
	chartInput := chartValuesInputFixture(t)
	inventory, err := Build(model, "")
	require.NoError(t, err)

	for _, release := range inventory.Releases {
		for _, workload := range release.Workloads {
			key := runtimeWorkloadInputKey(release.Key, workload.PSID, workload.Pool)
			value := chartInput.Runtime.Workloads[key]
			value.MaterializerConfig = materializerInputFixture(release, workload)
			chartInput.Runtime.Workloads[key] = value
		}
	}

	inputJSON, err := marshalChartValuesInput(chartInput)
	require.NoError(t, err)

	return func(path string) ([]byte, error) {
		switch path {
		case defaultPlannerPaths().model:
			return modelJSON, nil
		case testBootstrapInputPath:
			return inputJSON, nil
		case testBootstrapLifecycleFile:
			return lifecycleJSON, nil
		default:
			return os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
		}
	}
}

type bootstrapCLIErrorWriter struct{}

func (bootstrapCLIErrorWriter) Write([]byte) (int, error) {
	return 0, errors.New("injected bootstrap output failure")
}

func TestRunBootstrapDockerCommandMissingExecutable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{name: "missing executable", command: "cryptoutil-preflight-executable-that-does-not-exist", wantErr: true},
		{name: "bounded read-only command", command: "go", args: []string{cryptoutilSharedMagic.CLIVersionCommand}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			output, err := runBootstrapDockerCommand(ctx, test.command, test.args...)
			if test.wantErr {
				require.ErrorContains(t, err, "execute bootstrap probe")
			} else {
				require.NoError(t, err)
				require.Contains(t, string(output), cryptoutilSharedMagic.DevSetupGoCheckCmd)
			}
		})
	}
}
