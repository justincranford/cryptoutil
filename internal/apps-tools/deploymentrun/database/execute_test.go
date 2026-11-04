// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func connectionFixture(secret string) Connection {
	base := SecretDirectory + secret + "/"

	return Connection{
		DSNFile: base + "dsn", UsernameFile: base + "username", PasswordFile: base + "password",
		CAFile: base + "ca.crt", CertificateFile: base + cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate, KeyFile: base + cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey,
	}
}

func inputFixture(operation string) Input {
	target := Target{PSID: psidKMS, Database: testDatabase, Role: testRole, Leader: connectionFixture("sm-kms-migration-db")}
	if operation != Migrate {
		target.Follower = connectionFixture("postgres-follower-bootstrap")
		target.Publication = testPublication
		target.Subscription = "sub_sm_kms"
		target.Slot = "slot_sm_kms"
	}

	if operation == Replicate {
		target.SubscriberTLS = SubscriberTLS{CA: "/server/ca.crt", Certificate: "/server/tls.crt", Key: "/server/tls.key"}
	}

	input := Input{Backend: Postgres, Targets: []Target{target}}
	if operation == Initialize {
		input.LeaderAdmin = connectionFixture("postgres-bootstrap")
		input.FollowerAdmin = connectionFixture("postgres-follower-bootstrap")
	}

	return input
}

type fakeClient struct {
	name   string
	calls  *[]string
	fail   string
	schema string
}

func (c *fakeClient) record(operation string) error {
	value := c.name + ":" + operation

	*c.calls = append(*c.calls, value)
	if value == c.fail {
		return errors.New("DO-NOT-LOG")
	}

	return nil
}
func (c *fakeClient) Close() error { return c.record(testClose) }
func (c *fakeClient) RequireDatabase(context.Context, string, string) error {
	return c.record(testRequire)
}

func (c *fakeClient) EnsureDatabase(context.Context, string, string) error {
	return c.record("create-db")
}

func (c *fakeClient) EnsurePublication(context.Context, string) error { return c.record("create-pub") }

func (c *fakeClient) RequirePublication(context.Context, string) error {
	return c.record("require-pub")
}
func (c *fakeClient) Migrate(context.Context, string, string) error { return c.record(Migrate) }
func (c *fakeClient) Schema(context.Context) (string, error)        { return c.schema, c.record("schema") }

func (c *fakeClient) PrepareReplica(_ context.Context, _ string, expected string) error {
	if err := c.record("migrate"); err != nil {
		return err
	}

	if expected != c.schema {
		return errors.New("schema bootstrap mismatch")
	}

	return nil
}

func (c *fakeClient) EnsureSubscription(context.Context, Target, Client) error {
	return c.record("subscription")
}

func TestOrderedDatabaseLifecycle(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{Initialize, Migrate, Replicate} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			input := inputFixture(operation)
			calls := []string{}
			err := Execute(context.Background(), operation, input, func(_ context.Context, connection Connection, _ string) (Client, error) {
				name := testLeader
				if reflect.DeepEqual(connection, input.FollowerAdmin) && operation == Initialize {
					name = testFollowerAdmin
				} else if reflect.DeepEqual(connection, input.LeaderAdmin) && operation == Initialize {
					name = testLeaderAdmin
				} else if reflect.DeepEqual(connection, input.Targets[0].Follower) {
					name = testFollower
				}

				calls = append(calls, "open:"+name)

				return &fakeClient{name: name, calls: &calls, schema: testSameSchema}, nil
			})
			require.NoError(t, err)

			expected := map[string][]string{
				Migrate:    {testOpenLeader, testLeaderRequire, testLeaderMigration, testLeaderClose},
				Initialize: {testOpenLeaderAdmin, testOpenFollowerAdmin, testLeaderAdminCreate, testFollowerAdminCreate, testOpenLeader, testLeaderRequire, testLeaderCreatePublication, "leader-admin:close", "follower-admin:close", testLeaderClose},
				Replicate:  {testOpenLeader, testLeaderRequire, testOpenFollower, testFollowerRequire, testLeaderRequirePublication, testLeaderSchema, testFollowerMigration, testFollowerSubscription, testFollowerSchema, testLeaderClose, "follower:close"},
			}
			require.Equal(t, expected[operation], calls)
		})
	}
}

func TestDatabaseLifecycleStopsAtEveryBarrier(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{Initialize, Migrate, Replicate} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			barriers := map[string][]string{
				Migrate:    {testOpenLeader, testLeaderRequire, testLeaderMigration},
				Initialize: {testOpenLeaderAdmin, testOpenFollowerAdmin, testLeaderAdminCreate, testFollowerAdminCreate, testOpenLeader, testLeaderRequire, testLeaderCreatePublication},
				Replicate:  {testOpenLeader, testLeaderRequire, testOpenFollower, testFollowerRequire, testLeaderRequirePublication, testLeaderSchema, testFollowerMigration, testFollowerSchema, testFollowerSubscription, testSchemaMismatch},
			}
			for _, fail := range barriers[operation] {
				t.Run(fail, func(t *testing.T) {
					t.Parallel()

					input := inputFixture(operation)
					calls := []string{}
					err := Execute(context.Background(), operation, input, func(_ context.Context, connection Connection, _ string) (Client, error) {
						name := testLeader
						if operation == Initialize && connection == input.FollowerAdmin {
							name = testFollowerAdmin
						} else if operation == Initialize && connection == input.LeaderAdmin {
							name = testLeaderAdmin
						} else if connection == input.Targets[0].Follower {
							name = testFollower
						}

						calls = append(calls, "open:"+name)
						if fail == "open:"+name {
							return nil, errors.New("DO-NOT-LOG")
						}

						schema := testSameSchema
						if fail == testSchemaMismatch && name == testFollower {
							schema = testChanged
						}

						return &fakeClient{name: name, calls: &calls, fail: fail, schema: schema}, nil
					})
					require.Error(t, err)
					require.NotContains(t, err.Error(), "DO-NOT-LOG")

					if fail == testSchemaMismatch {
						require.NotContains(t, calls, testFollowerSubscription)
					}
				})
			}
		})
	}
}
