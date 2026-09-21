package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Dependency represents a single tool/library that must be installed.
type Dependency struct {
	// Name of the dependency (e.g., "pre-commit", "golangci-lint")
	Name string

	// Type categorizes how the tool should be installed
	Type string

	// MinVersion is the minimum required version (used for display/validation)
	MinVersion string

	// CheckCmd is the command to verify installation (e.g., "pre-commit --version")
	CheckCmd string

	// InstallCmd is the command to install the tool
	InstallCmd string

	// Description explains what the tool does
	Description string

	// DetectVersionCmd extracts the version from the output of CheckCmd
	// If empty, version detection is skipped
	DetectVersionCmd string
}

// Group represents a logical grouping of dependencies (ordering matters).
type Group struct {
	// Name of the group (e.g., "Python Runtime", "Go Toolchain", "Pre-commit Hooks")
	Name string

	// Description explains the group's purpose
	Description string

	// Dependencies in this group (processed in order)
	Dependencies []*Dependency

	// BlockOnFailure: if true, subsequent groups are skipped on failure
	BlockOnFailure bool
}

// Config contains all dependency groups in installation order.
type Config struct {
	Groups []*Group
}

// Load returns the hardcoded dependency configuration.
func Load() (*Config, error) {
	return &Config{
		Groups: []*Group{
			pythonRuntimeGroup(),
			uvToolGroup(),
			pythonDependenciesGroup(),
			goToolchainGroup(),
			golangciLintGroup(),
			preCommitGroup(),
		},
	}, nil
}

// pythonRuntimeGroup ensures Python is available (baseline for everything).
func pythonRuntimeGroup() *Group {
	return &Group{
		Name:        cryptoutilSharedMagic.DevSetupTestGroupPythonRuntime,
		Description: "Python 3.14+ is required for all other tools",
		Dependencies: []*Dependency{
			{
				Name:             cryptoutilSharedMagic.DevSetupToolTypePython,
				Type:             cryptoutilSharedMagic.DevSetupToolTypeSystemBinary,
				MinVersion:       cryptoutilSharedMagic.DevSetupTestVersionPythonRequired,
				CheckCmd:         cryptoutilSharedMagic.DevSetupPythonCheckCmd,
				Description:      "Python runtime (required for pre-commit, semgrep, and linters)",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupPythonCheckCmd,
			},
		},
		BlockOnFailure: true,
	}
}

// uvToolGroup installs uv package manager.
func uvToolGroup() *Group {
	return &Group{
		Name:        cryptoutilSharedMagic.DevSetupTestGroupUVPackageManager,
		Description: "uv is a fast Python package installer used for managing dev dependencies",
		Dependencies: []*Dependency{
			{
				Name:             "uv",
				Type:             cryptoutilSharedMagic.DevSetupToolTypeSystemBinary,
				MinVersion:       "0.12.0",
				CheckCmd:         cryptoutilSharedMagic.DevSetupUVCheckCmd,
				Description:      "Fast Python package installer and virtualenv manager",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupUVCheckCmd,
				// No InstallCmd: uv has no portable, non-shell bootstrap command; install manually
				// from https://docs.astral.sh/uv/getting-started/ if this check fails.
			},
		},
		BlockOnFailure: true,
	}
}

// pythonDependenciesGroup installs Python packages via uv.
func pythonDependenciesGroup() *Group {
	return &Group{
		Name:        cryptoutilSharedMagic.DevSetupTestGroupPythonDependencies,
		Description: "Python packages needed for development and CI/CD",
		Dependencies: []*Dependency{
			{
				Name:             "pre-commit",
				Type:             cryptoutilSharedMagic.DevSetupToolTypePythonModule,
				MinVersion:       "3.0.0",
				CheckCmd:         cryptoutilSharedMagic.DevSetupPreCommitCheckCmd,
				Description:      "Git hook framework for managing and maintaining multi-language pre-commit hooks",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupPreCommitCheckCmd,
				InstallCmd:       cryptoutilSharedMagic.DevSetupPreCommitInstallCmd,
			},
			{
				Name:             "semgrep",
				Type:             cryptoutilSharedMagic.DevSetupToolTypePythonModule,
				MinVersion:       "1.45.0",
				CheckCmd:         cryptoutilSharedMagic.DevSetupSemgrepCheckCmd,
				Description:      "Static analysis tool for code security and quality",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupSemgrepCheckCmd,
				InstallCmd:       cryptoutilSharedMagic.DevSetupSemgrepInstallCmd,
			},
			{
				Name:             "yamllint",
				Type:             cryptoutilSharedMagic.DevSetupToolTypePythonModule,
				MinVersion:       "1.26.0",
				CheckCmd:         cryptoutilSharedMagic.DevSetupYamllintCheckCmd,
				Description:      "YAML linter for consistent YAML formatting",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupYamllintCheckCmd,
				InstallCmd:       cryptoutilSharedMagic.DevSetupYamllintInstallCmd,
			},
		},
		BlockOnFailure: false,
	}
}

// goToolchainGroup ensures Go is available.
func goToolchainGroup() *Group {
	return &Group{
		Name:        cryptoutilSharedMagic.DevSetupTestGroupGoToolchain,
		Description: "Go runtime and build tools",
		Dependencies: []*Dependency{
			{
				Name:             "go",
				Type:             cryptoutilSharedMagic.DevSetupToolTypeSystemBinary,
				MinVersion:       cryptoutilSharedMagic.CICDTemplateGoVersion,
				CheckCmd:         cryptoutilSharedMagic.DevSetupGoCheckCmd,
				Description:      "Go programming language (required for building and testing)",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupGoCheckCmd,
			},
		},
		BlockOnFailure: true,
	}
}

// golangciLintGroup installs golangci-lint for Go linting.
func golangciLintGroup() *Group {
	return &Group{
		Name:        cryptoutilSharedMagic.DevSetupTestGroupGoLinting,
		Description: "Go static analysis and linting tools",
		Dependencies: []*Dependency{
			{
				Name:             "golangci-lint",
				Type:             cryptoutilSharedMagic.DevSetupToolTypeGoModule,
				MinVersion:       cryptoutilSharedMagic.DevSetupTestGolangciVersion,
				CheckCmd:         cryptoutilSharedMagic.DevSetupGolangciCheckCmd,
				Description:      "Fast Go linter aggregator (pinned per repo policy)",
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupGolangciCheckCmd,
				InstallCmd:       cryptoutilSharedMagic.DevSetupGolangciInstallCmd,
			},
		},
		BlockOnFailure: false,
	}
}

// preCommitGroup ensures git hooks are installed into .git/hooks (idempotent:
// re-running "pre-commit install" when hooks already exist is a no-op).
func preCommitGroup() *Group {
	return &Group{
		Name:        cryptoutilSharedMagic.DevSetupTestGroupPreCommitHooks,
		Description: "Git hooks for code quality checks",
		Dependencies: []*Dependency{
			{
				Name:        "pre-commit-hooks",
				Type:        cryptoutilSharedMagic.DevSetupToolTypeSystemBinary,
				CheckCmd:    cryptoutilSharedMagic.DevSetupGitHooksCheckCmd,
				Description: "Git hook scripts installed in .git/hooks",
				InstallCmd:  cryptoutilSharedMagic.DevSetupGitHooksInstallCmd,
			},
		},
		BlockOnFailure: false,
	}
}
