// Copyright (c) 2025-2026 Justin Cranford.
//

package cli

import (
	"fmt"
	"io"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
)

// ServiceIdentity holds the minimal service-identity constants for a PS-ID CLI entry point.
// All 10 PS-ID services pass these five magic constants to RouteServiceFromIdentity, which
// derives all usage strings and standard subcommand handlers internally.
type ServiceIdentity struct {
	// ServiceID is the combined product-service identifier (e.g., "sm-kms").
	ServiceID string
	// ProductName is the product name (e.g., "sm").
	ProductName string
	// ServiceName is the service name within the product (e.g., "kms").
	ServiceName string
	// DisplayName is the human-readable service name (e.g., "Key Management Service").
	DisplayName string
	// ServicePort is the default public service port for health checks (e.g., 8000).
	ServicePort uint16
}

// BuildServerUsage returns the server subcommand usage string for the given ServiceIdentity.
// The config file path follows the convention: configs/{ServiceID}/{ServiceID}-framework.yml.
func BuildServerUsage(id ServiceIdentity) string {
	configFilePath := fmt.Sprintf("configs/%s/%s-framework.yml", id.ServiceID, id.ServiceID)

	return cryptoutilUsage.BuildUsageServer(id.ProductName, id.ServiceName, id.DisplayName, configFilePath)
}

// RouteServiceFromIdentity is the single-call entry point for all PS-ID service CLIs.
// It builds all 8 usage strings from the service identity, provides standard client and
// init subcommand handlers, and delegates only the service-specific server subcommand to
// serverFn. serverFn should call StartServiceServer with the service-specific settings type.
func RouteServiceFromIdentity(id ServiceIdentity, args []string, stdout, stderr io.Writer, serverFn SubcommandFunc) int {
	configFilePath := fmt.Sprintf("configs/%s/%s-framework.yml", id.ServiceID, id.ServiceID)
	portStr := fmt.Sprintf("%d", id.ServicePort)

	usageMain := cryptoutilUsage.BuildUsageMain(id.ProductName, id.ServiceName, id.DisplayName)
	usageServer := cryptoutilUsage.BuildUsageServer(id.ProductName, id.ServiceName, id.DisplayName, configFilePath)
	usageClient := cryptoutilUsage.BuildUsageClient(id.ProductName, id.ServiceName, id.DisplayName)
	usageInit := cryptoutilUsage.BuildUsageInit(id.ProductName, id.ServiceName, id.DisplayName, configFilePath)
	usageHealth := cryptoutilUsage.BuildUsageHealth(id.ProductName, id.ServiceName, portStr)
	usageLivez := cryptoutilUsage.BuildUsageLivez(id.ProductName, id.ServiceName)
	usageReadyz := cryptoutilUsage.BuildUsageReadyz(id.ProductName, id.ServiceName)
	usageShutdown := cryptoutilUsage.BuildUsageShutdown(id.ProductName, id.ServiceName)

	clientFn := func(clientArgs []string, _ io.Writer, clientStderr io.Writer) int {
		if IsHelpRequest(clientArgs, ClientNotImplementedMessageConfig{Stderr: clientStderr, ServiceID: id.ServiceID, UsageText: usageClient}) {
			return 0
		}

		return 1
	}

	initFn := func(initArgs []string, initStdout, initStderr io.Writer) int {
		if IsHelpRequest(initArgs, ClientNotImplementedMessageConfig{Stderr: initStderr, UsageText: usageInit}) {
			return 0
		}

		return cryptoutilAppsFrameworkTls.InitForService(id.ServiceID, initArgs, initStdout, initStderr)
	}

	return RouteService(
		ServiceConfig{
			ServiceID:         id.ServiceID,
			ProductName:       id.ProductName,
			ServiceName:       id.ServiceName,
			DefaultPublicPort: id.ServicePort,
			UsageMain:         usageMain,
			UsageServer:       usageServer,
			UsageClient:       usageClient,
			UsageInit:         usageInit,
			UsageHealth:       usageHealth,
			UsageLivez:        usageLivez,
			UsageReadyz:       usageReadyz,
			UsageShutdown:     usageShutdown,
		},
		args, stdout, stderr,
		serverFn, clientFn, initFn,
	)
}
