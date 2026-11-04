// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateModelSchema(t *testing.T) {
	t.Parallel()

	schema, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-model.schema.json"))
	require.NoError(t, err)
	model, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, k8sModelFileName))
	require.NoError(t, err)

	tests := []struct {
		name       string
		modelJSON  []byte
		schemaJSON []byte
		want       string
	}{
		{name: "canonical versioned model", modelJSON: model, schemaJSON: schema},
		{name: "unsupported model version", modelJSON: []byte(strings.Replace(string(model), `"version": 2`, `"version": 3`, 1)), schemaJSON: schema, want: testModelSchemaError},
		{name: "missing bootstrap block", modelJSON: []byte(strings.Replace(string(model), `"bootstrap": {`, `"obsoleteBootstrap": {`, 1)), schemaJSON: schema, want: testModelSchemaError},
		{name: "unknown model property", modelJSON: []byte(strings.Replace(string(model), `"run": "review"`, `"run": "review", "unknown": true`, 1)), schemaJSON: schema, want: testModelSchemaError},
		{name: "invalid loopback endpoint", modelJSON: []byte(strings.Replace(string(model), `"apiHost": "127.0.0.1"`, `"apiHost": "`+testNonLoopbackIPv4+`"`, 1)), schemaJSON: schema, want: testModelSchemaError},
		{name: "invalid service CIDR", modelJSON: []byte(strings.Replace(string(model), `"serviceCIDR": "10.43.0.0/16"`, `"serviceCIDR": "invalid"`, 1)), schemaJSON: schema, want: testModelSchemaError},
		{name: "duplicate model field", modelJSON: []byte(strings.Replace(string(model), `"version": 2`, `"version": 2, "version": 2`, 1)), schemaJSON: schema, want: testDuplicateJSONKey},
		{name: "schema has duplicate key", modelJSON: model, schemaJSON: []byte(strings.Replace(string(schema), `"type": "object"`, `"type": "object", "type": "object"`, 1)), want: testDuplicateJSONKey},
		{name: "schema reference cannot compile", modelJSON: model, schemaJSON: []byte(`{"$ref":"#/$defs/missing","$defs":{}}`), want: "compile model schema"},
		{name: "malformed model JSON", modelJSON: []byte("{"), schemaJSON: schema, want: "model:"},
		{name: "malformed schema JSON", modelJSON: model, schemaJSON: []byte("{"), want: "model schema:"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateModelSchema(test.schemaJSON, test.modelJSON)
			if test.want == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, test.want)
		})
	}
}
