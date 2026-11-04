// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChartValuesInputSchema(t *testing.T) {
	t.Parallel()

	input := chartValuesInputFixture(t)
	inputJSON, err := marshalChartValuesInput(input)
	require.NoError(t, err)

	inputJSON = bytes.TrimSuffix(inputJSON, []byte{'\n'})

	schemaPath := filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-chart-values-input.schema.json")
	schemaJSON, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	require.NoError(t, ValidateChartValuesInputSchema(schemaJSON, inputJSON))

	decoded, err := DecodeChartValuesInput(inputJSON)
	require.NoError(t, err)
	require.Equal(t, input, decoded)
}

func chartValuesInputFixture(t *testing.T) ChartValuesInput {
	t.Helper()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	input := ChartValuesInput{
		Images:          chartImageFixture(scopes),
		PKIMaterializer: chartMaterializerFixture(),
		Storage:         ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		Runtime:         RuntimeChartInputs{Workloads: make(map[string]RuntimeWorkloadInput)},
	}
	for _, release := range inventory.Releases {
		input.Runtime.Workloads = mergeRuntimeChartInputs(
			input.Runtime.Workloads,
			chartRuntimeInputFixture(inventory, release.Key).Workloads,
		)
	}

	return input
}

func marshalChartValuesInput(input ChartValuesInput) ([]byte, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return append(data, '\n'), nil
}

func TestDecodeChartValuesInputRejectsMalformedDocuments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{name: "top level must be object", input: []byte("[]"), want: "expected a top-level JSON object"},
		{name: "duplicate key", input: []byte(`{"images":{},"images":{}}`), want: "duplicate JSON object key"},
		{name: "trailing JSON value", input: []byte(`{} {}`), want: testExactlyOneJSONDocument},
		{name: testUnknownJSONField, input: []byte(`{"unknown":true}`), want: testUnknownJSONField},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := DecodeChartValuesInput(test.input)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func mergeRuntimeChartInputs(
	aggregate, incoming map[string]RuntimeWorkloadInput,
) map[string]RuntimeWorkloadInput {
	for key, input := range incoming {
		aggregate[key] = input
	}

	return aggregate
}
