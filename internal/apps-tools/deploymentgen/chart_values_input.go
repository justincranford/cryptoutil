// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"fmt"
)

// ChartValuesInput supplies the non-secret image, storage, and per-workload
// settings required to project a complete deployment inventory into Helm values.
type ChartValuesInput struct {
	Images          map[string]ContainerImage `json:"images"`
	PKIMaterializer PKIMaterializerImage      `json:"pkiMaterializer"`
	Storage         ChartStorage              `json:"storage"`
	Runtime         RuntimeChartInputs        `json:"runtime"`
}

// DecodeChartValuesInput strictly decodes one complete JSON input document.
func DecodeChartValuesInput(data []byte) (ChartValuesInput, error) {
	var input ChartValuesInput
	if err := validateJSONDocument(data); err != nil {
		return input, fmt.Errorf("chart values input: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return ChartValuesInput{}, fmt.Errorf("chart values input: %w", err)
	}

	return input, nil
}
