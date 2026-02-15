package lint_deployments

import (
"fmt"
"os"
)

// Main is the CLI entry point for the deployment linter.
// Accepts optional base directory argument for testing.
func Main(args []string) int {
baseDir := "deployments"

// Allow injecting custom directory for testing
if len(args) > 0 && args[0] != "" {
baseDir = args[0]
}

// Validate directory exists
if _, err := os.Stat(baseDir); os.IsNotExist(err) {
fmt.Fprintf(os.Stderr, "ERROR: Directory does not exist: %s\n", baseDir)
return 1
}

// Run validation
results, err := ValidateAllDeployments(baseDir)
if err != nil {
fmt.Fprintf(os.Stderr, "ERROR: Validation failed: %v\n", err)
return 1
}

// Format and print results
output := FormatResults(results)
fmt.Print(output)

// Return non-zero if any validation failed
for _, r := range results {
if !r.Valid {
return 1
}
}

return 0
}
