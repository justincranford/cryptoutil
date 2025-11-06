// Package magic provides magic values and constants for the cryptoutil project.
//
// This file contains constants related to CI/CD operations.
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

// Regex patterns for test enforcement.

const (
	// UI constants.
	SeparatorLength = 50

	// Minimum number of regex match groups for action parsing.
	MinActionMatchGroups = 3

	// Cache file permissions (owner read/write only).
	CacheFilePermissions = 0o600

	// Dependency cache file name.
	DepCacheFileName = ".cicd-dep-cache.json"

	// Dependency check mode names.
	ModeNameDirect = "direct"
	ModeNameAll    = "all"

	// Time format for logging and timestamps.
	TimeFormat = "2006-01-02T15:04:05.999999999Z07:00"

	// Number of worker threads for concurrent file processing operations.
	Utf8EnforceWorkerPoolSize = 4

	// Usage message for the cicd command.
	UsageCICD = `Usage: cicd <command> [command...]

	Commands:
	  all-enforce-utf8                       - Enforce UTF-8 encoding without BOM
	  go-enforce-test-patterns               - Enforce test patterns (UUIDv7 usage, testify assertions)
	  go-enforce-any                         - Custom Go source code fixes (any -> any, etc.)
	  go-check-circular-package-dependencies - Check for circular dependencies in Go packages
	  go-update-direct-dependencies          - Check direct Go dependencies only
	  go-update-all-dependencies             - Check all Go dependencies (direct + transitive)
	  github-workflow-lint                   - Validate GitHub Actions workflow naming and structure, and check for outdated actions`
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
}

// Regex patterns for test enforcement.
var (
	// TestErrorfPattern - Regex pattern to match test.Errorf calls.
	TestErrorfPattern = regexp.MustCompile(`\bt\.Errorf\s*\(`)
	// TestFatalfPattern - Regex pattern to match test.Fatalf calls.
	TestFatalfPattern = regexp.MustCompile(`\bt\.Fatalf\s*\(`)
)

// File patterns for CI/CD enforcement commands.
var (
	// GoEnforceAnyFileExcludePatterns - Files excluded from go-enforce-any command to prevent self-modification.
	GoEnforceAnyFileExcludePatterns = []string{
		`internal[/\\]cmd[/\\]cicd[/\\]cicd_enforce_any\.go$`,          // Exclude this file itself to prevent self-modification
		`internal[/\\]cmd[/\\]cicd[/\\]cicd_enforce_any_test\.go$`,     // Exclude test file to preserve deliberate test patterns
		`internal[/\\]cmd[/\\]cicd[/\\]file_patterns_enforce_any\.go$`, // Exclude pattern definitions to prevent self-modification
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

	// EnforceUtf8FileExcludePatterns - File patterns to exclude from UTF-8 encoding checks.
	EnforceUtf8FileExcludePatterns = []string{
		`_gen\.go$`,     // Generated files
		`\.pb\.go$`,     // Protocol buffer files
		`vendor/`,       // Vendored dependencies
		`api/client`,    // Generated API client
		`api/model`,     // Generated API models
		`api/server`,    // Generated API server
		`.git/`,         // Git directory
		`node_modules/`, // Node.js dependencies
		// NOTE: cicd_checks.go and cicd_checks_test.go are intentionally NOT excluded from all-enforce-utf8
		// as these files should validate their own UTF-8 encoding compliance
	}
)
