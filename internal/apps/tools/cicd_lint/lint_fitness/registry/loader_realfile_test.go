// Copyright (c) 2025 Justin Cranford
//

package registry_test

import (
	"os"
	"path/filepath"
	"testing"

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
	// migrationRangeTestBelowMinStart is a migration start value below the required minimum.
	migrationRangeTestBelowMinStart = 100
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

	require.Len(t, r.Suites, 1, "expect exactly 1 suite")
	require.Len(t, r.Products, numExpectedProducts, "expect exactly 5 products")
	require.Len(t, r.ProductServices, numExpectedProductServices, "expect exactly 10 product-services")
}

func TestLoadRegistry_ValidMinimal(t *testing.T) {
	t.Parallel()

	path := writeRegistryYAML(t, minimalValidYAML())

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err)
	require.Len(t, r.Suites, 1)
	require.Len(t, r.Products, 1)
	require.Len(t, r.ProductServices, 1)
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
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSMKMS, first.PSID)
	require.Equal(t, "sm", first.Product)
	require.Equal(t, cryptoutilSharedMagic.KMSServiceName, first.Service)
	require.NotEmpty(t, first.DisplayName)
	require.Equal(t, "sm-kms/", first.InternalAppsDir)
	require.NotEmpty(t, first.MagicFile)
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

	require.Contains(t, productIDs, "sm")
	require.Contains(t, productIDs, cryptoutilSharedMagic.JoseProductName)
	require.Contains(t, productIDs, cryptoutilSharedMagic.PKIProductName)
	require.Contains(t, productIDs, cryptoutilSharedMagic.IdentityProductName)
	require.Contains(t, productIDs, cryptoutilSharedMagic.SkeletonProductName)
}

func TestLoadRegistry_ToSuites(t *testing.T) {
	t.Parallel()

	root := findProjectRoot(t)
	path := filepath.Join(root, "api", "cryptosuite-registry", "registry.yaml")

	r, err := lintFitnessRegistry.LoadRegistry(path)
	require.NoError(t, err)

	suites := r.ToSuites()
	require.Len(t, suites, 1)
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, suites[0].ID)
}
