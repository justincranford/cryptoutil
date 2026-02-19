// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cli_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

var testProductCfg = cryptoutilAppsTemplateCli.ProductConfig{ //nolint:gochecknoglobals // test fixture
	ProductName: "testproduct",
	UsageText:   "Usage: testproduct <service> [options]",
	VersionText: "testproduct v1.0.0",
}

func makeTestServiceEntry(name string, exitCode int) cryptoutilAppsTemplateCli.ServiceEntry {
	return cryptoutilAppsTemplateCli.ServiceEntry{
		Name: name,
		Handler: func(_ []string, _ io.Reader, _, _ io.Writer) int {
			return exitCode
		},
	}
}

func TestRouteProduct_NoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, nil, nil, &stdout, &stderr, nil)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Usage: testproduct")
}

func TestRouteProduct_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: "help"},
		{name: "help_long", arg: "--help"},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, []string{tt.arg}, nil, &stdout, &stderr, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stderr.String(), "Usage: testproduct")
		})
	}
}

func TestRouteProduct_VersionFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "version_word", arg: "version"},
		{name: "version_long", arg: "--version"},
		{name: "version_short", arg: "-v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, []string{tt.arg}, nil, &stdout, &stderr, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout.String(), "testproduct v1.0.0")
		})
	}
}

func TestRouteProduct_UnknownService(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, []string{"unknown-svc"}, nil, &stdout, &stderr, nil)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown service: unknown-svc")
	require.Contains(t, stderr.String(), "Usage: testproduct")
}

func TestRouteProduct_RoutesToService(t *testing.T) {
	t.Parallel()

	services := []cryptoutilAppsTemplateCli.ServiceEntry{
		makeTestServiceEntry("svc1", 0),
		makeTestServiceEntry("svc2", 42),
	}

	t.Run("routes_to_svc1", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer

		exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, []string{"svc1", "server"}, nil, &stdout, &stderr, services)
		require.Equal(t, 0, exitCode)
	})

	t.Run("routes_to_svc2", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer

		exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, []string{"svc2"}, nil, &stdout, &stderr, services)
		require.Equal(t, 42, exitCode)
	})
}

func TestRouteProduct_MultipleServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		serviceName  string
		expectedCode int
	}{
		{serviceName: "authz", expectedCode: 1},
		{serviceName: "idp", expectedCode: 2},
		{serviceName: "rs", expectedCode: 3},
		{serviceName: "rp", expectedCode: 4},
		{serviceName: "spa", expectedCode: 5},
	}

	services := make([]cryptoutilAppsTemplateCli.ServiceEntry, 0, len(tests))

	for i, tt := range tests {
		code := tests[i].expectedCode

		_ = tt

		services = append(services, cryptoutilAppsTemplateCli.ServiceEntry{
			Name: tests[i].serviceName,
			Handler: func(_ []string, _ io.Reader, _, _ io.Writer) int {
				return code
			},
		})
	}

	for _, tt := range tests {
		t.Run(tt.serviceName, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.RouteProduct(testProductCfg, []string{tt.serviceName}, nil, &stdout, &stderr, services)
			require.Equal(t, tt.expectedCode, exitCode)
		})
	}
}
