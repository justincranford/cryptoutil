// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	replicaTestSchema           = `{"migration":[{"version":2000,"dirty":false}],"columns":[],"constraints":[]}`
	replicaTestEmptySchema      = `{"migration":null,"columns":[],"constraints":[]}`
	replicaTablesQuery          = "SELECT tablename FROM pg_tables"
	replicaCountQuery           = "SELECT count(*) FROM pg_tables"
	testReplicaBegin            = "BEGIN"
	testReplicaCommit           = "COMMIT"
	testReplicaRollback         = "ROLLBACK"
	testReplicaCursorTable      = "schema_migrations"
	testReplicaTenantTable      = "tenants"
	testReplicaExistsQuery      = "SELECT EXISTS(SELECT 1 FROM"
	testReplicaBeginError       = "begin error"
	testReplicaCountError       = "count error"
	testReplicaMigrationError   = "migration error"
	testReplicaFreshMismatch    = "fresh mismatch"
	testReplicaTruncateError    = "truncate error"
	testReplicaCommitError      = "commit error"
	testReplicaExistingMismatch = "existing mismatch"
	testReplicaExistingStale    = "existing populated stale"
	testReplicaFresh            = "fresh schema"
	testReplicaVerify           = "verify populated matching"
	testReplicaStaged           = "verify staged empty"
)

func TestReplicaBootstrapFreshOwnershipAndVerifyOnly(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testReplicaFresh, testReplicaVerify, testReplicaStaged, testReplicaBeginError, testReplicaCountError, testReplicaMigrationError, testReplicaFreshMismatch, testReplicaTruncateError, testReplicaCommitError, testReplicaExistingMismatch, testReplicaExistingStale} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			steps := []sqlStep{
				{query: testReplicaBegin},
				{query: replicaCountQuery, rows: [][]driver.Value{{int64(0)}}},
				{query: testSchemaQuery, rows: [][]driver.Value{{replicaTestSchema}}},
				{query: testDirtyQuery, rows: [][]driver.Value{{false}}},
				{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaCursorTable}, {testReplicaTenantTable}}},
				{query: `TRUNCATE TABLE "public"."schema_migrations","public"."tenants" RESTART IDENTITY`},
				{query: testReplicaCommit},
			}

			switch test {
			case testReplicaBeginError:
				steps = []sqlStep{{query: testReplicaBegin, err: true}}
			case testReplicaCountError:
				steps = []sqlStep{{query: testReplicaBegin}, {query: replicaCountQuery, err: true}, {query: testReplicaRollback}}
			case testReplicaMigrationError:
				steps = append(steps[:2], sqlStep{query: testReplicaRollback})
			case testReplicaFreshMismatch:
				steps[2].rows = [][]driver.Value{{replicaTestEmptySchema}}
				steps = append(steps[:4], sqlStep{query: testReplicaRollback})
			case testReplicaTruncateError:
				steps[5].err = true
				steps[6].query = testReplicaRollback
			case testReplicaCommitError:
				steps[6].err = true
			case testReplicaVerify:
				steps = []sqlStep{{query: testReplicaBegin}, {query: replicaCountQuery, rows: [][]driver.Value{{int64(2)}}}, {query: testSchemaQuery, rows: [][]driver.Value{{replicaTestSchema}}}, {query: testReplicaCommit}}
			case testReplicaStaged:
				steps = []sqlStep{
					{query: testReplicaBegin},
					{query: replicaCountQuery, rows: [][]driver.Value{{int64(2)}}},
					{query: testSchemaQuery, rows: [][]driver.Value{{replicaTestEmptySchema}}},
					{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaCursorTable}}},
					{query: testReplicaExistsQuery, rows: [][]driver.Value{{false}}},
					{query: testReplicaCommit},
				}
			case testReplicaExistingMismatch:
				steps = []sqlStep{{query: testReplicaBegin}, {query: replicaCountQuery, rows: [][]driver.Value{{int64(2)}}}, {query: testSchemaQuery, rows: [][]driver.Value{{`{"columns":[{"unowned":"column"}],"constraints":[]}`}}}, {query: testReplicaRollback}}
			case testReplicaExistingStale:
				steps = []sqlStep{
					{query: testReplicaBegin},
					{query: replicaCountQuery, rows: [][]driver.Value{{int64(2)}}},
					{query: testSchemaQuery, rows: [][]driver.Value{{replicaTestEmptySchema}}},
					{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaCursorTable}}},
					{query: testReplicaExistsQuery, rows: [][]driver.Value{{true}}},
					{query: testReplicaRollback},
				}
			}

			client, script := sqlFixture(t, steps)
			applied := false
			err := client.prepareReplicaWith(context.Background(), psidKMS, replicaTestSchema, func(_ context.Context, tx *sql.Tx, psid string) error {
				applied = true

				require.NotNil(t, tx)
				require.Equal(t, psidKMS, psid)

				if test == testReplicaMigrationError {
					return errors.New("DO-NOT-LOG")
				}

				return nil
			})

			happy := test == testReplicaFresh || strings.HasPrefix(test, "verify ")
			if happy {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
			}

			if strings.HasPrefix(test, "verify ") || strings.HasPrefix(test, "existing ") {
				require.False(t, applied, "existing schemas must never be re-migrated or truncated")
			}

			require.Empty(t, script.steps)
			require.NoError(t, client.Close())
		})
	}
}

func TestReplicaBootstrapRejectsUnsupportedInputs(t *testing.T) {
	t.Parallel()

	var absent context.Context

	client := &SQLClient{DB: sharedSQLite, backend: SQLite}
	require.Error(t, client.PrepareReplica(absent, psidKMS, replicaTestSchema))
	require.Error(t, client.PrepareReplica(context.Background(), psidKMS, replicaTestSchema))
	client.backend = Postgres
	require.Error(t, client.PrepareReplica(context.Background(), "unowned-service", replicaTestSchema))

	for _, input := range []string{cryptoutilSharedMagic.DevSetupTestInvalidName, "null", `{"columns":[]}`} {
		_, err := replicaSchemaShape(input)
		require.Error(t, err)
	}
}

func TestInitialReplicaCopyNeverOverwritesRows(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		steps []sqlStep
		want  bool
	}{
		{cryptoutilSharedMagic.EMPTY, []sqlStep{{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaCursorTable}}}, {query: testReplicaExistsQuery, rows: [][]driver.Value{{false}}}}, false},
		{"populated", []sqlStep{{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaTenantTable}}}, {query: testReplicaExistsQuery, rows: [][]driver.Value{{true}}}}, true},
		{"inventory error", []sqlStep{{query: replicaTablesQuery, err: true}}, true},
		{"table identity error", []sqlStep{{query: replicaTablesQuery, rows: [][]driver.Value{{nil}}}}, true},
		{"data read error", []sqlStep{{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaTenantTable}}}, {query: testReplicaExistsQuery, err: true}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client, script := sqlFixture(t, test.steps)

			err := client.requireEmptyReplica(context.Background())
			if test.want {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Empty(t, script.steps)
			require.NoError(t, client.Close())
		})
	}
}
