// Copyright (c) 2025 Justin Cranford
//
//

// Package tls provides TLS certificate initialization for the framework.
// This package is used by Docker Compose E2E deployments to generate TLS
// certificate hierarchies into a shared volume, enabling proper TLS
// verification in tests and deployments.
package tls

import (
	"context"
	"fmt"
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// Init executes the pki-init CLI command.
// It expects exactly 2 positional args: tierID and targetDir.
func Init(args []string, _ io.Reader, stdout io.Writer, stderr io.Writer) int {
	return initRun(args, stdout, stderr, productionNewTelemetryService, productionNewGenerator)
}

// InitForSuite executes the init subcommand for a named suite.
func InitForSuite(_ string, args []string, stdout, stderr io.Writer) int {
	return initRun(args, stdout, stderr, productionNewTelemetryService, productionNewGenerator)
}

// InitForProduct executes the init subcommand for a named product.
func InitForProduct(_ string, args []string, stdout, stderr io.Writer) int {
	return initRun(args, stdout, stderr, productionNewTelemetryService, productionNewGenerator)
}

// InitForService executes the init subcommand for a named PS-ID service.
func InitForService(_ string, args []string, stdout, stderr io.Writer) int {
	return initRun(args, stdout, stderr, productionNewTelemetryService, productionNewGenerator)
}

// productionNewTelemetryService creates the real telemetry service for production use.
func productionNewTelemetryService(ctx context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error) {
	ts, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, &cryptoutilSharedTelemetry.TelemetrySettings{
		OTLPService: cryptoutilSharedMagic.PSIDPKIInit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry service: %w", err)
	}

	return ts, nil
}

// productionNewGenerator creates the real generator for production use.
func productionNewGenerator(ctx context.Context, ts *cryptoutilSharedTelemetry.TelemetryService) (*Generator, error) {
	return NewGenerator(ctx, ts)
}

// initRun is the shared implementation for all Init* functions.
func initRun(
	args []string,
	stdout, stderr io.Writer,
	telemetryFn func(context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error),
	generatorFn func(context.Context, *cryptoutilSharedTelemetry.TelemetryService) (*Generator, error),
) int {
	expectedArgCount := 2
	if len(args) != expectedArgCount {
		_, _ = fmt.Fprintf(stderr, "pki-init: usage: pki-init <tier-id> <target-dir>\n")

		return 1
	}

	tierID := args[0]
	targetDir := args[1]

	ctx := context.Background()

	ts, err := telemetryFn(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: failed to create telemetry service: %v\n", err)

		return 1
	}

	defer ts.Shutdown()

	gen, err := generatorFn(ctx, ts)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: failed to create generator: %v\n", err)

		return 1
	}

	defer gen.Shutdown()

	if err := gen.Generate(tierID, targetDir); err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprintf(stdout, "pki-init: certificates written to %q for tier %q\n", targetDir, tierID)

	return 0
}
