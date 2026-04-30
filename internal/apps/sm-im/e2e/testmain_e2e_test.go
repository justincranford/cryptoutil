// Copyright (c) 2025-2026 Justin Cranford.
//
//

//go:build e2e

package e2e_test

import (
	"fmt"
	http "net/http"
	"os"
	"testing"

	cryptoutilAppsFrameworkTestingE2eInfra "cryptoutil/internal/apps-framework/service/testing/e2e_infra"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient       *http.Client // InsecureSkipVerify — used for health checks / compose readiness.
	sharedHTTPClientWithCA *http.Client // CA-validated — used for TLS chain verification tests.
	composeManager         *cryptoutilAppsFrameworkTestingE2eInfra.ComposeManager

	// Three sm-im instances with different backends (actual container names).
	sqliteContainer    = cryptoutilSharedMagic.IME2ESQLiteContainer      // "sm-im-app-sqlite-1"
	postgres1Container = cryptoutilSharedMagic.IME2EPostgreSQL1Container // "sm-im-app-postgres-1"
	postgres2Container = cryptoutilSharedMagic.IME2EPostgreSQL2Container // "sm-im-app-postgres-2"

	// Service URLs (mapped from container ports to host ports).
	sqlitePublicURL    = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.IME2ESQLitePublicPort)      // "https://127.0.0.1:8700"
	postgres1PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.IME2EPostgreSQL1PublicPort) // "https://127.0.0.1:8701"
	postgres2PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.IME2EPostgreSQL2PublicPort) // "https://127.0.0.1:8702"
	grafanaURL         = fmt.Sprintf("http://127.0.0.1:%d", cryptoutilSharedMagic.IME2EGrafanaPort)                                          // "http://127.0.0.1:3000"

	healthChecks = map[string]string{
		sqliteContainer:    sqlitePublicURL + cryptoutilSharedMagic.IME2EHealthEndpoint,
		postgres1Container: postgres1PublicURL + cryptoutilSharedMagic.IME2EHealthEndpoint,
		postgres2Container: postgres2PublicURL + cryptoutilSharedMagic.IME2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running on Windows.
// Without Docker Desktop, the tests will fail with errors like:
// - "unable to get image... open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified"
// - "Failed to start docker compose: exit status 1"
// This is an environmental requirement, not a code issue. The integration tests (in ../integration/)
// provide sufficient coverage using SQLite in-memory and do not require Docker.
func TestMain(m *testing.M) {
	os.Exit(cryptoutilAppsFrameworkTestingE2eInfra.SetupE2ETestMain(m,
		cryptoutilAppsFrameworkTestingE2eInfra.E2ETestConfig{
			ComposeFile:    cryptoutilSharedMagic.IME2EComposeFile,
			Profiles:       []string{cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, cryptoutilSharedMagic.DockerServicePostgres},
			HealthChecks:   healthChecks,
			HealthTimeout:  cryptoutilSharedMagic.IME2EHealthTimeout,
			CACertPath:     cryptoutilSharedMagic.IME2EPublicCACertPath,
			ServiceLogName: cryptoutilSharedMagic.OTLPServiceSMIM,
		},
		func(env *cryptoutilAppsFrameworkTestingE2eInfra.E2ETestEnv) {
			sharedHTTPClient = env.InsecureClient
			sharedHTTPClientWithCA = env.SecureClient
			composeManager = env.ComposeManager
		}))
}
