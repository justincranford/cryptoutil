// Copyright (c) 2025 Justin Cranford

package registry_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// numExpectedProducts is the number of products declared in registry.yaml.
	numExpectedProducts = 5
	// numExpectedProductServices is the number of product-services declared in registry.yaml.
	numExpectedProductServices = 10
	// migrationRangeTestEnd is a migration range end value used in validation test cases.
	migrationRangeTestEnd = 3000
)

// findProjectRoot walks up from the test working directory until it finds a go.mod file.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "get working directory")

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		require.NotEqualf(t, parent, dir, "go.mod not found walking up from %s", dir)

		dir = parent
	}
}

// writeRegistryYAML writes a YAML string to a temp file and returns the path.
func writeRegistryYAML(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "registry.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	return path
}

// minimalValidYAML returns a registry YAML with exactly 1 suite, 1 product, and 1 PS-ID.
// Used as the happy-path baseline for table-driven invalid-field tests.
func minimalValidYAML() string {
	return `
suites:
  - id: test-suite
    display_name: "Test Suite"
    cmd_dir: test-suite/
products:
  - id: ex
    display_name: "Example Product"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources:
      - /items
`
}

func TestLoadRegistry_RealFile(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	path := filepath.Join(root, "api", "cryptosuite-registry", "registry.yaml")

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err, "LoadRegistry must succeed on the real registry.yaml")

	assert.Len(t, r.Suites, 1, "expect exactly 1 suite")
	assert.Len(t, r.Products, numExpectedProducts, "expect exactly 5 products")
	assert.Len(t, r.ProductServices, numExpectedProductServices, "expect exactly 10 product-services")
}

func TestLoadRegistry_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := lintFitnessRegistry.LoadRegistry("/nonexistent/path/that/does/not/exist/registry.yaml")
	require.Error(t, err, "LoadRegistry must fail when file does not exist")
	assert.Contains(t, err.Error(), "read registry")
}

func TestLoadRegistry_InvalidYAMLSyntax(t *testing.T) {
	t.Parallel()

	path := writeRegistryYAML(t, "suites: [\n  bad: yaml\n  unclosed:")
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err, "LoadRegistry must fail on invalid YAML syntax")
}

func TestLoadRegistry_EmptySuites(t *testing.T) {
	t.Parallel()

	yaml := `
suites: []
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "suites: must contain at least one entry")
}

func TestLoadRegistry_EmptyProducts(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products: []
product_services:
  - ps_id: s-svc
    product: s
    service: svc
    display_name: "S Service"
    internal_apps_dir: s-svc/
    magic_file: magic_s.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "products: must contain at least one entry")
}

func TestLoadRegistry_EmptyProductServices(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product_services: must contain at least one entry")
}

func TestLoadRegistry_DuplicateSuiteID(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: dup
    display_name: "Dup Suite"
    cmd_dir: dup/
  - id: dup
    display_name: "Dup Suite 2"
    cmd_dir: dup/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate id dup")
}

func TestLoadRegistry_DuplicatePSID(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service Dup"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8001
    pg_host_port: 54321
    migration_range_start: 3001
    migration_range_end: 3999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate ps_id ex-svc")
}

func TestLoadRegistry_OverlappingBasePorts(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
  - ps_id: ex-alt
    product: ex
    service: alt
    display_name: "Example Alt"
    internal_apps_dir: ex-alt/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54321
    migration_range_start: 3001
    migration_range_end: 3999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "8000 already assigned to ex-svc")
}

func TestLoadRegistry_OverlappingPGPorts(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
  - ps_id: ex-alt
    product: ex
    service: alt
    display_name: "Example Alt"
    internal_apps_dir: ex-alt/
    magic_file: magic_ex.go
    base_port: 8001
    pg_host_port: 54320
    migration_range_start: 3001
    migration_range_end: 3999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "54320 already assigned to ex-svc")
}

func TestLoadRegistry_OverlappingMigrationRanges(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 3999
    api_resources: []
  - ps_id: ex-alt
    product: ex
    service: alt
    display_name: "Example Alt"
    internal_apps_dir: ex-alt/
    magic_file: magic_ex.go
    base_port: 8001
    pg_host_port: 54321
    migration_range_start: 3000
    migration_range_end: 4999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overlaps with ex-svc")
}

func TestLoadRegistry_MigrationRangeEndNotGreaterThanStart(t *testing.T) {
	t.Parallel()

	type tc struct {
		name  string
		start int
		end   int
	}

	tests := []tc{
		{"end equals start", 2001, 2001},
		{"end less than start", migrationRangeTestEnd, 2001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Replace migration range values.
			yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: ` + strconv.Itoa(tt.start) + `
    migration_range_end: ` + strconv.Itoa(tt.end) + `
    api_resources: []
`

			path := writeRegistryYAML(t, yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "migration_range_end")
		})
	}
}

func TestLoadRegistry_MigrationRangeStartBelowMinimum(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 100
    migration_range_end: 999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "migration_range_start")
}

func TestLoadRegistry_PSIDNotMatchProductService(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: wrong-name
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: wrong-name/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `expected "ex-svc"`)
}

func TestLoadRegistry_InternalAppsDirNotMatchPSID(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: wrong-dir/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `expected "ex-svc/"`)
}

func TestLoadRegistry_InternalAppsDirMissingTrailingSlash(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`
	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must end with '/'")
}

func TestLoadRegistry_EmptyRequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "empty suite id",
			yaml: `
suites:
  - id: ""
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "suite",
		},
		{
			name: "empty product display_name",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: ""
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "display_name",
		},
		{
			name: "empty PS magic_file",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: ""
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "magic_file",
		},
		{
			name: "empty PS ps_id",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ""
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "ps_id",
		},
		{
			name: "empty PS product",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ""
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".product:",
		},
		{
			name: "empty PS service",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: ""
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".service:",
		},
		{
			name: "empty PS display_name",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: ""
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".display_name:",
		},
		{
			name: "empty PS internal_apps_dir",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ""
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".internal_apps_dir:",
		},
		{
			name: "PS internal_apps_dir missing trailing slash",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "must end with '/'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := writeRegistryYAML(t, tt.yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantContains)
		})
	}
}

func TestLoadRegistry_ToProductServices(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	path := filepath.Join(root, "api", "cryptosuite-registry", "registry.yaml")

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err)

	pss := r.ToProductServices()
	require.Len(t, pss, numExpectedProductServices)

	// Verify the first PS matches expected fields.
	first := pss[0]
	assert.Equal(t, cryptoutilSharedMagic.OTLPServiceSMKMS, first.PSID)
	assert.Equal(t, "sm", first.Product)
	assert.Equal(t, cryptoutilSharedMagic.KMSServiceName, first.Service)
	assert.NotEmpty(t, first.DisplayName)
	assert.Equal(t, "sm-kms/", first.InternalAppsDir)
	assert.NotEmpty(t, first.MagicFile)
}

func TestLoadRegistry_ToProducts(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	path := filepath.Join(root, "api", "cryptosuite-registry", "registry.yaml")

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err)

	products := r.ToProducts()
	require.Len(t, products, numExpectedProducts)

	productIDs := make([]string, len(products))

	for i, p := range products {
		productIDs[i] = p.ID
	}

	assert.Contains(t, productIDs, "sm")
	assert.Contains(t, productIDs, cryptoutilSharedMagic.JoseProductName)
	assert.Contains(t, productIDs, cryptoutilSharedMagic.PKIProductName)
	assert.Contains(t, productIDs, cryptoutilSharedMagic.IdentityProductName)
	assert.Contains(t, productIDs, cryptoutilSharedMagic.SkeletonProductName)
}

func TestLoadRegistry_ToSuites(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	path := filepath.Join(root, "api", "cryptosuite-registry", "registry.yaml")

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err)

	suites := r.ToSuites()
	require.Len(t, suites, 1)
	assert.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, suites[0].ID)
}

func TestLoadRegistry_ValidMinimal(t *testing.T) {
	t.Parallel()

	path := writeRegistryYAML(t, minimalValidYAML())

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err)
	assert.Len(t, r.Suites, 1)
	assert.Len(t, r.Products, 1)
	assert.Len(t, r.ProductServices, 1)
}

func TestLoadRegistry_SuiteEmptyDisplayName(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: ""
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "display_name")
}

func TestLoadRegistry_SuiteEmptyCmdDir(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: ""
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cmd_dir")
}

func TestLoadRegistry_SuiteCmdDirMissingTrailingSlash(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must end with '/'")
}

func TestLoadRegistry_DuplicateProductID(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
  - id: ex
    display_name: "Example Dup"
    internal_apps_dir: ex2/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate id ex")
}

func TestLoadRegistry_ProductEmptyID(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ""
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".id:")
}

func TestLoadRegistry_ProductEmptyInternalAppsDir(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ""
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal_apps_dir")
}

func TestLoadRegistry_ProductEmptyCmdDir(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ""
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cmd_dir")
}

func TestLoadRegistry_ProductCmdDirMissingTrailingSlash(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must end with '/'")
}

func TestLoadRegistry_PSNegativeBasePort(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: -1
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base_port")
}

func TestLoadRegistry_PSNegativePGHostPort(t *testing.T) {
	t.Parallel()

	yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: -1
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

	path := writeRegistryYAML(t, yaml)
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pg_host_port")
}
