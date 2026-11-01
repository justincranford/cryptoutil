//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	json "encoding/json"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testModelContext              = "model:"
	testRegistryContext           = "registry:"
	testReplicasKey               = "replicas"
	testUnknownJSONField          = "unknown field"
	testTopLevelJSONObject        = "top-level JSON object"
	testOpenJSONObject            = "{"
	testNullDocument              = "null"
	testDuplicateJSONKey          = "duplicate JSON object key"
	testDuplicateYAMLKey          = "duplicate YAML mapping key"
	testSingleJSONDocument        = "exactly one"
	testNonNullYAMLMapping        = "non-null YAML mapping"
	testUnexpectedYAMLField       = "field unexpected not found"
	testSingleYAMLDocument        = "exactly one YAML document"
	testInvalidDNSName            = "invalid DNS name"
	testDNSNameError              = "DNS name"
	testExampleValue              = "example"
	testDefaultReplicaJSON        = `"replicas":{"min":2,"max":2}`
	testTruncatedJSONArray        = `{"value": [1`
	testSuitePGSelector           = "u:cryptoutil:pg"
	testProductSMMemSelector      = "p:sm:mem"
	testServiceSMKMSPGSelector    = "s:sm-kms:pg"
	testServiceSMKMSMemSelector   = "s:sm-kms:mem"
	testInvalidVersion            = "invalid version"
	testInvalidDomain             = "invalid domain"
	testPoolError                 = "pool"
	testNonemptyError             = "nonempty"
	testPSIDError                 = "PS-ID"
	testProductIdentity           = "product"
	testSuiteIdentity             = "suite"
	testExtraIdentity             = "extra"
	testEmptyProductError         = "empty product"
	testDuplicateIdentityError    = "duplicate identity"
	testExpandedPodCapacityError  = "expanded pod capacity"
	testSQLIdentifierError        = "SQL identifier"
	testInvalidKubernetesIdentity = "Invalid"
	testThree                     = 3
)

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
			after:  `"run":"` + m.Run + `","Run":"other"`,
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
			before: `"name":"` + poolPG + `"`,
			after:  `"Name":"` + poolPG + `"`,
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
		{name: "unknown field", input: `{"unexpected": true}`, want: testUnknownJSONField},
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
			input: `{"overrides":{"sm-kms/` + poolPG + `":{"Min":1}}}`,
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

func TestRegistryDecodingStrictness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "malformed YAML", input: "suites: [\n", want: testRegistryContext},
		{name: "null document", input: "null\n", want: testNonNullYAMLMapping},
		{name: "sequence document", input: "- suites\n", want: testNonNullYAMLMapping},
		{name: "scalar document", input: "registry\n", want: testNonNullYAMLMapping},
		{name: "duplicate top-level YAML key", input: "suites: []\nsuites: []\n", want: testDuplicateYAMLKey},
		{name: "duplicate nested YAML key", input: "suites: [{id: first, id: second}]\n", want: testDuplicateYAMLKey},
		{name: "unknown top-level field", input: "unexpected: true\n", want: testUnexpectedYAMLField},
		{name: "unknown nested field", input: "suites: [{id: first, unexpected: true}]\n", want: testUnexpectedYAMLField},
		{name: "field from another mapping type", input: "suites: [{id: first, internal_apps_dir: first/}]\n", want: "field internal_apps_dir not found"},
		{name: "multiple YAML documents", input: "suites: []\n---\nproducts: []\n", want: testSingleYAMLDocument},
		{name: "malformed second document", input: "suites: []\n---\n[\n", want: testSingleYAMLDocument},
		{name: "single mapping", input: "suites: []\nproducts: []\nproduct_services: []\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := decodeRegistry([]byte(tc.input))
			if tc.want == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestExpandedDNSNameBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		extraByte   int
		domainBytes int
		wantError   string
		wantDomain  bool
	}{
		{name: "full name at DNS limit"},
		{name: "expanded name exceeds limit", extraByte: 1, wantError: testInvalidDNSName},
		{name: "suffix above legacy cutoff remains valid", domainBytes: 201, wantError: testInvalidDNSName, wantDomain: true},
		{name: "maximum suffix length is permitted", domainBytes: 253, wantError: testInvalidDNSName, wantDomain: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			prefix := "sm-kms-pg-0.sm-kms-pg-pods.cu-" + m.Run + "-s-sm-kms-pg.svc."

			domainLength := 253 - len(prefix) + tc.extraByte
			if tc.domainBytes > 0 {
				domainLength = tc.domainBytes
			}

			var labels []string
			for remaining := domainLength; remaining > 0; {
				if len(labels) > 0 {
					remaining--
				}

				labelLength := min(63, remaining)
				labels = append(labels, "1"+strings.Repeat("a", labelLength-1))
				remaining -= labelLength
			}

			m.ClusterDomain = strings.Join(labels, ".")
			if tc.wantDomain {
				require.Greater(t, len(m.ClusterDomain), 200)
				require.NoError(t, validate(m, r))
			}

			inv, err := build(m, r, testServiceSMKMSPGSelector)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.Equal(t, 253, len(inv.Releases[0].Workloads[0].Pods[0].FQDN))
			require.True(t, dnsLabel(strings.Split(m.ClusterDomain, ".")[0]))
		})
	}
}

func TestDNSValidationHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		label       string
		dnsName     string
		wantLabel   bool
		wantDNSName bool
	}{
		{name: "numeric label", label: "123", dnsName: "123.example", wantLabel: true, wantDNSName: true},
		{name: "zero-length name", dnsName: "", wantDNSName: false},
		{name: "missing domain label", label: "", dnsName: "example..local", wantDNSName: false},
		{name: "label too long", label: strings.Repeat("a", cryptoutilSharedMagic.FQDNLabelMaxLength+1), dnsName: "example.local", wantDNSName: true},
		{name: "name too long", label: testExampleValue, dnsName: strings.Repeat("a", cryptoutilSharedMagic.FQDNMaxLength+1), wantLabel: true, wantDNSName: false},
		{name: "leading hyphen", label: "-example", dnsName: "-example.local", wantDNSName: false},
		{name: "trailing hyphen", label: "example-", dnsName: "example-.local", wantDNSName: false},
		{name: "uppercase", label: "Example", dnsName: "Example.local", wantDNSName: false},
		{name: "invalid character", label: "bad_name", dnsName: "bad_name.local", wantDNSName: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantLabel, dnsLabel(tc.label))
			require.Equal(t, tc.wantDNSName, dnsName(tc.dnsName))
		})
	}
}
