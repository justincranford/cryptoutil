// Copyright (c) 2025 Justin Cranford

// Package lint_compose provides linting for Docker Compose files.
package lint_compose

import (
	"fmt"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintComposeAdminPortExposure "cryptoutil/internal/apps/cicd/lint_compose/admin_port_exposure"
	lintComposeDockerSecrets "cryptoutil/internal/apps/cicd/lint_compose/docker_secrets"
)

// LinterFunc is a function type for individual Docker Compose linters.
// Each linter receives a logger and the files map, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredLinters holds all linters to run as part of lint-compose.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"admin-port-exposure", lintComposeAdminPortExposure.Check},
	{"docker-secrets", lintComposeDockerSecrets.Check},
}

// Lint runs all registered Docker Compose linters.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Docker Compose linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, filesByExtension); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-compose completed with %d errors", len(errors)))

		msgs := make([]string, len(errors))

		for i, e := range errors {
			msgs[i] = e.Error()
		}

		return fmt.Errorf("lint-compose failed with %d errors: %s", len(errors), strings.Join(msgs, "; "))
	}

	logger.Log("lint-compose completed successfully")

	return nil
}
