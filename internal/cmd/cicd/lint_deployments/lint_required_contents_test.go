package lint_deployments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetExpectedConfigsContents validates the config contents map.
func TestGetExpectedConfigsContents(t *testing.T) {
	t.Parallel()

	contents := GetExpectedConfigsContents()

	require.NotEmpty(t, contents, "expected configs contents must not be empty")

	// All entries should be OPTIONAL (configs are less strict than deployments).
	for path, status := range contents {
		assert.Equal(t, OptionalFileStatus, status, "config path %s should be OPTIONAL", path)
	}

	// Verify expected config directories exist.
	expectedDirs := []string{
		"cryptoutil/",
		"cipher/", "cipher/im/",
		"identity/", "identity/authz", "identity/idp", "identity/rp", "identity/rs", "identity/spa", "identity/policies/", "identity/profiles/",
		"jose/", "jose/ja/",
		"pki/", "pki/ca/",
		"sm/", "sm/kms/",
	}

	for _, dir := range expectedDirs {
		_, ok := contents[dir]
		assert.True(t, ok, "expected config directory %q not found", dir)
	}

	assert.Len(t, contents, len(expectedDirs), "unexpected extra entries in configs contents")
}

// TestGetDeploymentDirectories validates deployment directory lists.
func TestGetDeploymentDirectories(t *testing.T) {
	t.Parallel()

	suite, product, productService, infrastructure, template := GetDeploymentDirectories()

	// Suite should contain exactly cryptoutil.
	require.Len(t, suite, 1)
	assert.Equal(t, "cryptoutil", suite[0])

	// Products should include all 5 products.
	expectedProducts := []string{"identity", "sm", "pki", "cipher", "jose"}
	assert.ElementsMatch(t, expectedProducts, product)

	// Product-services should include all 9 services.
	expectedServices := []string{
		"jose-ja", "cipher-im", "pki-ca", "sm-kms",
		"identity-authz", "identity-idp", "identity-rp", "identity-rs", "identity-spa",
	}
	assert.ElementsMatch(t, expectedServices, productService)

	// Infrastructure should have at least one entry.
	require.NotEmpty(t, infrastructure, "infrastructure deployments must not be empty")

	// Template should have exactly one entry.
	require.Len(t, template, 1)
	assert.Equal(t, "template", template[0])
}

// TestGetExpectedDeploymentsContents validates the full deployments contents map.
func TestGetExpectedDeploymentsContents(t *testing.T) {
	t.Parallel()

	contents := GetExpectedDeploymentsContents()
	require.NotEmpty(t, contents, "expected deployments contents must not be empty")

	// Verify product-service entries exist with compose.yml.
	services := []string{
		"jose-ja", "cipher-im", "pki-ca", "sm-kms",
		"identity-authz", "identity-idp", "identity-rp", "identity-rs", "identity-spa",
	}

	for _, svc := range services {
		key := svc + "/compose.yml"
		status, ok := contents[key]
		assert.True(t, ok, "expected deployment entry %q not found", key)
		assert.Equal(t, RequiredFileStatus, status, "compose.yml for %s should be REQUIRED", svc)
	}

	// Verify template compose.yml.
	templateKey := "template/compose.yml"
	status, ok := contents[templateKey]
	assert.True(t, ok, "expected template compose.yml entry not found")
	assert.Equal(t, RequiredFileStatus, status, "template compose.yml should be REQUIRED")

	// Verify only valid statuses are used.
	validStatuses := map[string]bool{
		RequiredFileStatus:  true,
		OptionalFileStatus:  true,
		ForbiddenFileStatus: true,
	}

	for path, fileStatus := range contents {
		assert.True(t, validStatuses[fileStatus],
			"invalid status %q for path %q", fileStatus, path)
	}
}

// TestAddProductServiceFiles validates that addProductServiceFiles populates expected entries.
func TestAddProductServiceFiles(t *testing.T) {
	t.Parallel()

	contents := make(map[string]string)
	addProductServiceFiles(&contents, "jose-ja")

	// Should have compose.yml as required.
	assert.Equal(t, RequiredFileStatus, contents["jose-ja/compose.yml"])

	// Should have Dockerfile as required.
	assert.Equal(t, RequiredFileStatus, contents["jose-ja/Dockerfile"])

	// Should have hash_pepper secret.
	assert.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/jose-ja-hash_pepper.secret"])

	// Should have unseal secrets.
	assert.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/jose-ja-unseal_1of5.secret"])

	// Should have postgres secrets.
	assert.Equal(t, RequiredFileStatus, contents["jose-ja/secrets/jose-ja-postgres_username.secret"])

	// Should have config files.
	assert.Equal(t, RequiredFileStatus, contents["jose-ja/config/jose-ja-app-common.yml"])

	// Should have forbidden deprecated files.
	assert.Equal(t, ForbiddenFileStatus, contents["jose-ja/config/demo-seed.yml"])
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
			assert.Equal(t, RequiredFileStatus, contents[tc.infraName+"/compose.yml"])

			if tc.wantExtra != "" {
				_, ok := contents[tc.wantExtra]
				assert.True(t, ok, "expected extra file %q not found", tc.wantExtra)
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
	assert.Equal(t, RequiredFileStatus, contents["template/compose.yml"])

	// Should have template secrets.
	assert.Equal(t, RequiredFileStatus, contents["template/secrets/hash_pepper_v3.secret"])
	assert.Equal(t, RequiredFileStatus, contents["template/secrets/unseal_1of5.secret"])
	assert.Equal(t, RequiredFileStatus, contents["template/secrets/postgres_username.secret"])
	assert.Equal(t, RequiredFileStatus, contents["template/secrets/postgres_url.secret"])
}
