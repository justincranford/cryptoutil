// Copyright (c) 2025-2026 Justin Cranford.
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

	// Three skeleton-template instances with different backends (actual compose service names).
	sqliteContainer    = cryptoutilSharedMagic.SkeletonTemplateE2ESQLiteContainer      // "skeleton-template-app-sqlite-1"
	postgres1Container = cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL1Container // "skeleton-template-app-postgres-1"
	postgres2Container = cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL2Container // "skeleton-template-app-postgres-2"

	// Service URLs (mapped from container ports to host ports).
	sqlitePublicURL    = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.SkeletonTemplateE2ESQLitePublicPort)      // "https://127.0.0.1:18900"
	postgres1PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL1PublicPort) // "https://127.0.0.1:18901"
	postgres2PublicURL = cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + fmt.Sprintf("%d", cryptoutilSharedMagic.SkeletonTemplateE2EPostgreSQL2PublicPort) // "https://127.0.0.1:18902"

	healthChecks = map[string]string{
		sqliteContainer:    sqlitePublicURL + cryptoutilSharedMagic.SkeletonTemplateE2EHealthEndpoint,
		postgres1Container: postgres1PublicURL + cryptoutilSharedMagic.SkeletonTemplateE2EHealthEndpoint,
		postgres2Container: postgres2PublicURL + cryptoutilSharedMagic.SkeletonTemplateE2EHealthEndpoint,
	}
)

// TestMain orchestrates docker compose lifecycle for skeleton-template E2E tests.
// This validates production-ready deployment with PostgreSQL, telemetry, and multiple instances.
//
// ENVIRONMENTAL NOTE: These E2E tests require Docker Desktop to be running.
// Without Docker Desktop, the tests will fail with compose errors.
func TestMain(m *testing.M) {
	os.Exit(cryptoutilAppsFrameworkTestingE2eInfra.SetupE2ETestMain(m,
		cryptoutilAppsFrameworkTestingE2eInfra.E2ETestConfig{
			ComposeFile:    cryptoutilSharedMagic.SkeletonTemplateE2EComposeFile,
			Profiles:       []string{cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, cryptoutilSharedMagic.DockerServicePostgres},
			HealthChecks:   healthChecks,
			HealthTimeout:  cryptoutilSharedMagic.SkeletonTemplateE2EHealthTimeout,
			CACertPath:     cryptoutilSharedMagic.SkeletonTemplateE2EPublicCACertPath,
			ServiceLogName: cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
		},
		func(env *cryptoutilAppsFrameworkTestingE2eInfra.E2ETestEnv) {
			sharedHTTPClient = env.InsecureClient
			sharedHTTPClientWithCA = env.SecureClient
			composeManager = env.ComposeManager
		}))
}
