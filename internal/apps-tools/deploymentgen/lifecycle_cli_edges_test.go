// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleRuntimeFile = "runtime-input.json"
	testLifecycleInputFlag   = "-input"
	testLifecycleSubcommand  = "lifecycle"
)

func TestLifecycleCLIQuotaFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, format string
	}{
		{"runtime with job headroom", runtimeValuesFormat},
		{"infrastructure with job headroom", infrastructureValuesFormat},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
			lifecycle, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)

			runtimeInput := ChartValuesInput{
				Images: chartImageFixture(RegisteredScopes()), PKIMaterializer: chartMaterializerFixture(),
				Storage: ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
				Runtime: chartRuntimeInputFixture(inventory, inventory.Releases[0].Key),
			}
			runtimeJSON, err := json.Marshal(runtimeInput)
			require.NoError(t, err)
			inputJSON, err := json.Marshal(input)
			require.NoError(t, err)
			modelJSON, err := json.Marshal(model)
			require.NoError(t, err)

			read := func(file string) ([]byte, error) {
				switch file {
				case testLifecycleOperatorFile:
					return inputJSON, nil
				case defaultPlannerPaths().model:
					return modelJSON, nil
				case testLifecycleRuntimeFile:
					return runtimeJSON, nil
				default:
					return os.ReadFile(filepath.Join("..", "..", "..", file))
				}
			}

			args := []string{testLifecycleSubcommand, testLifecycleInputFlag, testLifecycleOperatorFile, cliSelectFlag, testServiceSMKMSPGSelector, cliFormatFlag, test.format}
			if test.format == runtimeValuesFormat {
				args = append(args, cliRuntimeInputFlag, testLifecycleRuntimeFile)
			}

			var stdout, stderr bytes.Buffer

			require.Zero(t, Main(args, &stdout, &stderr, read, nil), stderr.String())
			expected, err := lifecycleOutput(test.format, testLifecycleRuntimeFile, model, inventory, lifecycle, read)
			require.NoError(t, err)
			require.Equal(t, expected, stdout.Bytes())
			require.Contains(t, stdout.String(), `"pods":`)
		})
	}
}

func TestLifecycleCLIErrorBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		args                  []string
		failPath, replacePath string
		replacement           []byte
		failOutput            bool
		want                  int
	}{
		{name: "lifecycle help", args: []string{cryptoutilSharedMagic.CLIHelpShortFlag}},
		{name: "lifecycle unknown flag", args: []string{"-not-a-flag"}, want: 2},
		{name: "missing input", args: []string{}, want: 2},
		{name: "lifecycle positional argument", args: []string{testLifecycleInputFlag, testLifecycleOperatorFile, "trailing"}, want: 2},
		{name: "malformed input", replacePath: testLifecycleOperatorFile, replacement: []byte(testOpenJSONObject), want: 1},
		{name: "lifecycle missing model", failPath: defaultPlannerPaths().model, want: 1},
		{name: "malformed model", replacePath: defaultPlannerPaths().model, replacement: []byte(testOpenJSONObject), want: 1},
		{name: "lifecycle unknown selector", args: []string{testLifecycleInputFlag, testLifecycleOperatorFile, cliSelectFlag, testUnknownReleaseKey}, want: 1},
		{name: "missing schema", failPath: lifecycleValuesSchemaPath, want: 1},
		{name: "invalid schema", replacePath: lifecycleValuesSchemaPath, replacement: []byte(testOpenJSONObject), want: 1},
		{name: "revision write failure", args: []string{testLifecycleInputFlag, testLifecycleOperatorFile, cliFormatFlag, lifecycleRevisionFormat}, failOutput: true, want: 1},
		{name: "values write failure", failOutput: true, want: 1},
		{name: "runtime missing input", args: []string{testLifecycleInputFlag, testLifecycleOperatorFile, cliFormatFlag, runtimeValuesFormat}, want: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _, input := lifecycleDeployFixture(t, "")
			inputJSON, err := json.Marshal(input)
			require.NoError(t, err)
			modelJSON, err := json.Marshal(model)
			require.NoError(t, err)

			read := func(file string) ([]byte, error) {
				if file == test.failPath {
					return nil, errors.New("injected read failure")
				}

				if file == test.replacePath {
					return test.replacement, nil
				}

				switch file {
				case testLifecycleOperatorFile:
					return inputJSON, nil
				case defaultPlannerPaths().model:
					return modelJSON, nil
				default:
					return os.ReadFile(filepath.Join("..", "..", "..", file))
				}
			}

			args := test.args
			if args == nil {
				args = []string{testLifecycleInputFlag, testLifecycleOperatorFile}
			}

			var stdout, stderr bytes.Buffer

			var writer io.Writer = &stdout
			if test.failOutput {
				writer = failingWriter{}
			}

			require.Equal(t, test.want, lifecycleMain(args, writer, &stderr, read, nil), stderr.String())
		})
	}
}

func TestLifecycleOutputRejectsBadProjectionInputs(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
	lifecycle, err := BuildLifecycleChartValues(model, inventory, input)
	require.NoError(t, err)

	for _, format := range []string{runtimeValuesFormat, infrastructureValuesFormat} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			_, err := lifecycleOutput(format, testLifecycleRuntimeFile, model, inventory, lifecycle, func(string) ([]byte, error) {
				return nil, errors.New("injected missing projection")
			})
			require.Error(t, err)
			_, err = lifecycleOutput(format, testLifecycleRuntimeFile, model, inventory, lifecycle, func(string) ([]byte, error) {
				return []byte(testOpenJSONObject), nil
			})
			require.Error(t, err)
		})
	}
}
