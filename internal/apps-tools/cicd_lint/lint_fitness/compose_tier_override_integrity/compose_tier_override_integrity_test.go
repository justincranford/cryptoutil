// Copyright (c) 2025-2026 Justin Cranford.
package compose_tier_override_integrity_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessComposeTierOverrideIntegrity "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/compose_tier_override_integrity"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testUsername = "username"
	testPassword = "password"
)

func testLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissions))
}

func setupTierFixture(t *testing.T, root string) {
	t.Helper()

	composeTemplate := `services:
  app: {}
secrets:
  postgres-url.secret: {file: ./secrets/postgres-url.secret}
  postgres-username.secret: {file: ./secrets/postgres-username.secret}
  postgres-password.secret: {file: ./secrets/postgres-password.secret}
  postgres-database.secret: {file: ./secrets/postgres-database.secret}
`

	for _, tier := range []string{cryptoutilSharedMagic.DefaultOTLPServiceDefault, "sm"} {
		writeFile(t, filepath.Join(root, "deployments", tier, cryptoutilSharedMagic.COMPOSE_YML), composeTemplate)
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-username.secret"), tier+"_user")
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-password.secret"), tier+"_pass")
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-database.secret"), tier+"_db")
		writeFile(t, filepath.Join(root, "deployments", tier, "secrets", "postgres-url.secret"), "postgres://"+tier+"_user:"+tier+"_pass@postgres-leader:5432/"+tier+"_db?sslmode=disable")
	}
}

func TestCheckInDir_Valid(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.NoError(t, err)
}

func TestCheckInDir_ForbiddenBuilderService(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	writeFile(t, filepath.Join(root, "deployments", cryptoutilSharedMagic.DefaultOTLPServiceDefault, cryptoutilSharedMagic.COMPOSE_YML),
		fmt.Sprintf("services:\n  %s: {}\nsecrets:\n  postgres-url.secret: {file: ./secrets/postgres-url.secret}\n  postgres-username.secret: {file: ./secrets/postgres-username.secret}\n  postgres-password.secret: {file: ./secrets/postgres-password.secret}\n  postgres-database.secret: {file: ./secrets/postgres-database.secret}\n", cryptoutilSharedMagic.DockerJobBuilderCryptoutil))

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.Error(t, err)
	require.Contains(t, err.Error(), cryptoutilSharedMagic.DockerJobBuilderCryptoutil)
}

func TestCheckInDir_MissingPostgresSecretOverride(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	writeFile(t, filepath.Join(root, "deployments", "sm", cryptoutilSharedMagic.COMPOSE_YML), `services:
  app: {}
secrets:
  postgres-url.secret: {file: ./secrets/postgres-url.secret}
  postgres-username.secret: {file: ./secrets/postgres-username.secret}
`)

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required postgres secret override")
}

func TestCheckInDir_PostgresURLNotSynced(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)

	writeFile(t, filepath.Join(root, "deployments", "sm", "secrets", "postgres-url.secret"), "postgres://wrong:wrong@postgres-leader:5432/sm_db")

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match")
}

func TestCheckInDir_PostgresURLValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{name: "invalid url", url: "://bad-url", wantErr: "not a valid URL"},
		{name: "wrong scheme", url: "mysql://sm_user:sm_pass@postgres-leader:5432/sm_db", wantErr: "must use postgres/postgresql scheme"},
		{name: "missing credentials", url: "postgres://postgres-leader:5432/sm_db", wantErr: "is missing credentials"},
		{name: "missing password", url: "postgres://sm_user@postgres-leader:5432/sm_db", wantErr: "is missing password"},
		{name: "username mismatch", url: "postgres://other:sm_pass@postgres-leader:5432/sm_db", wantErr: testUsername},
		{name: "password mismatch", url: "postgres://sm_user:other@postgres-leader:5432/sm_db", wantErr: testPassword},
		{name: "database mismatch", url: "postgres://sm_user:sm_pass@postgres-leader:5432/other", wantErr: cryptoutilSharedMagic.RealmStorageTypeDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			setupTierFixture(t, root)
			writeFile(t, filepath.Join(root, "deployments", "sm", "secrets", "postgres-url.secret"), tt.url)

			err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestCheckInDir_ComposeReadFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents string
		remove   bool
		wantErr  string
	}{
		{name: "missing compose file", remove: true, wantErr: "cannot read"},
		{name: "invalid compose yaml", contents: "services: [", wantErr: "cannot parse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			setupTierFixture(t, root)
			composePath := filepath.Join(root, "deployments", "sm", cryptoutilSharedMagic.COMPOSE_YML)

			if tt.remove {
				require.NoError(t, os.Remove(composePath))
			} else {
				writeFile(t, composePath, tt.contents)
			}

			err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestCheckInDir_EmptyComposeMapsAndMissingSecrets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	setupTierFixture(t, root)
	writeFile(t, filepath.Join(root, "deployments", "sm", cryptoutilSharedMagic.COMPOSE_YML), "{}\n")

	err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
	require.ErrorContains(t, err, "missing required postgres secret override")
}

func TestCheckInDir_MissingSecretValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		wantErr  string
	}{
		{name: testUsername, filename: "postgres-username.secret", wantErr: "cannot read postgres username secret"},
		{name: testPassword, filename: "postgres-password.secret", wantErr: "cannot read postgres password secret"},
		{name: cryptoutilSharedMagic.RealmStorageTypeDatabase, filename: "postgres-database.secret", wantErr: "cannot read postgres database secret"},
		{name: "url", filename: "postgres-url.secret", wantErr: "cannot read postgres URL secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			setupTierFixture(t, root)
			secretPath := filepath.Join(root, "deployments", "sm", "secrets", tt.filename)
			require.NoError(t, os.Remove(secretPath))

			err := lintFitnessComposeTierOverrideIntegrity.CheckInDir(testLogger(), root)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
