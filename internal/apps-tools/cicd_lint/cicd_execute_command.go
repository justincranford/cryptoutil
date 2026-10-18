// Copyright (c) 2025-2026 Justin Cranford.
//
//

package cicd_lint

import (
	"fmt"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilCmdCicdFormatGo "cryptoutil/internal/apps-tools/cicd_lint/format_go"
	cryptoutilCmdCicdFormatGotest "cryptoutil/internal/apps-tools/cicd_lint/format_gotest"
	cryptoutilCmdCicdFormatText "cryptoutil/internal/apps-tools/cicd_lint/format_text"
	cryptoutilGitHubCleanup "cryptoutil/internal/apps-tools/cicd_lint/github_cleanup"
	cryptoutilCmdCicdLintCompose "cryptoutil/internal/apps-tools/cicd_lint/lint_compose"
	cryptoutilLintDeployments "cryptoutil/internal/apps-tools/cicd_lint/lint_deployments"
	cryptoutilLintDocs "cryptoutil/internal/apps-tools/cicd_lint/lint_docs"
	cryptoutilLintFitness "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness"
	cryptoutilCmdCicdLintGo "cryptoutil/internal/apps-tools/cicd_lint/lint_go"
	cryptoutilCmdCicdLintGoMod "cryptoutil/internal/apps-tools/cicd_lint/lint_go_mod"
	cryptoutilCmdCicdLintGolangci "cryptoutil/internal/apps-tools/cicd_lint/lint_golangci"
	cryptoutilCmdCicdLintGotest "cryptoutil/internal/apps-tools/cicd_lint/lint_gotest"
	cryptoutilCmdCicdLintJavaTest "cryptoutil/internal/apps-tools/cicd_lint/lint_javatest"
	cryptoutilCmdCicdLintOpenAPI "cryptoutil/internal/apps-tools/cicd_lint/lint_openapi"
	cryptoutilCmdCicdLintPorts "cryptoutil/internal/apps-tools/cicd_lint/lint_ports"
	cryptoutilCmdCicdLintPythonTest "cryptoutil/internal/apps-tools/cicd_lint/lint_pythontest"
	cryptoutilCmdCicdLintText "cryptoutil/internal/apps-tools/cicd_lint/lint_text"
	cryptoutilCmdCicdLintWorkflow "cryptoutil/internal/apps-tools/cicd_lint/lint_workflow"
)

func executeCommand(
	logger *cryptoutilCmdCicdCommon.Logger,
	command string,
	filesByExtension map[string][]string,
	extraArgs []string,
) cryptoutilCmdCicdCommon.CommandResult {
	return executeCommandWithFormatters(
		logger,
		command,
		filesByExtension,
		extraArgs,
		cryptoutilCmdCicdFormatGo.Format,
		cryptoutilCmdCicdFormatGotest.Format,
	)
}

func executeCommandWithFormatters(
	logger *cryptoutilCmdCicdCommon.Logger,
	command string,
	filesByExtension map[string][]string,
	extraArgs []string,
	formatGo func(*cryptoutilCmdCicdCommon.Logger, map[string][]string) error,
	formatGoTest func(*cryptoutilCmdCicdCommon.Logger) error,
) cryptoutilCmdCicdCommon.CommandResult {
	cmdStart := time.Now().UTC()

	logger.Log(fmt.Sprintf("Executing command: %s", command))

	var cmdErr error

	switch command {
	case cmdLintText:
		cmdErr = cryptoutilCmdCicdLintText.Lint(logger, filesByExtension)
	case cmdLintGo:
		cmdErr = cryptoutilCmdCicdLintGo.Lint(logger)
	case cmdLintCompose:
		cmdErr = cryptoutilCmdCicdLintCompose.Lint(logger, filesByExtension)
	case cmdFormatGo:
		cmdErr = formatGo(logger, filesByExtension)
	case cmdLintGoTest:
		cmdErr = cryptoutilCmdCicdLintGotest.Lint(logger, filesByExtension)
	case cmdFormatGoTest:
		cmdErr = formatGoTest(logger)
	case cmdFormatText:
		cmdErr = cryptoutilCmdCicdFormatText.Format(logger, filesByExtension)
	case cmdLintWorkflow:
		cmdErr = cryptoutilCmdCicdLintWorkflow.Lint(logger, filesByExtension)
	case cmdLintGoMod:
		cmdErr = cryptoutilCmdCicdLintGoMod.Lint(logger)
	case cmdLintPorts:
		cmdErr = cryptoutilCmdCicdLintPorts.Lint(logger, filesByExtension)
	case cmdLintGolangci:
		cmdErr = cryptoutilCmdCicdLintGolangci.Lint(logger, filesByExtension)
	case cmdLintDocs:
		cmdErr = cryptoutilLintDocs.Lint(logger)
	case cmdLintDeployments:
		cmdErr = cryptoutilLintDeployments.Lint(logger)
	case cmdLintFitness:
		cmdErr = cryptoutilLintFitness.Lint(logger)
	case cmdLintOpenAPI:
		cmdErr = cryptoutilCmdCicdLintOpenAPI.Lint(logger, filesByExtension)
	case cmdLintJavaTest:
		cmdErr = cryptoutilCmdCicdLintJavaTest.Lint(logger, filesByExtension)
	case cmdLintPythonTest:
		cmdErr = cryptoutilCmdCicdLintPythonTest.Lint(logger, filesByExtension)
	case cmdGitHubCleanup:
		cmdErr = cryptoutilGitHubCleanup.Cleanup(logger, extraArgs)
	}

	cmdDuration := time.Since(cmdStart)
	logger.Log(fmt.Sprintf("Command '%s' completed in %.2fs", command, cmdDuration.Seconds()))

	return cryptoutilCmdCicdCommon.CommandResult{
		Command:  command,
		Duration: cmdDuration,
		Error:    cmdErr,
	}
}
