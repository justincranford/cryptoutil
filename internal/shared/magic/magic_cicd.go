// Copyright (c) 2025 Justin Cranford
//
//

package magic

import (
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	// CurrentDirectory is fallback when project root not found.
	CurrentDirectory = "."

	// CICDExcludeDirGit is the name of the git metadata directory to exclude from scans.
	CICDExcludeDirGit = ".git"

	// CICDExcludeDirVendor is the name of the vendor directory to exclude from scans.
	CICDExcludeDirVendor = "vendor"
)

// getProjectRoot finds the project root by walking up the directory tree to find .git directory.
// Returns absolute path to project root, or "." as fallback if .git not found.
func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return CurrentDirectory // Fallback to current directory
	}

	// Walk up directory tree until .git found.
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir // Found project root
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding .git.
			return CurrentDirectory // Fallback to current directory
		}

		dir = parent
	}
}

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
	// SeparatorLength is the UI separator length constant.
	SeparatorLength = 50

	// MinActionMatchGroups is the minimum number of regex match groups for action parsing.
	MinActionMatchGroups = 3

	// PropagateMarkerMatchGroups is the expected submatch count for @propagate/@source regex (full match + 2 capture groups).
	PropagateMarkerMatchGroups = 3

	// CacheFilePermissions is the cache file permissions (owner read/write only).
	CacheFilePermissions = 0o600

	// CacheDuration is the duration after which cache entries expire.
	CacheDuration = 5 * time.Minute

	// CICDOutputDirPermissions is the directory permissions (owner read/write/execute, group and others execute).
	CICDOutputDirPermissions = 0o711

	// CICDOutputDir is the CICD output directory for all generated files, reports, and cache files.
	CICDOutputDir = ".cicd"

	// CICDOutputFilePermissions is the CICD output file permissions (owner read/write, group/others read).
	CICDOutputFilePermissions = 0o644

	// CICDIdentityProjectStatusPath is the identity project status path.
	CICDIdentityProjectStatusPath = "docs/02-identityV2/PROJECT-STATUS.md"
	// CICDIdentityRequirementsCoveragePath is the identity requirements coverage path.
	CICDIdentityRequirementsCoveragePath = "docs/02-identityV2/REQUIREMENTS-COVERAGE.md"
	// CICDIdentityTaskDocsDir is the identity task docs directory path.
	CICDIdentityTaskDocsDir = "docs/02-identityV2/passthru5/"

	// RequirementsTotalPatternGroups is the requirements coverage regex pattern groups count.
	RequirementsTotalPatternGroups = 4
	// RequirementsPriorityPatternGroups is the requirements priority pattern groups count.
	RequirementsPriorityPatternGroups = 4
	// RequirementsTaskCoveragePatternGroups is the requirements task coverage pattern groups count.
	RequirementsTaskCoveragePatternGroups = 3
	// RequirementsUncoveredPatternGroups is the requirements uncovered pattern groups count.
	RequirementsUncoveredPatternGroups = 2

	// TestCoveragePatternGroups is the test coverage regex pattern groups count.
	TestCoveragePatternGroups = 2

	// GitShortHashLength is the git short hash length.
	GitShortHashLength = 8

	// GitRecentActivityDays is the git recent activity days lookback.
	GitRecentActivityDays = 7

	// DateFormatYYYYMMDD is the date format for PROJECT-STATUS.md timestamps.
	DateFormatYYYYMMDD = "2006-01-02"

	// PercentMultiplier is the percent multiplier for coverage calculations.
	PercentMultiplier = 100.0

	// RequirementsProductionReadyThreshold is the production readiness threshold for requirements.
	RequirementsProductionReadyThreshold = 85.0
	// TestCoverageProductionReadyThreshold is the production readiness threshold for test coverage.
	TestCoverageProductionReadyThreshold = 85.0
	// RequirementsConditionalThreshold is the conditional threshold for requirements.
	RequirementsConditionalThreshold = 80.0
	// TestCoverageConditionalThreshold is the conditional threshold for test coverage.
	TestCoverageConditionalThreshold = 80.0
	// RequirementsTaskMinimumThreshold is the minimum threshold for requirements tasks.
	RequirementsTaskMinimumThreshold = 90.0

	// DepCacheFileName is the dependency cache file name.
	DepCacheFileName = ".cicd/dep-cache.json"

	// CircularDepCacheFileName is the circular dependency cache file name.
	CircularDepCacheFileName = ".cicd/circular-dep-cache.json"

	// ModeNameDirect is the dependency check mode name for direct dependencies.
	ModeNameDirect = "direct"
	// ModeNameAll is the dependency check mode name for all dependencies.
	ModeNameAll = "all"

        // SuiteServiceCount is the total number of individual services in the cryptoutil suite.
        // Services: sm-kms, sm-im, jose-ja, pki-ca, identity-authz, identity-idp,
        // identity-rp, identity-rs, identity-spa, skeleton-template.
        SuiteServiceCount = 10
)

// ListAllFilesStartDirectory is the ListAllFiles start directory.
// Uses absolute path to project root (found by walking up to .git directory).
// This ensures file paths are always project-root-relative regardless of working directory.
// Fixes path relativity issue where CLI (run from root) and tests (run from package dir)
// produced different path formats, causing exclusion pattern mismatches.
var ListAllFilesStartDirectory = getProjectRoot()

const (
	// TestCacheValidMinutes is the time constants for dependency cache testing.
	TestCacheValidMinutes = 30
	// TestCacheExpiredHours is the time constants for expired cache testing.
	TestCacheExpiredHours = 2

	// DepCacheValidDuration is the dependency cache validity duration.
	DepCacheValidDuration = 30 * time.Minute

	// CircularDepCacheValidDuration is the circular dependency cache validity duration.
	CircularDepCacheValidDuration = 60 * time.Minute

	// TestGitHubAPICacheExpiredHours is the time constants for GitHub API cache testing.
	TestGitHubAPICacheExpiredHours = 1

	// TimeFormat is the time format for logging and timestamps.
	TimeFormat = "2006-01-02T15:04:05.999999999Z07:00"

	// Utf8EnforceWorkerPoolSize is the number of worker threads for concurrent file processing operations.
	Utf8EnforceWorkerPoolSize = 6

	// UsageCICD is the usage message for the cicd command.
	UsageCICD = `Usage: cicd <command> [command...]

	Commands:
	  format-go       - [Formatter] Auto-fix Go files (any -> any, loop var copies)
	  format-go-test  - [Formatter] Auto-fix Go test files (add t.Helper() to helpers)
	  lint-compose    - [Linter] Detect admin port 9090 exposure in Docker Compose files
	  lint-go         - [Linter] Check for circular dependencies in Go packages
	  lint-go-mod     - [Linter] Check direct Go dependencies for updates
	  lint-go-test    - [Linter] Enforce test patterns (UUIDv7 usage, testify assertions)
	  lint-golangci   - [Linter] Validate golangci-lint config files for v2 compatibility
	  lint-ports      - [Linter] Enforce standardized port assignments (no legacy ports)
	  lint-text       - [Linter] Enforce UTF-8 encoding without BOM for text files
	  lint-workflow   - [Linter] Validate GitHub Actions workflow naming and versions`
)

// ValidCommands defines the set of valid cicd commands.
var ValidCommands = map[string]bool{
	"format-go":      true,
	"format-go-test": true,
	"lint-compose":   true,
	"lint-go":        true,
	"lint-go-mod":    true,
	"lint-go-test":   true,
	"lint-golangci":  true,
	"lint-ports":     true,
	"lint-text":      true,
	"lint-workflow":  true,
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
	// Keys match command names: lint-text, lint-go, lint-compose, format-go, lint-go-test, format-go-test, lint-golangci, lint-ports, lint-workflow, lint-go-mod.
	CICDSelfExclusionPatterns = map[string]string{
		"format-go":      `internal[/\\]apps[/\\]cicd[/\\]format_go[/\\].*\.go$`,
		"format-go-test": `internal[/\\]apps[/\\]cicd[/\\]format_gotest[/\\].*\.go$`,
		"lint-compose":   `internal[/\\]apps[/\\]cicd[/\\]lint_compose[/\\].*\.go$`,
		"lint-go":        `internal[/\\]apps[/\\]cicd[/\\]lint_go[/\\].*\.go$`,
		"lint-go-mod":    `internal[/\\]apps[/\\]cicd[/\\]lint_go_mod[/\\].*\.go$`,
		"lint-go-test":   `internal[/\\]apps[/\\]cicd[/\\]lint_gotest[/\\].*\.go$`,
		"lint-golangci":  `internal[/\\]apps[/\\]cicd[/\\]lint_golangci[/\\].*\.go$`,
		"lint-ports":     `internal[/\\]apps[/\\]cicd[/\\]lint_ports[/\\].*\.go$`,
		"lint-text":      `internal[/\\]apps[/\\]cicd[/\\]lint_text[/\\].*\.go$`,
		"lint-workflow":  `internal[/\\]apps[/\\]cicd[/\\]lint_workflow[/\\].*\.go$`,
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
