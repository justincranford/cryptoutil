// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	json "encoding/json"
	"errors"
	"reflect"

	"gopkg.in/yaml.v3"
)

func yamlVerificationFields(raw []byte, value *map[string]any) error {
	if err := yaml.Unmarshal(raw, value); err != nil {
		return errors.New("duplicate verification properties")
	}

	return nil
}

func bytesVerificationEqual(actual, expected []byte) bool { return bytes.Equal(actual, expected) }

func verificationFieldsMatch(raw []byte, input VerificationInput) bool {
	typed, _ := json.Marshal(input)

	var original, normalized any
	if json.Unmarshal(raw, &original) != nil || json.Unmarshal(typed, &normalized) != nil {
		return false
	}

	return reflect.DeepEqual(original, normalized)
}
