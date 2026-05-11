// Copyright (c) 2025-2026 Justin Cranford.
package listener_test

import (
	"os"
	"testing"

	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps-framework/service/config/tls_generator"
	cryptoutilAppsFrameworkServiceTestHelpBootstrap "cryptoutil/internal/apps-framework/service/test_help_bootstrap"
	cryptoutilAppsFrameworkServiceTestHelpTLS "cryptoutil/internal/apps-framework/service/test_help_tls"

	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps-framework/service/server/testutil"
)

// TestMain initializes shared test fixtures to avoid Windows firewall prompts from multiple server starts.
func TestMain(m *testing.M) {
	settings := cryptoutilAppsFrameworkServiceTestHelpBootstrap.NewTestServerSettingsForTestMain()
	publicTLS := cryptoutilAppsFrameworkServiceTestHelpTLS.NewTestTLSSettingsForTestMain()
	privateTLS := cryptoutilAppsFrameworkServiceTestHelpTLS.NewTestTLSSettingsForTestMain()

	publicMaterial, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(publicTLS)
	if err != nil {
		panic("failed to generate public TLS material: " + err.Error())
	}

	privateMaterial, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateTLSMaterial(privateTLS)
	if err != nil {
		panic("failed to generate private TLS material: " + err.Error())
	}

	cryptoutilAppsFrameworkServiceServerTestutil.ConfigureTestFixtures(settings, publicTLS, privateTLS, publicMaterial.RootCAPool, privateMaterial.RootCAPool)

	os.Exit(m.Run())
}
