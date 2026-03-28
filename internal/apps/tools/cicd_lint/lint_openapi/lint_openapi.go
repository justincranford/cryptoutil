// Copyright (c) 2025 Justin Cranford

// Package lint_openapi provides OpenAPI linting for CI/CD pipelines.
// Sub-linters validate OpenAPI spec versions and oapi-codegen configuration.
package lint_openapi

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintOpenAPICodegenConfig "cryptoutil/internal/apps/tools/cicd_lint/lint_openapi/codegen_config"
	lintOpenAPIVersion "cryptoutil/internal/apps/tools/cicd_lint/lint_openapi/openapi_version"
)

// LinterFunc is a function type for individual OpenAPI linters.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredLinters holds all linters to run as part of lint-openapi.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"openapi-version", lintOpenAPIVersion.Check},
	{"codegen-config", lintOpenAPICodegenConfig.Check},
}

// Lint runs all registered OpenAPI linters on the provided files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running OpenAPI linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, filesByExtension); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-openapi completed with %d errors", len(errors)))

		return fmt.Errorf("lint-openapi failed with %d errors", len(errors))
	}

	logger.Log("lint-openapi completed successfully")

	return nil
}
