// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplicaInventoryFailuresRollbackBeforeClearing(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		queryError bool
		readError  bool
		rows       [][]driver.Value
		existing   bool
	}{
		{name: "inventory query", queryError: true},
		{name: "inventory scan", rows: [][]driver.Value{{nil}}},
		{name: "inventory empty"},
		{name: "inventory oversized", rows: oversizedReplicaRows()},
		{name: "inventory read", readError: true},
		{name: "existing schema query", existing: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			steps := []sqlStep{
				{query: testReplicaBegin},
				{query: replicaCountQuery, rows: [][]driver.Value{{int64(0)}}},
				{query: testSchemaQuery, rows: [][]driver.Value{{replicaTestSchema}}},
				{query: testDirtyQuery, rows: [][]driver.Value{{false}}},
				{query: replicaTablesQuery, rows: test.rows, err: test.queryError},
				{query: testReplicaRollback},
			}
			if test.existing {
				steps = []sqlStep{
					{query: testReplicaBegin},
					{query: replicaCountQuery, rows: [][]driver.Value{{int64(1)}}},
					{query: testSchemaQuery, err: true},
					{query: testReplicaRollback},
				}
			}

			client, script := sqlFixture(t, steps)
			if test.readError {
				script.inventoryReadErr = errors.New("DO-NOT-LOG")
			}

			applied := false
			err := client.prepareReplicaWith(context.Background(), psidKMS, replicaTestSchema, func(context.Context, *sql.Tx, string) error {
				applied = true

				return nil
			})
			require.Error(t, err)
			require.NotContains(t, err.Error(), "DO-NOT-LOG")
			require.Equal(t, !test.existing, applied)
			require.Empty(t, script.steps, "rollback must occur without any truncate or commit")
			require.NoError(t, client.Close())
		})
	}
}

func oversizedReplicaRows() [][]driver.Value {
	rows := make([][]driver.Value, maximumReplicaTables+1)
	for index := range rows {
		rows[index] = []driver.Value{testReplicaTenantTable}
	}

	return rows
}

func TestInitialReplicaInventoryBoundAndReadErrors(t *testing.T) {
	t.Parallel()

	for _, readError := range []bool{true, false} {
		t.Run(map[bool]string{true: "read failure", false: "table bound"}[readError], func(t *testing.T) {
			t.Parallel()

			client, script := sqlFixture(t, []sqlStep{{query: replicaTablesQuery, rows: oversizedReplicaRows()}})
			if readError {
				script.inventoryReadErr = errors.New("DO-NOT-LOG")
			}

			err := client.requireEmptyReplica(context.Background())
			require.Error(t, err)
			require.NotContains(t, err.Error(), "DO-NOT-LOG")
			require.Empty(t, script.steps, "no table contents may be read for an incomplete inventory")
			require.NoError(t, client.Close())
		})
	}
}

func TestReplicaMalformedExpectationDoesNotBeginTransaction(t *testing.T) {
	t.Parallel()

	client, script := sqlFixture(t, nil)
	err := client.prepareReplicaWith(context.Background(), psidKMS, cryptoutilSharedMagic.DevSetupTestInvalidName, func(context.Context, *sql.Tx, string) error {
		t.Fatal("invalid expectation must never reach migrations")

		return nil
	})
	require.ErrorContains(t, err, "schema description rejected")
	require.Empty(t, script.steps)
	require.NoError(t, client.Close())
}
