// Copyright (c) 2025-2026 Justin Cranford.
//

package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testUnexpectedInitFlag = "--invalid-option"

type testIdentitySettings struct{}

func (testIdentitySettings) GetBindAddresses() (string, uint16, string, uint16) {
	return cryptoutilSharedMagic.IPv4Loopback, 8443, cryptoutilSharedMagic.IPv4Loopback, 9443
}

// testIdentityBase holds the 5 service-identity constants used across multiple tests.
// Tests that need a custom ServerFn build a local ServiceIdentity with these same fields.
var testIdentityBase = cryptoutilAppsFrameworkCli.ServiceIdentity{ //nolint:gochecknoglobals // test fixture
	ServiceID:   testServiceID,
	ProductName: testProductNameService,
	ServiceName: testServiceNameConst,
	DisplayName: testTestService,
	ServicePort: testDefaultPort,
}

func TestBuildServerUsage(t *testing.T) {
	t.Parallel()

	got := cryptoutilAppsFrameworkCli.BuildServerUsage(testIdentityBase)
	require.Contains(t, got, testProductNameService)
	require.Contains(t, got, testServiceNameConst)
	require.Contains(t, got, testTestService)
	require.Contains(t, got, "configs/test-svc/test-svc-framework.yml")
}

func TestNewServiceIdentity_ServerFn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		args            []string
		parseErr        error
		newServerErr    error
		wantExitCode    int
		wantStderr      string
		wantReady       bool
		wantPublicURL   string
		wantParseCalls  int
		wantServerCalls int
	}{
		{
			name:         cryptoutilSharedMagic.CLIHelpCommand,
			args:         []string{cryptoutilSharedMagic.CLIHelpFlag},
			wantStderr:   testServiceNameConst,
			wantExitCode: 0,
		},
		{
			name:           "parse error",
			parseErr:       errors.New("injected configuration parse failure"),
			wantStderr:     "Failed to parse configuration",
			wantExitCode:   1,
			wantParseCalls: 1,
		},
		{
			name:            "server construction error",
			newServerErr:    errors.New("injected server construction failure"),
			wantStderr:      "Failed to create server",
			wantExitCode:    1,
			wantParseCalls:  1,
			wantServerCalls: 1,
		},
		{
			name:            "starts server with settings bind addresses",
			wantExitCode:    0,
			wantReady:       true,
			wantPublicURL:   "https://127.0.0.1:8443",
			wantParseCalls:  1,
			wantServerCalls: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := &testServer{}

			var parseCalls, newServerCalls int

			identity := cryptoutilAppsFrameworkCli.NewServiceIdentity(
				testServiceID,
				testProductNameService,
				testServiceNameConst,
				testTestService,
				testDefaultPort,
				func(_ *pflag.FlagSet, args []string, exitIfHelp bool) (testIdentitySettings, error) {
					parseCalls++

					require.Equal(t, []string{"start"}, args)
					require.True(t, exitIfHelp)

					return testIdentitySettings{}, testCase.parseErr
				},
				func(_ context.Context, _ testIdentitySettings) (*testServer, error) {
					newServerCalls++

					if testCase.newServerErr != nil {
						return nil, testCase.newServerErr
					}

					return server, nil
				},
			)

			var stdout, stderr bytes.Buffer

			exitCode := identity.ServerFn(testCase.args, &stdout, &stderr)

			require.Equal(t, testCase.wantExitCode, exitCode)

			if testCase.wantStderr != "" {
				require.Contains(t, stderr.String(), testCase.wantStderr)
			}

			require.Equal(t, testCase.wantReady, server.ready)

			if testCase.wantPublicURL != "" {
				require.Contains(t, stdout.String(), testCase.wantPublicURL)
			}

			require.Equal(t, testCase.wantParseCalls, parseCalls)
			require.Equal(t, testCase.wantServerCalls, newServerCalls)
		})
	}
}

func TestServerFlagSetName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "test-svc-server", cryptoutilAppsFrameworkCli.ServerFlagSetName(testServiceID))
}

func TestRouteServiceFromIdentity_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: testHelpWord, arg: cryptoutilSharedMagic.CLIHelpCommand},
		{name: testHelpLong, arg: cryptoutilSharedMagic.CLIHelpFlag},
		{name: testHelpShort, arg: cryptoutilSharedMagic.CLIHelpShortFlag},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
				testIdentityBase,
				[]string{tt.arg},
				&stdout, &stderr,
			)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout.String(), testServiceNameConst)
		})
	}
}

func TestRouteServiceFromIdentity_VersionSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentityBase,
		[]string{cryptoutilSharedMagic.CLIVersionCommand},
		&stdout, &stderr,
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), testServiceID)
}

func TestRouteServiceFromIdentity_ServerSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	serverCalled := false
	id := cryptoutilAppsFrameworkCli.ServiceIdentity{
		ServiceID:   testServiceID,
		ProductName: testProductNameService,
		ServiceName: testServiceNameConst,
		DisplayName: testTestService,
		ServicePort: testDefaultPort,
		ServerFn: func(_ []string, _, _ io.Writer) int {
			serverCalled = true

			return 0
		},
	}

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		id,
		[]string{cryptoutilSharedMagic.CMD_SERVER},
		&stdout, &stderr,
	)
	require.Equal(t, 0, exitCode)
	require.True(t, serverCalled)
}

func TestRouteServiceFromIdentity_ClientHelpSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentityBase,
		[]string{testClient, cryptoutilSharedMagic.CLIHelpFlag},
		&stdout, &stderr,
	)
	require.Equal(t, 0, exitCode)
}

func TestRouteServiceFromIdentity_ClientUnknownArg(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentityBase,
		[]string{testClient, testSomeArg},
		&stdout, &stderr,
	)
	require.Equal(t, 1, exitCode)
}

func TestRouteServiceFromIdentity_InitHelpSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentityBase,
		[]string{testInit, cryptoutilSharedMagic.CLIHelpFlag},
		&stdout, &stderr,
	)
	require.Equal(t, 0, exitCode)
}

func TestRouteServiceFromIdentity_InitFlagError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentityBase,
		[]string{testInit, testUnexpectedInitFlag},
		&stdout, &stderr,
	)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), cryptoutilSharedMagic.PSIDPKIInit)
	require.Contains(t, stderr.String(), "unknown flag")
}

func TestRouteServiceFromIdentity_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentityBase,
		[]string{"unknown-subcommand"},
		&stdout, &stderr,
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown subcommand")
}

func TestRouteServiceFromIdentity_EmptyArgsDefaultsToServer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	serverCalled := false
	id := cryptoutilAppsFrameworkCli.ServiceIdentity{
		ServiceID:   testServiceID,
		ProductName: testProductNameService,
		ServiceName: testServiceNameConst,
		DisplayName: testTestService,
		ServicePort: testDefaultPort,
		ServerFn: func(_ []string, _, _ io.Writer) int {
			serverCalled = true

			return 0
		},
	}

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		id,
		nil,
		&stdout, &stderr,
	)
	require.Equal(t, 0, exitCode)
	require.True(t, serverCalled)
}

func TestBuildServerUsage_ContainsConfigPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		serviceID string
		wantInfix string
	}{
		{name: "svc_sm_kms", serviceID: cryptoutilSharedMagic.OTLPServiceSMKMS, wantInfix: "configs/sm-kms/sm-kms-framework.yml"},
		{name: "svc_skeleton_template", serviceID: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, wantInfix: "configs/skeleton-template/skeleton-template-framework.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id := cryptoutilAppsFrameworkCli.ServiceIdentity{
				ServiceID:   tt.serviceID,
				ProductName: "p",
				ServiceName: "s",
				DisplayName: "D",
				ServicePort: cryptoutilSharedMagic.KMSServicePort,
			}
			got := cryptoutilAppsFrameworkCli.BuildServerUsage(id)
			require.True(t, strings.Contains(got, tt.wantInfix),
				"expected usage to contain %q, got: %q", tt.wantInfix, got)
		})
	}
}
