//go:build e2e

// Copyright (c) 2025-2026 Justin Cranford.

package test_orch_e2e

import (
	"testing"

	cryptoutilAppsFrameworkTestingE2eInfra "cryptoutil/internal/apps-framework/service/testing/e2e_infra"
)

// E2EComposeManager aliases the docker-compose manager used by e2e_infra.
type E2EComposeManager = cryptoutilAppsFrameworkTestingE2eInfra.ComposeManager

// E2ETestConfig aliases the canonical E2E config struct.
type E2ETestConfig = cryptoutilAppsFrameworkTestingE2eInfra.E2ETestConfig

// E2ETestEnv aliases the canonical E2E environment struct.
type E2ETestEnv = cryptoutilAppsFrameworkTestingE2eInfra.E2ETestEnv

// SetupE2ETestMain delegates to testing/e2e_infra and provides a pass-through
// mode for minimal E2E packages that only need standard TestMain behavior.
func SetupE2ETestMain(m *testing.M, cfg E2ETestConfig, onReady func(*E2ETestEnv)) int {
	if cfg.ComposeFile == "" {
		if onReady != nil {
			onReady(&E2ETestEnv{})
		}

		return m.Run()
	}

	if onReady == nil {
		onReady = func(*E2ETestEnv) {}
	}

	return cryptoutilAppsFrameworkTestingE2eInfra.SetupE2ETestMain(m, cfg, onReady)
}
