// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

const (
	// Command name for enforcing UTF-8 encoding on all files.
	cmdAllEnforceUTF8 = "all-enforce-utf8"

	// Command name for enforcing Go any usage instead of interface{}.
	cmdGoEnforceAny = "go-enforce-any"

	// Command name for enforcing test patterns in Go files.
	cmdGoEnforceTestPatterns = "go-enforce-test-patterns"

	// Command name for linting GitHub workflow files.
	cmdGitHubWorkflowLint = "github-workflow-lint"

	// Command name for auto-fixing staticcheck error strings.
	cmdGoFixStaticcheckErrorStrings = "go-fix-staticcheck-error-strings"

	// Command name for auto-fixing copyloopvar issues.
	cmdGoFixCopyLoopVar = "go-fix-copyloopvar"

	// Command name for auto-fixing thelper issues.
	cmdGoFixTHelper = "go-fix-thelper"

	// Command name for running all auto-fix commands.
	cmdGoFixAll = "go-fix-all"
)

// Run executes the specified CI/CD check commands.
// Commands are executed sequentially, collecting results for each.
// Returns an error if any command fails, but continues executing all commands.
func Run(commands []string) error {
	logger := common.NewLogger("Run")
	startTime := time.Now()

	err := validateCommands(commands)
	if err != nil {
		return err
	}

	logger.Log("validateCommands completed")

	var allFiles []string

	doListAllFiles := false

	for _, cmd := range commands {
		if cmd == cmdAllEnforceUTF8 || cmd == cmdGoEnforceTestPatterns || cmd == cmdGoEnforceAny || cmd == cmdGitHubWorkflowLint || cmd == cmdGoFixStaticcheckErrorStrings || cmd == cmdGoFixCopyLoopVar || cmd == cmdGoFixTHelper || cmd == cmdGoFixAll {
			doListAllFiles = true

			break
		}
	}

	if doListAllFiles {
		listFilesStart := time.Now()

		allFiles, err = cryptoutilFiles.ListAllFiles(".")
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}

		logger.Log(fmt.Sprintf("collectAllFiles completed in %.2fs", time.Since(listFilesStart).Seconds()))
	}

	logger.Log(fmt.Sprintf("Executing %d commands", len(commands)))

	// Execute all commands and collect results
	results := make([]common.CommandResult, 0, len(commands))

	for i, command := range commands {
		cmdStart := time.Now()

		logger.Log(fmt.Sprintf("Executing command %d/%d: %s", i+1, len(commands), command))

		var cmdErr error

		switch command {
		case cmdAllEnforceUTF8:
			cmdErr = allEnforceUtf8(logger, allFiles)
		case cmdGoEnforceTestPatterns:
			cmdErr = goEnforceTestPatterns(logger, allFiles)
		case cmdGoEnforceAny:
			cmdErr = goEnforceAny(logger, allFiles)
		case "go-check-circular-package-dependencies":
			cmdErr = goCheckCircularPackageDeps(logger)
		case "go-check-identity-imports":
			cmdErr = goCheckIdentityImports(logger)
		case "go-update-direct-dependencies":
			cmdErr = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
		case "go-update-all-dependencies":
			cmdErr = goUpdateDeps(logger, cryptoutilMagic.DepCheckAll)
		case cmdGitHubWorkflowLint:
			cmdErr = checkWorkflowLintWithError(logger, allFiles)
		case cmdGoFixStaticcheckErrorStrings:
			cmdErr = goFixStaticcheckErrorStrings(logger, allFiles)
		case cmdGoFixCopyLoopVar:
			cmdErr = goFixCopyLoopVar(logger, allFiles)
		case cmdGoFixTHelper:
			cmdErr = goFixTHelper(logger, allFiles)
		case cmdGoFixAll:
			cmdErr = goFixAll(logger, allFiles)
		}

		cmdDuration := time.Since(cmdStart)
		results = append(results, common.CommandResult{
			Command:  command,
			Duration: cmdDuration,
			Error:    cmdErr,
		})

		logger.Log(fmt.Sprintf("Command '%s' completed in %.2fs", command, cmdDuration.Seconds()))

		// Add a separator between multiple commands
		if i < len(commands)-1 {
			common.PrintCommandSeparator()
		}
	}

	// Print summary
	totalDuration := time.Since(startTime)
	common.PrintExecutionSummary(results, totalDuration)

	// Collect all errors
	failedCommands := common.GetFailedCommands(results)

	if len(failedCommands) > 0 {
		return fmt.Errorf("failed commands: %s", strings.Join(failedCommands, ", "))
	}

	logger.Log("Run completed successfully")

	return nil
}

func validateCommands(commands []string) error {
	logger := common.NewLogger("validateCommands")

	if len(commands) == 0 {
		logger.Log("validateCommands: empty commands")

		return fmt.Errorf("%s", cryptoutilMagic.UsageCICD)
	}

	var errs []error

	commandCounts := make(map[string]int)

	for _, command := range commands {
		if cryptoutilMagic.ValidCommands[command] {
			commandCounts[command]++
		} else {
			errs = append(errs, fmt.Errorf("unknown command: %s\n\n%s", command, cryptoutilMagic.UsageCICD))
		}
	}

	// Check for duplicate commands
	for command, count := range commandCounts {
		if count > 1 {
			errs = append(errs, fmt.Errorf("command '%s' specified %d times - each command can only be used once", command, count))
		}
	}

	// Check for mutually exclusive commands
	if commandCounts["go-update-direct-dependencies"] > 0 && commandCounts["go-update-all-dependencies"] > 0 {
		errs = append(errs, fmt.Errorf("commands 'go-update-direct-dependencies' and 'go-update-all-dependencies' cannot be used together - choose one dependency update mode"))
	}

	if len(errs) > 0 {
		logger.Log("validateCommands: validation errors")

		return fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	logger.Log("validateCommands: success")

	return nil
}
