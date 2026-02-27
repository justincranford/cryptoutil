// Copyright (c) 2025 Justin Cranford

package format_go

import (
	"fmt"
	"runtime"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	formatGoCopyLoopVar "cryptoutil/internal/apps/cicd/format_go/copyloopvar"
	formatGoEnforceAny "cryptoutil/internal/apps/cicd/format_go/enforce_any"
	formatGoEnforceTimeNowUTC "cryptoutil/internal/apps/cicd/format_go/enforce_time_now_utc"
)

// FormatterFunc is a function type for individual Go formatters.
// Each formatter receives a logger, root directory, and Go version, returning counts and error.
type FormatterFunc func(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, goVersion string) (processed, modified, fixed int, err error)

// FormatterFuncSimple is a function type for formatters that work on files by extension.
type FormatterFuncSimple func(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error

// registeredFormatters holds all formatters to run as part of format-go.
var registeredFormatters = []struct {
	name      string
	formatter FormatterFunc
}{
	{"copyloopvar", formatGoCopyLoopVar.Fix},
}

// registeredSimpleFormatters holds formatters that work on file lists.
var registeredSimpleFormatters = []struct {
	name      string
	formatter FormatterFuncSimple
}{
	{"enforce-any", formatGoEnforceAny.Enforce},
	{"enforce-time-now-utc", formatGoEnforceTimeNowUTC.Enforce},
}

// Format runs all registered Go formatters.
// Returns an error if any formatter finds issues.
func Format(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running Go formatters...")

	var errors []error

	goVersion := runtime.Version()
	rootDir := "."

	// Run formatters that use root directory.
	for _, f := range registeredFormatters {
		logger.Log(fmt.Sprintf("Running formatter: %s", f.name))

		processed, modified, fixed, err := f.formatter(logger, rootDir, goVersion)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", f.name, err))
		} else {
			logger.Log(fmt.Sprintf("%s: processed=%d, modified=%d, fixed=%d", f.name, processed, modified, fixed))
		}
	}

	// Run formatters that use file lists.
	for _, f := range registeredSimpleFormatters {
		logger.Log(fmt.Sprintf("Running formatter: %s", f.name))

		if err := f.formatter(logger, filesByExtension); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", f.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("format-go completed with %d errors", len(errors)))

		return fmt.Errorf("format-go completed with modifications - please commit the changes")
	}

	logger.Log("format-go completed successfully")

	return nil
}
