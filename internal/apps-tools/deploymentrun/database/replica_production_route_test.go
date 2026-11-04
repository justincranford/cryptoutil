// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplicaProductionRouteUsesActualMigrationDriver(t *testing.T) {
	t.Parallel()
	client, script := sqlFixture(t, []sqlStep{
		{query: testReplicaBegin},
		{query: replicaCountQuery, rows: [][]driver.Value{{int64(0)}}},
		{query: "SELECT CURRENT_DATABASE()", err: true},
		{query: testReplicaRollback},
	})
	client.backend = Postgres
	require.ErrorContains(t, client.PrepareReplica(context.Background(), psidKMS, replicaTestSchema), "transactional replica migrations rejected")
	require.Empty(t, script.steps, "the public suite route must reach the production PostgreSQL driver, not an injected fixture migration")
	require.NoError(t, client.Close())
}
