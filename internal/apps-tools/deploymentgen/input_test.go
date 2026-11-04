// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	json "encoding/json"
	"errors"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testModelContext                     = "model:"
	testReplicasKey                      = "replicas"
	testUnknownJSONField                 = "unknown field"
	testExactlyOneJSONDocument           = "exactly one JSON document"
	testInjectedChartReadError           = "injected chart read failure"
	testTopLevelJSONObject               = "top-level JSON object"
	testOpenJSONObject                   = "{"
	testNullDocument                     = "null"
	testDuplicateJSONKey                 = "duplicate JSON object key"
	testSingleJSONDocument               = "exactly one"
	testInvalidDNSName                   = "invalid DNS name"
	testDNSNameError                     = "DNS name"
	testExampleValue                     = "example"
	testDefaultReplicaJSON               = `"replicas":{"min":2,"max":2}`
	testTruncatedJSONArray               = `{"value": [1`
	testSuitePGSelector                  = "u:cryptoutil:pg"
	testProductSMMemSelector             = "p:sm:mem"
	testServiceSMKMSPGSelector           = "s:sm-kms:pg"
	testServiceSMKMSMemSelector          = "s:sm-kms:mem"
	testInvalidVersion                   = "invalid version"
	testInvalidDomain                    = "invalid domain"
	testPoolError                        = "pool"
	testNonemptyError                    = "nonempty"
	testPSIDError                        = "PS-ID"
	testProductIdentity                  = "product"
	testSuiteIdentity                    = "suite"
	testExtraIdentity                    = "extra"
	testEmptyProductError                = "empty product"
	testDuplicateIdentityError           = "duplicate identity"
	testExpandedPodCapacityError         = "expanded pod capacity"
	testSQLIdentifierError               = "SQL identifier"
	testInvalidKubernetesIdentity        = "Invalid"
	testThree                            = 3
	testProductIdentityMixedDiskSelector = "p:identity:mixed-disk"
	testNoMatchedReleaseError            = "matched no release"
	testUnknownScopeSelector             = "p:pki:pg"
	testEmptyRegistryCase                = "empty registry"
	testInvalidScopeIDCase               = "invalid scope ID"
	testDuplicateScopeCase               = "duplicate scope"
	testUnknownRegistryTierCase          = "unknown registry tier"
	testInvalidSchemaVersion             = "invalid schema version"
	testDigestPinned                     = "digest-pinned"
	testNonLoopbackIPv4                  = "192.0.2.1"
	testLoopbackIPv4Error                = "loopback IPv4"
	testRenderedLFError                  = "LF and end with LF"
	testStorageSize                      = "1Gi"
	testStorageClassName                 = "cu-local-retain"
	testMaterializerRequirements         = "PKI materializer requires"
	testInvalidKubernetesValues          = "valid Kubernetes values"
	testDifferentIdentity                = "other"
	testIdentityMismatchError            = "identities differ"
	testNonOverlappingIPv4Error          = "non-overlapping IPv4"
	testMissingRelease                   = "selected release must exist"
	testProbeCommandError                = "startup/liveness/readiness"
	testModelSchemaError                 = "model does not match schema"
	testOverridesJSONField               = "overrides"
)

type scriptedJSONToken struct {
	value json.Token
	err   error
}

type scriptedJSONTokenReader struct {
	moreValues []bool
	tokens     []scriptedJSONToken
	moreIndex  int
	tokenIndex int
}

func (r *scriptedJSONTokenReader) More() bool {
	if r.moreIndex >= len(r.moreValues) {
		return false
	}

	value := r.moreValues[r.moreIndex]
	r.moreIndex++

	return value
}

func (r *scriptedJSONTokenReader) Token() (json.Token, error) {
	if r.tokenIndex >= len(r.tokens) {
		return nil, errors.New("scripted JSON token reader exhausted")
	}

	result := r.tokens[r.tokenIndex]
	r.tokenIndex++

	return result.value, result.err
}

func TestScanJSONValueReaderErrors(t *testing.T) {
	t.Parallel()

	injected := errors.New("injected token reader error")

	tests := []struct {
		name   string
		reader *scriptedJSONTokenReader
		token  json.Token
		want   string
	}{
		{
			name:   "unexpected delimiter",
			reader: &scriptedJSONTokenReader{},
			token:  json.Delim('x'),
			want:   "unexpected JSON delimiter",
		},
		{
			name:   "non-string object key",
			reader: &scriptedJSONTokenReader{moreValues: []bool{true}, tokens: []scriptedJSONToken{{value: json.Number("1")}}},
			token:  json.Delim('{'),
			want:   "expected a JSON object key",
		},
		{
			name:   "object key read failure",
			reader: &scriptedJSONTokenReader{moreValues: []bool{true}, tokens: []scriptedJSONToken{{err: injected}}},
			token:  json.Delim('{'),
			want:   "read JSON object key",
		},
		{
			name:   "object value read failure",
			reader: &scriptedJSONTokenReader{moreValues: []bool{true}, tokens: []scriptedJSONToken{{value: "key"}, {err: injected}}},
			token:  json.Delim('{'),
			want:   "read JSON object value",
		},
		{
			name:   "object end read failure",
			reader: &scriptedJSONTokenReader{moreValues: []bool{false}, tokens: []scriptedJSONToken{{err: injected}}},
			token:  json.Delim('{'),
			want:   "read JSON object end",
		},
		{
			name:   "wrong object end",
			reader: &scriptedJSONTokenReader{moreValues: []bool{false}, tokens: []scriptedJSONToken{{value: json.Delim(']')}}},
			token:  json.Delim('{'),
			want:   "expected end of JSON object",
		},
		{
			name:   "array value read failure",
			reader: &scriptedJSONTokenReader{moreValues: []bool{true}, tokens: []scriptedJSONToken{{err: injected}}},
			token:  json.Delim('['),
			want:   "read JSON array value",
		},
		{
			name:   "array end read failure",
			reader: &scriptedJSONTokenReader{moreValues: []bool{false}, tokens: []scriptedJSONToken{{err: injected}}},
			token:  json.Delim('['),
			want:   "read JSON array end",
		},
		{
			name:   "wrong array end",
			reader: &scriptedJSONTokenReader{moreValues: []bool{false}, tokens: []scriptedJSONToken{{value: json.Delim('}')}}},
			token:  json.Delim('['),
			want:   "expected end of JSON array",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := scanJSONValue(tc.reader, tc.token)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestModelRejectsCaseAliases(t *testing.T) {
	t.Parallel()

	m, _ := fixture(t)
	modelJSON, err := json.Marshal(m)
	require.NoError(t, err)

	tests := []struct {
		name   string
		before string
		after  string
	}{
		{
			name:   "top-level run alias",
			before: `"run":"` + m.Run + `"`,
			after:  `"run":"` + m.Run + `","Run":"` + testDifferentIdentity + `"`,
		},
		{
			name:   "recognized top-level alias alone",
			before: `"run":"` + m.Run + `"`,
			after:  `"Run":"` + m.Run + `"`,
		},
		{
			name:   "nested replicas min alias",
			before: testDefaultReplicaJSON,
			after:  `"replicas":{"min":2,"Min":1,"max":2}`,
		},
		{
			name:   "recognized nested alias alone",
			before: testDefaultReplicaJSON,
			after:  `"replicas":{"Min":2,"max":2}`,
		},
		{
			name:   "mode name alias",
			before: `"name":"` + cryptoutilSharedMagic.DeploymentPoolPG + `"`,
			after:  `"Name":"` + cryptoutilSharedMagic.DeploymentPoolPG + `"`,
		},
		{
			name:   "infrastructure field alias",
			before: `"postgresLeader":1`,
			after:  `"PostgresLeader":1`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := strings.Replace(string(modelJSON), tc.before, tc.after, 1)
			require.NotEqual(t, string(modelJSON), input, "fixture replacement must match")

			_, err := decodeModel([]byte(input))
			require.ErrorContains(t, err, "JSON field names are case-sensitive")
		})
	}
}

func TestOverrideKeysRemainExact(t *testing.T) {
	t.Parallel()

	m, r := fixture(t)
	m.Overrides["SM-KMS/pg"] = bounds{Min: 2, Max: 2}
	modelJSON, err := json.Marshal(m)
	require.NoError(t, err)

	decoded, err := decodeModel(modelJSON)
	require.NoError(t, err)
	require.Equal(t, bounds{Min: 2, Max: 2}, decoded.Overrides["SM-KMS/pg"])
	require.ErrorContains(t, validate(decoded, r), "unknown override")
}

func TestModelDecoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "zero-length input", input: "", want: testModelContext},
		{name: "truncated object", input: testOpenJSONObject, want: testModelContext},
		{name: "unknown registry field", input: `{"unexpected": true}`, want: testUnknownJSONField},
		{name: "wrong field type", input: `{"version": "one"}`, want: testModelContext},
		{name: testNullDocument, input: testNullDocument, want: testTopLevelJSONObject},
		{name: "array", input: `[]`, want: testTopLevelJSONObject},
		{name: "string", input: `"value"`, want: testTopLevelJSONObject},
		{name: "duplicate top-level JSON key", input: `{"version": 1, "version": 1}`, want: testDuplicateJSONKey},
		{name: "duplicate nested JSON key", input: `{"overrides": {"sm-kms/pg": {"min": 1, "min": 2}}}`, want: testDuplicateJSONKey},
		{name: "invalid array value", input: `{"value": [1,]}`, want: testModelContext},
		{name: "truncated array value", input: testTruncatedJSONArray, want: testModelContext},
		{name: "truncated object value", input: `{"value": 1`, want: testModelContext},
		{name: "missing object value", input: `{"value":`, want: testModelContext},
		{name: "mismatched object delimiter", input: `{"value": 1]`, want: testModelContext},
		{name: "truncated nested array", input: testTruncatedJSONArray, want: testModelContext},
		{name: "mismatched array delimiter", input: `{"value": [1}`, want: testModelContext},
		{name: "nested array value error", input: `{"value": [[1,]]}`, want: testModelContext},
		{name: "unexpected array value delimiter", input: `{"value": ]}`, want: testModelContext},
		{name: "multiple JSON documents", input: `{} {}`, want: testSingleJSONDocument},
		{name: "trailing content", input: `{} trailing`, want: testSingleJSONDocument},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := decodeModel([]byte(tc.input))
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestJSONSchemaValidationHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		validate func([]byte) error
		wantErr  bool
	}{
		{
			name:  "malformed model object",
			input: testOpenJSONObject,
			validate: func(input []byte) error {
				return validateModelJSONFieldNames(input)
			},
			wantErr: true,
		},
		{
			name:  "non-object model",
			input: `null`,
			validate: func(input []byte) error {
				return validateModelJSONFieldNames(input)
			},
		},
		{
			name:  "non-object replicas",
			input: `{"replicas":null}`,
			validate: func(input []byte) error {
				return validateModelJSONFieldNames(input)
			},
		},
		{
			name:  "non-object overrides",
			input: `{"overrides":null}`,
			validate: func(input []byte) error {
				return validateModelJSONFieldNames(input)
			},
		},
		{
			name:  "non-array modes",
			input: `{"modes":null}`,
			validate: func(input []byte) error {
				return validateModelJSONFieldNames(input)
			},
		},
		{
			name:  "override field alias",
			input: `{"overrides":{"sm-kms/` + cryptoutilSharedMagic.DeploymentPoolPG + `":{"Min":1}}}`,
			validate: func(input []byte) error {
				return validateModelJSONFieldNames(input)
			},
			wantErr: true,
		},
		{
			name:  "malformed schema object",
			input: testOpenJSONObject,
			validate: func(input []byte) error {
				return validateJSONSchemaFields(input, "test", "field")
			},
			wantErr: true,
		},
		{
			name:  "malformed array",
			input: `[`,
			validate: func(input []byte) error {
				_, _, err := jsonArrayElements(input)

				return err
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.validate([]byte(tc.input))
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestModelFieldNameValidationNestedReaderErrors(t *testing.T) {
	t.Parallel()

	injected := errors.New("injected nested JSON reader error")

	t.Run("override object reader", func(t *testing.T) {
		t.Parallel()

		calls := 0
		objectFields := func(data []byte) (map[string]json.RawMessage, bool, error) {
			calls++
			if calls == 1 {
				return map[string]json.RawMessage{
					testOverridesJSONField: json.RawMessage(`{}`),
				}, true, nil
			}

			return nil, false, injected
		}

		err := validateModelJSONFieldNamesWithReaders(
			[]byte(`{}`),
			objectFields,
			jsonArrayElements,
		)
		require.ErrorIs(t, err, injected)
	})
	t.Run("mode array reader", func(t *testing.T) {
		t.Parallel()

		objectFields := func([]byte) (map[string]json.RawMessage, bool, error) {
			return map[string]json.RawMessage{
				"modes": json.RawMessage(`[]`),
			}, true, nil
		}
		arrayElements := func([]byte) ([]json.RawMessage, bool, error) {
			return nil, false, injected
		}

		err := validateModelJSONFieldNamesWithReaders(
			[]byte(`{}`),
			objectFields,
			arrayElements,
		)
		require.ErrorIs(t, err, injected)
	})
	t.Run("bootstrap object reader", func(t *testing.T) {
		t.Parallel()

		calls := 0
		objectFields := func([]byte) (map[string]json.RawMessage, bool, error) {
			calls++
			if calls == 1 {
				return map[string]json.RawMessage{
					"bootstrap": json.RawMessage(`{`),
				}, true, nil
			}

			return nil, false, injected
		}

		err := validateModelJSONFieldNamesWithReaders(
			[]byte(`{}`),
			objectFields,
			jsonArrayElements,
		)
		require.ErrorContains(t, err, "decode JSON object fields")
	})
}
