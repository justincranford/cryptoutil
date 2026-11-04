// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductionMigrationSelection(t *testing.T) {
	t.Parallel()

	for _, psid := range []string{psidAuthz, psidIDP, psidRP, psidRS, psidSPA, psidSkeleton, psidKMS, testUnowned} {
		t.Run(psid, func(t *testing.T) {
			t.Parallel()

			files, err := Migrations(psid)
			if psid == testUnowned {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			entries, err := fs.ReadDir(files, "migrations")
			require.NoError(t, err)
			require.Greater(t, len(entries), 8)
		})
	}
}
