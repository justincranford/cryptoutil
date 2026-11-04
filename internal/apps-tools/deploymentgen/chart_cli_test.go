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

func TestBuildChartValuesFromCLIInputs(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	fullInventory, err := build(model, scopes, "")
	require.NoError(t, err)

	inputJSON, err := marshalChartValuesInput(chartValuesInputFixture(t))
	require.NoError(t, err)

	readFile := func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
	}

	for _, release := range fullInventory.Releases {
		t.Run(release.Key, func(t *testing.T) {
			t.Parallel()

			selectedInventory, buildErr := build(model, scopes, release.Key)
			require.NoError(t, buildErr)

			valuesJSON, valuesErr := buildRuntimeChartValuesFromInput(
				model, selectedInventory, release.Key, inputJSON, readFile,
			)
			require.NoError(t, valuesErr)
			require.NoError(t, ValidateRuntimeChartValuesSchema(
				mustReadFile(t, runtimeChartValuesSchemaPath), valuesJSON,
			))
		})
	}

	infrastructureJSON, err := buildInfrastructureChartValuesFromInventory(model, fullInventory, readFile)
	require.NoError(t, err)
	require.NoError(t, ValidateInfrastructureChartValuesSchema(
		mustReadFile(t, infrastructureChartValuesSchemaPath), infrastructureJSON,
	))
}

func TestBuildRuntimeChartValuesFromInputFailures(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	selected := inventory.Releases[0]
	selectedInventory := inventoryWithRelease(inventory, selected.Key)

	validInput, err := marshalChartValuesInput(chartValuesInputFixture(t))
	require.NoError(t, err)
	invalidResources := chartValuesInputFixture(t)
	workload := selected.Workloads[0]
	inputKey := runtimeWorkloadInputKey(selected.Key, workload.PSID, workload.Pool)
	workloadInput := invalidResources.Runtime.Workloads[inputKey]
	workloadInput.Resources.Requests.CPU = "999999999999999999999m"
	invalidResources.Runtime.Workloads[inputKey] = workloadInput
	invalidResourcesJSON, err := marshalChartValuesInput(invalidResources)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		inputJSON            []byte
		failRead             string
		overrideOutputSchema []byte
		wantErr              string
	}{
		{
			name: "input schema read failure", inputJSON: validInput, failRead: chartValuesInputSchemaPath,
			wantErr: chartValuesInputSchemaReadError,
		},
		{
			name: "semantic resource validation failure", inputJSON: invalidResourcesJSON,
			wantErr: testRuntimeWorkloadInvalidResourcesError,
		},
		{
			name: "runtime output schema read failure", inputJSON: validInput, failRead: runtimeChartValuesSchemaPath,
			wantErr: "read runtime chart values schema",
		},
		{
			name: "runtime output schema rejects projection", inputJSON: validInput,
			overrideOutputSchema: []byte(`{"type":"null"}`), wantErr: testSchemaMismatchError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			readFile := func(path string) ([]byte, error) {
				if path == test.failRead {
					return nil, errors.New(testInjectedChartReadError)
				}

				if path == runtimeChartValuesSchemaPath && len(test.overrideOutputSchema) > 0 {
					return test.overrideOutputSchema, nil
				}

				return os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
			}

			_, err := buildRuntimeChartValuesFromInput(
				model, selectedInventory, selected.Key, test.inputJSON, readFile,
			)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestBuildInfrastructureChartValuesFromInventoryFailures(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	tests := []struct {
		name                 string
		model                Model
		failRead             string
		badDefaults          bool
		overrideOutputSchema []byte
		wantErr              string
	}{
		{
			name: "defaults read failure", model: model, failRead: infrastructureChartDefaultsPath,
			wantErr: "read infrastructure chart defaults",
		},
		{
			name: "malformed canonical defaults", model: model, badDefaults: true,
			wantErr: infrastructureValuesYAMLDecodeError,
		},
		{
			name: testModelAndInventoryMismatchCaseName, model: model,
			wantErr: testModelInventoryIdentityMismatch,
		},
		{
			name: "infrastructure output schema read failure", model: model,
			failRead: infrastructureChartValuesSchemaPath, wantErr: "read infrastructure chart values schema",
		},
		{
			name: "infrastructure output schema rejects projection", model: model,
			overrideOutputSchema: []byte(`{"type":"null"}`), wantErr: testSchemaMismatchError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			modelForTest := test.model
			if test.name == testModelAndInventoryMismatchCaseName {
				modelForTest.Run = "other-run"
			}

			readFile := func(path string) ([]byte, error) {
				if path == test.failRead {
					return nil, errors.New(testInjectedChartReadError)
				}

				if path == infrastructureChartDefaultsPath && test.badDefaults {
					return []byte("components: ["), nil
				}

				if path == infrastructureChartValuesSchemaPath && len(test.overrideOutputSchema) > 0 {
					return test.overrideOutputSchema, nil
				}

				return os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
			}

			_, err := buildInfrastructureChartValuesFromInventory(modelForTest, inventory, readFile)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
	require.NoError(t, err)

	return data
}

func TestGeneratorCLIChartValuesModes(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	modelJSON, err := json.Marshal(model)
	require.NoError(t, err)

	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	inputJSON, err := marshalChartValuesInput(chartValuesInputFixture(t))
	require.NoError(t, err)

	readCanonical := func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
	}

	selectedRelease := inventory.Releases[0]
	selectedInventory, err := build(model, scopes, selectedRelease.Key)
	require.NoError(t, err)

	runtimeValues, err := buildRuntimeChartValuesFromInput(
		model, selectedInventory, selectedRelease.Key, inputJSON, readCanonical,
	)
	require.NoError(t, err)
	infrastructureValues, err := buildInfrastructureChartValuesFromInventory(model, inventory, readCanonical)
	require.NoError(t, err)

	runtimePath := runtimeValuesOutputPath(model.Run, selectedRelease.Helm)
	infrastructurePath := infrastructureValuesOutputPath(model.Run)

	tests := []struct {
		name           string
		args           []string
		wantCode       int
		wantOutput     []byte
		wantPath       string
		inputFault     bool
		badInput       bool
		failSchemaRead string
		failWrite      bool
		failStdout     bool
		wantMessage    string
		wantWriteCount int
	}{
		{
			name:       "runtime values preview",
			args:       []string{cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			wantOutput: runtimeValues,
		},
		{
			name:     "runtime values write",
			args:     []string{cliWriteFlag, cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			wantPath: runtimePath, wantOutput: runtimeValues, wantWriteCount: 1, wantMessage: cliGeneratedMessage,
		},
		{
			name:     "runtime values check",
			args:     []string{cliCheckFlag, cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			wantPath: runtimePath, wantOutput: runtimeValues, wantMessage: cliPassMessage,
		},
		{
			name:       "infrastructure values preview",
			args:       []string{cliFormatFlag, infrastructureValuesFormat},
			wantOutput: infrastructureValues,
		},
		{
			name:     "infrastructure values write",
			args:     []string{cliWriteFlag, cliFormatFlag, infrastructureValuesFormat},
			wantPath: infrastructurePath, wantOutput: infrastructureValues, wantWriteCount: 1, wantMessage: cliGeneratedMessage,
		},
		{
			name:     "infrastructure values check",
			args:     []string{cliCheckFlag, cliFormatFlag, infrastructureValuesFormat},
			wantPath: infrastructurePath, wantOutput: infrastructureValues, wantMessage: cliPassMessage,
		},
		{
			name:     "runtime values require selector",
			args:     []string{cliFormatFlag, runtimeValuesFormat, cliRuntimeInputFlag, cliRuntimeInputPath},
			wantCode: 2, wantMessage: cliInvalidFlagsMessage,
		},
		{
			name:     "runtime values require input",
			args:     []string{cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key},
			wantCode: 2, wantMessage: cliInvalidFlagsMessage,
		},
		{
			name:     "runtime input rejected for other format",
			args:     []string{cliRuntimeInputFlag, cliRuntimeInputPath},
			wantCode: 2, wantMessage: cliInvalidFlagsMessage,
		},
		{
			name:     "partial infrastructure artifact prohibited",
			args:     []string{cliWriteFlag, cliFormatFlag, infrastructureValuesFormat, cliSelectFlag, selectedRelease.Key},
			wantCode: 2, wantMessage: cliInvalidFlagsMessage,
		},
		{
			name:       "runtime input read failure",
			args:       []string{cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			inputFault: true, wantCode: 1, wantMessage: "read runtime chart input",
		},
		{
			name:     "runtime input schema failure",
			args:     []string{cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			badInput: true, wantCode: 1, wantMessage: "additional properties",
		},
		{
			name:           "runtime input schema unavailable",
			args:           []string{cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			failSchemaRead: chartValuesInputSchemaPath, wantCode: 1, wantMessage: chartValuesInputSchemaReadError,
		},
		{
			name:     "runtime write failure",
			args:     []string{cliWriteFlag, cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			wantPath: runtimePath, wantOutput: runtimeValues, wantWriteCount: 1, failWrite: true, wantCode: 1, wantMessage: cliWriteArtifactMessage,
		},
		{
			name:     "status write failure",
			args:     []string{cliWriteFlag, cliFormatFlag, runtimeValuesFormat, cliSelectFlag, selectedRelease.Key, cliRuntimeInputFlag, cliRuntimeInputPath},
			wantPath: runtimePath, wantOutput: runtimeValues, wantWriteCount: 1, failStdout: true, wantCode: 1, wantMessage: cliWriteStatusMessage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			readFile := func(path string) ([]byte, error) {
				if filepath.Base(path) == k8sModelFileName {
					return modelJSON, nil
				}

				if path == test.failSchemaRead {
					return nil, errors.New("injected schema read failure")
				}

				if path == cliRuntimeInputPath {
					if test.inputFault {
						return nil, errors.New("injected input read failure")
					}

					if test.badInput {
						return []byte(`{"unknown":true}`), nil
					}

					return inputJSON, nil
				}

				if path == test.wantPath && test.wantPath != "" && containsFlag(test.args, "-check") {
					return test.wantOutput, nil
				}

				return readCanonical(path)
			}

			writes := 0
			writeFile := func(path string, data []byte, mode os.FileMode) error {
				writes++

				require.Equal(t, test.wantPath, path)
				require.Equal(t, test.wantOutput, data)
				require.Equal(t, os.FileMode(cryptoutilSharedMagic.CICDOutputFilePermissions), mode)

				if test.failWrite {
					return errors.New("injected output write failure")
				}

				return nil
			}

			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			stdoutWriter := interface{ Write([]byte) (int, error) }(&stdout)
			if test.failStdout {
				stdoutWriter = failingWriter{}
			}

			code := internalMain(test.args, stdoutWriter, &stderr, readFile, writeFile)
			require.Equal(t, test.wantCode, code, stderr.String())
			require.Equal(t, test.wantWriteCount, writes)

			if test.wantMessage != "" {
				require.Contains(t, stdout.String()+stderr.String(), test.wantMessage)
			}

			if len(test.wantOutput) > 0 && test.wantPath == "" {
				require.Equal(t, test.wantOutput, stdout.Bytes())
			}
		})
	}

	t.Run("invalid selector is not a process panic", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer

		code := internalMain(
			[]string{cliFormatFlag, runtimeValuesFormat, cliSelectFlag, testUnknownScopeSelector, cliRuntimeInputFlag, cliRuntimeInputPath},
			&stdout,
			&stderr,
			func(path string) ([]byte, error) {
				if filepath.Base(path) == k8sModelFileName {
					return modelJSON, nil
				}

				return nil, fmt.Errorf("unexpected path: %s", path)
			},
			func(string, []byte, os.FileMode) error { return nil },
		)
		require.Equal(t, 1, code)
		require.Contains(t, stderr.String(), testNoMatchedReleaseError)
	})
}
