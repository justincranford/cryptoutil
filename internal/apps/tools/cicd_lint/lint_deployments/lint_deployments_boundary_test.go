package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestValidatePorts_ExactBoundaryPorts verifies that ports at the exact min and max
// of each deployment range are accepted, killing BOUNDARY mutants on < and > operators.
func TestValidatePorts_ExactBoundaryPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		port  int
		level string
		valid bool
	}{
		// Service range [8000-8999]: exact min and max are valid.
		{name: "service exact min", port: servicePortMin, level: DeploymentTypeProductService, valid: true},
		{name: "service exact max", port: servicePortMax, level: DeploymentTypeProductService, valid: true},
		// One below min and one above max are invalid.
		{name: "service below min", port: servicePortMin - 1, level: DeploymentTypeProductService, valid: false},
		{name: "service above max", port: servicePortMax + 1, level: DeploymentTypeProductService, valid: false},
		// Product range [18000-18999]: exact boundaries.
		{name: "product exact min", port: productPortMin, level: DeploymentTypeProduct, valid: true},
		{name: "product exact max", port: productPortMax, level: DeploymentTypeProduct, valid: true},
		{name: "product below min", port: productPortMin - 1, level: DeploymentTypeProduct, valid: false},
		{name: "product above max", port: productPortMax + 1, level: DeploymentTypeProduct, valid: false},
		// Suite range [28000-28999]: exact boundaries.
		{name: "suite exact min", port: suitePortMin, level: DeploymentTypeSuite, valid: true},
		{name: "suite exact max", port: suitePortMax, level: DeploymentTypeSuite, valid: true},
		{name: "suite below min", port: suitePortMin - 1, level: DeploymentTypeSuite, valid: false},
		{name: "suite above max", port: suitePortMax + 1, level: DeploymentTypeSuite, valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			compose := fmt.Sprintf("services:\n  svc:\n    ports:\n      - \"%d:8080\"\n", tc.port)
			dir := createDeploymentWithCompose(t, compose)

			result, err := ValidatePorts(dir, "test-svc", tc.level)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.valid, result.Valid, "port %d in %s range", tc.port, tc.level)
		})
	}
}

// TestValidateConfigPortValue_ExactBoundary verifies that config bind-public-port
// at exact min/max values is accepted, killing BOUNDARY mutants on < and > operators.
func TestValidateConfigPortValue_ExactBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		port  int
		level string
		valid bool
	}{
		{name: "service exact min", port: servicePortMin, level: DeploymentTypeProductService, valid: true},
		{name: "service exact max", port: servicePortMax, level: DeploymentTypeProductService, valid: true},
		{name: "service below min", port: servicePortMin - 1, level: DeploymentTypeProductService, valid: false},
		{name: "service above max", port: servicePortMax + 1, level: DeploymentTypeProductService, valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := map[string]any{"bind-public-port": tc.port}
			result := &PortValidationResult{Valid: true}

			validateConfigPortValue(config, "/fake/config.yml", "test-svc", tc.level, result)
			require.Equal(t, tc.valid, result.Valid, "config port %d in %s range", tc.port, tc.level)
		})
	}
}

// TestCheckOTLPProtocolOverride_LineNumber verifies the line number in OTLP protocol
// warnings is correct, killing INCREMENT_DECREMENT mutant on lineNumber++.
func TestCheckOTLPProtocolOverride_LineNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		content        string
		wantLine       int
		wantNoWarnings bool
	}{
		{
			name:     "protocol on line 3",
			content:  "key1: value1\nkey2: value2\notlp-endpoint: grpc://collector:4317\n",
			wantLine: 3,
		},
		{
			name:     "protocol on line 5",
			content:  "line1: a\nline2: b\nline3: c\nline4: d\notlp-endpoint: http://collector:4318\n",
			wantLine: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
		},
		{
			name:     "protocol on line 1",
			content:  "otlp-endpoint: grpc://collector:4317\n",
			wantLine: 1,
		},
		{
			name:           "no protocol prefix - no warning",
			content:        "otlp-endpoint: collector:4317\n",
			wantNoWarnings: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			configDir := filepath.Join(dir, "config")
			require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, "config-test.yml"), []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			result := &ValidationResult{Valid: true}
			checkOTLPProtocolOverride(dir, "test-svc", DeploymentTypeProductService, result)

			if tc.wantNoWarnings {
				require.Empty(t, result.Warnings)

				return
			}

			require.Len(t, result.Warnings, 1)

			expectedLineRef := fmt.Sprintf("config-test.yml:%d:", tc.wantLine)
			require.Contains(t, result.Warnings[0], expectedLineRef,
				"warning should reference exact line %d", tc.wantLine)
		})
	}
}

// TestFormatResults_SortOrderBoundary verifies exact output order when
// multiple invalid and valid items are present, killing NEGATION mutants
// on the sort comparator.
func TestFormatResults_SortOrderBoundary(t *testing.T) {
	t.Parallel()

	results := []ValidationResult{
		{Path: "/deploy/b-valid", Type: "PRODUCT-SERVICE", Valid: true},
		{Path: "/deploy/z-invalid", Type: "PRODUCT-SERVICE", Valid: false, Errors: []string{"err1"}},
		{Path: "/deploy/a-valid", Type: "PRODUCT-SERVICE", Valid: true},
		{Path: "/deploy/m-invalid", Type: "PRODUCT-SERVICE", Valid: false, Errors: []string{"err2"}},
	}

	output := FormatResults(results)

	// All invalid items must appear before all valid items.
	firstValid := strings.Index(output, "a-valid")
	lastInvalid := strings.LastIndex(output, "z-invalid")
	require.Greater(t, firstValid, lastInvalid, "all invalid items must come before all valid items")

	// Within each group, items should be sorted alphabetically by path.
	mInvalidIdx := strings.Index(output, "m-invalid")
	zInvalidIdx := strings.Index(output, "z-invalid")
	require.Less(t, mInvalidIdx, zInvalidIdx, "invalid items should be sorted by path")

	aValidIdx := strings.Index(output, "a-valid")
	bValidIdx := strings.Index(output, "b-valid")
	require.Less(t, aValidIdx, bValidIdx, "valid items should be sorted by path")
}

// TestCountFailed verifies countFailed counts ValidatorResult entries where Passed==false.
func TestCountFailed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results []ValidatorResult
		want    int
	}{
		{
			name:    "all pass",
			results: []ValidatorResult{{Passed: true}, {Passed: true}},
			want:    0,
		},
		{
			name:    "one fail",
			results: []ValidatorResult{{Passed: true}, {Passed: false}},
			want:    1,
		},
		{
			name:    "all fail",
			results: []ValidatorResult{{Passed: false}, {Passed: false}},
			want:    2,
		},
		{
			name:    "empty results",
			results: nil,
			want:    0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := &AllValidationResult{Results: tc.results}
			require.Equal(t, tc.want, countFailed(result))
		})
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestLint_WithRealDirs(t *testing.T) {
	original, err := os.Getwd()
	require.NoError(t, err)

	// Navigate five directories up from the package dir to reach the project root.
	// Package path: internal/apps/tools/cicd_lint/lint_deployments/ (5 levels below module root).
	root := filepath.Join(original, "../../../../../")

	if _, statErr := os.Stat(filepath.Join(root, "deployments")); os.IsNotExist(statErr) {
		t.Skip("cannot locate project deployments/ directory — not running from expected workspace")
	}

	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { require.NoError(t, os.Chdir(original)) })

	logger := cryptoutilCicdCommon.NewLogger("test-lint-real-dirs")
	require.NoError(t, Lint(logger), "Lint must succeed on real project files")
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestLint_WithInvalidDir(t *testing.T) {
	original, err := os.Getwd()
	require.NoError(t, err)

	// Chdir to a temp directory where deployments/ and configs/ do not exist.
	// ValidateNaming will return Valid=false for non-existent paths, causing Lint to error.
	emptyDir := t.TempDir()
	require.NoError(t, os.Chdir(emptyDir))
	t.Cleanup(func() { require.NoError(t, os.Chdir(original)) })

	logger := cryptoutilCicdCommon.NewLogger("test-lint-invalid-dir")
	err = Lint(logger)
	require.Error(t, err, "Lint must fail when deployments/ and configs/ do not exist")
	require.Contains(t, err.Error(), "lint-deployments failed")
}

// TestFormatResults_EmptySlicesSectionsOmitted verifies that empty MissingDirs
// and MissingSecrets don't produce output, killing BOUNDARY mutant on len > 0.
func TestFormatResults_EmptySlicesSectionsOmitted(t *testing.T) {
	t.Parallel()

	// Result with all optional fields empty.
	results := []ValidationResult{
		{
			Path:           "/deploy/test-svc",
			Type:           "PRODUCT-SERVICE",
			Valid:          true,
			MissingDirs:    []string{},
			MissingFiles:   []string{},
			MissingSecrets: []string{},
			Errors:         []string{},
			Warnings:       []string{},
		},
	}

	output := FormatResults(results)

	require.NotContains(t, output, "Missing directories")
	require.NotContains(t, output, "Missing files")
	require.NotContains(t, output, "Missing secrets")
	require.NotContains(t, output, "ERROR:")
	require.NotContains(t, output, "WARN:")
}

// TestValidateStructuralMirror_DirArgIsFile verifies that passing a regular file
// as the deploymentsDir or configsDir triggers the expected error without requiring
// os.Chmod (which does not restrict access on Windows NTFS).
func TestValidateStructuralMirror_DirArgIsFile(t *testing.T) {
	t.Parallel()

	validDir := t.TempDir()
	fileInDir := filepath.Join(t.TempDir(), "not-a-dir.txt")
	require.NoError(t, os.WriteFile(fileInDir, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

	tests := []struct {
		name            string
		deploymentsDir  string
		configsDir      string
		wantErrContains string
	}{
		{
			name:            "deployments dir is a file",
			deploymentsDir:  fileInDir,
			configsDir:      validDir,
			wantErrContains: "failed to list deployment directories",
		},
		{
			name:            "configs dir is a file",
			deploymentsDir:  validDir,
			configsDir:      fileInDir,
			wantErrContains: "failed to list config directories",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ValidateStructuralMirror(tc.deploymentsDir, tc.configsDir)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErrContains)
		})
	}
}
