// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"strings"
	"testing"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestOrdinaryDatabaseRolesMatchEveryCanonicalPostgreSQLGraphRelease(t *testing.T) {
	t.Parallel()
	model, _ := fixture(t)
	model.Run = DevelopmentSmokeRun
	inventory, err := Build(model, "")
	require.NoError(t, err)

	for _, release := range inventory.Releases {
		if len(release.Databases) == 0 {
			continue
		}

		t.Run(release.Key, func(t *testing.T) {
			t.Parallel()

			connection := db.Connection{
				DSNFile: "/run/cryptoutil/secrets/owned-db/dsn", UsernameFile: "/run/cryptoutil/secrets/owned-db/username",
				PasswordFile: "/run/cryptoutil/secrets/owned-db/password", CAFile: "/run/cryptoutil/secrets/owned-db/ca.crt",
				CertificateFile: "/run/cryptoutil/secrets/owned-db/tls.crt", KeyFile: "/run/cryptoutil/secrets/owned-db/tls.key",
			}

			input := db.Input{Backend: db.Postgres, LeaderAdmin: connection, FollowerAdmin: connection}
			for _, database := range release.Databases {
				input.Targets = append(input.Targets, db.Target{
					PSID: database.PSID, Database: database.Name, Role: database.Role,
					Publication: database.Publication, Subscription: database.Subscription,
					Slot: "slot_" + strings.TrimPrefix(database.Subscription, "sub_"), Leader: connection, Follower: connection,
				})
			}

			bound, err := BuildRoleProvisionInput(model, release.Key, input)
			require.NoError(t, err)
			require.NoError(t, bound.Validate())

			for index, target := range bound.Database.Targets {
				require.Equal(t, release.Databases[index].Role, target.Role)
				require.Equal(t, release.Databases[index].Name, target.Database)
			}

			_, err = BuildRoleProvisionInput(Model{}, release.Key, input)
			require.Error(t, err)
			_, err = BuildRoleProvisionInput(model, "", input)
			require.Error(t, err)

			forged := input
			forged.Targets = append([]db.Target(nil), input.Targets...)
			forged.Targets[0].Role = "not_the_graph_owner"
			_, err = BuildRoleProvisionInput(model, release.Key, forged)
			require.Error(t, err)

			incomplete := input
			incomplete.Targets = append([]db.Target(nil), input.Targets[:len(input.Targets)-1]...)
			_, err = BuildRoleProvisionInput(model, release.Key, incomplete)
			require.Error(t, err)

			parts := strings.Split(release.Key, ":")
			if parts[0] == tierProduct && parts[1] != "sm" {
				foreign := input
				foreign.Targets = append([]db.Target(nil), input.Targets...)
				target := &foreign.Targets[0]
				target.PSID = cryptoutilSharedMagic.IMJWTIssuer
				stem := strings.ReplaceAll(strings.Join([]string{model.Run, parts[0], parts[1], parts[2], target.PSID}, "_"), "-", "_")
				target.Role, target.Database = "u_"+stem, "d_"+stem
				target.Publication, target.Subscription, target.Slot = "pub_"+stem, "sub_"+stem, "slot_"+stem

				require.NoError(t, (db.RoleProvisionInput{Run: model.Run, Tier: parts[0], Scope: parts[1], Variant: parts[2], Database: foreign}).Validate())
				_, err = BuildRoleProvisionInput(model, release.Key, foreign)
				require.ErrorContains(t, err, "outside owning graph")
			}
		})
	}
}
