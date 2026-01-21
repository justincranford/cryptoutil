// Copyright (c) 2025 Justin Cranford
//
//

package util

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
)

// YAML2JSON converts a YAML string to a JSON string.
func YAML2JSON(y string) (string, error) {
	object, err := ParseYAML(y)
	if err != nil {
		return "", err
	}

	return EncodeJSON(object)
}

// JSON2YAML converts a JSON string to a YAML string.
func JSON2YAML(j string) (string, error) {
	object, err := ParseJSON(j)
	if err != nil {
		return "", err
	}

	return EncodeYAML(object)
}

// ParseYAML parses a YAML string into a Go object.
func ParseYAML(y string) (any, error) {
	var object any
	if err := yaml.Unmarshal([]byte(y), &object); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return object, nil
}

// ParseJSON parses a JSON string into a Go object.
func ParseJSON(j string) (any, error) {
	var object any
	if err := json.Unmarshal([]byte(j), &object); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return object, nil
}

// EncodeYAML encodes a Go object into a YAML string.
func EncodeYAML(object any) (string, error) {
	yamlContent, err := yaml.Marshal(object)
	if err != nil {
		return "", fmt.Errorf("failed to encode YAML: %w", err)
	}

	return string(yamlContent), nil
}

// EncodeJSON encodes a Go object into a JSON string.
func EncodeJSON(object any) (string, error) {
	jsonContent, err := json.Marshal(object)
	if err != nil {
		return "", fmt.Errorf("failed to encode JSON: %w", err)
	}

	return string(jsonContent), nil
}
