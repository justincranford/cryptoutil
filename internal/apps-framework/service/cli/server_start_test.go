// Copyright (c) 2025-2026 Justin Cranford.
//

package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	cryptoutilAppsFrameworkCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	ready bool
}

func (s *testServer) SetReady(ready bool) {
	s.ready = ready
}

func (*testServer) Start(_ context.Context) error {
	return nil
}

func (*testServer) Shutdown(_ context.Context) error {
	return nil
}

func TestStartServiceServer_Help(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.StartServiceServer(
		[]string{cryptoutilSharedMagic.CLIHelpFlag},
		&stdout,
		&stderr,
		cryptoutilAppsFrameworkCli.ServerStartOptions[*struct{}]{
			UsageServer:  "Usage: test server",
			ServiceLabel: "test-svc",
			FlagSetName:  "test-server",
			ParseConfig: func(_ *pflag.FlagSet, _ []string, _ bool) (*struct{}, error) {
				return &struct{}{}, nil
			},
			NewServer: func(_ context.Context, _ *struct{}) (cryptoutilAppsFrameworkCli.ReadyStarter, error) {
				return &testServer{}, nil
			},
			BindAddresses: func(_ *struct{}) (string, uint16, string, uint16) {
				return cryptoutilSharedMagic.IPv4Loopback, 8443, cryptoutilSharedMagic.IPv4Loopback, 9443
			},
		},
	)

	require.Equal(t, 0, exitCode)
	require.Contains(t, stderr.String(), "Usage: test server")
	require.Empty(t, stdout.String())
}

func TestStartServiceServer_ParseError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.StartServiceServer(
		nil,
		&stdout,
		&stderr,
		cryptoutilAppsFrameworkCli.ServerStartOptions[*struct{}]{
			UsageServer:  "Usage: test server",
			ServiceLabel: "test-svc",
			FlagSetName:  "test-server",
			ParseConfig: func(_ *pflag.FlagSet, _ []string, _ bool) (*struct{}, error) {
				return nil, fmt.Errorf("parse failed")
			},
			NewServer: func(_ context.Context, _ *struct{}) (cryptoutilAppsFrameworkCli.ReadyStarter, error) {
				return &testServer{}, nil
			},
			BindAddresses: func(_ *struct{}) (string, uint16, string, uint16) {
				return cryptoutilSharedMagic.IPv4Loopback, 8443, cryptoutilSharedMagic.IPv4Loopback, 9443
			},
		},
	)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Failed to parse configuration")
	require.Contains(t, stderr.String(), "parse failed")
	require.Empty(t, stdout.String())
}

func TestStartServiceServer_Success(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	createdServer := &testServer{}

	exitCode := cryptoutilAppsFrameworkCli.StartServiceServer(
		nil,
		&stdout,
		&stderr,
		cryptoutilAppsFrameworkCli.ServerStartOptions[*struct{}]{
			UsageServer:  "Usage: test server",
			ServiceLabel: "test-svc",
			FlagSetName:  "test-server",
			ParseConfig: func(_ *pflag.FlagSet, _ []string, _ bool) (*struct{}, error) {
				return &struct{}{}, nil
			},
			NewServer: func(_ context.Context, _ *struct{}) (cryptoutilAppsFrameworkCli.ReadyStarter, error) {
				return createdServer, nil
			},
			BindAddresses: func(_ *struct{}) (string, uint16, string, uint16) {
				return cryptoutilSharedMagic.IPv4Loopback, 8443, cryptoutilSharedMagic.IPv4Loopback, 9443
			},
		},
	)

	require.Equal(t, 0, exitCode)
	require.True(t, createdServer.ready)
	require.Contains(t, stdout.String(), "Starting test-svc service")
	require.Contains(t, stdout.String(), "Public Server: https://127.0.0.1:8443")
	require.Contains(t, stdout.String(), "Admin Server:  https://127.0.0.1:9443")
	require.Contains(t, stdout.String(), "test-svc service stopped")
	require.Empty(t, stderr.String())
}
