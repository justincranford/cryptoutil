// Copyright (c) 2025 Justin Cranford

// Package lint_ports validates port assignments across cryptoutil codebase.
// Ensures legacy ports are not used and ports match the standardized scheme.
package lint_ports

import (
	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintPortsHealthPaths "cryptoutil/internal/apps/cicd/lint_ports/health_paths"
	lintPortsHostPortRanges "cryptoutil/internal/apps/cicd/lint_ports/host_port_ranges"
	lintPortsLegacyPorts "cryptoutil/internal/apps/cicd/lint_ports/legacy_ports"
)

// LinterFunc is a function type for individual port linters.
// Each linter receives a logger and the files map, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredLinters holds all linters to run as part of lint-ports.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"legacy-ports", lintPortsLegacyPorts.Check},
	{"host-port-ranges", lintPortsHostPortRanges.Check},
	{"health-paths", lintPortsHealthPaths.Check},
}

// Lint checks all relevant files for legacy port usage violations.
// Returns an error if any legacy ports are found.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running port validation lint...")

	var allErrors []error

	for _, l := range registeredLinters {
		if err := l.linter(logger, filesByExtension); err != nil {
			allErrors = append(allErrors, err)
		}
	}

	if len(allErrors) > 0 {
		// Return the first error to preserve specific error messages for backwards compatibility.
		return allErrors[0]
	}

	logger.Log("✅ lint-ports passed: all validations successful")

	return nil
}
