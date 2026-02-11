// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides the unified demo CLI implementation for cryptoutil.
// It supports subcommands for KMS, Identity, and integrated demos with
// structured error handling, progress display, and configurable output formats.
//
// Usage:
//
//	demo kms       - Run KMS demo
//	demo identity  - Run Identity demo
//	demo all       - Run full integration demo
//	demo help      - Show help
//
// Reference: Session 3 Q11-15, Session 5 Q1, Q15.
package demo

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Exit codes following Session 5 Q1 decision: simple 0/1/2 pattern.
const (
	// ExitSuccess indicates all demo steps completed successfully.
	ExitSuccess = 0
	// ExitPartialFailure indicates some demo steps failed but execution continued.
	ExitPartialFailure = 1
	// ExitFailure indicates a critical failure that stopped execution.
	ExitFailure = 2
)

// OutputFormat specifies the output format for demo results.
type OutputFormat string

const (
	// OutputHuman produces human-readable output with colors and emojis.
	OutputHuman OutputFormat = "human"
	// OutputJSON produces machine-readable JSON output.
	OutputJSON OutputFormat = "json"
	// OutputStructured produces structured log output.
	OutputStructured OutputFormat = "structured"
)

// DefaultHealthTimeout is the default timeout for health checks (Session 3 Q14).
var DefaultHealthTimeout = cryptoutilSharedMagic.DefaultDemoTimeout

// Config holds the demo CLI configuration.
type Config struct {
	// OutputFormat specifies the output format (human, json, structured).
	OutputFormat OutputFormat

	// HealthTimeout is the timeout for health check waiting.
	HealthTimeout time.Duration

	// ContinueOnError specifies whether to continue after errors.
	ContinueOnError bool

	// RetryCount is the number of retries for failed operations.
	RetryCount int

	// RetryDelay is the delay between retries.
	RetryDelay time.Duration

	// Verbose enables verbose output.
	Verbose bool

	// NoColor disables colored output (Session 5 Q15).
	NoColor bool

	// Quiet suppresses non-essential output.
	Quiet bool
}

// DefaultConfig returns the default demo configuration.
func DefaultConfig() *Config {
	return &Config{
		OutputFormat:    OutputHuman,
		HealthTimeout:   DefaultHealthTimeout,
		ContinueOnError: true,
		RetryCount:      cryptoutilSharedMagic.DefaultDemoRetryCount,
		RetryDelay:      cryptoutilSharedMagic.DefaultDemoRetryDelay,
		Verbose:         false,
		NoColor:         detectNoColor(),
		Quiet:           false,
	}
}

// detectNoColor checks if color should be disabled (CI environment, NO_COLOR env var).
func detectNoColor() bool {
	// Check NO_COLOR environment variable (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}

	// Check common CI environment variables
	ciEnvVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "TRAVIS"}
	for _, envVar := range ciEnvVars {
		if _, ok := os.LookupEnv(envVar); ok {
			return true
		}
	}

	return false
}

// Demo runs the demo CLI with command-line arguments.
func Demo(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage()

		return ExitFailure
	}

	command := args[1]
	cmdArgs := args[2:]
	config := parseArgs(cmdArgs)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDemoTimeout)
	defer cancel()

	exitCode := ExitSuccess

	switch command {
	case "kms":
		exitCode = runKMSDemo(ctx, config)
	case "identity":
		exitCode = runIdentityDemo(ctx, config)
	case "jose":
		exitCode = runJOSEDemo(ctx, config)
	case "ca":
		exitCode = runCADemo(ctx, config)
	case "all":
		exitCode = runIntegrationDemo(ctx, config)
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		printVersion()
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)

		printUsage()

		exitCode = ExitFailure
	}

	return exitCode
}

// parseArgs parses command-line arguments into a Config.
func parseArgs(args []string) *Config {
	config := DefaultConfig()

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--output", "-o":
			if i+1 < len(args) {
				i++

				switch args[i] {
				case "human":
					config.OutputFormat = OutputHuman
				case "json":
					config.OutputFormat = OutputJSON
				case "structured":
					config.OutputFormat = OutputStructured
				}
			}
		case "--no-color":
			config.NoColor = true
		case "--verbose":
			config.Verbose = true
		case "--quiet", "-q":
			config.Quiet = true
		case "--continue-on-error":
			config.ContinueOnError = true
		case "--fail-fast":
			config.ContinueOnError = false
		case "--health-timeout":
			if i+1 < len(args) {
				i++

				if d, err := time.ParseDuration(args[i]); err == nil {
					config.HealthTimeout = d
				}
			}
		case "--retry":
			if i+1 < len(args) {
				i++
				// Attempt to parse retry count; ignore invalid values.
				_, _ = fmt.Sscanf(args[i], "%d", &config.RetryCount)
			}
		}
	}

	return config
}

// printUsage prints the CLI usage information.
func printUsage() {
	fmt.Println("Usage: demo <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  kms        Run KMS demo (key pools, encryption, signing)")
	fmt.Println("  identity   Run Identity demo (OAuth 2.1, tokens)")
	fmt.Println("  jose       Run JOSE Authority demo (JWK, JWS, JWE, JWT)")
	fmt.Println("  ca         Run CA demo (certificates, revocation)")
	fmt.Println("  all        Run full integration demo (KMS + Identity)")
	fmt.Println("  help       Show this help message")
	fmt.Println("  version    Show version information")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --output, -o <format>  Output format: human, json, structured (default: human)")
	fmt.Println("  --no-color             Disable colored output")
	fmt.Println("  --verbose              Enable verbose output")
	fmt.Println("  --quiet, -q            Suppress non-essential output")
	fmt.Println("  --continue-on-error    Continue after errors (default)")
	fmt.Println("  --fail-fast            Stop on first error")
	fmt.Println("  --health-timeout <dur> Health check timeout (default: 30s)")
	fmt.Println("  --retry <n>            Number of retries for failed operations (default: 3)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  demo kms                          # Run KMS demo with default settings")
	fmt.Println("  demo identity --output json       # Run Identity demo with JSON output")
	fmt.Println("  demo jose --verbose               # Run JOSE Authority demo with verbose logging")
	fmt.Println("  demo ca --fail-fast               # Run CA demo, stop on first error")
	fmt.Println("  demo all --verbose --no-color     # Run full demo with verbose logging")
	fmt.Println()
}

// printVersion prints version information.
func printVersion() {
	fmt.Println("cryptoutil demo CLI")
	fmt.Println("Version: 0.1.0")
	fmt.Println("Reference: passthru2 Session 3 Q11")
}
