package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractHostPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "host:container", input: "8080:8080", expected: "8080"},
		{name: "ip:host:container", input: "127.0.0.1:8080:8080", expected: "127.0.0.1:8080"},
		{name: "container only", input: "8080", expected: ""},
		{name: "quoted", input: "\"8080:8080\"", expected: "8080"},
		{name: "empty", input: "", expected: ""},
		{name: "too many parts", input: "a:b:c:d", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, extractHostPort(tc.input))
		})
	}
}

func TestExtractDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		svc      composeService
		expected []string
	}{
		{name: "nil depends_on", svc: composeService{}, expected: nil},
		{
			name:     "list format",
			svc:      composeService{DependsOn: []interface{}{"svc1", "svc2"}},
			expected: []string{"svc1", "svc2"},
		},
		{
			name:     "map format",
			svc:      composeService{DependsOn: map[string]interface{}{"svc1": map[string]interface{}{"condition": "service_started"}}},
			expected: []string{"svc1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractDependencies(&tc.svc)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractSecretName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{name: "string", input: "my_secret.secret", expected: "my_secret.secret"},
		{name: "map with source", input: map[string]interface{}{"source": "my_secret.secret"}, expected: "my_secret.secret"},
		{name: "map without source", input: map[string]interface{}{"target": "/run/secrets/foo"}, expected: ""},
		{name: "nil", input: nil, expected: ""},
		{name: "integer", input: 42, expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, extractSecretName(tc.input))
		})
	}
}

func TestExtractEnvironmentVars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		svc      composeService
		expected map[string]string
	}{
		{name: "nil", svc: composeService{}, expected: map[string]string{}},
		{
			name:     "map format",
			svc:      composeService{Environment: map[string]interface{}{"KEY1": "value1", "KEY2": nil}},
			expected: map[string]string{"KEY1": "value1", "KEY2": ""},
		},
		{
			name:     "list format",
			svc:      composeService{Environment: []interface{}{"KEY1=value1", "KEY2=value2", "NOVALUE"}},
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2", "NOVALUE": ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractEnvironmentVars(&tc.svc)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsExemptFromHealthcheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		svcName  string
		svc      composeService
		expected bool
	}{
		{name: "builder prefix", svcName: "builder-myapp", svc: composeService{}, expected: true},
		{name: "healthcheck prefix", svcName: "healthcheck-secrets", svc: composeService{}, expected: true},
		{name: "echo entrypoint", svcName: "init", svc: composeService{Entrypoint: []interface{}{"sh", "-c", "echo done"}}, expected: true},
		{name: "regular service", svcName: "myapp", svc: composeService{}, expected: false},
		{name: "non-echo entrypoint", svcName: "myapp", svc: composeService{Entrypoint: []interface{}{"app", "start"}}, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, isExemptFromHealthcheck(tc.svcName, &tc.svc))
		})
	}
}

func TestFormatComposeValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *ComposeValidationResult
		contains []string
	}{
		{
			name:     "pass",
			result:   &ComposeValidationResult{Path: "test.yml", Valid: true},
			contains: []string{"test.yml", "PASS"},
		},
		{
			name: "fail with errors and warnings",
			result: &ComposeValidationResult{
				Path: "test.yml", Valid: false,
				Errors: []string{"port conflict"}, Warnings: []string{"dep warning"},
			},
			contains: []string{"FAIL", "ERROR: port conflict", "WARNING: dep warning"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatComposeValidationResult(tc.result)
			for _, s := range tc.contains {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestSortedServiceNames(t *testing.T) {
	t.Parallel()

	compose := &composeFile{
		Services: map[string]composeService{"zulu": {}, "alpha": {}, "mike": {}},
	}
	names := sortedServiceNames(compose)
	assert.Equal(t, []string{"alpha", "mike", "zulu"}, names)
}

func TestParseComposeWithIncludes_MissingInclude(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `include:
  - path: nonexistent/compose.yml
services:
  myapp:
    image: myapp:latest
`
	composePath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	compose, err := parseComposeWithIncludes(composePath)
	require.NoError(t, err)
	require.NotNil(t, compose)
	assert.Len(t, compose.Services, 1)
}

func TestParseComposeWithIncludes_InvalidIncludeYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	includeDir := filepath.Join(dir, "shared")
	require.NoError(t, os.MkdirAll(includeDir, dirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(includeDir, "compose.yml"),
		[]byte("invalid: [yaml"), filePermissions))

	content := `include:
  - path: shared/compose.yml
services:
  myapp:
    image: myapp:latest
`
	composePath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	compose, err := parseComposeWithIncludes(composePath)
	require.NoError(t, err)
	require.NotNil(t, compose)
}

func TestMergeIncludedFile_EmptyPath(t *testing.T) {
	t.Parallel()

	compose := &composeFile{}
	mergeIncludedFile("/tmp", "", compose)
	assert.Nil(t, compose.Secrets)
}

func TestMergeIncludedFile_NilMaps(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	includeContent := `secrets:
  new.secret:
    file: ./secrets/new.secret
services:
  new-svc:
    image: new:latest
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "include.yml"),
		[]byte(includeContent), filePermissions))

	// compose starts with nil Secrets and nil Services.
	compose := &composeFile{}

	mergeIncludedFile(dir, "include.yml", compose)
	assert.NotNil(t, compose.Secrets)
	assert.NotNil(t, compose.Services)
	assert.Equal(t, "./secrets/new.secret", compose.Secrets["new.secret"].File)
	assert.Equal(t, "new:latest", compose.Services["new-svc"].Image)
}

func TestMergeIncludedFile_OverlappingSecrets(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	includeContent := `secrets:
  existing.secret:
    file: ./secrets/included.secret
  new.secret:
    file: ./secrets/new.secret
services:
  existing-svc:
    image: existing:latest
  new-svc:
    image: new:latest
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "include.yml"),
		[]byte(includeContent), filePermissions))

	compose := &composeFile{
		Secrets:  map[string]composeSecret{"existing.secret": {File: "./original.secret"}},
		Services: map[string]composeService{"existing-svc": {Image: "original:latest"}},
	}

	mergeIncludedFile(dir, "include.yml", compose)

	// Existing should NOT be overwritten.
	assert.Equal(t, "./original.secret", compose.Secrets["existing.secret"].File)
	assert.Equal(t, "original:latest", compose.Services["existing-svc"].Image)

	// New should be added.
	assert.Equal(t, "./secrets/new.secret", compose.Secrets["new.secret"].File)
	assert.Equal(t, "new:latest", compose.Services["new-svc"].Image)
}

func TestValidateComposeFile_ContainerOnlyPort(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `services:
  myapp:
    image: myapp:latest
    ports:
      - "8080"
    healthcheck:
      test: ["CMD", "true"]
`
	composePath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	result, err := ValidateComposeFile(composePath)
	require.NoError(t, err)
	assert.True(t, result.Valid, "container-only port should not cause conflicts: %v", result.Errors)
}

func TestValidateComposeFile_EmptySecretNameSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `services:
  myapp:
    image: myapp:latest
    secrets:
      - 42
    healthcheck:
      test: ["CMD", "true"]
secrets:
  real.secret:
    file: ./secrets/real.secret
`
	composePath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	result, err := ValidateComposeFile(composePath)
	require.NoError(t, err)
	assert.True(t, result.Valid, "integer secret refs should be skipped: %v", result.Errors)
}
