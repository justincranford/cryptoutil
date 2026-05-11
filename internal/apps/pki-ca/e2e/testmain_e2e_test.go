//go:build e2e

// Copyright (c) 2025-2026 Justin Cranford.
package e2e_test

import (
	"fmt"
	http "net/http"
	"os"
	"testing"

	cryptoutilTestOrchE2e "cryptoutil/internal/apps-framework/service/test_orch_e2e"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient       *http.Client // InsecureSkipVerify — used for health checks / compose readiness.
	sharedHTTPClientWithCA *http.Client // CA-validated — used for TLS chain verification tests.
	composeManager         *cryptoutilTestOrchE2e.E2EComposeManager

	// Four pki-ca instances with different backends (actual compose service names).
	sqlite1Container   = cryptoutilSharedMagic.PKICAE2ESQLiteContainer      // "pki-ca-app-sqlite-1"
	sqlite2Container   = cryptoutilSharedMagic.PKICAE2ESQLite2Container     // "pki-ca-app-sqlite-2"
	postgres1Container = cryptoutilSharedMagic.PKICAE2EPostgreSQL1Container // "pki-ca-app-postgresql-1"
	postgres2Container = cryptoutilSharedMagic.PKICAE2EPostgreSQL2Container // "pki-ca-app-postgresql-2"

	// Service URLs (mapped from container ports to host ports).
	sqlite1PublicURL   = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.PKICAE2ESQLitePublicPort)      // "https://127.0.0.1:8300"
	sqlite2PublicURL   = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.PKICAE2ESQLite2PublicPort)     // "https://127.0.0.1:8301"
	postgres1PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.PKICAE2EPostgreSQL1PublicPort) // "https://127.0.0.1:8302"
	postgres2PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.PKICAE2EPostgreSQL2PublicPort) // "https://127.0.0.1:8303"

	healthChecks = map[string]string{
		sqlite1Container:   sqlite1PublicURL + cryptoutilSharedMagic.PKICAE2EHealthEndpoint,
		sqlite2Container:   sqlite2PublicURL + cryptoutilSharedMagic.PKICAE2EHealthEndpoint,
		postgres1Container: postgres1PublicURL + cryptoutilSharedMagic.PKICAE2EHealthEndpoint,
		postgres2Container: postgres2PublicURL + cryptoutilSharedMagic.PKICAE2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for pki-ca E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running.
// Without Docker Desktop, the tests will fail with compose errors.
func TestMain(m *testing.M) {
	os.Exit(cryptoutilTestOrchE2e.SetupE2ETestMain(m,
		cryptoutilTestOrchE2e.E2ETestConfig{
			ComposeFile:    cryptoutilSharedMagic.PKICAE2EComposeFile,
			Profiles:       []string{cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, cryptoutilSharedMagic.DockerServicePostgres},
			HealthChecks:   healthChecks,
			HealthTimeout:  cryptoutilSharedMagic.PKICAE2EHealthTimeout,
			CACertPath:     cryptoutilSharedMagic.PKICAE2EPublicCACertPath,
			ServiceLogName: cryptoutilSharedMagic.OTLPServicePKICA,
		},
		func(env *cryptoutilTestOrchE2e.E2ETestEnv) {
			sharedHTTPClient = env.InsecureClient
			sharedHTTPClientWithCA = env.SecureClient
			composeManager = env.ComposeManager
		}))
}
