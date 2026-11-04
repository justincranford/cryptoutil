// Copyright (c) 2025-2026 Justin Cranford.
package cli_test

import (
	"bytes"
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilAppsFrameworkCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestStartupzCommandUsesExpectedEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		wantExit   int
		wantText   string
	}{
		{name: "startup complete", statusCode: http.StatusOK, wantText: "Service startup complete"},
		{name: "startup incomplete", statusCode: http.StatusServiceUnavailable, wantExit: 1, wantText: "Service startup is incomplete"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			paths := make(chan string, 1)

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				paths <- request.URL.Path

				writer.WriteHeader(testCase.statusCode)
			}))
			defer server.Close()

			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			exitCode := cryptoutilAppsFrameworkCli.StartupzCommand(
				[]string{cryptoutilSharedMagic.CLIURLFlag, server.URL}, &stdout, &stderr, "startupz usage",
			)

			require.Equal(t, testCase.wantExit, exitCode)
			require.Equal(t, cryptoutilSharedMagic.DefaultAdminStartupzEndpoint, <-paths)

			if testCase.wantExit == 0 {
				require.Contains(t, stdout.String(), testCase.wantText)
			} else {
				require.Contains(t, stderr.String(), testCase.wantText)
			}
		})
	}
}

func TestHTTPGetUsesBoundedProbeContext(t *testing.T) {
	t.Parallel()

	requestCanceled := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
		close(requestCanceled)
	}))
	defer server.Close()

	_, _, err := cryptoutilAppsFrameworkCli.HTTPGet(server.URL, "", "", "")
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	<-requestCanceled
}
