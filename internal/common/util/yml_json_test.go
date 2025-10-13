package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	validYAML       = "key1: value1\nkey2: value2\n"
	invalidYAML     = "key1 value1\nkey2: value2\n"
	validJSON       = `{"key1":"value1","key2":"value2"}`
	invalidJSON     = `{"key1":"value1", "key2":}`
	singleParsedObj = map[string]any{"key": "value"}
	validYAMLSingle = "key: value\n"
	validJSONSingle = `{"key":"value"}`
)

func TestHappyPathYAML2JSON(t *testing.T) {
	result, err := YAML2JSON(validYAML)
	require.NoError(t, err)
	require.Equal(t, validJSON, result)
}

func TestSadPathYAML2JSON(t *testing.T) {
	_, err := YAML2JSON(invalidYAML)
	require.Error(t, err)
	require.EqualError(t, err, "failed to parse YAML: [1:1] unexpected key name\n>  1 | key1 value1\n   2 | key2: value2\n       ^\n")
}

func TestHappyPathJSON2YAML(t *testing.T) {
	result, err := JSON2YAML(validJSON)
	require.NoError(t, err)
	require.Equal(t, validYAML, result)
}

func TestSadPathJSON2YAML(t *testing.T) {
	_, err := JSON2YAML(invalidJSON)
	require.Error(t, err)
	require.EqualError(t, err, "failed to parse JSON: invalid character '}' looking for beginning of value")
}

func TestHappyPathParseYAML(t *testing.T) {
	object, err := ParseYAML(validYAMLSingle)
	require.NoError(t, err)
	objMap, ok := object.(map[string]any)
	require.True(t, ok, "object should be a map[string]any")
	require.Equal(t, "value", objMap["key"])
}

func TestSadPathParseYAML(t *testing.T) {
	_, err := ParseYAML(invalidYAML)
	require.Error(t, err)
	require.EqualError(t, err, "failed to parse YAML: [1:1] unexpected key name\n>  1 | key1 value1\n   2 | key2: value2\n       ^\n")
}

func TestHappyPathParseJSON(t *testing.T) {
	object, err := ParseJSON(validJSONSingle)
	require.NoError(t, err)
	objMap, ok := object.(map[string]any)
	require.True(t, ok, "object should be a map[string]any")
	require.Equal(t, "value", objMap["key"])
}

func TestSadPathParseJSON(t *testing.T) {
	_, err := ParseJSON(invalidJSON)
	require.Error(t, err)
	require.EqualError(t, err, "failed to parse JSON: invalid character '}' looking for beginning of value")
}

func TestHappyPathEncodeYAML(t *testing.T) {
	result, err := EncodeYAML(singleParsedObj)
	require.NoError(t, err)
	require.Equal(t, validYAMLSingle, result)
}

func TestSadPathEncodeYAML(t *testing.T) {
	_, err := EncodeYAML(func() {})
	require.Error(t, err)
	require.EqualError(t, err, "failed to encode YAML: unknown value type func()")
}

func TestHappyPathEncodeJSON(t *testing.T) {
	result, err := EncodeJSON(singleParsedObj)
	require.NoError(t, err)
	require.Equal(t, validJSONSingle, result)
}

func TestSadPathEncodeJSON(t *testing.T) {
	_, err := EncodeJSON(func() {})
	require.Error(t, err)
	require.EqualError(t, err, "failed to encode JSON: json: unsupported type: func()")
}
