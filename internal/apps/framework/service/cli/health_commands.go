// Copyright (c) 2025 Justin Cranford
//

package cli

import (
	"fmt"
	"io"
	http "net/http"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// HealthCommand implements the health subcommand.
// Calls GET /service/api/v1/health on the public server.
// usageText is shown when --help is passed.
// defaultPublicPort is the service's default public port (e.g., 8700 for sm-im).
func HealthCommand(args []string, stdout, stderr io.Writer, usageText string, defaultPublicPort uint16) int {
	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", defaultPublicPort, cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath)

	return httpGetCommand(args, stdout, stderr, usageText, defaultURL, "/health",
		http.StatusOK, "Service is healthy", "Health check failed", "Service is unhealthy")
}

// LivezCommand implements the livez subcommand.
// Calls GET /admin/api/v1/livez on the admin server.
// usageText is shown when --help is passed.
func LivezCommand(args []string, stdout, stderr io.Writer, usageText string) int {
	livezPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath
	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", cryptoutilSharedMagic.DefaultPrivatePortCryptoutil, livezPath)

	return httpGetCommand(args, stdout, stderr, usageText, defaultURL, livezPath,
		http.StatusOK, "Service is alive", "Liveness check failed", "Service is not alive")
}

// ReadyzCommand implements the readyz subcommand.
// Calls GET /admin/api/v1/readyz on the admin server.
// usageText is shown when --help is passed.
func ReadyzCommand(args []string, stdout, stderr io.Writer, usageText string) int {
	readyzPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath
	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", cryptoutilSharedMagic.DefaultPrivatePortCryptoutil, readyzPath)

	return httpGetCommand(args, stdout, stderr, usageText, defaultURL, readyzPath,
		http.StatusOK, "Service is ready", "Readiness check failed", "Service is not ready")
}

// ShutdownCommand implements the shutdown subcommand.
// Calls POST /admin/api/v1/shutdown on the admin server.
// usageText is shown when --help is passed.
func ShutdownCommand(args []string, stdout, stderr io.Writer, usageText string) int {
	shutdownPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath
	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", cryptoutilSharedMagic.DefaultPrivatePortCryptoutil, shutdownPath)

	return httpPostCommand(args, stdout, stderr, usageText, defaultURL, shutdownPath,
		"Shutdown initiated", "Shutdown request failed", "Shutdown request failed")
}

// httpGetCommand is the shared implementation for GET-based health check commands.
// It parses --url and --cacert flags, calls HTTPGet, and displays results.
// successCode is the expected HTTP status code for success.
// successMsg, requestErrMsg, and failureMsg are the display messages.
func httpGetCommand(
	args []string,
	stdout, stderr io.Writer,
	usageText string,
	defaultURL string,
	urlSuffix string,
	successCode int,
	successMsg string,
	requestErrMsg string,
	failureMsg string,
) int {
	if IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, usageText)

		return 0
	}

	url, cacertPath := parseURLAndCACert(args, defaultURL, urlSuffix)

	statusCode, body, err := HTTPGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ %s: %v\n", requestErrMsg, err)

		return 1
	}

	return displayResult(stdout, stderr, statusCode, body, successCode, successMsg, failureMsg)
}

// httpPostCommand is the shared implementation for POST-based admin commands (shutdown).
// It parses --url and --cacert flags, calls HTTPPost, and displays results.
func httpPostCommand(
	args []string,
	stdout, stderr io.Writer,
	usageText string,
	defaultURL string,
	urlSuffix string,
	successMsg string,
	requestErrMsg string,
	failureMsg string,
) int {
	if IsHelpRequest(args) {
		_, _ = fmt.Fprintln(stderr, usageText)

		return 0
	}

	url, cacertPath := parseURLAndCACert(args, defaultURL, urlSuffix)

	statusCode, body, err := HTTPPost(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "❌ %s: %v\n", requestErrMsg, err)

		return 1
	}

	// ShutdownCommand accepts both 200 OK and 202 Accepted as success.
	if statusCode == http.StatusOK || statusCode == http.StatusAccepted {
		_, _ = fmt.Fprintf(stdout, "✅ %s (HTTP %d)\n", successMsg, statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "❌ %s (HTTP %d)\n", failureMsg, statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// parseURLAndCACert parses --url and --cacert flags from args.
// defaultURL is used when --url is not provided.
// urlSuffix ensures the URL ends with the correct path suffix.
func parseURLAndCACert(args []string, defaultURL, urlSuffix string) (url, cacertPath string) {
	url = defaultURL

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case cryptoutilSharedMagic.CLIURLFlag:
			if i+1 < len(args) && url == defaultURL {
				baseURL := args[i+1]

				if strings.HasSuffix(baseURL, urlSuffix) {
					url = baseURL
				} else {
					url = baseURL + urlSuffix
				}

				i++
			}
		case cryptoutilSharedMagic.CLICACertFlag:
			if i+1 < len(args) && cacertPath == "" {
				cacertPath = args[i+1]
				i++
			}
		}
	}

	return url, cacertPath
}

// displayResult writes success or failure output based on the HTTP status code.
func displayResult(stdout, stderr io.Writer, statusCode int, body string, successCode int, successMsg, failureMsg string) int {
	if statusCode == successCode {
		_, _ = fmt.Fprintf(stdout, "✅ %s (HTTP %d)\n", successMsg, statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "❌ %s (HTTP %d)\n", failureMsg, statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}
