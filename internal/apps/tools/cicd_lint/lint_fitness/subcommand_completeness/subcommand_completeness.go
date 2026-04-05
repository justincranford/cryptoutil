// Copyright (c) 2025 Justin Cranford

// Package subcommand_completeness validates that all registry-declared product-services
// use the standard RouteService framework, which guarantees the mandatory subcommands
// (server, client, init, health, livez, readyz, shutdown) are all present.
//
// The check scans internal/apps/{ps-id}/*.go for calls to RouteService.
// Any service that wires up RouteService inherits all mandatory subcommands from
// the framework CLI package (internal/apps/framework/service/cli).
package subcommand_completeness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilLintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

// routeServiceCall is the canonical marker that a service uses the standard CLI framework.
// All mandatory subcommands (server, client, init, health, livez, readyz, shutdown) are
// provided automatically when a service calls RouteService.
const routeServiceCall = "RouteService"

// Check validates subcommand completeness from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.ReadDir, os.ReadFile)
}

// CheckInDir validates that each registry PS-ID service uses RouteService under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	logger.Log("Checking subcommand completeness for all registry services...")

	var violations []string

	for _, ps := range cryptoutilLintFitnessRegistry.AllProductServices() {
		serviceDir := filepath.Join(rootDir, "internal", "apps", ps.PSID)

		if err := checkServiceUsesRouteService(serviceDir, ps.PSID, readDirFn, readFileFn); err != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", ps.PSID, err))
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("subcommand-completeness violations:\n%s", strings.Join(violations, "\n"))
	}

	return nil
}

// checkServiceUsesRouteService verifies that at least one Go file in serviceDir
// contains a call to RouteService, indicating the service uses the standard CLI framework.
func checkServiceUsesRouteService(serviceDir, psID string, readDirFn func(string) ([]os.DirEntry, error), readFileFn func(string) ([]byte, error)) error {
	entries, err := readDirFn(serviceDir)
	if err != nil {
		return fmt.Errorf("cannot read service directory %s: %w", serviceDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(serviceDir, name)

		content, readErr := readFileFn(filePath)
		if readErr != nil {
			return fmt.Errorf("cannot read %s: %w", filePath, readErr)
		}

		if strings.Contains(string(content), routeServiceCall) {
			return nil
		}
	}

	return fmt.Errorf("no Go file in internal/apps/%s/ calls RouteService; service lacks mandatory subcommands (server, client, init, health, livez, readyz, shutdown)", psID)
}
