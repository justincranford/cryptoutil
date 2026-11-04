// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	json "encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	modelSchemaID               = "https://cryptoutil.dev/schemas/k8s-model.schema.json"
	chartValuesInputSchemaID    = "https://cryptoutil.dev/schemas/k8s-chart-values-input.schema.json"
	runtimeChartValuesSchemaID  = "https://cryptoutil.dev/schemas/k8s-runtime-values.schema.json"
	infrastructureChartSchemaID = "https://cryptoutil.dev/schemas/k8s-infrastructure-values.schema.json"
)

// ValidateModelSchema validates a model against its committed JSON Schema draft.
func ValidateModelSchema(schemaJSON, modelJSON []byte) error {
	return validateJSONSchemaDocument(modelSchemaID, "model schema", "model", schemaJSON, modelJSON)
}

// ValidateChartValuesInputSchema validates explicit non-secret chart projection inputs.
func ValidateChartValuesInputSchema(schemaJSON, inputJSON []byte) error {
	return validateJSONSchemaDocument(
		chartValuesInputSchemaID, "chart values input schema", "chart values input", schemaJSON, inputJSON,
	)
}

// ValidateRuntimeChartValuesSchema validates generated runtime chart values against
// the strict JSON Schema consumed by the canonical Helm chart.
func ValidateRuntimeChartValuesSchema(schemaJSON, valuesJSON []byte) error {
	return validateJSONSchemaDocument(
		runtimeChartValuesSchemaID, "runtime values schema", "runtime values", schemaJSON, valuesJSON,
	)
}

// ValidateInfrastructureChartValuesSchema validates values against the canonical infrastructure chart schema.
func ValidateInfrastructureChartValuesSchema(schemaJSON, valuesJSON []byte) error {
	return validateJSONSchemaDocument(
		infrastructureChartSchemaID, "infrastructure values schema", "infrastructure values", schemaJSON, valuesJSON,
	)
}

// ValidateLifecycleChartValuesSchema validates the non-secret lifecycle chart contract.
func ValidateLifecycleChartValuesSchema(schemaJSON, valuesJSON []byte) error {
	return validateJSONSchemaDocument(
		"https://cryptoutil.dev/schemas/k8s-lifecycle-values.schema.json",
		"lifecycle values schema", "lifecycle values", schemaJSON, valuesJSON,
	)
}

func validateJSONSchemaDocument(
	schemaID, schemaName, documentName string,
	schemaJSON, documentJSON []byte,
) error {
	var schemaDocument any
	if err := json.Unmarshal(schemaJSON, &schemaDocument); err != nil {
		return fmt.Errorf("%s: %w", schemaName, err)
	}

	if err := validateJSONDocument(schemaJSON); err != nil {
		return fmt.Errorf("%s: %w", schemaName, err)
	}

	var document any
	if err := json.Unmarshal(documentJSON, &document); err != nil {
		return fmt.Errorf("%s: %w", documentName, err)
	}

	if err := validateJSONDocument(documentJSON); err != nil {
		return fmt.Errorf("%s: %w", documentName, err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()

	if err := compiler.AddResource(schemaID, schemaDocument); err != nil {
		return fmt.Errorf("register %s: %w", schemaName, err)
	}

	schema, err := compiler.Compile(schemaID)
	if err != nil {
		return fmt.Errorf("compile %s: %w", schemaName, err)
	}

	if err := schema.Validate(document); err != nil {
		return fmt.Errorf("%s does not match schema: %w", documentName, err)
	}

	return nil
}
