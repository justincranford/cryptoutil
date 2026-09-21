// Copyright (c) 2025-2026 Justin Cranford.
// Package cicd_dev_setup implements the developer environment setup CLI orchestration
// (dependency verification and idempotent installation for local dev/CI setup).
package cicd_dev_setup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	cryptoutilAppsToolsDevSetupChecker "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/checker"
	cryptoutilAppsToolsDevSetupConfig "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/config"
	cryptoutilAppsToolsDevSetupInstaller "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/installer"
	cryptoutilAppsToolsDevSetupReporter "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/reporter"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// exitCodeSuccess and exitCodeFailure are the process exit codes returned by Main.
const (
	exitCodeSuccess = 0
	exitCodeFailure = 1
)

// Setup holds all injected dependencies for the dev-setup orchestration.
// Using this struct allows tests to inject mock implementations.
type Setup struct {
	Config    *cryptoutilAppsToolsDevSetupConfig.Config
	Checker   cryptoutilAppsToolsDevSetupChecker.Checker
	Installer cryptoutilAppsToolsDevSetupInstaller.Installer
	Reporter  cryptoutilAppsToolsDevSetupReporter.Reporter
}

// Main is the CLI entry point invoked by cmd/cicd-dev-setup/main.go.
// It accepts args/stdin/stdout/stderr for testability and returns a process exit code.
func Main(_ []string, _ io.Reader, stdout, stderr io.Writer) int {
	if err := internalMain(stdout, stderr); err != nil {
		_, _ = fmt.Fprintf(stderr, "ERROR: %v\n", err)

		return exitCodeFailure
	}

	return exitCodeSuccess
}

// internalMain is the testable main function that accepts injected dependencies via Setup.
// This enables unit tests to inject mock implementations of Checker, Installer, and Reporter.
func internalMain(stdout, stderr io.Writer) error {
	ctx := context.Background()

	// Load dependency configuration
	cfg, err := cryptoutilAppsToolsDevSetupConfig.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine project root
	projectRoot, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to determine project root: %w", err)
	}

	// Initialize components with real implementations (seam pattern).
	// Tests can call internalMainWithSetup() instead to inject mocks.
	exec := cryptoutilAppsToolsDevSetupInstaller.NewSystemExecutor()
	chk := cryptoutilAppsToolsDevSetupChecker.NewDependencyChecker(exec)
	inst := cryptoutilAppsToolsDevSetupInstaller.NewDependencyInstaller(exec, projectRoot)
	rep := cryptoutilAppsToolsDevSetupReporter.NewConsoleReporter(stdout, stderr)

	setup := &Setup{
		Config:    cfg,
		Checker:   chk,
		Installer: inst,
		Reporter:  rep,
	}

	return internalMainWithSetup(ctx, setup)
}

// internalMainWithSetup is the core logic that can accept injected dependencies.
// This is used by internalMain() in production and by tests that need to inject mocks.
func internalMainWithSetup(ctx context.Context, setup *Setup) error {
	if setup == nil {
		return fmt.Errorf("setup is required")
	}

	// Run dependency verification and installation
	summary, err := runSetup(ctx, setup.Config, setup.Checker, setup.Installer, setup.Reporter)
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Report results
	if err := setup.Reporter.ReportSummary(summary); err != nil {
		return fmt.Errorf("failed to report results: %w", err)
	}

	// Exit gracefully with status 0 even if dependencies have issues.
	// This allows post-checkout hooks to continue without blocking.
	// Users can see the warnings and manually install/fix tools as needed.
	return nil
}

func runSetup(
	ctx context.Context,
	cfg *cryptoutilAppsToolsDevSetupConfig.Config,
	chk cryptoutilAppsToolsDevSetupChecker.Checker,
	inst cryptoutilAppsToolsDevSetupInstaller.Installer,
	rep cryptoutilAppsToolsDevSetupReporter.Reporter,
) (*cryptoutilAppsToolsDevSetupReporter.Summary, error) {
	summary := &cryptoutilAppsToolsDevSetupReporter.Summary{
		Groups: make(map[string]*cryptoutilAppsToolsDevSetupReporter.GroupResult),
	}

	// Process each dependency group in order
	for _, group := range cfg.Groups {
		groupResult := &cryptoutilAppsToolsDevSetupReporter.GroupResult{
			Name:         group.Name,
			Dependencies: make([]*cryptoutilAppsToolsDevSetupReporter.DepResult, 0),
		}

		// Check and install each dependency in the group
		for _, dep := range group.Dependencies {
			depResult := &cryptoutilAppsToolsDevSetupReporter.DepResult{
				Name:    dep.Name,
				Version: dep.MinVersion,
			}

			// Check if already installed (idempotent: skip install entirely when found)
			installed, version, err := chk.Check(ctx, dep)
			if err != nil {
				depResult.Status = cryptoutilSharedMagic.DevSetupStatusCheckFailed
				depResult.Error = err
				groupResult.HasErrors = true
				summary.HasErrors = true
			} else if installed {
				depResult.Status = cryptoutilSharedMagic.DevSetupStatusInstalled
				depResult.ActualVersion = version
			} else {
				// Not installed - attempt installation
				if err := inst.Install(ctx, dep); err != nil {
					depResult.Status = cryptoutilSharedMagic.DevSetupStatusInstallFailed
					depResult.Error = err
					groupResult.HasErrors = true
					summary.HasErrors = true
				} else if installed, version, err := chk.Check(ctx, dep); err == nil && installed {
					depResult.Status = cryptoutilSharedMagic.DevSetupStatusInstalled
					depResult.ActualVersion = version
				} else {
					depResult.Status = cryptoutilSharedMagic.DevSetupStatusVerifyFailed
					depResult.Error = fmt.Errorf("installation succeeded but tool not found on PATH")
					groupResult.HasErrors = true
					summary.HasErrors = true
				}
			}

			groupResult.Dependencies = append(groupResult.Dependencies, depResult)
		}

		summary.Groups[group.Name] = groupResult
	}

	return summary, nil
}

func getProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd, nil
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", fmt.Errorf("project root not found (no go.mod)")
		}

		cwd = parent
	}
}
