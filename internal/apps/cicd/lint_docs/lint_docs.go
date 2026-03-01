// Copyright (c) 2025 Justin Cranford

// Package lint_docs provides documentation linting: chunk verification, chunk
// content validation, and propagation reference validation.
package lint_docs

import (
	"errors"
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilCheckChunkVerification "cryptoutil/internal/apps/cicd/lint_docs/check_chunk_verification"
	cryptoutilValidateChunks "cryptoutil/internal/apps/cicd/lint_docs/validate_chunks"
	cryptoutilValidatePropagation "cryptoutil/internal/apps/cicd/lint_docs/validate_propagation"
)

// LinterFunc is the signature for lint_docs sub-linters.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds the ordered list of documentation linters.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"check-chunk-verification", cryptoutilCheckChunkVerification.Check},
	{"validate-chunks", cryptoutilValidateChunks.Check},
	{"validate-propagation", cryptoutilValidatePropagation.Check},
}

// Lint runs all registered documentation linters sequentially.
// Continues on failure, collecting all errors before returning.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	var errs []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running %s", l.name))

		if err := l.linter(logger); err != nil {
			logger.LogError(err)
			errs = append(errs, fmt.Errorf("%s: %w", l.name, err))
		} else {
			logger.Log(fmt.Sprintf("  \u2705 %s passed", l.name))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("lint-docs failed: %w", errors.Join(errs...))
	}

	return nil
}
