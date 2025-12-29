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

	// ListAllFiles start directory.
	ListAllFilesStartDirectory = "."

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
	  lint-text      - [Linter] Enforce UTF-8 encoding without BOM for text files
	  lint-go        - [Linter] Check for circular dependencies in Go packages
	  format-go      - [Formatter] Auto-fix Go files (any -> any, loop var copies)
	  lint-go-test   - [Linter] Enforce test patterns (UUIDv7 usage, testify assertions)
	  format-go-test - [Formatter] Auto-fix Go test files (add t.Helper() to helpers)
	  lint-workflow  - [Linter] Validate GitHub Actions workflow naming and versions
	  lint-go-mod    - [Linter] Check direct Go dependencies for updates`
)

// ValidCommands defines the set of valid cicd commands.
var ValidCommands = map[string]bool{
	"lint-text":           true,
	"lint-go":             true,
	"format-go":           true,
	"lint-go-test":        true,
	"format-go-test":      true,
	"lint-workflow":       true,
	"lint-go-mod":         true,
	"check-no-cgo-sqlite": true,
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
	// CICDSelfExclusionPatterns - Self-exclusion patterns for each cicd command.
	// CRITICAL: Each command excludes its own subdirectory to prevent self-modification.
	// Keys match command names: lint-text, lint-go, format-go, lint-go-test, format-go-test, lint-workflow, lint-go-mod.
	CICDSelfExclusionPatterns = map[string]string{
		"lint-text":      `internal[/\\]cmd[/\\]cicd[/\\]lint_text[/\\].*\.go$`,
		"lint-go":        `internal[/\\]cmd[/\\]cicd[/\\]lint_go[/\\].*\.go$`,
		"format-go":      `internal[/\\]cmd[/\\]cicd[/\\]format_go[/\\].*\.go$`,
		"lint-go-test":   `internal[/\\]cmd[/\\]cicd[/\\]lint_gotest[/\\].*\.go$`,
		"format-go-test": `internal[/\\]cmd[/\\]cicd[/\\]format_gotest[/\\].*\.go$`,
		"lint-workflow":  `internal[/\\]cmd[/\\]cicd[/\\]lint_workflow[/\\].*\.go$`,
		"lint-go-mod":    `internal[/\\]cmd[/\\]cicd[/\\]lint_go_mod[/\\].*\.go$`,
	}

	// GeneratedFileExcludePatterns - File patterns for generated files that should be excluded from linting.
	// These are excluded in addition to directory-level exclusions.
	GeneratedFileExcludePatterns = []string{
		`_gen\.go$`, // Generated files.
		`\.pb\.go$`, // Protocol buffer files.
	}

	// EnforceUtf8FileIncludePatterns - File patterns to include in UTF-8 encoding checks.
	EnforceUtf8FileIncludePatterns = []string{
		// Source code files.
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
		// Database and configuration files.
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

	// TextFilenameExtensionInclusions defines which file extensions to include when scanning.
	// Used by ListAllFiles to filter files by extension.
	TextFilenameExtensionInclusions = []string{
		"go",
		"yml",
		"yaml",
		"mod",
		"sum",
		"json",
		"md",
		"txt",
		"toml",
		"ps1",
		"sh",
		"sql",
		"gitignore",
		"dockerignore",
		"properties",
		"log",
		"out",
		"pem",
	}

	// DirectoryNameExclusions defines directories to skip when scanning.
	// Used by ListAllFiles to exclude generated/vendored directories.
	DirectoryNameExclusions = []string{
		".git",
		"api/client",
		"api/model",
		"api/server",
		"api/idp",
		"api/authz",
		"node_modules",
		"test-output",
		"vendor",
		"workflow-reports",
	}
)
