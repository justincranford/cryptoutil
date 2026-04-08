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
	// importPrefixSplitN splits import path into PS-ID and rest.
	importPrefixSplitN = 2

	// toolsDirName is the tools directory name; excluded from isolation checks.
	toolsDirName = "tools"
)

var importLinePattern = regexp.MustCompile(`^\s+(?:\w+ )?"([^"]+)"`)

// Check verifies cross-service import isolation from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir verifies cross-service import isolation under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDir(logger, rootDir, filepath.Walk)
}

// checkInDir is the internal implementation that accepts a walkFn for testing.
func checkInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, walkFn func(string, filepath.WalkFunc) error) error {
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
		svcDir := filepath.Join(appsDir, svc.psid)

		if walkErr := walkServiceImports(projectRoot, svcDir, svc, services, &violations, walkFn); walkErr != nil {
			return fmt.Errorf("failed to scan service %s: %w", svc.psid, walkErr)
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

// serviceRef identifies a PS-ID and its owning product.
type serviceRef struct {
	psid    string
	product string
}

// knownProducts lists all product names used to extract product from PS-ID.
// PS-IDs follow the pattern "<product>-<service>" (e.g., "sm-im", "pki-ca").
var knownProducts = []string{
	cryptoutilSharedMagic.IdentityProductName,
	cryptoutilSharedMagic.JoseProductName,
	cryptoutilSharedMagic.PKIProductName,
	cryptoutilSharedMagic.SkeletonProductName,
	cryptoutilSharedMagic.SMProductName,
}

// productFromPSID extracts the product name from a PS-ID.
// Returns the PS-ID itself if no known product prefix matches (e.g., product dirs).
func productFromPSID(psid string) string {
	for _, product := range knownProducts {
		if strings.HasPrefix(psid, product+"-") {
			return product
		}
	}

	return psid
}

// collectServices discovers all flat PS-ID service directories under appsDir.
// Only directories with a server/ subdirectory qualify as real services.
// Directories starting with _ (archived) are excluded.
// The tools, skeleton, framework, and cryptoutil directories are excluded.
func collectServices(appsDir string) ([]serviceRef, error) {
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, fmt.Errorf("read apps dir: %w", err)
	}

	services := make([]serviceRef, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()

		// Skip archived directories, skeleton product/services (templates), and tools.
		// framework, cryptoutil, and product dirs are filtered by the server/ check below.
		switch {
		case strings.HasPrefix(dirName, "_"):
			continue
		case dirName == toolsDirName:
			continue
		case dirName == cryptoutilSharedMagic.SkeletonProductName || strings.HasPrefix(dirName, cryptoutilSharedMagic.SkeletonProductName+"-"):
			continue
		}

		// Only directories with server/ are real services.
		// Product dirs and shared packages lack server/.
		serverDir := filepath.Join(appsDir, dirName, "server")
		if _, statErr := os.Stat(serverDir); os.IsNotExist(statErr) {
			continue
		}

		product := productFromPSID(dirName)

		services = append(services, serviceRef{psid: dirName, product: product})
	}

	return services, nil
}

func walkServiceImports(projectRoot, svcDir string, self serviceRef, allServices []serviceRef, violations *[]string, walkFn func(string, filepath.WalkFunc) error) error {
	err := walkFn(svcDir, func(path string, info os.FileInfo, err error) error {
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
// may share packages (e.g., identity-idp importing identity-authz/clientauth).
// Cross-product imports are forbidden: e.g., jose-ja importing pki-ca internals.
func isViolation(importPath string, self serviceRef, allServices []serviceRef) bool {
	const appsPrefix = "cryptoutil/internal/apps/"

	if !strings.HasPrefix(importPath, appsPrefix) {
		return false
	}

	rest := strings.TrimPrefix(importPath, appsPrefix)
	parts := strings.SplitN(rest, "/", importPrefixSplitN)

	if len(parts) < importPrefixSplitN {
		return false
	}

	importPSID := parts[0]

	// Allow framework imports (all services may import the service framework) and tools (cicd tooling).
	if importPSID == cryptoutilSharedMagic.FrameworkProductName || importPSID == toolsDirName {
		return false
	}

	// Allow self-imports.
	if importPSID == self.psid {
		return false
	}

	// Determine imported PS-ID's product.
	importProduct := productFromPSID(importPSID)

	// Allow same-product imports (e.g., identity-idp importing identity-authz).
	if importProduct == self.product {
		return false
	}

	// Cross-product: check if the import points to another product's service.
	for _, svc := range allServices {
		if svc.psid == importPSID {
			return true
		}
	}

	return false
}
