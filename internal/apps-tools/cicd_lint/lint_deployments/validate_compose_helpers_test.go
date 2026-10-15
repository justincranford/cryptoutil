package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
		{name: cryptoutilSharedMagic.EMPTY, input: "", expected: ""},
		{name: "too many parts", input: "a:b:c:d", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, extractHostPort(tc.input))
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
			name:     testListFormat,
			svc:      composeService{DependsOn: []any{testSvc1, testSvc2Value24BCAB51}},
			expected: []string{testSvc1, testSvc2Value24BCAB51},
		},
		{
			name:     testMapFormat,
			svc:      composeService{DependsOn: map[string]any{testSvc1: map[string]any{"condition": "service_started"}}},
			expected: []string{testSvc1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractDependencies(&tc.svc)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractSecretName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{name: testString, input: testMySecretSecret, expected: testMySecretSecret},
		{name: "map with source", input: map[string]any{"source": testMySecretSecret}, expected: testMySecretSecret},
		{name: "map without source", input: map[string]any{"target": "/run/secrets/foo"}, expected: ""},
		{name: testNil, input: nil, expected: ""},
		{name: "integer", input: cryptoutilSharedMagic.AnswerToLifeUniverseEverything, expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, extractSecretName(tc.input))
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
		{name: testNil, svc: composeService{}, expected: map[string]string{}},
		{
			name:     testMapFormat,
			svc:      composeService{Environment: map[string]any{testKey1: testValue1, testKey2: nil}},
			expected: map[string]string{testKey1: testValue1, testKey2: ""},
		},
		{
			name:     testListFormat,
			svc:      composeService{Environment: []any{"KEY1=value1", "KEY2=value2", testNovalue}},
			expected: map[string]string{testKey1: testValue1, testKey2: "value2", testNovalue: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := extractEnvironmentVars(&tc.svc)
			require.Equal(t, tc.expected, result)
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
		{name: "healthcheck prefix", svcName: cryptoutilSharedMagic.DockerJobHealthcheckSecrets, svc: composeService{}, expected: true},
		{name: "setup suffix", svcName: "migration-setup", svc: composeService{}, expected: true},
		{name: "init suffix", svcName: "db-init", svc: composeService{}, expected: true},
		{name: "echo entrypoint", svcName: "init", svc: composeService{Entrypoint: []any{"sh", "-c", "echo done"}}, expected: true},
		{name: "override-only port service", svcName: cryptoutilSharedMagic.KMSE2ESQLiteContainer, svc: composeService{Ports: []string{"18000:8080"}}, expected: true},
		{name: "regular service", svcName: testMyapp, svc: composeService{}, expected: false},
		{name: "non-echo entrypoint", svcName: testMyapp, svc: composeService{Entrypoint: []any{"app", "start"}}, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, isExemptFromHealthcheck(tc.svcName, &tc.svc))
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
			result:   &ComposeValidationResult{Path: testTestYml, Valid: true},
			contains: []string{testTestYml, cryptoutilSharedMagic.TestStatusPass},
		},
		{
			name: "fail with errors and warnings",
			result: &ComposeValidationResult{
				Path: testTestYml, Valid: false,
				Errors: []string{testPortConflict}, Warnings: []string{"dep warning"},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, "ERROR: port conflict", "WARNING: dep warning"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatComposeValidationResult(tc.result)
			for _, s := range tc.contains {
				require.Contains(t, output, s)
			}
		})
	}
}

func TestSortedServiceNames(t *testing.T) {
	t.Parallel()

	compose := &composeFile{
		Services: map[string]composeService{testZulu: {}, testAlpha: {}, testMike: {}},
	}
	names := sortedServiceNames(compose)
	require.Equal(t, []string{testAlpha, testMike, testZulu}, names)
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
	composePath := filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	compose, err := parseComposeWithIncludes(composePath)
	require.NoError(t, err)
	require.NotNil(t, compose)
	require.Len(t, compose.Services, 1)
}

func TestParseComposeWithIncludes_InvalidIncludeYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	includeDir := filepath.Join(dir, "shared")
	require.NoError(t, os.MkdirAll(includeDir, dirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(includeDir, cryptoutilSharedMagic.COMPOSE_YML),
		[]byte("invalid: [yaml"), filePermissions))

	content := `include:
  - path: shared/compose.yml
services:
  myapp:
    image: myapp:latest
`
	composePath := filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	compose, err := parseComposeWithIncludes(composePath)
	require.NoError(t, err)
	require.NotNil(t, compose)
}

func TestMergeIncludedFile_EmptyPath(t *testing.T) {
	t.Parallel()

	compose := &composeFile{}
	mergeIncludedFile("/tmp", "", compose)
	require.Nil(t, compose.Secrets)
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
	require.NotNil(t, compose.Secrets)
	require.NotNil(t, compose.Services)
	require.Equal(t, "./secrets/new.secret", compose.Secrets["new.secret"].File)
	require.Equal(t, "new:latest", compose.Services["new-svc"].Image)
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
	require.Equal(t, "./original.secret", compose.Secrets["existing.secret"].File)
	require.Equal(t, "original:latest", compose.Services["existing-svc"].Image)

	// New should be added.
	require.Equal(t, "./secrets/new.secret", compose.Secrets["new.secret"].File)
	require.Equal(t, "new:latest", compose.Services["new-svc"].Image)
}

func TestParseComposeWithIncludes_NestedServiceDependencies(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serviceDir := filepath.Join(dir, "service")
	productDir := filepath.Join(dir, "product")

	suiteDir := filepath.Join(dir, "suite")
	for _, nestedDir := range []string{serviceDir, productDir, suiteDir} {
		require.NoError(t, os.MkdirAll(nestedDir, dirPermissions))
	}

	serviceCompose := fmt.Sprintf(`services:
  app:
    depends_on:
      cert-copy:
        condition: service_completed_successfully
  cert-copy:
    depends_on:
      %s:
        condition: service_completed_successfully
  %s:
    command: initialize
`, cryptoutilSharedMagic.PSIDPKIInit, cryptoutilSharedMagic.PSIDPKIInit)
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, composeFileName), []byte(serviceCompose), filePermissions))

	productCompose := fmt.Sprintf(`include:
  - path: ../service/%s
services:
  app:
    volumes:
      - product-certs:/certs:ro
`, composeFileName)
	require.NoError(t, os.WriteFile(filepath.Join(productDir, composeFileName), []byte(productCompose), filePermissions))

	suiteCompose := fmt.Sprintf(`include:
  - path: ../product/%s
services:
  app:
    volumes:
      - suite-certs:/certs:ro
`, composeFileName)
	suiteComposePath := filepath.Join(suiteDir, composeFileName)
	require.NoError(t, os.WriteFile(suiteComposePath, []byte(suiteCompose), filePermissions))

	compose, err := parseComposeWithIncludes(suiteComposePath)
	require.NoError(t, err)

	app := compose.Services["app"]
	require.True(t, dependsOnSuccessfulPKIInit(compose, "app", make(map[string]struct{})))
	require.Equal(t, []composeVolume{{
		Raw:    "suite-certs:/certs:ro",
		Type:   cryptoutilSharedMagic.DockerComposeVolumeMountTypeVolume,
		Source: "suite-certs",
		Target: "/certs",
	}}, app.Volumes)
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
	composePath := filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	result, err := ValidateComposeFile(composePath)
	require.NoError(t, err)
	require.True(t, result.Valid, "container-only port should not cause conflicts: %v", result.Errors)
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
	composePath := filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

	result, err := ValidateComposeFile(composePath)
	require.NoError(t, err)
	require.True(t, result.Valid, "integer secret refs should be skipped: %v", result.Errors)
}
