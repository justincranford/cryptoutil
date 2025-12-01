// Copyright (c) 2025 Justin Cranford
//
//

package magic

import (
	"regexp"
	"time"
)

// DepCheckMode represents the mode for dependency checking operations.
type DepCheckMode int

// Dependency check modes for go-update-dependencies commands.
const (
	// DepCheckDirect - Check only direct dependencies.
	DepCheckDirect DepCheckMode = iota
	// DepCheckAll - Check all dependencies (direct + transitive).
	DepCheckAll
)

// DepCache represents cached dependency check results.
type DepCache struct {
	LastCheck    time.Time `json:"last_check"`
	GoModModTime time.Time `json:"go_mod_mod_time"`
	GoSumModTime time.Time `json:"go_sum_mod_time"`
	OutdatedDeps []string  `json:"outdated_deps"`
	Mode         string    `json:"mode"`
}

// CircularDepCache represents cached circular dependency check results.
type CircularDepCache struct {
	LastCheck       time.Time `json:"last_check"`
	GoModModTime    time.Time `json:"go_mod_mod_time"`
	HasCircularDeps bool      `json:"has_circular_deps"`
	CircularDeps    []string  `json:"circular_deps"` // Flattened representation of cycles
}

// Regex patterns for test enforcement.

const (
	// UI constants.
	SeparatorLength = 50

	// MinActionMatchGroups is the minimum number of regex match groups for action parsing.
	MinActionMatchGroups = 3

	// Cache file permissions (owner read/write only).
	CacheFilePermissions = 0o600

	// CacheDuration is the duration after which cache entries expire.
	CacheDuration = 5 * time.Minute

	// Directory permissions (owner read/write/execute, group and others execute).
	CICDOutputDirPermissions = 0o711

	// CICD output directory for all generated files, reports, and cache files.
	// Centralizes all cicd outputs to declutter repo root, simplify .gitignore patterns,
	// and simplify VS Code settings.json file exclusions.
	CICDOutputDir = ".cicd"

	// CICD output file permissions (owner read/write, group/others read).
	CICDOutputFilePermissions = 0o644

	// Identity project paths.
	CICDIdentityProjectStatusPath        = "docs/02-identityV2/PROJECT-STATUS.md"
	CICDIdentityRequirementsCoveragePath = "docs/02-identityV2/REQUIREMENTS-COVERAGE.md"
	CICDIdentityTaskDocsDir              = "docs/02-identityV2/passthru5/"

	// Requirements coverage regex pattern groups.
	RequirementsTotalPatternGroups        = 4
	RequirementsPriorityPatternGroups     = 4
	RequirementsTaskCoveragePatternGroups = 3
	RequirementsUncoveredPatternGroups    = 2

	// Test coverage regex pattern groups.
	TestCoveragePatternGroups = 2

	// Git short hash length.
	GitShortHashLength = 8

	// Git recent activity days lookback.
	GitRecentActivityDays = 7

	// Date format for PROJECT-STATUS.md timestamps.
	DateFormatYYYYMMDD = "2006-01-02"

	// Percent multiplier for coverage calculations.
	PercentMultiplier = 100.0

	// Production readiness thresholds.
	RequirementsProductionReadyThreshold = 85.0
	TestCoverageProductionReadyThreshold = 85.0
	RequirementsConditionalThreshold     = 80.0
	TestCoverageConditionalThreshold     = 80.0
	RequirementsTaskMinimumThreshold     = 90.0

	// Dependency cache file name.
	DepCacheFileName = ".cicd/dep-cache.json"

	// Circular dependency cache file name.
	CircularDepCacheFileName = ".cicd/circular-dep-cache.json"

	// Dependency check mode names.
	ModeNameDirect = "direct"
	ModeNameAll    = "all"

	// Time constants for dependency cache testing.
	TestCacheValidMinutes = 30
	TestCacheExpiredHours = 2

	// Dependency cache validity duration.
	DepCacheValidDuration = 30 * time.Minute

	// Circular dependency cache validity (longer since go.mod changes less frequently).
	CircularDepCacheValidDuration = 60 * time.Minute

	// Time constants for GitHub API cache testing.
	TestGitHubAPICacheExpiredHours = 1

	// Time format for logging and timestamps.
	TimeFormat = "2006-01-02T15:04:05.999999999Z07:00"

	// Number of worker threads for concurrent file processing operations.
	Utf8EnforceWorkerPoolSize = 6

	// Usage message for the cicd command.
	UsageCICD = `Usage: cicd <command> [command...]

	Commands:
	  all-enforce-utf8                       - [Linter] Enforce UTF-8 encoding without BOM
	  go-enforce-test-patterns               - [Linter] Enforce test patterns (UUIDv7 usage, testify assertions)
	  go-enforce-any                         - [Formatter] Custom Go source code fixes (any -> any, etc.)
	  go-check-circular-package-dependencies - [Linter] Check for circular dependencies in Go packages
	  go-update-direct-dependencies          - [Linter] Check direct Go dependencies only
	  go-update-all-dependencies             - [Linter] Check all Go dependencies (direct + transitive)
	  github-workflow-lint                   - [Linter] Validate GitHub Actions workflow naming and structure, and check for outdated actions
	  go-fix-copyloopvar                     - [Formatter] Auto-fix: Remove unnecessary loop variable copies (Go 1.25+)
	  go-fix-thelper                         - [Formatter] Auto-fix: Add t.Helper() to test helper functions
	  go-fix-all                             - [Formatter] Auto-fix: Run all go-fix-* commands in sequence`
)

// ValidCommands defines the set of valid cicd commands.
var ValidCommands = map[string]bool{
	"all-enforce-utf8":                       true,
	"go-enforce-test-patterns":               true,
	"go-enforce-any":                         true,
	"go-check-circular-package-dependencies": true,
	"go-update-direct-dependencies":          true,
	"go-update-all-dependencies":             true,
	"github-workflow-lint":                   true,
	"go-fix-copyloopvar":                     true,
	"go-fix-thelper":                         true,
	"go-fix-all":                             true,
}

// Regex patterns for test enforcement.
var (
	// TestErrorfPattern - Regex pattern to match test.Errorf calls.
	TestErrorfPattern = regexp.MustCompile(`\bt\.Errorf\s*\(`)
	// TestFatalfPattern - Regex pattern to match test.Fatalf calls.
	TestFatalfPattern = regexp.MustCompile(`\bt\.Fatalf\s*\(`)

	// Test validation regex patterns for cicd test files.
	TestErrorfValidationPattern  = regexp.MustCompile(`^t\.Errorf\([^)]+\)$`)
	TestFErrorfValidationPattern = regexp.MustCompile(`^f\.Errorf\([^)]+\)$`)
	TestFatalfValidationPattern  = regexp.MustCompile(`t\.Fatalf\([^)]+\)`)
)

// File patterns for CI/CD enforcement commands.
var (
	// GoEnforceAnyFileExcludePatterns - Files excluded from go-enforce-any command to prevent self-modification.
	// CRITICAL: Test files containing deliberate `interface {}` (note space) patterns MUST be excluded to prevent modification.
	// When adding new tests for `interface {}` to `any` enforcement, add them to cicd_enforce_any_test.go.
	GoEnforceAnyFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_enforce_any[/\\].*\.go$`, // Exclude files in itself to prevent self-modification and self-destruction
		`api/client`,    // Generated API client
		`api/model`,     // Generated API models
		`api/server`,    // Generated API server
		`_gen\.go$`,     // Generated files
		`\.pb\.go$`,     // Protocol buffer files
		`vendor/`,       // Vendored dependencies
		`.git/`,         // Git directory
		`node_modules/`, // Node.js dependencies
	}

	// EnforceUtf8FileIncludePatterns - File patterns to include in UTF-8 encoding checks.
	EnforceUtf8FileIncludePatterns = []string{
		// Source code files
		"*.go",    // Go source files
		"*.java",  // Java source files
		"*.sh",    // Shell scripts
		"*.py",    // Python scripts and utilities
		"*.ps1",   // PowerShell scripts
		"*.psm1",  // PowerShell module files
		"*.psd1",  // PowerShell data files
		"*.bat",   // Windows batch files
		"*.cmd",   // Windows command files
		"*.c",     // C source files
		"*.cpp",   // C++ source files
		"*.h",     // C/C++ header files
		"*.php",   // PHP files
		"*.rb",    // Ruby files
		"*.rs",    // Rust source files
		"*.js",    // JavaScript files
		"*.ts",    // TypeScript files
		"*.tsx",   // TypeScript React files
		"*.vue",   // Vue.js files
		"*.kt",    // Kotlin source files
		"*.kts",   // Kotlin script files
		"*.swift", // Swift source files
		// Database and configuration files
		"*.sql",        // SQL files
		"*.xml",        // XML configuration and data files
		"*.yml",        // YAML files
		"*.yaml",       // YAML files
		"*.json",       // JSON files
		"*.toml",       // TOML configuration files
		"*.tmpl",       // Template files
		"*.properties", // Properties files
		"*.ini",        // INI files
		"*.cfg",        // Configuration files
		"*.conf",       // Configuration files
		"*.config",     // Configuration files
		"config",       // Generic config files
		".env",         // Environment variable files
		// Build files
		"Dockerfile", // Dockerfiles
		"Makefile",   // Makefiles
		"*.mk",       // Makefiles
		"*.cmake",    // CMake files
		"*.gradle",   // Gradle build files
		// Data files
		"*.csv",    // CSV data files
		"*.pem",    // PEM files
		"*.secret", // Secret files
		// Documentation and markup files
		"*.html",     // HTML files
		"*.css",      // CSS files
		"*.md",       // Markdown files
		"*.txt",      // Text files
		"*.asciidoc", // AsciiDoc files
		"*.adoc",     // AsciiDoc files
	}

	// AllEnforceUtf8FileExcludePatterns - File patterns to exclude from UTF-8 encoding checks.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	AllEnforceUtf8FileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]all_enforce_utf8[/\\].*\.go$`, // Exclude files in itself to prevent self-modification
		`_gen\.go$`,     // Generated files
		`\.pb\.go$`,     // Protocol buffer files
		`vendor/`,       // Vendored dependencies
		`api/client`,    // Generated API client
		`api/model`,     // Generated API models
		`api/server`,    // Generated API server
		`.git/`,         // Git directory
		`node_modules/`, // Node.js dependencies
	}

	// GoEnforceTestPatternsFileExcludePatterns - Files excluded from go-enforce-test-patterns command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoEnforceTestPatternsFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_enforce_test_patterns[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GoCheckCircularPackageDependenciesFileExcludePatterns - Files excluded from go-check-circular-package-dependencies command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoCheckCircularPackageDependenciesFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_check_circular_package_dependencies[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GoFixCopyLoopVarFileExcludePatterns - Files excluded from go-fix-copyloopvar command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoFixCopyLoopVarFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_fix_copyloopvar[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GoFixTHelperFileExcludePatterns - Files excluded from go-fix-thelper command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoFixTHelperFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_fix_thelper[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GoFixAllFileExcludePatterns - Files excluded from go-fix-all command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoFixAllFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_fix_all[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GoUpdateDirectDependenciesFileExcludePatterns - Files excluded from go-update-direct-dependencies command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoUpdateDirectDependenciesFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_update_direct_dependencies[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GoUpdateAllDependenciesFileExcludePatterns - Files excluded from go-update-all-dependencies command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GoUpdateAllDependenciesFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]go_update_all_dependencies[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}

	// GithubWorkflowLintFileExcludePatterns - Files excluded from github-workflow-lint command.
	// CRITICAL: Exclude own subdirectory to prevent self-modification.
	GithubWorkflowLintFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]github_workflow_lint[/\\].*\.go$`,
		`api/client`, `api/model`, `api/server`,
		`_gen\.go$`, `\.pb\.go$`, `vendor/`, `.git/`, `node_modules/`,
	}
)
