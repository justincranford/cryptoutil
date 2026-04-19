// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestE2E_PostgreSQLMTLS verifies that the sm-kms app connects to PostgreSQL
// using mTLS (client certificate, Cat 12) by querying pg_stat_ssl from inside
// the postgres-leader container.
//
// This test runs `docker exec` to query pg_stat_ssl for connections from the
// sm-kms app instances (postgres-1 and postgres-2).
func TestE2E_PostgreSQLMTLS(t *testing.T) {
	t.Parallel()

	// Query pg_stat_ssl to verify SSL connections from sm-kms app instances.
	// We look for connections where ssl=true and client_dn is non-empty (client cert present).
	query := "SELECT application_name, ssl, client_dn FROM pg_stat_ssl " +
		"JOIN pg_stat_activity ON pg_stat_ssl.pid = pg_stat_activity.pid " +
		"WHERE application_name LIKE 'sm-kms%' AND ssl = true AND client_dn IS NOT NULL;"

	args := composeManager.BuildDockerExecArgs(cryptoutilSharedMagic.KMSPostgresLeaderContainer,
		"psql", "--username=sm_kms_database_user", "--dbname=sm_kms_database",
		"--tuples-only", "--command", query)

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("docker", args...) //nolint:gosec // docker exec with known args
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "docker exec psql pg_stat_ssl should succeed\nstderr: %s", stderr.String())

	output := stdout.String()
	t.Logf("pg_stat_ssl output:\n%s", output)

	// At least one sm-kms app connection should show ssl=true with client cert.
	require.True(t, strings.Contains(output, "t"), "Expected ssl=true (t) in pg_stat_ssl output, got:\n%s", output)
}
