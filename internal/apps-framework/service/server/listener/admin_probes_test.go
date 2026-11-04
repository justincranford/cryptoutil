// Copyright (c) 2025-2026 Justin Cranford.
package listener

import (
	"context"
	"errors"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAdminProbeStateTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		transition func(*AdminServer)
		wantStatus int
	}{
		{
			name:       "startup remains closed before initialization",
			path:       cryptoutilSharedMagic.PrivateAdminStartupzRequestPath,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "startup latch opens",
			path: cryptoutilSharedMagic.PrivateAdminStartupzRequestPath,
			transition: func(server *AdminServer) {
				server.MarkStartupComplete()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "readiness requires startup completion",
			path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			transition: func(server *AdminServer) {
				server.SetReady(true)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "readiness opens after startup and dependencies",
			path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			transition: func(server *AdminServer) {
				server.MarkStartupComplete()
				server.SetReady(true)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "readiness closes while draining",
			path: cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			transition: func(server *AdminServer) {
				server.MarkStartupComplete()
				server.SetReady(true)
				server.BeginDrain()
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "liveness remains open while draining",
			path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			transition: func(server *AdminServer) {
				server.BeginDrain()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "startup latch remains open while draining",
			path: cryptoutilSharedMagic.PrivateAdminStartupzRequestPath,
			transition: func(server *AdminServer) {
				server.MarkStartupComplete()
				server.BeginDrain()
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := &AdminServer{app: fiber.New()}
			server.registerRoutes()

			if testCase.transition != nil {
				testCase.transition(server)
			}

			request := httptest.NewRequest(http.MethodGet, cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+testCase.path, nil)
			response, err := server.app.Test(request, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, testCase.wantStatus, response.StatusCode)
		})
	}
}

func TestAdminReadyzUsesSanitizedDependencySnapshot(t *testing.T) {
	t.Parallel()

	const privateError = "sensitive-dependency-error"

	monitor, err := cryptoutilAppsFrameworkServiceServerReadiness.New(cryptoutilAppsFrameworkServiceServerReadiness.Config{
		Checks: []cryptoutilAppsFrameworkServiceServerReadiness.Check{{
			Name:       cryptoutilSharedMagic.RealmStorageTypeDatabase,
			ReasonCode: "database_unavailable",
			Run: func(context.Context) error {
				return errors.New(privateError)
			},
		}},
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)
	require.NoError(t, monitor.Start(t.Context()))
	require.ErrorIs(t, monitor.WaitInitial(t.Context()), cryptoutilAppsFrameworkServiceServerReadiness.ErrInitialChecksFailed)

	server := &AdminServer{app: fiber.New()}
	server.registerRoutes()
	server.MarkStartupComplete()
	server.SetReady(true)
	server.SetReadinessMonitor(monitor)

	request := httptest.NewRequest(http.MethodGet,
		cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath+cryptoutilSharedMagic.PrivateAdminReadyzRequestPath, nil)
	response, err := server.app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, response.StatusCode)
	require.Contains(t, string(body), "database_unavailable")
	require.NotContains(t, string(body), privateError)
	require.NoError(t, monitor.Stop(context.Background()))
}
