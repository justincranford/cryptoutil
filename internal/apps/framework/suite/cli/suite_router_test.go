// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cli_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkSuiteCli "cryptoutil/internal/apps/framework/suite/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var testSuiteCfg = cryptoutilAppsFrameworkSuiteCli.SuiteConfig{ //nolint:gochecknoglobals // test fixture
	UsageText: "Usage: testsuite <product> [service] [options]",
}

func makeTestProductEntry(name string, exitCode int) cryptoutilAppsFrameworkSuiteCli.ProductEntry {
	return cryptoutilAppsFrameworkSuiteCli.ProductEntry{
		Name: name,
		Handler: func(_ []string, _ io.Reader, _, _ io.Writer) int {
			return exitCode
		},
	}
}

func TestRouteSuite_NoArgs(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkSuiteCli.RouteSuite(testSuiteCfg, nil, nil, &stdout, &stderr, nil)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Usage: testsuite")
}

func TestRouteSuite_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: cryptoutilSharedMagic.CLIHelpCommand},
		{name: "help_long", arg: cryptoutilSharedMagic.CLIHelpFlag},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsFrameworkSuiteCli.RouteSuite(testSuiteCfg, []string{tt.arg}, nil, &stdout, &stderr, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stderr.String(), "Usage: testsuite")
		})
	}
}

func TestRouteSuite_UnknownProduct(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkSuiteCli.RouteSuite(testSuiteCfg, []string{"unknown-product"}, nil, &stdout, &stderr, nil)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown product: unknown-product")
	require.Contains(t, stderr.String(), "Usage: testsuite")
}

func TestRouteSuite_RoutesToProduct(t *testing.T) {
	t.Parallel()

	products := []cryptoutilAppsFrameworkSuiteCli.ProductEntry{
		makeTestProductEntry(cryptoutilSharedMagic.SMProductName, 0),
		makeTestProductEntry(cryptoutilSharedMagic.JoseProductName, cryptoutilSharedMagic.AnswerToLifeUniverseEverything),
	}

	t.Run("routes_to_sm", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer

		exitCode := cryptoutilAppsFrameworkSuiteCli.RouteSuite(testSuiteCfg, []string{cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.KMSServiceName, "server"}, nil, &stdout, &stderr, products)
		require.Equal(t, 0, exitCode)
	})

	t.Run("routes_to_jose", func(t *testing.T) {
		t.Parallel()

		var stdout, stderr bytes.Buffer

		exitCode := cryptoutilAppsFrameworkSuiteCli.RouteSuite(testSuiteCfg, []string{cryptoutilSharedMagic.JoseProductName}, nil, &stdout, &stderr, products)
		require.Equal(t, cryptoutilSharedMagic.AnswerToLifeUniverseEverything, exitCode)
	})
}

func TestRouteSuite_MultipleProducts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		productName  string
		expectedCode int
	}{
		{productName: cryptoutilSharedMagic.IdentityProductName, expectedCode: 1},
		{productName: cryptoutilSharedMagic.JoseProductName, expectedCode: 2},
		{productName: cryptoutilSharedMagic.PKIProductName, expectedCode: 3},
		{productName: cryptoutilSharedMagic.SkeletonProductName, expectedCode: 4},
		{productName: cryptoutilSharedMagic.SMProductName, expectedCode: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
	}

	products := make([]cryptoutilAppsFrameworkSuiteCli.ProductEntry, 0, len(tests))

	for i := range tests {
		code := tests[i].expectedCode

		products = append(products, cryptoutilAppsFrameworkSuiteCli.ProductEntry{
			Name: tests[i].productName,
			Handler: func(_ []string, _ io.Reader, _, _ io.Writer) int {
				return code
			},
		})
	}

	for _, tt := range tests {
		t.Run(tt.productName, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsFrameworkSuiteCli.RouteSuite(testSuiteCfg, []string{tt.productName}, nil, &stdout, &stderr, products)
			require.Equal(t, tt.expectedCode, exitCode)
		})
	}
}
