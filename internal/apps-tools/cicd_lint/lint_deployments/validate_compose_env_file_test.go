// Copyright (c) 2025-2026 Justin Cranford.
package lint_deployments

import (
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestExtractComposeEnvFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		want    []composeEnvFile
		wantErr string
	}{
		{name: "nil", value: nil},
		{name: testSinglePath, value: testComposeEnvFile, want: []composeEnvFile{{path: testComposeEnvFile, required: true}}},
		{
			name:  "sequence with required defaults and optional mapping",
			value: []any{testComposeEnvFile, map[string]any{testComposePathField: testOptionalEnvFile, testComposeRequiredField: false}},
			want: []composeEnvFile{
				{path: testComposeEnvFile, required: true},
				{path: testOptionalEnvFile, required: false},
			},
		},
		{name: "empty path", value: "", wantErr: testMustNotBeEmpty},
		{name: "empty sequence path", value: []any{""}, wantErr: testMustNotBeEmpty},
		{name: "mapping missing path", value: []any{map[string]any{testComposeRequiredField: false}}, wantErr: testNonEmptyStringPath},
		{name: "mapping with non-string path", value: []any{map[string]any{testComposePathField: 42}}, wantErr: testNonEmptyStringPath},
		{name: "mapping with invalid required value", value: []any{map[string]any{testComposePathField: testComposeEnvFile, testComposeRequiredField: "false"}}, wantErr: testMustBeABoolean},
		{name: "sequence with invalid entry", value: []any{testComposeEnvFile, 42}, wantErr: testStringOrMapping},
		{name: "unsupported type", value: 42, wantErr: testStringOrSequence},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := extractComposeEnvFiles(tc.value)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestResolveComposeEnvFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr string
	}{
		{name: "relative path", path: testComposeEnvFile, want: filepath.Join("compose", testComposeEnvFile)},
		{name: "absolute path", path: filepath.Join(t.TempDir(), testComposeEnvFile)},
		{name: "unresolved interpolation", path: "${ENV_FILE}", wantErr: testUnresolvedVariableInterpolation},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveComposeEnvFilePath("compose", tc.path)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			if tc.want != "" {
				require.Equal(t, tc.want, got)

				return
			}

			require.Equal(t, tc.path, got)
		})
	}
}

func TestParseComposeEnvFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents string
		want     []composeEnvironmentVariable
	}{
		{name: "empty and comments", contents: " # comment\r\n \t \r\n"},
		{
			name:     "exported quoted value and missing separator",
			contents: " export " + testKey1 + "=\"" + testValue1 + "\"\r\n" + testNovalue + "\n",
			want: []composeEnvironmentVariable{
				{key: testKey1, value: testValue1},
				{key: testNovalue},
			},
		},
		{
			name:     "colon separator with single quotes",
			contents: testKey2 + ": '" + testValue1 + "'\n",
			want:     []composeEnvironmentVariable{{key: testKey2, value: testValue1}},
		},
		{
			name:     "mismatched quotes are preserved",
			contents: testKey1 + "=\"" + testValue1 + "'\n",
			want:     []composeEnvironmentVariable{{key: testKey1, value: "\"" + testValue1 + "'"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, parseComposeEnvFile(tc.contents))
		})
	}
}

func TestDockerSecretTargetName(t *testing.T) {
	t.Parallel()

	targetName := testDatabasePassword
	tests := []struct {
		name string
		ref  any
		want string
	}{
		{name: "string reference", ref: testMySecretSecret, want: testMySecretSecret},
		{name: "configured target", ref: map[string]any{testComposeSourceField: testMySecretSecret, testComposeTargetField: targetName}, want: targetName},
		{name: "secrets directory target", ref: map[string]any{testComposeSourceField: testMySecretSecret, testComposeTargetField: cryptoutilSharedMagic.DockerSecretsDir + "/" + targetName}, want: targetName},
		{name: "empty target uses source", ref: map[string]any{testComposeSourceField: testMySecretSecret, testComposeTargetField: ""}, want: testMySecretSecret},
		{name: "non-string target uses source", ref: map[string]any{testComposeSourceField: testMySecretSecret, testComposeTargetField: 42}, want: testMySecretSecret},
		{name: "traversal target rejected", ref: map[string]any{testComposeSourceField: testMySecretSecret, testComposeTargetField: "../" + targetName}},
		{name: "nested secret name rejected", ref: "nested/" + testMySecretSecret},
		{name: "unsupported reference rejected", ref: 42},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, dockerSecretTargetName(tc.ref))
		})
	}
}

func TestIsDockerSecretFileReference(t *testing.T) {
	t.Parallel()

	targetName := testDatabasePassword
	service := composeService{
		Secrets: []any{
			testMySecretSecret,
			map[string]any{testComposeSourceField: testExternalSecret, testComposeTargetField: targetName},
			map[string]any{testComposeSourceField: testUnconfiguredSecret},
			42,
		},
	}
	compose := composeFile{
		Secrets: map[string]composeSecret{
			testMySecretSecret:     {File: "./secrets/" + testMySecretSecret},
			testExternalSecret:     {External: true},
			testUnconfiguredSecret: {},
		},
	}
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "mounted file secret", value: cryptoutilSharedMagic.DockerSecretsDir + "/" + testMySecretSecret, want: true},
		{name: "externally managed secret target", value: cryptoutilSharedMagic.DockerSecretsDir + "/" + targetName, want: true},
		{name: "wrong directory", value: "/other/" + testMySecretSecret},
		{name: "invalid secret path", value: cryptoutilSharedMagic.DockerSecretsDir + "/../" + testMySecretSecret},
		{name: "undeclared secret", value: cryptoutilSharedMagic.DockerSecretsDir + "/missing-secret"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, isDockerSecretFileReference(tc.value, &service, &compose))
		})
	}
}
