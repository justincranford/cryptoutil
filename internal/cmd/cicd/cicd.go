// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	cryptoutilCmdCicdAllEnforceUtf8 "cryptoutil/internal/cmd/cicd/all_enforce_utf8"
	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilCmdCicdGithubWorkflowLint "cryptoutil/internal/cmd/cicd/github_workflow_lint"
	cryptoutilCmdCicdGoCheckCircularPackageDependencies "cryptoutil/internal/cmd/cicd/go_check_circular_package_dependencies"
	cryptoutilCmdCicdGoEnforceAny "cryptoutil/internal/cmd/cicd/go_enforce_any"
	cryptoutilCmdCicdGoEnforceTestPatterns "cryptoutil/internal/cmd/cicd/go_enforce_test_patterns"
	cryptoutilCmdCicdGoFixAll "cryptoutil/internal/cmd/cicd/go_fix_all"
	cryptoutilCmdCicdGoFixCopyLoopVar "cryptoutil/internal/cmd/cicd/go_fix_copyloopvar"
	cryptoutilCmdCicdGoFixTHelper "cryptoutil/internal/cmd/cicd/go_fix_thelper"
	cryptoutilCmdCicdGoUpdateDirectDependencies "cryptoutil/internal/cmd/cicd/go_update_direct_dependencies"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

const (
	cmdAllEnforceUTF8                     = "all-enforce-utf8"                       // [Linter] Enforces UTF-8 encoding without BOM - works on all text files
	cmdGoCheckCircularPackageDependencies = "go-check-circular-package-dependencies" // [Linter] Checks for circular Go package dependencies - works on *.go files
	cmdGoEnforceAny                       = "go-enforce-any"                         // [Formatter] Replaces interface{} with any - works on *.go files
	cmdGoFixCopyLoopVar                   = "go-fix-copyloopvar"                     // [Formatter] Fixes Go 1.22+ loop variable scoping - works on *.go files
	cmdGoFixAll                           = "go-fix-all"                             // [Formatter] Runs all go-fix-* commands in sequence - works on *.go files
	cmdGoEnforceTestPatterns              = "go-enforce-test-patterns"               // [Linter] Enforces test patterns (UUIDv7 usage, testify assertions) - works on *_test.go files
	cmdGoFixTHelper                       = "go-fix-thelper"                         // [Formatter] Adds t.Helper() to test helper functions - works on *_test.go files
	cmdGitHubWorkflowLint                 = "github-workflow-lint"                   // [Linter] Validates GitHub Actions workflow naming and structure - works on *.yml, *.yaml files
	cmdGoUpdateDirectDependencies         = "go-update-direct-dependencies"          // [Linter] Checks direct Go dependencies for updates - works on go.mod, go.sum
	cmdGoUpdateAllDependencies            = "go-update-all-dependencies"             // [Linter] Checks all Go dependencies for updates - works on go.mod, go.sum
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

	exclusions := []string{
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

	doListAllFiles := false

	for _, cmd := range commands {
		if cmd == cmdAllEnforceUTF8 || cmd == cmdGoEnforceTestPatterns || cmd == cmdGoEnforceAny || cmd == cmdGitHubWorkflowLint || cmd == cmdGoFixCopyLoopVar || cmd == cmdGoFixTHelper || cmd == cmdGoFixAll {
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
		case cmdAllEnforceUTF8:
			cmdErr = cryptoutilCmdCicdAllEnforceUtf8.Enforce(logger, allFiles)
		case cmdGoEnforceTestPatterns:
			cmdErr = cryptoutilCmdCicdGoEnforceTestPatterns.Enforce(logger, allFiles)
		case cmdGoEnforceAny:
			cmdErr = cryptoutilCmdCicdGoEnforceAny.Enforce(logger, allFiles)
		case cmdGoCheckCircularPackageDependencies:
			cmdErr = cryptoutilCmdCicdGoCheckCircularPackageDependencies.Check(logger)
		case cmdGoUpdateDirectDependencies:
			cmdErr = cryptoutilCmdCicdGoUpdateDirectDependencies.Update(logger, cryptoutilMagic.DepCheckDirect)
		case cmdGoUpdateAllDependencies:
			cmdErr = cryptoutilCmdCicdGoUpdateDirectDependencies.Update(logger, cryptoutilMagic.DepCheckAll)
		case cmdGitHubWorkflowLint:
			cmdErr = cryptoutilCmdCicdGithubWorkflowLint.Lint(logger, allFiles)
		case cmdGoFixCopyLoopVar:
			_, _, _, cmdErr = cryptoutilCmdCicdGoFixCopyLoopVar.Fix(logger, ".", runtime.Version())
		case cmdGoFixTHelper:
			_, _, _, cmdErr = cryptoutilCmdCicdGoFixTHelper.Fix(logger, ".")
		case cmdGoFixAll:
			_, _, _, cmdErr = cryptoutilCmdCicdGoFixAll.Fix(logger, ".", runtime.Version())
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

	// Check for duplicate commands
	for command, count := range commandCounts {
		if count > 1 {
			errs = append(errs, fmt.Errorf("command '%s' specified %d times - each command can only be used once", command, count))
		}
	}

	// Check for mutually exclusive commands
	if commandCounts[cmdGoUpdateDirectDependencies] > 0 && commandCounts[cmdGoUpdateAllDependencies] > 0 {
		errs = append(errs, fmt.Errorf("commands '%s' and '%s' cannot be used together - choose one dependency update mode", cmdGoUpdateDirectDependencies, cmdGoUpdateAllDependencies))
	}

	if len(errs) > 0 {
		logger.Log("validateCommands: validation errors")

		return fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	logger.Log("validateCommands: success")

	return nil
}
