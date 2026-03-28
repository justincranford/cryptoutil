package lint_deployments

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestGetExpectedConfigsContents validates the config contents map.
func TestGetExpectedConfigsContents(t *testing.T) {
	t.Parallel()

	contents := GetExpectedConfigsContents()

	require.NotEmpty(t, contents, "expected configs contents must not be empty")

	// All entries should be OPTIONAL (configs are less strict than deployments).
	for path, status := range contents {
		require.Equal(t, OptionalFileStatus, status, "config path %s should be OPTIONAL", path)
	}

	// Verify expected config directories exist.
	expectedDirs := []string{
		"cryptoutil/",
		"identity-authz/", "identity-authz/domain/policies/",
		"identity-idp/",
		"identity-rp/",
		"identity-rs/",
		"identity-spa/",
		"jose-ja/",
		"pki-ca/", "pki-ca/profiles/",
		"skeleton-template/",
		"sm-im/",
		"sm-kms/",
	}

	for _, dir := range expectedDirs {
		_, ok := contents[dir]
		require.True(t, ok, "expected config directory %q not found", dir)
	}

	require.Len(t, contents, len(expectedDirs), "unexpected extra entries in configs contents")
}

// TestGetDeploymentDirectories validates deployment directory lists.
func TestGetDeploymentDirectories(t *testing.T) {
	t.Parallel()

	suite, product, productService, infrastructure, template := GetDeploymentDirectories()

	// Suite should contain exactly cryptoutil.
	require.Len(t, suite, 1)
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, suite[0])

	// Products should include all 5 products.
	expectedProducts := []string{cryptoutilSharedMagic.IdentityProductName, cryptoutilSharedMagic.SMProductName, cryptoutilSharedMagic.PKIProductName, cryptoutilSharedMagic.JoseProductName, cryptoutilSharedMagic.SkeletonProductName}
	require.ElementsMatch(t, expectedProducts, product)

	// Product-services should include all 10 services.
	expectedServices := []string{
		cryptoutilSharedMagic.OTLPServiceJoseJA, cryptoutilSharedMagic.OTLPServiceSMIM, cryptoutilSharedMagic.OTLPServicePKICA, cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.OTLPServiceIdentityIDP, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.OTLPServiceIdentityRS, cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
	}
	require.ElementsMatch(t, expectedServices, productService)

	// Infrastructure should have at least one entry.
	require.NotEmpty(t, infrastructure, "infrastructure deployments must not be empty")

	// Template should have exactly one entry.
	require.Len(t, template, 1)
	require.Equal(t, cryptoutilSharedMagic.SkeletonTemplateServiceName, template[0])
}

// TestGetExpectedDeploymentsContents validates the full deployments contents map.
func TestGetExpectedDeploymentsContents(t *testing.T) {
	t.Parallel()

	contents := GetExpectedDeploymentsContents()
	require.NotEmpty(t, contents, "expected deployments contents must not be empty")

	// Verify product-service entries exist with compose.yml.
	services := []string{
		cryptoutilSharedMagic.OTLPServiceJoseJA, cryptoutilSharedMagic.OTLPServiceSMIM, cryptoutilSharedMagic.OTLPServicePKICA, cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.OTLPServiceIdentityIDP, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.OTLPServiceIdentityRS, cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
	}

	for _, svc := range services {
		key := svc + "/compose.yml"
		status, ok := contents[key]
		require.True(t, ok, "expected deployment entry %q not found", key)
		require.Equal(t, RequiredFileStatus, status, "compose.yml for %s should be REQUIRED", svc)
	}

	// Verify template compose.yml.
	templateKey := "template/compose.yml"
	status, ok := contents[templateKey]
	require.True(t, ok, "expected template compose.yml entry not found")
	require.Equal(t, RequiredFileStatus, status, "template compose.yml should be REQUIRED")

	// Verify only valid statuses are used.
	validStatuses := map[string]bool{
		RequiredFileStatus:  true,
		OptionalFileStatus:  true,
		ForbiddenFileStatus: true,
	}

	for path, fileStatus := range contents {
		require.True(t, validStatuses[fileStatus],
			"invalid status %q for path %q", fileStatus, path)
	}
}

// TestAddProductServiceFiles validates that addProductServiceFiles populates expected entries.
func TestAddProductServiceFiles(t *testing.T) {
	t.Parallel()

	contents := make(map[string]string)
	addProductServiceFiles(&contents, cryptoutilSharedMagic.OTLPServiceJoseJA)

	// Should have compose.yml as required.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/compose.yml"])

	// Should have Dockerfile as required.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/Dockerfile"])

	// Should have hash-pepper secret (hyphenated, no service prefix).
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/hash-pepper-v3.secret"])

	// Should have unseal secrets (hyphenated, no service prefix).
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/unseal-1of5.secret"])

	// Should have postgres secrets (hyphenated, no service prefix).
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/postgres-username.secret"])

	// Should have browser/service credential secrets.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/browser-password.secret"])
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/service-password.secret"])

	// Should have config files.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/config/jose-ja-app-common.yml"])

	// Should have forbidden deprecated files.
	require.Equal(t, ForbiddenFileStatus, contents["jose-ja/config/demo-seed.yml"])
}

// TestAddInfrastructureFiles validates infrastructure file entries.
func TestAddInfrastructureFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		infraName string
		wantExtra string
	}{
		{
			name:      "observability has no extras",
			infraName: "observability",
		},
		{
			name:      "shared-postgres has init-db.sql",
			infraName: "shared-postgres",
			wantExtra: "shared-postgres/init-db.sql",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			contents := make(map[string]string)
			addInfrastructureFiles(&contents, tc.infraName)

			// All infra should have compose.yml.
			require.Equal(t, RequiredFileStatus, contents[tc.infraName+"/compose.yml"])

			if tc.wantExtra != "" {
				_, ok := contents[tc.wantExtra]
				require.True(t, ok, "expected extra file %q not found", tc.wantExtra)
			}
		})
	}
}

// TestAddTemplateFiles validates template file entries.
func TestAddTemplateFiles(t *testing.T) {
	t.Parallel()

	contents := make(map[string]string)
	addTemplateFiles(&contents)

	// Should have template compose.yml as required.
	require.Equal(t, RequiredFileStatus, contents["template/compose.yml"])

	// Should have template secrets (hyphenated, no prefix).
	require.Equal(t, RequiredFileStatus, contents["template/secrets/hash-pepper-v3.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/unseal-1of5.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/postgres-username.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/postgres-url.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/browser-password.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/service-password.secret"])
}

// TestAddSuiteProductSecrets validates suite/product-level secret entries.
func TestAddSuiteProductSecrets(t *testing.T) {
	t.Parallel()

	contents := make(map[string]string)
	addSuiteProductSecrets(&contents, cryptoutilSharedMagic.DefaultOTLPServiceDefault)

	// Should have hash-pepper secret.
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/hash-pepper-v3.secret"])

	// Should have unseal secrets.
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/unseal-1of5.secret"])
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/unseal-5of5.secret"])

	// Should have postgres secrets.
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/postgres-username.secret"])
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/postgres-url.secret"])

	// Browser/service credentials use .secret.never at suite/product level.
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/browser-password.secret.never"])
	require.Equal(t, RequiredFileStatus, contents[cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/secrets/service-password.secret.never"])

	// Should have exactly 14 entries (no extras).
	require.Len(t, contents, 14)
}
