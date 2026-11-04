// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMigrationBorrowHonorsOwningAndCallerContext(t *testing.T) {
	t.Parallel()

	for _, ownerCancelled := range []bool{true, false} {
		t.Run(map[bool]string{true: "owning operation", false: "caller"}[ownerCancelled], func(t *testing.T) {
			t.Parallel()

			db := sql.OpenDB(transportTestConnector{connection: &transportTestConnection{}})

			db.SetMaxOpenConns(1)
			defer func() { _ = db.Close() }()

			// Exhaust the pool; a Background constructor lease must not outlive
			// the owning deadline, and a live owner must respect caller cancel.
			held, err := db.Conn(context.Background())
			require.NoError(t, err)

			defer func() { _ = held.Close() }()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			owner, caller := context.Background(), ctx
			if ownerCancelled {
				owner, caller = ctx, context.Background()
			}

			connector := &migrationConnector{ctx: owner, db: db}
			result := make(chan error, 1)

			go func() {
				lease, err := connector.Connect(caller)
				if lease != nil {
					_ = lease.Close()
				}

				result <- err
			}()

			require.Eventually(t, func() bool { return db.Stats().WaitCount > 0 }, 5*time.Second, time.Millisecond, "cancellation must interrupt a real owner-pool wait")
			cancel()

			select {
			case err := <-result:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(5 * time.Second):
				// Always unblock even a regressed implementation before failing.
				require.NoError(t, held.Close())
				require.ErrorIs(t, <-result, context.Canceled)

				t.Fatal("migration lease ignored cancellation while owner pool was exhausted")
			}

			require.Equal(t, 1, db.Stats().InUse, "the test still owns its sole pool lease")
			require.Empty(t, connector.connections, "a cancelled borrow cannot register a migration lease")
			_, err = connector.Connect(ctx)
			require.ErrorIs(t, err, context.Canceled, "a previously cancelled caller must fail before a lease")
		})
	}
}
