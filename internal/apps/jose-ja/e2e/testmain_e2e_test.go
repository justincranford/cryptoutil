// Copyright (c) 2025 Justin Cranford
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

	// Three jose-ja instances with different backends (actual compose service names).
	sqliteContainer    = cryptoutilSharedMagic.JoseJAE2ESQLiteContainer      // "jose-ja-app-sqlite-1"
	postgres1Container = cryptoutilSharedMagic.JoseJAE2EPostgreSQL1Container // "jose-ja-app-postgres-1"
	postgres2Container = cryptoutilSharedMagic.JoseJAE2EPostgreSQL2Container // "jose-ja-app-postgres-2"

	// Service URLs (mapped from container ports to host ports).
	sqlitePublicURL    = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.JoseJAE2ESQLitePublicPort)      // "https://127.0.0.1:18800"
	postgres1PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.JoseJAE2EPostgreSQL1PublicPort) // "https://127.0.0.1:18801"
	postgres2PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.JoseJAE2EPostgreSQL2PublicPort) // "https://127.0.0.1:18802"

	healthChecks = map[string]string{
		sqliteContainer:    sqlitePublicURL + cryptoutilSharedMagic.JoseJAE2EHealthEndpoint,
		postgres1Container: postgres1PublicURL + cryptoutilSharedMagic.JoseJAE2EHealthEndpoint,
		postgres2Container: postgres2PublicURL + cryptoutilSharedMagic.JoseJAE2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for jose-ja E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running.
// Without Docker Desktop, the tests will fail with compose errors.
func TestMain(m *testing.M) {
	os.Exit(cryptoutilAppsFrameworkTestingE2eInfra.SetupE2ETestMain(m,
		cryptoutilAppsFrameworkTestingE2eInfra.E2ETestConfig{
			ComposeFile:    cryptoutilSharedMagic.JoseJAE2EComposeFile,
			Profiles:       []string{cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, cryptoutilSharedMagic.DockerServicePostgres},
			HealthChecks:   healthChecks,
			HealthTimeout:  cryptoutilSharedMagic.JoseJAE2EHealthTimeout,
			CACertPath:     cryptoutilSharedMagic.JoseJAE2EPublicCACertPath,
			ServiceLogName: cryptoutilSharedMagic.OTLPServiceJoseJA,
		},
		func(env *cryptoutilAppsFrameworkTestingE2eInfra.E2ETestEnv) {
			sharedHTTPClient = env.InsecureClient
			sharedHTTPClientWithCA = env.SecureClient
			composeManager = env.ComposeManager
		}))
}
