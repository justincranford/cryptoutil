// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"database/sql"
	"database/sql/driver"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrdinaryRoleCreationIsTransactionalAndCollisionGuarded(t *testing.T) {
	t.Parallel()

	for mode := range testRoleCase8 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			role := testOrdinaryRole
			comment := "cu.io/run=p2sm1103;cu.io/role=" + role
			steps := []sqlStep{{query: testReplicaBegin}, {query: testRoleAdvisoryQuery}, {query: testOrdinaryLookup, rows: [][]driver.Value{{true, comment}}}}

			switch mode {
			case 1, testRoleCase4, testRoleCase5, testRoleCase6:
				steps[2].rows = nil

				steps = append(steps, sqlStep{query: "CREATE ROLE", err: mode == testRoleCase4})
				if mode != testRoleCase4 {
					steps = append(steps, sqlStep{query: "COMMENT ON ROLE", err: mode == testRoleCase5})
				}

				if mode != testRoleCase4 && mode != testRoleCase5 {
					steps = append(steps, sqlStep{query: testOrdinaryLookup, rows: [][]driver.Value{{mode != testRoleCase6, comment}}})
				}
			case 2:
				steps[2].rows = [][]driver.Value{{true, "another-owner"}}
			case testRoleCase3:
				steps[2].rows = [][]driver.Value{{false, comment}}
			}

			success := mode == 0 || mode == 1
			if success || mode == testRoleCase7 {
				steps = append(steps, sqlStep{query: testReplicaCommit, err: mode == testRoleCase7})
			} else {
				steps = append(steps, sqlStep{query: testReplicaRollback})
			}

			script := &sqlConnector{t: t, steps: steps}

			database := sql.OpenDB(script)
			defer func() { require.NoError(t, database.Close()) }()

			client := &SQLClient{DB: database, backend: Postgres}

			err := client.EnsureOrdinaryRole(t.Context(), testOrdinaryRun, role, []byte(testOrdinarySQLPassword))
			if success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
				require.NotContains(t, err.Error(), testOrdinarySQLPassword)
			}

			require.Empty(t, script.steps)
			require.Contains(t, script.queries[2], "member=pg_roles.oid OR roleid=pg_roles.oid")
		})
	}
}
