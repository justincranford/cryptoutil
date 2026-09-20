package main

import (
	"context"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/checker"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/config"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/installer"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/reporter"
	"fmt"
	"os"
	"path/filepath"
)

// Setup holds all injected dependencies for the dev-setup orchestration.
// Using this struct allows tests to inject mock implementations.
type Setup struct {
	Config    *config.Config
	Checker   checker.Checker
	Installer installer.Installer
	Reporter  reporter.Reporter
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

// internalMain is the testable main function that accepts injected dependencies via Setup.
// This enables unit tests to inject mock implementations of Checker, Installer, and Reporter.
func internalMain() error {
	ctx := context.Background()

	// Load dependency configuration
	cfg, err := config.Load()
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
	exec := installer.NewSystemExecutor()
	chk := checker.NewDependencyChecker(exec)
	inst := installer.NewDependencyInstaller(exec, projectRoot)
	rep := reporter.NewConsoleReporter(os.Stdout, os.Stderr)

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

	// Return error if any dependency check/install failed
	if summary.HasErrors {
		return fmt.Errorf("some dependencies failed to install")
	}

	return nil
}

func runSetup(
	ctx context.Context,
	cfg *config.Config,
	chk checker.Checker,
	inst installer.Installer,
	rep reporter.Reporter,
) (*reporter.Summary, error) {
	summary := &reporter.Summary{
		Groups: make(map[string]*reporter.GroupResult),
	}

	// Process each dependency group in order
	for _, group := range cfg.Groups {
		groupResult := &reporter.GroupResult{
			Name:         group.Name,
			Dependencies: make([]*reporter.DepResult, 0),
		}

		// Check and install each dependency in the group
		for _, dep := range group.Dependencies {
			depResult := &reporter.DepResult{
				Name:    dep.Name,
				Version: dep.MinVersion,
			}

			// Check if already installed
			installed, version, err := chk.Check(ctx, dep)
			if err != nil {
				depResult.Status = reporter.StatusCheckFailed
				depResult.Error = err
				groupResult.HasErrors = true
				summary.HasErrors = true
			} else if installed {
				depResult.Status = reporter.StatusInstalled
				depResult.ActualVersion = version
			} else {
				// Not installed - attempt installation
				if err := inst.Install(ctx, dep); err != nil {
					depResult.Status = reporter.StatusInstallFailed
					depResult.Error = err
					groupResult.HasErrors = true
					summary.HasErrors = true
				} else {
					// Verify installation
					if installed, version, err := chk.Check(ctx, dep); err != nil {
						depResult.Status = reporter.StatusVerifyFailed
						depResult.Error = fmt.Errorf("post-install verification failed: %w", err)
						groupResult.HasErrors = true
						summary.HasErrors = true
					} else if installed {
						depResult.Status = reporter.StatusInstalled
						depResult.ActualVersion = version
					} else {
						depResult.Status = reporter.StatusInstallFailed
						depResult.Error = fmt.Errorf("installation succeeded but tool not found")
						groupResult.HasErrors = true
						summary.HasErrors = true
					}
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
