// Copyright (c) 2025-2026 Justin Cranford.
//

package cli

import (
	"context"
	"fmt"
	"io"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
)

// BindAddresser is implemented by all service settings types via the embedded
// *ServiceFrameworkServerSettings.GetBindAddresses() method.
// It allows NewServiceIdentity to derive bind addresses without a per-service lambda.
type BindAddresser interface {
	GetBindAddresses() (publicAddress string, publicPort uint16, adminAddress string, adminPort uint16)
}

// ServiceIdentity holds all PS-ID-specific constants needed by RouteServiceFromIdentity.
// Construct it with NewServiceIdentity, which wires the type-safe ServerFn automatically.
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
	// ServerFn is the "server" subcommand handler. Set by NewServiceIdentity.
	ServerFn SubcommandFunc
}

// NewServiceIdentity creates a ServiceIdentity with a fully wired ServerFn.
// The ServerFn calls StartServiceServer using parseConfig and newServer, deriving
// bind addresses via the BindAddresser interface (all service settings embed
// *ServiceFrameworkServerSettings which implements BindAddresser).
//
// Type parameters are inferred from the function arguments:
//   - S: the service settings type (must embed *ServiceFrameworkServerSettings)
//   - R: the concrete server type returned by newServer (must implement ReadyStarter)
func NewServiceIdentity[S BindAddresser, R ReadyStarter](
	serviceID, productName, serviceName, displayName string,
	servicePort uint16,
	parseConfig ParseWithFlagSetFunc[S],
	newServer func(ctx context.Context, settings S) (R, error),
) ServiceIdentity {
	id := ServiceIdentity{
		ServiceID:   serviceID,
		ProductName: productName,
		ServiceName: serviceName,
		DisplayName: displayName,
		ServicePort: servicePort,
	}

	id.ServerFn = func(serverArgs []string, serverStdout, serverStderr io.Writer) int {
		return StartServiceServer(serverArgs, serverStdout, serverStderr, ServerStartOptions[S]{
			UsageServer:  BuildServerUsage(id),
			ServiceLabel: id.ServiceID,
			FlagSetName:  ServerFlagSetName(id.ServiceID),
			ParseConfig:  parseConfig,
			NewServer: func(ctx context.Context, settings S) (ReadyStarter, error) {
				return newServer(ctx, settings)
			},
			BindAddresses: func(settings S) (string, uint16, string, uint16) {
				return settings.GetBindAddresses()
			},
		})
	}

	return id
}

// BuildServerUsage returns the server subcommand usage string for the given ServiceIdentity.
// The config file path follows the convention: configs/{ServiceID}/{ServiceID}-framework.yml.
func BuildServerUsage(id ServiceIdentity) string {
	configFilePath := fmt.Sprintf("configs/%s/%s-framework.yml", id.ServiceID, id.ServiceID)

	return cryptoutilUsage.BuildUsageServer(id.ProductName, id.ServiceName, id.DisplayName, configFilePath)
}

// RouteServiceFromIdentity is the single-call entry point for all PS-ID service CLIs.
// It builds all 8 usage strings from the service identity, provides standard client and
// init subcommand handlers, and dispatches the "server" subcommand to id.ServerFn.
// Construct id using NewServiceIdentity to wire all service-specific settings automatically.
func RouteServiceFromIdentity(id ServiceIdentity, args []string, stdout, stderr io.Writer) int {
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
		id.ServerFn, clientFn, initFn,
	)
}
