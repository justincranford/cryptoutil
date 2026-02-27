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
		"identity/", "identity/authz", "identity/idp", "identity/rp", "identity/rs", "identity/spa", "identity/policies/", "identity/profiles/",
		"jose/", "jose/ja/",
		"pki/", "pki/ca/",
		"sm/", "sm/im/", "sm/kms/",
		"skeleton/", "skeleton/template/",
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

	// Suite should contain exactly cryptoutil-suite.
	require.Len(t, suite, 1)
	require.Equal(t, "cryptoutil-suite", suite[0])

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
	require.Equal(t, "template", template[0])
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

	// Should have hash_pepper secret.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/jose-ja-hash_pepper.secret"])

	// Should have unseal secrets.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/jose-ja-unseal_1of5.secret"])

	// Should have postgres secrets.
	require.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/jose-ja-postgres_username.secret"])

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
		{
			name:      "shared-citus has init-citus.sql",
			infraName: "shared-citus",
			wantExtra: "shared-citus/init-citus.sql",
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

	// Should have template secrets.
	require.Equal(t, RequiredFileStatus, contents["template/secrets/hash_pepper_v3.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/unseal_1of5.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/postgres_username.secret"])
	require.Equal(t, RequiredFileStatus, contents["template/secrets/postgres_url.secret"])
}
