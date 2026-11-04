// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"io"
	"slices"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func decodeModel(data []byte) (model, error) {
	return DecodeModel(data)
}

// DecodeModel validates one strict, case-sensitive JSON model document.
func DecodeModel(data []byte) (Model, error) {
	var m Model
	if err := validateJSONDocument(data); err != nil {
		return m, fmt.Errorf("model: %w", err)
	}

	if err := validateModelJSONFieldNames(data); err != nil {
		return m, fmt.Errorf("model: %w", err)
	}

	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()

	if err := d.Decode(&m); err != nil {
		return m, fmt.Errorf("model: %w", err)
	}

	return m, nil
}

func validateModelJSONFieldNames(data []byte) error {
	return validateModelJSONFieldNamesWithReaders(data, jsonObjectFields, jsonArrayElements)
}

func validateModelJSONFieldNamesWithReaders(
	data []byte,
	objectFieldsReader func([]byte) (map[string]json.RawMessage, bool, error),
	arrayElementsReader func([]byte) ([]json.RawMessage, bool, error),
) error {
	fields, isObject, err := objectFieldsReader(data)
	if err != nil {
		return err
	}

	if !isObject {
		return nil
	}

	if err := validateJSONSchemaFieldMap(fields, "model",
		cryptoutilSharedMagic.CLIVersionCommand, cryptoutilSharedMagic.WorkflowSubCmdRun, "revision", "clusterDomain", "localDomain",
		"replicas", "overrides", "modes", "infrastructure", "bootstrap",
	); err != nil {
		return err
	}

	if replicas, ok := fields["replicas"]; ok {
		if err := validateJSONSchemaFields(replicas, "replicas", "min", "max"); err != nil {
			return err
		}
	}

	if overrides, ok := fields["overrides"]; ok {
		overrideFields, isObject, err := objectFieldsReader(overrides)
		if err != nil {
			return err
		}

		if isObject {
			keys := make([]string, 0, len(overrideFields))
			for key := range overrideFields {
				keys = append(keys, key)
			}

			slices.Sort(keys)

			for _, key := range keys {
				if err := validateJSONSchemaFields(overrideFields[key], fmt.Sprintf("override %q", key), "min", "max"); err != nil {
					return err
				}
			}
		}
	}

	if modes, ok := fields["modes"]; ok {
		modeValues, isArray, err := arrayElementsReader(modes)
		if err != nil {
			return err
		}

		if isArray {
			for i, value := range modeValues {
				if err := validateJSONSchemaFields(value, fmt.Sprintf("modes[%d]", i), cryptoutilSharedMagic.ClaimName, "pools"); err != nil {
					return err
				}
			}
		}
	}

	if infrastructure, ok := fields["infrastructure"]; ok {
		if err := validateJSONSchemaFields(infrastructure, "infrastructure",
			"postgresLeader", "postgresFollower", "collector", "lgtm",
		); err != nil {
			return err
		}
	}

	if bootstrap, ok := fields["bootstrap"]; ok {
		if err := validateJSONSchemaFields(bootstrap, "bootstrap",
			"apiHost", "apiPort", "gatewayHost", "gatewayPort", "apiServiceIP", "serviceCIDR", "podCIDR",
			"serverNodes", "agentNodes",
		); err != nil {
			return err
		}
	}

	return nil
}

func jsonObjectFields(data []byte) (map[string]json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, false, nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil, false, fmt.Errorf("decode JSON object fields: %w", err)
	}

	return fields, true, nil
}

func jsonArrayElements(data []byte) ([]json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, false, nil
	}

	var values []json.RawMessage
	if err := json.Unmarshal(trimmed, &values); err != nil {
		return nil, false, fmt.Errorf("decode JSON array elements: %w", err)
	}

	return values, true, nil
}

func validateJSONSchemaFields(data []byte, context string, allowed ...string) error {
	fields, isObject, err := jsonObjectFields(data)
	if err != nil || !isObject {
		return err
	}

	return validateJSONSchemaFieldMap(fields, context, allowed...)
}

func validateJSONSchemaFieldMap(fields map[string]json.RawMessage, context string, allowed ...string) error {
	unknown := make([]string, 0)

	for key := range fields {
		if !slices.Contains(allowed, key) {
			unknown = append(unknown, key)
		}
	}

	if len(unknown) == 0 {
		return nil
	}

	slices.Sort(unknown)

	return fmt.Errorf("unknown field %q in %s (JSON field names are case-sensitive)", unknown[0], context)
}

func validateJSONDocument(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()

	first, err := d.Token()
	if err != nil {
		return fmt.Errorf("read first JSON token: %w", err)
	}

	if first != json.Delim('{') {
		return fmt.Errorf("expected a top-level JSON object")
	}

	if err := scanJSONValue(d, first); err != nil {
		return err
	}

	if _, err := d.Token(); err != io.EOF {
		if err != nil {
			return fmt.Errorf("expected exactly one JSON document: %w", err)
		}

		return fmt.Errorf("expected exactly one JSON document")
	}

	return nil
}

type jsonTokenReader interface {
	More() bool
	Token() (json.Token, error)
}

func scanJSONValue(d jsonTokenReader, token json.Token) error {
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		seen := map[string]struct{}{}

		for d.More() {
			keyToken, err := d.Token()
			if err != nil {
				return fmt.Errorf("read JSON object key: %w", err)
			}

			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("expected a JSON object key")
			}

			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate JSON object key: %q", key)
			}

			seen[key] = struct{}{}

			valueToken, err := d.Token()
			if err != nil {
				return fmt.Errorf("read JSON object value: %w", err)
			}

			if err := scanJSONValue(d, valueToken); err != nil {
				return err
			}
		}

		end, err := d.Token()
		if err != nil {
			return fmt.Errorf("read JSON object end: %w", err)
		}

		if end != json.Delim('}') {
			return fmt.Errorf("expected end of JSON object")
		}
	case '[':
		for d.More() {
			valueToken, err := d.Token()
			if err != nil {
				return fmt.Errorf("read JSON array value: %w", err)
			}

			if err := scanJSONValue(d, valueToken); err != nil {
				return err
			}
		}

		end, err := d.Token()
		if err != nil {
			return fmt.Errorf("read JSON array end: %w", err)
		}

		if end != json.Delim(']') {
			return fmt.Errorf("expected end of JSON array")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter: %q", delim)
	}

	return nil
}
