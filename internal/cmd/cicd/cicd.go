// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"cryptoutil/internal/cmd/cicd/all_enforce_utf8"
	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/github_workflow_lint"
	"cryptoutil/internal/cmd/cicd/go_check_circular_package_dependencies"
	"cryptoutil/internal/cmd/cicd/go_check_identity_imports"
	"cryptoutil/internal/cmd/cicd/go_enforce_any"
	"cryptoutil/internal/cmd/cicd/go_enforce_test_patterns"
	"cryptoutil/internal/cmd/cicd/go_fix_all"
	"cryptoutil/internal/cmd/cicd/go_fix_copyloopvar"
	"cryptoutil/internal/cmd/cicd/go_fix_staticcheck_error_strings"
	"cryptoutil/internal/cmd/cicd/go_fix_thelper"
	"cryptoutil/internal/cmd/cicd/go_identity_requirements_check"
	"cryptoutil/internal/cmd/cicd/go_update_direct_dependencies"
	"cryptoutil/internal/cmd/cicd/go_update_project_status"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

const (
	cmdAllEnforceUTF8                     = "all-enforce-utf8"
	cmdGoEnforceAny                       = "go-enforce-any"
	cmdGoEnforceTestPatterns              = "go-enforce-test-patterns"
	cmdGitHubWorkflowLint                 = "github-workflow-lint"
	cmdGoFixStaticcheckErrorStrings       = "go-fix-staticcheck-error-strings"
	cmdGoFixCopyLoopVar                   = "go-fix-copyloopvar"
	cmdGoFixTHelper                       = "go-fix-thelper"
	cmdGoFixAll                           = "go-fix-all"
	cmdGoCheckCircularPackageDependencies = "go-check-circular-package-dependencies"
	cmdGoCheckIdentityImports             = "go-check-identity-imports"
	cmdGoIdentityRequirementsCheck        = "go-identity-requirements-check"
	cmdGoUpdateDirectDependencies         = "go-update-direct-dependencies"
	cmdGoUpdateAllDependencies            = "go-update-all-dependencies"
	cmdGoUpdateProjectStatus              = "go-update-project-status"
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

	// Extract actual commands (skip flags starting with -)
	actualCommands := []string{}

	for _, cmd := range commands {
		if !strings.HasPrefix(cmd, "-") {
			actualCommands = append(actualCommands, cmd)
		}
	}

	logger.Log(fmt.Sprintf("Executing %d commands", len(actualCommands)))

	// Execute all commands and collect results
	results := make([]common.CommandResult, 0, len(actualCommands))

	// Find index of first actual command to get remaining args
	cmdStartIndex := 0

	for i, arg := range commands {
		if !strings.HasPrefix(arg, "-") {
			cmdStartIndex = i

			break
		}
	}

	for i, command := range actualCommands {
		cmdStart := time.Now()

		logger.Log(fmt.Sprintf("Executing command %d/%d: %s", i+1, len(actualCommands), command))

		var cmdErr error

		// Get remaining args after current command for commands that accept flags
		// Find this command in original args list
		cmdIndex := cmdStartIndex
		for j := cmdStartIndex; j < len(commands); j++ {
			if commands[j] == command {
				cmdIndex = j

				break
			}
		}

		// Remaining args = everything after this command
		remainingArgs := []string{}
		if cmdIndex < len(commands)-1 {
			remainingArgs = commands[cmdIndex+1:]
		}

		switch command {
		case cmdAllEnforceUTF8:
			cmdErr = all_enforce_utf8.Enforce(logger, allFiles)
		case cmdGoEnforceTestPatterns:
			cmdErr = go_enforce_test_patterns.Enforce(logger, allFiles)
		case cmdGoEnforceAny:
			cmdErr = go_enforce_any.Enforce(logger, allFiles)
		case cmdGoCheckCircularPackageDependencies:
			cmdErr = go_check_circular_package_dependencies.Check(logger)
		case cmdGoCheckIdentityImports:
			cmdErr = go_check_identity_imports.Check(logger)
		case cmdGoIdentityRequirementsCheck:
			// Pass remaining args for flag parsing (--strict, --task-threshold, etc.)
			cmdErr = go_identity_requirements_check.Enforce(context.Background(), logger, remainingArgs)
		case cmdGoUpdateProjectStatus:
			cmdErr = go_update_project_status.Update(context.Background(), logger, remainingArgs)
		case cmdGoUpdateDirectDependencies:
			cmdErr = go_update_direct_dependencies.Update(logger, cryptoutilMagic.DepCheckDirect)
		case cmdGoUpdateAllDependencies:
			cmdErr = go_update_direct_dependencies.Update(logger, cryptoutilMagic.DepCheckAll)
		case cmdGitHubWorkflowLint:
			cmdErr = github_workflow_lint.Lint(logger, allFiles)
		case cmdGoFixStaticcheckErrorStrings:
			_, _, _, cmdErr = go_fix_staticcheck_error_strings.Fix(logger, ".")
		case cmdGoFixCopyLoopVar:
			_, _, _, cmdErr = go_fix_copyloopvar.Fix(logger, ".", runtime.Version())
		case cmdGoFixTHelper:
			_, _, _, cmdErr = go_fix_thelper.Fix(logger, ".")
		case cmdGoFixAll:
			_, _, _, cmdErr = go_fix_all.Fix(logger, ".", runtime.Version())
		}

		cmdDuration := time.Since(cmdStart)
		results = append(results, common.CommandResult{
			Command:  command,
			Duration: cmdDuration,
			Error:    cmdErr,
		})

		logger.Log(fmt.Sprintf("Command '%s' completed in %.2fs", command, cmdDuration.Seconds()))

		// Add a separator between multiple commands
		if i < len(actualCommands)-1 {
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
		// Skip flag arguments (--strict, --task-threshold, etc.)
		// Flags are passed to subcommand Enforce functions, not validated as commands
		if strings.HasPrefix(command, "-") {
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
