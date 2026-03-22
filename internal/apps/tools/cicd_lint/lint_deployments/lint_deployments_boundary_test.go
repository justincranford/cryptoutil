package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
