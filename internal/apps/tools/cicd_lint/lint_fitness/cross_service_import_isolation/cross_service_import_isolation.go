// Copyright (c) 2025 Justin Cranford

// Package cross_service_import_isolation verifies that no service package imports
// another product's service internal package. This enforces product boundary
// isolation: service code in product A must not import internal packages
// from product B. Same-product cross-service imports are permitted since those
// are expected during migration and are handled per-product.
// Imports of internal/apps/framework/ and internal/shared/ are always allowed.
package cross_service_import_isolation

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// importPrefixSplitN splits import path into product, service, and rest.
	importPrefixSplitN = 3

	// cicdProductName is the cicd tooling product; excluded from isolation checks.
	cicdProductName = "cicd"
)

var importLinePattern = regexp.MustCompile(`^\s+(?:\w+ )?"([^"]+)"`)

// Test seams: replaceable in tests to exercise unreachable OS-level error paths.
// See ARCHITECTURE.md Section 10.2.4 (Test Seam Injection Pattern).
var crossServiceWalkFn = filepath.Walk

// Check verifies cross-service import isolation from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir verifies cross-service import isolation under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking cross-service import isolation...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	appsDir := filepath.Join(projectRoot, "internal", "apps")

	services, err := collectServices(appsDir)
	if err != nil {
		return fmt.Errorf("failed to collect services: %w", err)
	}

	var violations []string

	for _, svc := range services {
		svcDir := filepath.Join(appsDir, svc.product, svc.service)

		if walkErr := walkServiceImports(projectRoot, svcDir, svc, services, &violations); walkErr != nil {
			return fmt.Errorf("failed to scan service %s/%s: %w", svc.product, svc.service, walkErr)
		}
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d cross-service import isolation violations", len(violations))
	}

	logger.Log("Cross-service import isolation check passed")

	return nil
}

// serviceRef identifies a product+service pair.
type serviceRef struct {
	product string
	service string
}

// collectServices discovers all product/service directories under appsDir.
// Only directories with a server/ subdirectory qualify as real services.
// Directories starting with _ (archived) are excluded.
// The cicd and skeleton products are excluded as they are tooling/templates.
func collectServices(appsDir string) ([]serviceRef, error) {
	var services []serviceRef

	products, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, fmt.Errorf("read apps dir: %w", err)
	}

	for _, productEntry := range products {
		if !productEntry.IsDir() {
			continue
		}

		product := productEntry.Name()

		// Skip archived directories, cicd tooling, and skeleton templates.
		if strings.HasPrefix(product, "_") || product == cicdProductName || product == cryptoutilSharedMagic.SkeletonProductName {
			continue
		}

		productDir := filepath.Join(appsDir, product)

		serviceEntries, err := os.ReadDir(productDir)
		if err != nil {
			return nil, fmt.Errorf("read product dir %s: %w", product, err)
		}

		for _, svcEntry := range serviceEntries {
			if !svcEntry.IsDir() {
				continue
			}

			// Skip archived service directories.
			if strings.HasPrefix(svcEntry.Name(), "_") {
				continue
			}

			// Only directories with server/ are real services.
			// Shared packages (identity/domain, identity/config, etc.) lack server/.
			serverDir := filepath.Join(productDir, svcEntry.Name(), "server")
			if _, statErr := os.Stat(serverDir); os.IsNotExist(statErr) {
				continue
			}

			services = append(services, serviceRef{product: product, service: svcEntry.Name()})
		}
	}

	return services, nil
}

func walkServiceImports(projectRoot, svcDir string, self serviceRef, allServices []serviceRef, violations *[]string) error {
	err := crossServiceWalkFn(svcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		imports, parseErr := extractImports(path)
		if parseErr != nil {
			return parseErr
		}

		for _, imp := range imports {
			if isViolation(imp, self, allServices) {
				rel, _ := filepath.Rel(projectRoot, path)
				*violations = append(*violations, fmt.Sprintf(
					"%s: imports %s (cross-service isolation violation)", rel, imp))
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking service dir %s: %w", svcDir, err)
	}

	return nil
}

// extractImports returns all import paths from a Go source file.
func extractImports(filePath string) ([]string, error) {
	f, err := os.Open(filePath) //nolint:gosec // filePath from filepath.Walk, controlled
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	var imports []string

	inImport := false

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "import (" {
			inImport = true

			continue
		}

		if inImport && trimmed == ")" {
			inImport = false

			continue
		}

		if inImport || strings.HasPrefix(trimmed, `import "`) {
			if m := importLinePattern.FindStringSubmatch(line); len(m) > 1 {
				imports = append(imports, m[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return imports, nil
}

// isViolation returns true if importPath crosses a PRODUCT boundary.
// Same-product cross-service imports are permitted: services within a product
// may share packages (e.g., identity/idp importing identity/authz/clientauth).
// Cross-product imports are forbidden: e.g., jose/ja importing pki/ca internals.
func isViolation(importPath string, self serviceRef, allServices []serviceRef) bool {
	const appsPrefix = "cryptoutil/internal/apps/"

	if !strings.HasPrefix(importPath, appsPrefix) {
		return false
	}

	rest := strings.TrimPrefix(importPath, appsPrefix)
	parts := strings.SplitN(rest, "/", importPrefixSplitN)

	if len(parts) < importPrefixSplitN-1 {
		return false
	}

	importProduct := parts[0]
	importService := parts[1]

	// Allow framework imports (all services may import the service framework) and cicd (tooling).
	if importProduct == cryptoutilSharedMagic.FrameworkProductName || importProduct == cicdProductName {
		return false
	}

	// Allow self-imports and same-product imports.
	if importProduct == self.product {
		return false
	}

	// Cross-product: check if the import points to another product's service.
	for _, svc := range allServices {
		if svc.product == importProduct && svc.service == importService {
			return true
		}
	}

	return false
}
