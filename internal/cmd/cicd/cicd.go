// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilCmdCicdFormatGo "cryptoutil/internal/cmd/cicd/format_go"
	cryptoutilCmdCicdFormatGotest "cryptoutil/internal/cmd/cicd/format_gotest"
	cryptoutilCmdCicdLintGo "cryptoutil/internal/cmd/cicd/lint_go"
	cryptoutilCmdCicdLintGoMod "cryptoutil/internal/cmd/cicd/lint_go_mod"
	cryptoutilCmdCicdLintGotest "cryptoutil/internal/cmd/cicd/lint_gotest"
	cryptoutilCmdCicdLintText "cryptoutil/internal/cmd/cicd/lint_text"
	cryptoutilCmdCicdLintWorkflow "cryptoutil/internal/cmd/cicd/lint_workflow"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

const (
	cmdLintWorkflow = "lint-workflow"  // [Linter] Workflow file linters (GitHub Actions).
	cmdLintText     = "lint-text"      // [Linter] Text file linters (UTF-8 encoding).
	cmdLintGo       = "lint-go"        // [Linter] Go package linters (circular dependencies).
	cmdFormatGo     = "format-go"      // [Formatter] Go file formatters (any, copyloopvar).
	cmdLintGoTest   = "lint-go-test"   // [Linter] Go test file linters (test patterns).
	cmdFormatGoTest = "format-go-test" // [Formatter] Go test file formatters (t.Helper).
	cmdLintGoMod    = "lint-go-mod"    // [Linter] Go module linters (dependency updates).
)

var (
	exclusions = []string{
		"api/client",
		"api/model",
		"api/server",
		"api/idp",
		"api/authz",
		"test-output",
		"test-reports",
		"workflow-reports",
		"vendor",
	}
)


// Run executes the specified CI/CD check commands.
// Commands are executed sequentially, collecting results for each.
// Returns an error if any command fails, but continues executing all commands.
func Run(commands []string) error {
	logger := cryptoutilCmdCicdCommon.NewLogger("Run")
	startTime := time.Now()

	err := validateCommands(commands)
	if err != nil {
		return err
	}

	logger.Log("validateCommands completed")

	var allFiles []string

	doListAllFiles := false

	for _, cmd := range commands {
		if cmd == cmdLintText || cmd == cmdLintGoTest || cmd == cmdFormatGo || cmd == cmdLintWorkflow || cmd == cmdFormatGoTest {
			doListAllFiles = true

			break
		}
	}

	if doListAllFiles {
		listFilesStart := time.Now()

		allFiles, err = cryptoutilFiles.ListAllFiles(".", exclusions...)
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}

		logger.Log(fmt.Sprintf("collectAllFiles completed in %.2fs", time.Since(listFilesStart).Seconds()))
	}

	// Extract actual commands (skip flags starting with - and their values)
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

	logger.Log(fmt.Sprintf("Executing %d commands", len(actualCommands)))

	// Execute all commands and collect results
	results := make([]cryptoutilCmdCicdCommon.CommandResult, 0, len(actualCommands))

	for i, command := range actualCommands {
		cmdStart := time.Now()

		logger.Log(fmt.Sprintf("Executing command %d/%d: %s", i+1, len(actualCommands), command))

		var cmdErr error

		switch command {
		case cmdLintText:
			cmdErr = cryptoutilCmdCicdLintText.Lint(logger, allFiles)
		case cmdLintGo:
			cmdErr = cryptoutilCmdCicdLintGo.Lint(logger)
		case cmdFormatGo:
			cmdErr = cryptoutilCmdCicdFormatGo.Format(logger, allFiles)
		case cmdLintGoTest:
			cmdErr = cryptoutilCmdCicdLintGotest.Lint(logger, allFiles)
		case cmdFormatGoTest:
			cmdErr = cryptoutilCmdCicdFormatGotest.Format(logger)
		case cmdLintWorkflow:
			cmdErr = cryptoutilCmdCicdLintWorkflow.Lint(logger, allFiles)
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

		// Add a separator between multiple commands
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

func validateCommands(commands []string) error {
	logger := cryptoutilCmdCicdCommon.NewLogger("validateCommands")

	if len(commands) == 0 {
		logger.Log("validateCommands: empty commands")

		return fmt.Errorf("%s", cryptoutilMagic.UsageCICD)
	}

	var errs []error

	commandCounts := make(map[string]int)

	skipNext := false

	for _, command := range commands {
		// Skip flag values after flag names (e.g., skip "P5.01" after "--start-task")
		if skipNext {
			skipNext = false

			continue
		}

		// Skip flag arguments (--strict, --task-threshold, --start-task, --end-task, --output, etc.)
		// Flags are passed to subcommand Enforce functions, not validated as commands
		if strings.HasPrefix(command, "-") {
			skipNext = true // Next argument is flag value, skip it

			continue
		}

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

		return fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	logger.Log("validateCommands: success")

	return nil
}
