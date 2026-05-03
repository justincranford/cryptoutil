// Copyright (c) 2025-2026 Justin Cranford.
//

package cli

import (
	"context"
	"fmt"
	"io"

	cryptoutilAppsFrameworkServiceLifecycle "cryptoutil/internal/apps-framework/service/lifecycle"

	"github.com/spf13/pflag"
)

// ReadyStarter is the server contract used by StartServiceServer.
// Service server implementations expose SetReady plus lifecycle methods.
type ReadyStarter interface {
	SetReady(ready bool)
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// ParseWithFlagSetFunc parses CLI args into service settings.
type ParseWithFlagSetFunc[S any] func(fs *pflag.FlagSet, args []string, exitIfHelp bool) (S, error)

// NewServerFromConfigFunc constructs a service server from parsed settings.
type NewServerFromConfigFunc[S any] func(ctx context.Context, settings S) (ReadyStarter, error)

// BindAddressesFunc extracts public/admin bind addresses and ports from settings.
type BindAddressesFunc[S any] func(settings S) (publicAddress string, publicPort uint16, adminAddress string, adminPort uint16)

// ServerStartOptions configures StartServiceServer behavior for a product-service CLI.
type ServerStartOptions[S any] struct {
	UsageServer   string
	ServiceLabel  string
	FlagSetName   string
	ParseConfig   ParseWithFlagSetFunc[S]
	NewServer     NewServerFromConfigFunc[S]
	BindAddresses BindAddressesFunc[S]
}

// StartServiceServer implements the common "server" subcommand flow for all product-services.
func StartServiceServer[S any](args []string, stdout, stderr io.Writer, opts ServerStartOptions[S]) int {
	if IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, opts.UsageServer)

		return 0
	}

	ctx := context.Background()

	argsWithSubcommand := append([]string{"start"}, args...)

	fs := pflag.NewFlagSet(opts.FlagSetName, pflag.ContinueOnError)

	settings, err := opts.ParseConfig(fs, argsWithSubcommand, true)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to parse configuration: %v\n", err)

		return 1
	}

	srv, err := opts.NewServer(ctx, settings)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ Failed to create server: %v\n", err)

		return 1
	}

	srv.SetReady(true)

	publicAddress, publicPort, adminAddress, adminPort := opts.BindAddresses(settings)

	_, _ = fmt.Fprintf(stdout, "🚀 Starting %s service...\n", opts.ServiceLabel)
	_, _ = fmt.Fprintf(stdout, "   Public Server: https://%s:%d\n", publicAddress, publicPort)
	_, _ = fmt.Fprintf(stdout, "   Admin Server:  https://%s:%d\n", adminAddress, adminPort)

	exitCode := cryptoutilAppsFrameworkServiceLifecycle.RunService(ctx, stdout, stderr, srv)

	_, _ = fmt.Fprintf(stdout, "✅ %s service stopped\n", opts.ServiceLabel)

	return exitCode
}
