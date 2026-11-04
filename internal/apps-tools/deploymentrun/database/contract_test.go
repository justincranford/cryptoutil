// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	json "encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrictDatabaseInput(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"", strings.Repeat("x", maximumInput+1), "{", `{"backend":"postgres","backend":"sqlite"}`,
		`{"unknown":true}`, `{} {}`, `{"targets":[{"psid":"a","psid":"b"}]}`,
	} {
		t.Run(raw[:min(len(raw), 24)], func(t *testing.T) { t.Parallel(); _, err := Decode([]byte(raw)); require.Error(t, err) })
	}

	raw, err := json.Marshal(inputFixture(Migrate))
	require.NoError(t, err)
	got, err := Decode(raw)
	require.NoError(t, err)
	require.Equal(t, inputFixture(Migrate), got)
}

func TestDatabaseExplicitContractRejections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		edit      func(*Input)
	}{
		{"unknown operation", testUnknown, nil},
		{testUnknownBackend, Migrate, func(i *Input) { i.Backend = testUnknown }},
		{"empty catalog", Migrate, func(i *Input) { i.Targets = nil }},
		{"oversized catalog", Migrate, func(i *Input) { i.Targets = make([]Target, 129) }},
		{"SQLite init", Initialize, func(i *Input) { i.Backend = SQLite }},
		{"multi migration", Migrate, func(i *Input) { i.Targets = append(i.Targets, i.Targets[0]) }},
		{testUnknownService, Migrate, func(i *Input) { i.Targets[0].PSID = testUnknown }},
		{"bad DB name", Migrate, func(i *Input) { i.Targets[0].Database = testSQLInjection }},
		{"duplicate DB", Replicate, func(i *Input) { i.Targets = append(i.Targets, i.Targets[0]) }},
		{"missing DSN", Migrate, func(i *Input) { i.Targets[0].Leader.DSNFile = "" }},
		{"missing TLS", Migrate, func(i *Input) { i.Targets[0].Leader.CAFile = "" }},
		{"migration follower", Migrate, func(i *Input) { i.Targets[0].Follower = connectionFixture(testFollower) }},
		{"publication name", Replicate, func(i *Input) { i.Targets[0].Publication = testSQLInjection }},
		{"missing follower TLS", Replicate, func(i *Input) { i.Targets[0].Follower.KeyFile = "" }},
		{"server relative path", Replicate, func(i *Input) { i.Targets[0].SubscriberTLS.CA = testRelative }},
		{"server path traversal", Replicate, func(i *Input) { i.Targets[0].SubscriberTLS.Key = "/server/../key" }},
		{"init server transport", Initialize, func(i *Input) { i.Targets[0].SubscriberTLS.Key = "/server/key" }},
		{"init leader admin", Initialize, func(i *Input) { i.LeaderAdmin.DSNFile = "" }},
		{"init follower admin", Initialize, func(i *Input) { i.FollowerAdmin.DSNFile = "" }},
		{"migration admin", Migrate, func(i *Input) { i.LeaderAdmin = connectionFixture("admin") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			input := inputFixture(test.operation)
			if test.edit != nil {
				test.edit(&input)
			}

			require.Error(t, input.Validate(test.operation))
		})
	}

	for _, operation := range []string{Migrate, Initialize, Replicate} {
		t.Run(operation, func(t *testing.T) { t.Parallel(); require.NoError(t, inputFixture(operation).Validate(operation)) })
	}
}

func TestProtectedDatabasePaths(t *testing.T) {
	t.Parallel()

	for _, value := range []string{testRelative, "/etc/passwd", SecretDirectory + "one/../key", SecretDirectory + "one/two/three", SecretDirectory + "one/..", SecretDirectory + "one/key\\escape"} {
		t.Run(value, func(t *testing.T) { t.Parallel(); require.False(t, protectedPath(value)) })
	}

	require.True(t, protectedPath(SecretDirectory+"owned/key"))
	connection := Connection{DSNFile: SecretDirectory + "owned/dsn"}
	require.NoError(t, connection.Validate(SQLite))
	connection.KeyFile = SecretDirectory + "owned/key"
	require.Error(t, connection.Validate(SQLite))

	input := inputFixture(Replicate)
	require.Len(t, input.Files(), 12)
}

func TestDatabaseExecutionValidationAndCleanup(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testInvalid, testCancelled, testNilOpener, testNilClient, testCloseError} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			input := inputFixture(Migrate)
			if test == testInvalid {
				input.Backend = testUnknown
			}

			if test == testCancelled {
				cancel()
			}

			calls := []string{}

			var open Open = func(context.Context, Connection, string) (Client, error) {
				if test == testNilClient {
					return nil, nil
				}

				return &fakeClient{name: testLeader, calls: &calls, fail: testLeaderClose}, nil
			}
			if test == testNilOpener {
				open = nil
			}

			err := Execute(ctx, Migrate, input, open)
			require.Error(t, err)

			if test == testCloseError {
				require.Contains(t, calls, testLeaderMigration)
				require.ErrorContains(t, err, "closure")
			}
		})
	}
}

func TestDatabaseJobArguments(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"db-init", "replication", "sm-kms-migrate", "unowned-migrate", testUnowned} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			args, secret, err := JobArguments(name)
			if strings.HasPrefix(name, testUnowned) {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Len(t, args, 3)
			require.Equal(t, Command, args[0])
			require.Contains(t, args[2], secret+"/"+InputName)
		})
	}
}

func TestApplyDatabaseErrorIsRedacted(t *testing.T) {
	t.Parallel()
	client, script := sqlFixture(t, []sqlStep{{query: "CREATE TABLE IF NOT EXISTS schema_migrations", err: true}})
	err := Apply(context.Background(), client.DB, psidKMS, SQLite)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "DO-NOT-LOG")
	// The migration driver may use a different introspection statement across
	// versions; this test specifically asserts the current owning driver path.
	require.Empty(t, script.steps)
	require.NoError(t, client.Close())
}
