// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package authz provides the unified command interface for Identity Authorization Server service.
package authz

import (
	"context"
	"fmt"
	"io"
	"log"
	http "net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cryptoutilAppsIdentityAuthzServer "cryptoutil/internal/apps/identity-authz/server"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	configFlagShort  = "-c"
	valueStop        = "stop"
	valueHealth      = "health"
	defaultAdminPort = 9090
	httpTimeout      = 5 * time.Second
)

var osExit = os.Exit

type serviceLifecycle interface {
	SetReady(bool)
	Start(context.Context) error
}

type requestCreator func(context.Context, string, string, io.Reader) (*http.Request, error)

type requestDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type serviceRequest struct {
	progressMessage       string
	method                string
	path                  string
	requestFailureMessage string
	doFailureMessage      string
	successMessage        string
	statusFailureMessage  string
	contentType           string
	expectedStatus        int
}

type commandHandlers struct {
	start  func([]string)
	stop   func([]string)
	status func([]string)
	health func([]string)
	exit   func(int)
}

// Execute handles Authz service commands matching other service patterns.
// Supports: start, stop, status, health.
func Execute(parameters []string) {
	execute(parameters, commandHandlers{
		start: func(parameters []string) {
			startService(
				parameters,
				cryptoutilAppsIdentityAuthzServerConfig.Parse,
				func(ctx context.Context, settings *cryptoutilAppsIdentityAuthzServerConfig.IdentityAuthzServerSettings) (serviceLifecycle, error) {
					return cryptoutilAppsIdentityAuthzServer.NewFromConfig(ctx, settings)
				},
				log.Fatalf,
			)
		},
		stop:   stopService,
		status: statusService,
		health: healthService,
		exit:   osExit,
	})
}

func execute(parameters []string, handlers commandHandlers) {
	if len(parameters) < 1 {
		printUsage()
		handlers.exit(1)

		return
	}

	subcommand := parameters[0]
	cmdParams := parameters[1:]

	switch subcommand {
	case valueStart:
		handlers.start(cmdParams)
	case valueStop:
		handlers.stop(cmdParams)
	case cryptoutilSharedMagic.StringStatus:
		handlers.status(cmdParams)
	case valueHealth:
		handlers.health(cmdParams)
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()
		handlers.exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cryptoutil identity-authz <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  start    - Start Identity Authorization Server")
	fmt.Println("  stop     - Stop Identity Authorization Server")
	fmt.Println("  status   - Check service status")
	fmt.Println("  health   - Check service health")
	fmt.Println("\nOptions:")
	fmt.Println("  --config, -c <path>  - Config file path (default: /app/run/identity-authz-docker.yml)")
}

// startService starts the Identity Authorization Server.
func startService[T any](
	parameters []string,
	parseConfig func([]string, bool) (*T, error),
	newServer func(context.Context, *T) (serviceLifecycle, error),
	fatalf func(string, ...any),
) {
	configFile := parseConfigFlag(parameters, "/app/run/identity-authz-docker.yml")

	fmt.Fprintf(os.Stderr, "Starting Identity Authorization Server\n")
	fmt.Fprintf(os.Stderr, "Using config file: %s\n", configFile)

	// Load Authz-specific configuration from YAML file.
	parseArgs := []string{valueStart, cryptoutilSharedMagic.IdentityCLIFlagConfig, configFile}

	settings, err := parseConfig(parseArgs, false)
	if err != nil {
		fatalf("Failed to load config from %s: %v", configFile, err)

		return
	}

	// Create context.
	ctx := context.Background()

	// Create Authz server using template-based implementation.
	server, err := newServer(ctx, settings)
	if err != nil {
		fatalf("Failed to create Authz server: %v", err)

		return
	}

	// Mark server as ready.
	server.SetReady(true)

	fmt.Fprintf(os.Stderr, "Starting Authz server...\n")

	// Start server (blocks until shutdown).
	if err := server.Start(ctx); err != nil {
		fatalf("Authz server error: %v", err)
	}
}

// stopService sends shutdown request to Authz admin endpoint.
func stopService(parameters []string) {
	sendServiceRequest(parameters, serviceRequest{
		progressMessage:       "Sending shutdown request to admin endpoint (port %d)...\n",
		method:                http.MethodPost,
		path:                  "/admin/v1/shutdown",
		requestFailureMessage: cryptoutilSharedMagic.RequestCreationFailureMessage,
		doFailureMessage:      "Failed to send shutdown request",
		successMessage:        "Shutdown request sent successfully",
		statusFailureMessage:  "Shutdown request failed with status: %d\n",
		contentType:           "application/json",
		expectedStatus:        http.StatusOK,
	}, http.NewRequestWithContext, &http.Client{Timeout: httpTimeout}, osExit)
}

// statusService checks Authz service readiness.
func statusService(parameters []string) {
	sendServiceRequest(parameters, serviceRequest{
		progressMessage:       "Checking service status (port %d)...\n",
		method:                http.MethodGet,
		path:                  "/admin/v1/readyz",
		requestFailureMessage: cryptoutilSharedMagic.RequestCreationFailureMessage,
		doFailureMessage:      "Failed to check service status",
		successMessage:        "Service is ready",
		statusFailureMessage:  "Service is not ready (status: %d)\n",
		expectedStatus:        http.StatusOK,
	}, http.NewRequestWithContext, &http.Client{Timeout: httpTimeout}, osExit)
}

// healthService checks Authz service health.
func healthService(parameters []string) {
	sendServiceRequest(parameters, serviceRequest{
		progressMessage:       "Checking service health (port %d)...\n",
		method:                http.MethodGet,
		path:                  "/admin/v1/livez",
		requestFailureMessage: cryptoutilSharedMagic.RequestCreationFailureMessage,
		doFailureMessage:      "Failed to check service health",
		successMessage:        "Service is healthy",
		statusFailureMessage:  "Service is not healthy (status: %d)\n",
		expectedStatus:        http.StatusOK,
	}, http.NewRequestWithContext, &http.Client{Timeout: httpTimeout}, osExit)
}

func sendServiceRequest(
	parameters []string,
	request serviceRequest,
	newRequest requestCreator,
	client requestDoer,
	exit func(int),
) {
	adminPort := parseAdminPort(parameters, defaultAdminPort)

	fmt.Fprintf(os.Stderr, request.progressMessage, adminPort)

	url := fmt.Sprintf("https://127.0.0.1:%d%s", adminPort, request.path)

	req, err := newRequest(context.Background(), request.method, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", request.requestFailureMessage, err)
		exit(1)

		return
	}

	if request.contentType != "" {
		req.Header.Set("Content-Type", request.contentType)
	}

	resp, err := client.Do(req) // #nosec G107 - URL constructed from localhost and controlled adminPort
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", request.doFailureMessage, err)
		exit(1)

		return
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == request.expectedStatus {
		fmt.Fprintf(os.Stderr, "%s\n", request.successMessage)
	} else {
		fmt.Fprintf(os.Stderr, request.statusFailureMessage, resp.StatusCode)
		exit(1)
	}
}

// parseConfigFlag extracts config file path from parameters.
func parseConfigFlag(parameters []string, defaultValue string) string {
	for i, param := range parameters {
		// Handle --config=/path/to/file format (single element with =).
		if strings.HasPrefix(param, cryptoutilSharedMagic.IdentityCLIFlagConfig+"=") {
			value := strings.TrimPrefix(param, cryptoutilSharedMagic.IdentityCLIFlagConfig+"=")

			return strings.TrimPrefix(value, cryptoutilSharedMagic.FileURIScheme)
		}

		if strings.HasPrefix(param, configFlagShort+"=") {
			value := strings.TrimPrefix(param, configFlagShort+"=")

			return strings.TrimPrefix(value, cryptoutilSharedMagic.FileURIScheme)
		}
		// Handle --config /path/to/file format (two elements).
		if (param == cryptoutilSharedMagic.IdentityCLIFlagConfig || param == configFlagShort) && i+1 < len(parameters) {
			// Handle file:// prefix for Docker secrets.
			return strings.TrimPrefix(parameters[i+1], cryptoutilSharedMagic.FileURIScheme)
		}
	}

	return defaultValue
}

// parseAdminPort extracts admin port from parameters.
func parseAdminPort(parameters []string, defaultValue int) int {
	for i, param := range parameters {
		if param == "--admin-port" && i+1 < len(parameters) {
			port, err := strconv.Atoi(parameters[i+1])
			if err == nil {
				return port
			}
		}
	}

	return defaultValue
}
