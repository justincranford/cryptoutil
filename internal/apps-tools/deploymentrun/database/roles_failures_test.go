// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"bytes"
	"context"
	"database/sql"
	json "encoding/json"
	"errors"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type roleRejectedWriter struct{}

func (roleRejectedWriter) Write([]byte) (int, error) { return 0, errors.New("PRIVATE") }

func TestOrdinaryRoleManifestRejectsMalformedOrForgedEncoding(t *testing.T) {
	t.Parallel()

	for mode := range testRoleCase8 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(ordinaryRoleFixture())
			require.NoError(t, err)

			switch mode {
			case 0:
				raw = nil
			case 1:
				raw = []byte("{")
			case 2:
				raw = []byte(`{"run":"one","run":"two"}`)
			case testRoleCase3:
				raw = []byte(`{"unknown":true}`)
			case testRoleCase4:
				raw = append(raw, []byte(` {}`)...)
			case testRoleCase5:
				raw = make([]byte, maximumInput+1)
			case testRoleCase6:
				raw = []byte(`{}`)
			case testRoleCase7:
				raw = []byte(`[]`)
			}

			_, err = DecodeRoleProvision(raw)
			require.Error(t, err)
		})
	}
}

func TestOrdinaryRoleSQLRejectsInvalidDependenciesAndTransportFailures(t *testing.T) {
	t.Parallel()

	for mode := range testRoleCase7 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			steps := []sqlStep{{query: testReplicaBegin}}

			switch mode {
			case 0:
				steps[0].err = true
			case 1:
				steps = append(steps, sqlStep{query: testRoleAdvisoryQuery, err: true}, sqlStep{query: testReplicaRollback})
			case 2:
				steps = append(steps, sqlStep{query: testRoleAdvisoryQuery}, sqlStep{query: testOrdinaryLookup, err: true}, sqlStep{query: testReplicaRollback})
			default:
				steps = nil
			}

			script := &sqlConnector{t: t, steps: steps}

			database := sql.OpenDB(script)
			defer func() { require.NoError(t, database.Close()) }()

			client := &SQLClient{DB: database, backend: Postgres}
			password := []byte(testOrdinarySQLPassword)
			run := testOrdinaryRun

			switch mode {
			case testRoleCase3:
				client.backend = SQLite
			case testRoleCase4:
				run = testRoleForeignRun
			case testRoleCase5:
				password = []byte("one\nsecond")
			case testRoleCase6:
				client = nil
			}

			err := client.EnsureOrdinaryRole(t.Context(), run, testOrdinaryRole, password)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "DO-NOT-LOG")
			require.Empty(t, script.steps)
		})
	}
}

func TestOrdinaryRoleCommandReportsOutputFailureAndSupportedHelp(t *testing.T) {
	t.Parallel()

	for mode := range testRoleCase3 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			var stderr bytes.Buffer

			if mode < 2 {
				var stdout io.Writer = &bytes.Buffer{}
				if mode == 1 {
					stdout = roleRejectedWriter{}
				}

				code := Main([]string{ProvisionRoles, testOrdinaryHelp}, nil, stdout, &stderr)
				require.Equal(t, mode, code)

				return
			}

			input := ordinaryRoleFixture()
			input.Run = testRoleForeignRun
			calls := 0
			err := ProvisionOrdinaryRoles(t.Context(), input, func(_ context.Context, _ Connection, _ string) (Client, error) {
				calls++

				return nil, nil
			}, func(string) ([]byte, error) { return nil, nil })
			require.Error(t, err)
			require.Zero(t, calls)
		})
	}
}
