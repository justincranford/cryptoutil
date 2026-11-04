// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"path/filepath"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	runtimeValuesFormat                 = "runtime-values"
	infrastructureValuesFormat          = "infrastructure-values"
	chartValuesInputSchemaPath          = "api/cryptosuite-registry/k8s/k8s-chart-values-input.schema.json"
	runtimeChartValuesSchemaPath        = "api/cryptosuite-registry/templates/kubernetes/cryptoutil-runtime/values.schema.json"
	infrastructureChartValuesSchemaPath = "api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra/values.schema.json"
	infrastructureChartDefaultsPath     = "api/cryptosuite-registry/templates/kubernetes/cryptoutil-infra/values.yaml"
	chartValuesInputSchemaReadError     = "read chart values input schema"
)

func buildRuntimeChartValuesFromInput(
	model Model,
	inventory Inventory,
	releaseKey string,
	inputJSON []byte,
	readFile func(string) ([]byte, error),
) ([]byte, error) {
	inputSchema, err := readFile(chartValuesInputSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", chartValuesInputSchemaReadError, err)
	}

	if err := ValidateChartValuesInputSchema(inputSchema, inputJSON); err != nil {
		return nil, err
	}

	input, err := DecodeChartValuesInput(inputJSON)
	if err != nil {
		return nil, err
	}

	values, err := BuildRuntimeChartValues(
		model, inventory, releaseKey, input.Images, input.PKIMaterializer, input.Storage, input.Runtime,
	)
	if err != nil {
		return nil, err
	}

	valuesJSON, err := MarshalRuntimeChartValues(values)
	if err != nil {
		return nil, fmt.Errorf("marshal runtime chart values: %w", err)
	}

	valuesSchema, err := readFile(runtimeChartValuesSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("read runtime chart values schema: %w", err)
	}

	if err := ValidateRuntimeChartValuesSchema(valuesSchema, valuesJSON); err != nil {
		return nil, err
	}

	return valuesJSON, nil
}

func buildInfrastructureChartValuesFromInventory(
	model Model,
	inventory Inventory,
	readFile func(string) ([]byte, error),
) ([]byte, error) {
	defaultsYAML, err := readFile(infrastructureChartDefaultsPath)
	if err != nil {
		return nil, fmt.Errorf("read infrastructure chart defaults: %w", err)
	}

	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	if err != nil {
		return nil, err
	}

	values, err := BuildInfrastructureChartValues(model, inventory, defaults)
	if err != nil {
		return nil, err
	}

	valuesJSON, err := MarshalInfrastructureChartValues(values)
	if err != nil {
		return nil, fmt.Errorf("marshal infrastructure chart values: %w", err)
	}

	valuesSchema, err := readFile(infrastructureChartValuesSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("read infrastructure chart values schema: %w", err)
	}

	if err := ValidateInfrastructureChartValuesSchema(valuesSchema, valuesJSON); err != nil {
		return nil, err
	}

	return valuesJSON, nil
}

func runtimeValuesOutputPath(run, helmRelease string) string {
	return filepath.Join("deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "generated", run, helmRelease, "values.yaml")
}

func infrastructureValuesOutputPath(run string) string {
	return filepath.Join("deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "generated", run, cryptoutilSharedMagic.PKIInitEntityInfra, "values.yaml")
}
