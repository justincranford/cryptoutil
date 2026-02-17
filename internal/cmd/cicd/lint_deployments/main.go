package lint_deployments

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultDeploymentsDir = "deployments"
	defaultConfigsDir     = "configs"
)

// Main is the CLI entry point for the deployment linter.
// Accepts optional base directory argument for testing.
func Main(args []string) int {
	// Check for subcommands first.
	if len(args) > 0 {
		switch args[0] {
		case "generate-listings":
			return mainGenerateListings(args[1:])
		case "validate-mirror":
			return mainValidateMirror(args[1:])
		case "validate-compose":
			return mainValidateCompose(args[1:])
		}
	}

	baseDir := defaultDeploymentsDir

	// Allow injecting custom directory for testing.
	if len(args) > 0 && args[0] != "" {
		baseDir = args[0]
	}

	// Validate directory exists.
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Directory does not exist: %s\n", baseDir)

		return 1
	}

	// Run validation.
	results, err := ValidateAllDeployments(baseDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Validation failed: %v\n", err)

		return 1
	}

	// Format and print results.
	output := FormatResults(results)
	fmt.Print(output)

	// Return non-zero if any validation failed.
	for _, r := range results {
		if !r.Valid {
			return 1
		}
	}

	return 0
}

// mainGenerateListings handles the generate-listings subcommand.
func mainGenerateListings(args []string) int {
	deploymentsDir := defaultDeploymentsDir
	configsDir := defaultConfigsDir

	if len(args) >= 2 {
		deploymentsDir = args[0]
		configsDir = args[1]
	}

	deploymentsOutput := filepath.Join(deploymentsDir, "deployments_all_files.json")
	configsOutput := filepath.Join(configsDir, "configs_all_files.json")

	// Generate deployments listing.
	if err := WriteListingFile(deploymentsDir, deploymentsOutput); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Failed to generate deployments listing: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprintf(os.Stdout, "Generated: %s\n", deploymentsOutput)

	// Generate configs listing.
	if err := WriteListingFile(configsDir, configsOutput); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Failed to generate configs listing: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprintf(os.Stdout, "Generated: %s\n", configsOutput)

	return 0
}

// mainValidateMirror handles the validate-mirror subcommand.
func mainValidateMirror(args []string) int {
	deploymentsDir := defaultDeploymentsDir
	configsDir := defaultConfigsDir

	if len(args) >= 2 {
		deploymentsDir = args[0]
		configsDir = args[1]
	}

	result, err := ValidateStructuralMirror(deploymentsDir, configsDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Mirror validation failed: %v\n", err)

		return 1
	}

	fmt.Print(FormatMirrorResult(result))

	if !result.Valid {
		return 1
	}

	return 0
}

// mainValidateCompose handles the validate-compose subcommand.
func mainValidateCompose(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: validate-compose requires a compose file path\n")

		return 1
	}

	composePath := args[0]

	composeResult, err := ValidateComposeFile(composePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: Compose validation failed: %v\n", err)

		return 1
	}

	fmt.Print(FormatComposeValidationResult(composeResult))

	if !composeResult.Valid {
		return 1
	}

	return 0
}
