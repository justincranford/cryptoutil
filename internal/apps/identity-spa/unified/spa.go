// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package spa provides the unified command interface for Identity SPA service.
package spa

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

	cryptoutilAppsIdentitySpaServer "cryptoutil/internal/apps/identity-spa/server"
	cryptoutilAppsIdentitySpaServerConfig "cryptoutil/internal/apps/identity-spa/server/config"
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

// Execute handles SPA service commands matching other service patterns.
// Supports: start, stop, status, health.
func Execute(parameters []string) {
	execute(parameters, commandHandlers{
		start: func(parameters []string) {
			startService(
				parameters,
				cryptoutilAppsIdentitySpaServerConfig.Parse,
				func(ctx context.Context, settings *cryptoutilAppsIdentitySpaServerConfig.IdentitySPAServerSettings) (serviceLifecycle, error) {
					return cryptoutilAppsIdentitySpaServer.NewFromConfig(ctx, settings)
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
	fmt.Println("Usage: cryptoutil identity-spa <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  start    - Start Identity SPA service")
	fmt.Println("  stop     - Stop Identity SPA service")
	fmt.Println("  status   - Check service status")
	fmt.Println("  health   - Check service health")
	fmt.Println("\nOptions:")
	fmt.Println("  --config, -c <path>  - Config file path (default: /app/run/identity-spa-docker.yml)")
}

// startService starts the Identity SPA service.
func startService[T any](
	parameters []string,
	parseConfig func([]string, bool) (*T, error),
	newServer func(context.Context, *T) (serviceLifecycle, error),
	fatalf func(string, ...any),
) {
	configFile := parseConfigFlag(parameters, "/app/run/identity-spa-docker.yml")

	fmt.Fprintf(os.Stderr, "Starting Identity SPA service\n")
	fmt.Fprintf(os.Stderr, "Using config file: %s\n", configFile)

	// Load SPA-specific configuration from YAML file.
	parseArgs := []string{valueStart, cryptoutilSharedMagic.IdentityCLIFlagConfig, configFile}

	settings, err := parseConfig(parseArgs, false)
	if err != nil {
		fatalf("Failed to load config from %s: %v", configFile, err)

		return
	}

	// Create context.
	ctx := context.Background()

	// Create SPA server using template-based implementation.
	server, err := newServer(ctx, settings)
	if err != nil {
		fatalf("Failed to create SPA server: %v", err)

		return
	}

	// Mark server as ready.
	server.SetReady(true)

	fmt.Fprintf(os.Stderr, "Starting SPA server...\n")

	// Start server (blocks until shutdown).
	if err := server.Start(ctx); err != nil {
		fatalf("SPA server error: %v", err)
	}
}

// stopService sends shutdown request to SPA admin endpoint.
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

// statusService checks SPA service readiness.
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

// healthService checks SPA service health.
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
