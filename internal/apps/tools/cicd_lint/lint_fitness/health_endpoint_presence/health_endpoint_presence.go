// Copyright (c) 2025 Justin Cranford

// Package health_endpoint_presence verifies that all services register
// health check endpoints. Services must reference livez, readyz, and health
// paths to satisfy the ARCHITECTURE.md Section 5.5 health check pattern.
package health_endpoint_presence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// healthRequirements are string patterns that must appear somewhere in a service.
var healthRequirements = []string{"livez", "readyz"}

// Check verifies health endpoint presence from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.ReadDir, os.ReadFile)
}

// CheckInDir verifies health endpoint presence under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking health endpoint presence in services...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	appsDir := filepath.Join(projectRoot, "internal", "apps")

	services, err := discoverServices(appsDir, readDirFn)
	if err != nil {
		return fmt.Errorf("failed to discover services: %w", err)
	}

	var violations []string

	for _, svc := range services {
		svcViolations := checkServiceHealth(svc, appsDir, readFileFn)
		violations = append(violations, svcViolations...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d health endpoint presence violations", len(violations))
	}

	logger.Log("Health endpoint presence check passed")

	return nil
}

// serviceID holds product/service name pair.
type serviceID struct {
	product string
	service string
}

// discoverServices returns all product/service pairs under appsDir.
// Excludes: cicd, skeleton, template products; archived (_-prefixed) dirs;
// non-service dirs (those without a server/ subdirectory).
func discoverServices(appsDir string, readDirFn func(string) ([]os.DirEntry, error)) ([]serviceID, error) {
	var services []serviceID

	products, err := readDirFn(appsDir)
	if err != nil {
		return nil, fmt.Errorf("read apps dir: %w", err)
	}

	for _, p := range products {
		if !p.IsDir() {
			continue
		}

		product := p.Name()
		if product == "cicd" || product == cryptoutilSharedMagic.SkeletonProductName || product == cryptoutilSharedMagic.SkeletonTemplateServiceName {
			continue
		}

		productDir := filepath.Join(appsDir, product)

		svcEntries, err := os.ReadDir(productDir)
		if err != nil {
			return nil, fmt.Errorf("read product dir %s: %w", product, err)
		}

		for _, s := range svcEntries {
			if !s.IsDir() {
				continue
			}

			name := s.Name()
			if strings.HasPrefix(name, "_") {
				continue // Skip archived directories.
			}

			// Only include actual services (must have server/ subdirectory).
			serverGoFile := filepath.Join(productDir, name, "server", "server.go")
			if _, statErr := os.Stat(serverGoFile); statErr == nil {
				services = append(services, serviceID{product: product, service: name})
			}
		}
	}

	return services, nil
}

// checkServiceHealth checks that a service references all required health patterns.
func checkServiceHealth(svc serviceID, appsDir string, readFileFn func(string) ([]byte, error)) []string {
	svcDir := filepath.Join(appsDir, svc.product, svc.service)

	// Collect all Go file content in the service directory.
	foundPatterns := make(map[string]bool)

	_ = filepath.Walk(svcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, readErr := readFileFn(path) //nolint:gosec // path from filepath.Walk, controlled
		if readErr != nil {
			return fmt.Errorf("reading file %s: %w", path, readErr)
		}

		for _, pattern := range healthRequirements {
			if strings.Contains(string(content), pattern) {
				foundPatterns[pattern] = true
			}
		}

		return nil
	})

	var violations []string

	for _, pattern := range healthRequirements {
		if !foundPatterns[pattern] {
			violations = append(violations, fmt.Sprintf(
				"internal/apps/%s/%s: missing health pattern %q", svc.product, svc.service, pattern))
		}
	}

	return violations
}
