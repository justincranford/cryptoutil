// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	json "encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type sqlStep struct {
	query string
	rows  [][]driver.Value
	err   bool
}
type sqlConnector struct {
	t                *testing.T
	steps            []sqlStep
	closed           bool
	inventoryReadErr error
	queries          []string
}

func (c *sqlConnector) Connect(context.Context) (driver.Conn, error) {
	return &sqlConnection{script: c}, nil
}
func (c *sqlConnector) Driver() driver.Driver { return sqlScriptDriver{} }

type sqlScriptDriver struct{}

func (sqlScriptDriver) Open(string) (driver.Conn, error) { return nil, errors.New("connector only") }

type sqlConnection struct{ script *sqlConnector }

func (c *sqlConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("context queries only")
}

func (c *sqlConnection) Begin() (driver.Tx, error) {
	if _, err := c.step(testReplicaBegin); err != nil {
		return nil, err
	}

	return &scriptTransaction{connection: c}, nil
}

func (c *sqlConnection) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}

type scriptTransaction struct{ connection *sqlConnection }

func (tx *scriptTransaction) Commit() error {
	_, err := tx.connection.step(testReplicaCommit)

	return err
}

func (tx *scriptTransaction) Rollback() error {
	_, err := tx.connection.step(testReplicaRollback)

	return err
}

func (c *sqlConnection) Close() error {
	c.script.closed = true

	return nil
}

func (c *sqlConnection) step(query string) (sqlStep, error) {
	c.script.queries = append(c.script.queries, query)
	if strings.Contains(query, testPublicationQuery) {
		require.Contains(c.script.t, query, "puballtables AND pubinsert AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot")
		require.Contains(c.script.t, query, "pubowner=(SELECT oid FROM pg_roles WHERE rolname=current_user)")
	}

	require.NotEmpty(c.script.t, c.script.steps, "unexpected SQL operation")
	step := c.script.steps[0]
	c.script.steps = c.script.steps[1:]
	require.Contains(c.script.t, query, step.query)

	if step.err {
		return step, errors.New("DO-NOT-LOG")
	}

	return step, nil
}

func (c *sqlConnection) QueryContext(ctx context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	step, err := c.step(query)
	if err != nil {
		return nil, err
	}

	rows := &sqlRows{data: step.rows}
	if strings.Contains(query, replicaTablesQuery) {
		rows.readErr = c.script.inventoryReadErr
	}

	return rows, nil
}

func (c *sqlConnection) ExecContext(ctx context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	_, err := c.step(query)

	return driver.RowsAffected(1), err
}

type sqlRows struct {
	data    [][]driver.Value
	readErr error
}

func (r *sqlRows) Columns() []string {
	if len(r.data) == 0 {
		return []string{"missing"}
	}

	names := make([]string, len(r.data[0]))
	for index := range names {
		names[index] = "value"
	}

	return names
}
func (*sqlRows) Close() error { return nil }
func (r *sqlRows) Next(values []driver.Value) error {
	if r.readErr != nil {
		return r.readErr
	}

	if len(r.data) == 0 {
		return io.EOF
	}

	copy(values, r.data[0])
	r.data = r.data[1:]

	return nil
}

func sqlFixture(t *testing.T, steps []sqlStep) (*SQLClient, *sqlConnector) {
	t.Helper()
	script := &sqlConnector{t: t, steps: steps}
	db := sql.OpenDB(script)

	return &SQLClient{DB: db, backend: Postgres}, script
}

func TestSQLDatabaseIdentityAndCreation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		steps     []sqlStep
		operation string
		want      bool
	}{
		{testRequire, []sqlStep{{testCurrentDatabaseQuery, [][]driver.Value{{testDatabase, testRole}}, false}}, testRequire, false},
		{"wrong database", []sqlStep{{testCurrentDatabaseQuery, [][]driver.Value{{testWrong, testRole}}, false}}, testRequire, true},
		{"changed database owner", []sqlStep{{testCurrentDatabaseQuery, [][]driver.Value{{testDatabase, testWrong}}, false}}, testRequire, true},
		{"query failure", []sqlStep{{query: testCurrentDatabaseQuery, err: true}}, testRequire, true},
		{"existing owned", []sqlStep{{testRolesQuery, [][]driver.Value{{true}}, false}, {testDatabaseQuery, [][]driver.Value{{testRole}}, false}}, testCreate, false},
		{"existing wrong owner", []sqlStep{{testRolesQuery, [][]driver.Value{{true}}, false}, {testDatabaseQuery, [][]driver.Value{{testWrong}}, false}}, testCreate, true},
		{"absent role", []sqlStep{{testRolesQuery, [][]driver.Value{{false}}, false}}, testCreate, true},
		{"role lookup failure", []sqlStep{{query: testRolesQuery, err: true}}, testCreate, true},
		{"database lookup failure", []sqlStep{{testRolesQuery, [][]driver.Value{{true}}, false}, {query: testDatabaseQuery, err: true}}, testCreate, true},
		{testCreate, []sqlStep{{testRolesQuery, [][]driver.Value{{true}}, false}, {query: testDatabaseQuery}, {query: `CREATE DATABASE "d_sm_kms" OWNER "u_sm_kms"`}}, testCreate, false},
		{testCreateFailure, []sqlStep{{testRolesQuery, [][]driver.Value{{true}}, false}, {query: testDatabaseQuery}, {query: "CREATE DATABASE", err: true}}, testCreate, true},
		{"invalid identity", nil, testInvalid, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client, script := sqlFixture(t, test.steps)

			var err error

			switch test.operation {
			case testRequire:
				err = client.RequireDatabase(context.Background(), testDatabase, testRole)
			case testCreate:
				err = client.EnsureDatabase(context.Background(), testDatabase, testRole)
			default:
				err = client.EnsureDatabase(context.Background(), testSQLInjection, testRole)
			}

			if test.want {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
			} else {
				require.NoError(t, err)
			}

			require.Empty(t, script.steps)
			require.NoError(t, client.Close())
		})
	}
}

func TestSQLPublicationGuards(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		steps   []sqlStep
		create  bool
		invalid bool
		want    bool
	}{
		{"existing", []sqlStep{{testPublicationQuery, [][]driver.Value{{true, true}}, false}}, true, false, false},
		{testCreate, []sqlStep{{query: testPublicationQuery}, {query: `CREATE PUBLICATION "pub_sm_kms" FOR ALL TABLES`}}, true, false, false},
		{testCreateFailure, []sqlStep{{query: testPublicationQuery}, {query: "CREATE PUBLICATION", err: true}}, true, false, true},
		{"contract collision", []sqlStep{{testPublicationQuery, [][]driver.Value{{false, true}}, false}}, true, false, true},
		{"publication owner collision", []sqlStep{{testPublicationQuery, [][]driver.Value{{true, false}}, false}}, true, false, true},
		{"lookup failure", []sqlStep{{query: testPublicationQuery, err: true}}, true, false, true},
		{"require missing", []sqlStep{{query: testPublicationQuery}}, false, false, true},
		{"require present", []sqlStep{{testPublicationQuery, [][]driver.Value{{true, true}}, false}}, false, false, false},
		{"require foreign publication", []sqlStep{{testPublicationQuery, [][]driver.Value{{true, false}}, false}}, false, false, true},
		{testInvalid, nil, true, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client, script := sqlFixture(t, test.steps)

			name := testPublication
			if test.invalid {
				name = testSQLInjection
			}

			var err error
			if test.create {
				err = client.EnsurePublication(context.Background(), name)
			} else {
				err = client.RequirePublication(context.Background(), name)
			}

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

func TestSQLSchemaParityInspection(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		steps []sqlStep
		want  bool
	}{
		{"clean", []sqlStep{{testSchemaQuery, [][]driver.Value{{testSchema}}, false}, {testDirtyQuery, [][]driver.Value{{false}}, false}}, false},
		{"dirty", []sqlStep{{testSchemaQuery, [][]driver.Value{{testSchema}}, false}, {testDirtyQuery, [][]driver.Value{{true}}, false}}, true},
		{"schema query error", []sqlStep{{query: testSchemaQuery, err: true}}, true},
		{"cursor query error", []sqlStep{{testSchemaQuery, [][]driver.Value{{testSchema}}, false}, {query: testDirtyQuery, err: true}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client, script := sqlFixture(t, test.steps)

			got, err := client.Schema(context.Background())
			if test.want {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testSchema, got)
			}

			require.Empty(t, script.steps)
			require.NoError(t, client.Close())
		})
	}
}

func TestSQLGuardedSubscription(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testCreate, testVerify, testDisabled, testWrongOwner, testChangedTransport, testChangedSlot, testChangedPublication, testInvalidPublicationJSON, testLookupError, testCreateError, testReadinessError, testWaitCancel, testWrongSource} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			target := inputFixture(Replicate).Targets[0]
			sourceURL, err := url.Parse("postgres://replication-user:DO-NOT-LOG@leader:5432/d_sm_kms")
			require.NoError(t, err)

			leader := &SQLClient{sourceURL: sourceURL}
			expected := *sourceURL
			expected.RawQuery = url.Values{"sslmode": {"verify-full"}, "sslrootcert": {target.SubscriberTLS.CA}, "sslcert": {target.SubscriberTLS.Certificate}, "sslkey": {target.SubscriberTLS.Key}, "connect_timeout": {"10"}, "ssl_min_protocol_version": {"TLSv1.3"}}.Encode()
			pub, err := json.Marshal([]string{target.Publication})
			require.NoError(t, err)

			state := []driver.Value{true, true, expected.String(), target.Slot, string(pub)}

			switch test {
			case testDisabled:
				state[0] = false
			case testWrongOwner:
				state[1] = false
			case testChangedTransport:
				state[2] = testChanged
			case testChangedSlot:
				state[3] = testChanged
			case testChangedPublication:
				state[4] = `["changed"]`
			case testInvalidPublicationJSON:
				state[4] = testInvalid
			}

			steps := []sqlStep{{testSubscriptionQuery, [][]driver.Value{state}, false}}
			if test == testCreate || test == testCreateError {
				steps = []sqlStep{
					{query: testSubscriptionQuery},
					{query: replicaTablesQuery, rows: [][]driver.Value{{testReplicaCursorTable}}},
					{query: testReplicaExistsQuery, rows: [][]driver.Value{{false}}},
					{query: "CREATE SUBSCRIPTION", err: test == testCreateError},
				}
			}

			if test == testLookupError {
				steps = []sqlStep{{query: testSubscriptionQuery, err: true}}
			}

			if test == testCreate || test == testVerify || test == testReadinessError || test == testWaitCancel {
				steps = append(steps, sqlStep{query: "pg_stat_subscription", rows: [][]driver.Value{{test != testWaitCancel}}, err: test == testReadinessError})
			}

			if test == testWrongSource {
				steps = nil
			}

			client, script := sqlFixture(t, steps)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if test == testWaitCancel {
				// Cancel after the first not-ready result without timing races.
				ctx = &cancelAfterQueries{Context: ctx, script: script, cancel: cancel}
			}

			var source Client = leader
			if test == testWrongSource {
				source = &fakeClient{}
			}

			err = client.EnsureSubscription(ctx, target, source)

			for _, query := range script.queries {
				if strings.Contains(query, "pg_stat_subscription") {
					require.Equal(t, 3, strings.Count(query, "s.subdbid=(SELECT oid FROM pg_database WHERE datname=current_database())"))
					require.Contains(t, query, "s.oid=st.subid")
				} else if strings.Contains(query, "FROM pg_subscription") {
					require.Contains(t, query, "subdbid=(SELECT oid FROM pg_database WHERE datname=current_database())")
				}
			}

			if test == testCreate || test == testVerify {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
			}

			require.Empty(t, script.steps)
			require.NoError(t, client.Close())
		})
	}
}

type cancelAfterQueries struct {
	context.Context
	script *sqlConnector
	cancel context.CancelFunc
}

func (c *cancelAfterQueries) Done() <-chan struct{} {
	if len(c.script.steps) == 0 {
		c.cancel()
	}

	return c.Context.Done()
}

func TestSQLLiteralEscaping(t *testing.T) {
	t.Parallel()
	require.Equal(t, `'a''b'`, sqlLiteral("a'b"))
	require.True(t, strings.HasPrefix(sqlLiteral("value"), "'"))
}
