// Copyright (c) 2025 Justin Cranford
//

package cli

import (
	"fmt"
	"io"
)

// ServiceConfig holds configuration for a service CLI entrypoint.
// All SERVICE CLI entrypoints (product-service combinations) use this.
type ServiceConfig struct {
	// ServiceID is the combined product-service identifier (e.g., "sm-im").
	ServiceID string
	// ProductName is the product name (e.g., "sm").
	ProductName string
	// ServiceName is the service name within the product (e.g., "im").
	ServiceName string
	// DefaultPublicPort is the default public server port for health checks (e.g., 8700).
	DefaultPublicPort uint16
	// UsageMain is the usage text for the main service command (shown for --help / no args).
	UsageMain string
	// UsageServer is the usage text for the server subcommand.
	UsageServer string
	// UsageClient is the usage text for the client subcommand.
	UsageClient string
	// UsageInit is the usage text for the init subcommand.
	UsageInit string
	// UsageHealth is the usage text for the health subcommand.
	UsageHealth string
	// UsageLivez is the usage text for the livez subcommand.
	UsageLivez string
	// UsageReadyz is the usage text for the readyz subcommand.
	UsageReadyz string
	// UsageShutdown is the usage text for the shutdown subcommand.
	UsageShutdown string
}

// SubcommandFunc is a function that handles a CLI subcommand.
// It receives the remaining args after the subcommand, and returns an exit code.
type SubcommandFunc func(args []string, stdout, stderr io.Writer) int

// RouteService implements the standard service command router.
// It handles version/help flags and routes to the standard subcommands.
//
// Mandatory subcommands (all services MUST support):
//   - version:  Print version information.
//   - server:   Start the service server (via serverFn).
//   - client:   Run client operations (via clientFn).
//   - init:     Initialize database and configuration (via initFn).
//   - health:   Check service health via public API (template-provided).
//   - livez:    Check service liveness via admin API (template-provided).
//   - readyz:   Check service readiness via admin API (template-provided).
//   - shutdown: Trigger graceful shutdown via admin API (template-provided).
//
// The serverFn, clientFn, and initFn are service-specific implementations.
// The health/livez/readyz/shutdown commands are provided by the template.
func RouteService(cfg ServiceConfig, args []string, stdout, stderr io.Writer, serverFn, clientFn, initFn SubcommandFunc) int {
	// Default to "server" subcommand if no args provided (backward compatibility).
	if len(args) == 0 {
		args = []string{"server"}
	}

	// Check for help flags.
	if args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag {
		_, _ = fmt.Fprintln(stdout, cfg.UsageMain)

		return 0
	}

	// Route to subcommand.
	switch args[0] {
	case versionCommand:
		_, _ = fmt.Fprintf(stdout, "%s service\n", cfg.ServiceID)
		_, _ = fmt.Fprintf(stdout, "Part of cryptoutil %s product\n", cfg.ProductName)
		_, _ = fmt.Fprintln(stdout, "Version information available via Docker image tags")

		return 0
	case "server":
		return serverFn(args[1:], stdout, stderr)
	case "client":
		return clientFn(args[1:], stdout, stderr)
	case "init":
		return initFn(args[1:], stdout, stderr)
	case "health":
		return HealthCommand(args[1:], stdout, stderr, cfg.UsageHealth, cfg.DefaultPublicPort)
	case "livez":
		return LivezCommand(args[1:], stdout, stderr, cfg.UsageLivez)
	case "readyz":
		return ReadyzCommand(args[1:], stdout, stderr, cfg.UsageReadyz)
	case "shutdown":
		return ShutdownCommand(args[1:], stdout, stderr, cfg.UsageShutdown)
	default:
		_, _ = fmt.Fprintf(stderr, "Unknown subcommand: %s\n\n", args[0])

		_, _ = fmt.Fprintln(stdout, cfg.UsageMain)

		return 1
	}
}
