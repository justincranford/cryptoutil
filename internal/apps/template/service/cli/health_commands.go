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
// defaultPublicPort is the service's default public port (e.g., 8700 for cipher-im).
func HealthCommand(args []string, stdout, stderr io.Writer, usageText string, defaultPublicPort uint16) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, usageText)

		return 0
	}

	// Parse flags.
	defaultBase := fmt.Sprintf("https://127.0.0.1:%d%s", defaultPublicPort, cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath)

	url := defaultBase

	cacertPath := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultBase { // Only set if not already set
				baseURL := args[i+1]

				healthPath := cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath + "/health"
				if !strings.HasSuffix(baseURL, "/health") {
					if strings.HasSuffix(baseURL, healthPath) {
						url = baseURL
					} else {
						url = baseURL + "/health"
					}
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call health endpoint.
	statusCode, body, err := HTTPGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Health check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		_, _ = fmt.Fprintf(stdout, "\u2705 Service is healthy (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "\u274c Service is unhealthy (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// LivezCommand implements the livez subcommand.
// Calls GET /admin/api/v1/livez on the admin server.
// usageText is shown when --help is passed.
func LivezCommand(args []string, stdout, stderr io.Writer, usageText string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, usageText)

		return 0
	}

	// Parse flags.
	livezPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath

	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", cryptoutilSharedMagic.DefaultPrivatePortCryptoutil, livezPath)

	url := defaultURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultURL { // Only set if not already set
				baseURL := args[i+1]

				if !strings.HasSuffix(baseURL, livezPath) {
					url = baseURL + livezPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call livez endpoint.
	statusCode, body, err := HTTPGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Liveness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		_, _ = fmt.Fprintf(stdout, "\u2705 Service is alive (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "\u274c Service is not alive (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// ReadyzCommand implements the readyz subcommand.
// Calls GET /admin/api/v1/readyz on the admin server.
// usageText is shown when --help is passed.
func ReadyzCommand(args []string, stdout, stderr io.Writer, usageText string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, usageText)

		return 0
	}

	// Parse flags.
	readyzPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath

	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", cryptoutilSharedMagic.DefaultPrivatePortCryptoutil, readyzPath)

	url := defaultURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultURL { // Only set if not already set
				baseURL := args[i+1]

				if !strings.HasSuffix(baseURL, readyzPath) {
					url = baseURL + readyzPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call readyz endpoint.
	statusCode, body, err := HTTPGet(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Readiness check failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK {
		_, _ = fmt.Fprintf(stdout, "\u2705 Service is ready (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "\u274c Service is not ready (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}

// ShutdownCommand implements the shutdown subcommand.
// Calls POST /admin/api/v1/shutdown on the admin server.
// usageText is shown when --help is passed.
func ShutdownCommand(args []string, stdout, stderr io.Writer, usageText string) int {
	if len(args) > 0 && (args[0] == helpCommand || args[0] == helpFlag || args[0] == helpShortFlag) {
		_, _ = fmt.Fprintln(stderr, usageText)

		return 0
	}

	// Parse flags.
	shutdownPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath

	defaultURL := fmt.Sprintf("https://127.0.0.1:%d%s", cryptoutilSharedMagic.DefaultPrivatePortCryptoutil, shutdownPath)

	url := defaultURL

	var cacertPath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case urlFlag:
			if i+1 < len(args) && url == defaultURL { // Only set if not already set
				baseURL := args[i+1]

				if !strings.HasSuffix(baseURL, shutdownPath) {
					url = baseURL + shutdownPath
				} else {
					url = baseURL
				}

				i++ // Skip next arg
			}
		case cacertFlag:
			if i+1 < len(args) && cacertPath == "" { // Only set if not already set
				cacertPath = args[i+1]
				i++ // Skip next arg
			}
		}
	}

	// Call shutdown endpoint.
	statusCode, body, err := HTTPPost(url, cacertPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "\u274c Shutdown request failed: %v\n", err)

		return 1
	}

	// Display results.
	if statusCode == http.StatusOK || statusCode == http.StatusAccepted {
		_, _ = fmt.Fprintf(stdout, "\u2705 Shutdown initiated (HTTP %d)\n", statusCode)

		if body != "" {
			_, _ = fmt.Fprintln(stdout, body)
		}

		return 0
	}

	_, _ = fmt.Fprintf(stderr, "\u274c Shutdown request failed (HTTP %d)\n", statusCode)

	if body != "" {
		_, _ = fmt.Fprintln(stderr, body)
	}

	return 1
}
