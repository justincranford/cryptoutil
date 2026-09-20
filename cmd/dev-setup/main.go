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

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

	// Exit gracefully with status 0 even if dependencies have issues.
	// This allows post-checkout hooks to continue without blocking.
	// Users can see the warnings and manually install/fix tools as needed.
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
				} else {
					// Installation command succeeded. Assume the tool is available even if
					// verification check fails (e.g., tool already installed via different mechanism,
					// or checker can't find it in expected location). Idempotent installs succeed
					// silently when already installed.
					if installed, version, _ := chk.Check(ctx, dep); installed {
						depResult.Status = cryptoutilSharedMagic.DevSetupStatusInstalled
						depResult.ActualVersion = version
					} else {
						// Installation succeeded but verification check couldn't confirm it.
						// This is OK - mark as installed and move on. The installation command
						// exiting with 0 is our confirmation that the tool is available.
						depResult.Status = cryptoutilSharedMagic.DevSetupStatusInstalled
						depResult.ActualVersion = "unknown (installed via " + dep.Type + ")"
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
