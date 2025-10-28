package cicd

// to prevent self-modification and preserve deliberate test patterns.
var goEnforceAnyFileExcludePatterns = []string{
	`internal[/\\]cmd[/\\]cicd[/\\]cicd\.go$`,      // Exclude this file itself to avoid replacing the regex pattern
	`internal[/\\]cmd[/\\]cicd[/\\]cicd_test\.go$`, // Exclude test file to preserve deliberate bad patterns for testing
	`api/client`,    // Generated API client
	`api/model`,     // Generated API models
	`api/server`,    // Generated API server
	`_gen\.go$`,     // Generated files
	`\.pb\.go$`,     // Protocol buffer files
	`vendor/`,       // Vendored dependencies
	`.git/`,         // Git directory
	`node_modules/`, // Node.js dependencies
}
