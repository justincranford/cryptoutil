// Copyright (c) 2025 Justin Cranford
//
//

// Package cicd provides CI/CD workflow utilities and automation.
package cicd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilCmdCicdFormatGo "cryptoutil/internal/cmd/cicd/format_go"
	cryptoutilCmdCicdFormatGotest "cryptoutil/internal/cmd/cicd/format_gotest"
	cryptoutilCmdCicdLintCompose "cryptoutil/internal/cmd/cicd/lint_compose"
	cryptoutilCmdCicdLintGo "cryptoutil/internal/cmd/cicd/lint_go"
	cryptoutilCmdCicdLintGoMod "cryptoutil/internal/cmd/cicd/lint_go_mod"
	cryptoutilCmdCicdLintGotest "cryptoutil/internal/cmd/cicd/lint_gotest"
	cryptoutilCmdCicdLintText "cryptoutil/internal/cmd/cicd/lint_text"
	cryptoutilCmdCicdLintWorkflow "cryptoutil/internal/cmd/cicd/lint_workflow"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

const (
	cmdLintText     = "lint-text"      // [Linter] Text file linters (UTF-8 encoding).
	cmdLintWorkflow = "lint-workflow"  // [Linter] Workflow file linters (GitHub Actions).
	cmdLintGo       = "lint-go"        // [Linter] Go package linters (circular dependencies, CGO-free SQLite).
	cmdLintCompose  = "lint-compose"   // [Linter] Docker Compose file linters (admin port exposure).
	cmdFormatGo     = "format-go"      // [Formatter] Go file formatters (any, copyloopvar).
	cmdLintGoTest   = "lint-go-test"   // [Linter] Go test file linters (test patterns).
	cmdFormatGoTest = "format-go-test" // [Formatter] Go test file formatters (t.Helper).
	cmdLintGoMod    = "lint-go-mod"    // [Linter] Go module linters (dependency updates).
)

// Run executes the specified CI/CD check commands.
// Commands are executed sequentially, collecting results for each.
// Returns an error if any command fails, but continues executing all commands.
func Run(commands []string) error {
	logger := cryptoutilCmdCicdCommon.NewLogger("Run")
	startTime := time.Now()

	actualCommands, err := validateCommands(commands)
	if err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}

	logger.Log("validateCommands completed")

	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(cryptoutilMagic.ListAllFilesStartDirectory)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	logger.Log(fmt.Sprintf("Executing %d commands", len(actualCommands)))

	// Execute all commands and collect results.
	results := make([]cryptoutilCmdCicdCommon.CommandResult, 0, len(actualCommands))

	for i, command := range actualCommands {
		cmdStart := time.Now()

		logger.Log(fmt.Sprintf("Executing command %d/%d: %s", i+1, len(actualCommands), command))

		var cmdErr error

		switch command {
		case cmdLintText:
			cmdErr = cryptoutilCmdCicdLintText.Lint(logger, filesByExtension)
		case cmdLintGo:
			cmdErr = cryptoutilCmdCicdLintGo.Lint(logger)
		case cmdLintCompose:
			cmdErr = cryptoutilCmdCicdLintCompose.Lint(logger, filesByExtension)
		case cmdFormatGo:
			cmdErr = cryptoutilCmdCicdFormatGo.Format(logger, filesByExtension)
		case cmdLintGoTest:
			cmdErr = cryptoutilCmdCicdLintGotest.Lint(logger, filesByExtension)
		case cmdFormatGoTest:
			cmdErr = cryptoutilCmdCicdFormatGotest.Format(logger)
		case cmdLintWorkflow:
			cmdErr = cryptoutilCmdCicdLintWorkflow.Lint(logger, filesByExtension)
		case cmdLintGoMod:
			cmdErr = cryptoutilCmdCicdLintGoMod.Lint(logger)
		}

		cmdDuration := time.Since(cmdStart)
		results = append(results, cryptoutilCmdCicdCommon.CommandResult{
			Command:  command,
			Duration: cmdDuration,
			Error:    cmdErr,
		})

		logger.Log(fmt.Sprintf("Command '%s' completed in %.2fs", command, cmdDuration.Seconds()))

		// Add a separator between multiple commands.
		if i < len(actualCommands)-1 {
			cryptoutilCmdCicdCommon.PrintCommandSeparator()
		}
	}

	// Print summary
	totalDuration := time.Since(startTime)
	cryptoutilCmdCicdCommon.PrintExecutionSummary(results, totalDuration)

	// Collect all errors
	failedCommands := cryptoutilCmdCicdCommon.GetFailedCommands(results)

	if len(failedCommands) > 0 {
		return fmt.Errorf("failed commands: %s", strings.Join(failedCommands, ", "))
	}

	logger.Log("Run completed successfully")

	return nil
}

func validateCommands(commands []string) ([]string, error) {
	logger := cryptoutilCmdCicdCommon.NewLogger("validateCommands")

	if len(commands) == 0 {
		logger.Log("validateCommands: empty commands")

		return nil, fmt.Errorf("%s", cryptoutilMagic.UsageCICD)
	}

	// Extract actual commands (skip flags starting with - and their values).
	actualCommands := []string{}
	skipNext := false

	for _, cmd := range commands {
		if skipNext {
			skipNext = false

			continue
		}

		if strings.HasPrefix(cmd, "-") {
			skipNext = true // Next arg is flag value, skip it

			continue
		}

		actualCommands = append(actualCommands, cmd)
	}

	if len(actualCommands) == 0 {
		logger.Log("validateCommands: no actual commands after flag extraction")

		return nil, fmt.Errorf("%s", cryptoutilMagic.UsageCICD)
	}

	var errs []error

	commandCounts := make(map[string]int)

	for _, command := range actualCommands {
		if cryptoutilMagic.ValidCommands[command] {
			commandCounts[command]++
		} else {
			errs = append(errs, fmt.Errorf("unknown command: %s\n\n%s", command, cryptoutilMagic.UsageCICD))
		}
	}

	// Check for duplicate commands.
	for command, count := range commandCounts {
		if count > 1 {
			errs = append(errs, fmt.Errorf("command '%s' specified %d times - each command can only be used once", command, count))
		}
	}

	if len(errs) > 0 {
		logger.Log("validateCommands: validation errors")

		return nil, fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	logger.Log("validateCommands: success")

	return actualCommands, nil
}
