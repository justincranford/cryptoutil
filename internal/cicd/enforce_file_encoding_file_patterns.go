// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains file pattern definitions used by the CI/CD checks.
// These patterns define which files should be included or excluded from various checks.
package cicd

// File patterns for encoding checks (include patterns).
var enforceFileEncodingFileIncludePatterns = []string{
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

// Exclusion patterns for file processing (exclude patterns).
var enforceFileEncodingFileExcludePatterns = []string{
	`_gen\.go$`,     // Generated files
	`\.pb\.go$`,     // Protocol buffer files
	`vendor/`,       // Vendored dependencies
	`api/client`,    // Generated API client
	`api/model`,     // Generated API models
	`api/server`,    // Generated API server
	`.git/`,         // Git directory
	`node_modules/`, // Node.js dependencies
	// NOTE: cicd_checks.go and cicd_checks_test.go are intentionally NOT excluded from enforce-file-encoding
	// as these files should validate their own UTF-8 encoding compliance
}
