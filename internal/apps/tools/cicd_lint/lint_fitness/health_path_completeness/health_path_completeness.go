// Copyright (c) 2025 Justin Cranford

// Package health_path_completeness verifies that all services document all
// required health endpoint paths. Each service must reference all five standard
// health paths: /service/api/v1/health, /browser/api/v1/health,
// /admin/api/v1/livez, /admin/api/v1/readyz, and /admin/api/v1/shutdown.
// See ARCHITECTURE.md Section 5.5 for the health check pattern.
package health_path_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilLintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// requiredHealthPaths are the full path strings that must appear in each service's Go files.
// These correspond to the five standard health endpoints registered by the framework.
var requiredHealthPaths = []string{
	cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath + "/health",                                            // /service/api/v1/health
	cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath + "/health",                                            // /browser/api/v1/health
	cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,    // /admin/api/v1/livez
	cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,   // /admin/api/v1/readyz
	cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, // /admin/api/v1/shutdown
}

// Check verifies health path completeness from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.ReadDir, os.ReadFile)
}

// CheckInDir verifies health path completeness under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking health path completeness in services...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	appsDir := filepath.Join(projectRoot, "internal", "apps")

	var violations []string

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		if ps.PSID == cryptoutilSharedMagic.SkeletonTemplateServiceID {
			continue
		}

		svcDir := filepath.Join(appsDir, ps.InternalAppsDir)

		svcViolations, err := checkServiceHealthPaths(ps.PSID, svcDir, readDirFn, readFileFn)
		if err != nil {
			return fmt.Errorf("checking service %s: %w", ps.PSID, err)
		}

		violations = append(violations, svcViolations...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d health path completeness violations", len(violations))
	}

	logger.Log("Health path completeness check passed")

	return nil
}

// checkServiceHealthPaths checks that svcDir references all required health paths.
// Returns violations, one per missing path.
func checkServiceHealthPaths(psID, svcDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) ([]string, error) {
	foundPaths := make(map[string]bool)

	entries, err := readDirFn(svcDir)
	if err != nil {
		return nil, fmt.Errorf("read service dir %s: %w", psID, err)
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(svcDir, entry.Name())

		content, readErr := readFileFn(filePath) //nolint:gosec // path from controlled ReadDir
		if readErr != nil {
			return nil, fmt.Errorf("read file %s: %w", filePath, readErr)
		}

		for _, path := range requiredHealthPaths {
			if strings.Contains(string(content), path) {
				foundPaths[path] = true
			}
		}
	}

	var violations []string

	for _, path := range requiredHealthPaths {
		if !foundPaths[path] {
			violations = append(violations, fmt.Sprintf(
				"internal/apps/%s: missing health path %q in top-level Go files", psID, path))
		}
	}

	return violations, nil
}
