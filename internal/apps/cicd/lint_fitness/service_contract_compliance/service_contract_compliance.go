// Copyright (c) 2025 Justin Cranford

// Package service_contract_compliance verifies that all services have a
// compile-time interface assertion: `var _ ServiceServer = (*XxxServer)(nil)`.
// This assertion ensures services satisfy the ServiceServer contract defined in
// internal/apps/template/service/server/contract.go.
package service_contract_compliance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// compileTimeAssertionPattern is the pattern that must appear in a service's server.go.
const compileTimeAssertionPattern = "ServiceServer = (*"

// Check verifies service contract compliance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir verifies service contract compliance under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking service contract compliance (compile-time assertions)...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	appsDir := filepath.Join(projectRoot, "internal", "apps")

	services, err := discoverServices(appsDir)
	if err != nil {
		return fmt.Errorf("failed to discover services: %w", err)
	}

	var violations []string

	for _, svc := range services {
		serverFile := filepath.Join(appsDir, svc.product, svc.service, "server", "server.go")

		if checkErr := checkServerFile(serverFile, svc, &violations); checkErr != nil {
			return fmt.Errorf("failed to check %s/%s: %w", svc.product, svc.service, checkErr)
		}
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d service contract compliance violations", len(violations))
	}

	logger.Log("Service contract compliance check passed")

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
func discoverServices(appsDir string) ([]serviceID, error) {
	var services []serviceID

	products, err := os.ReadDir(appsDir)
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

// checkServerFile verifies that the server.go file contains the compile-time assertion.
func checkServerFile(serverFile string, svc serviceID, violations *[]string) error {
	content, err := os.ReadFile(serverFile) //nolint:gosec // serverFile is a constructed path, controlled
	if err != nil {
		if os.IsNotExist(err) {
			*violations = append(*violations, fmt.Sprintf(
				"internal/apps/%s/%s/server/server.go: missing (cannot check compile-time assertion)",
				svc.product, svc.service))

			return nil
		}

		return fmt.Errorf("read %s: %w", serverFile, err)
	}

	if !strings.Contains(string(content), compileTimeAssertionPattern) {
		*violations = append(*violations, fmt.Sprintf(
			"internal/apps/%s/%s/server/server.go: missing compile-time assertion %q",
			svc.product, svc.service, compileTimeAssertionPattern))
	}

	return nil
}
