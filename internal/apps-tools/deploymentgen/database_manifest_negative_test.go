// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	json "encoding/json"
	"testing"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDatabaseManifestRejectsForgedCatalog(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		edit func(*db.Input)
	}{
		{"SQLite shared Job", func(i *db.Input) { i.Backend = db.SQLite }},
		{"invalid connection", func(i *db.Input) { i.Targets[0].Leader.CAFile = "" }},
		{"empty catalog", func(i *db.Input) { i.Targets = nil }},
		{"extra valid catalog", func(i *db.Input) {
			extra := i.Targets[0]
			extra.Database = "extra_catalog"
			i.Targets = append(i.Targets, extra)
		}},
		{"wrong database", func(i *db.Input) { i.Targets[0].Database = "unowned_database" }},
		{"wrong service", func(i *db.Input) { i.Targets[0].PSID = cryptoutilSharedMagic.OTLPServiceIdentityIDP }},
		{"wrong role", func(i *db.Input) { i.Targets[0].Role = "unowned_role" }},
		{"wrong publication", func(i *db.Input) { i.Targets[0].Publication = "unowned_publication" }},
		{"wrong subscription", func(i *db.Input) { i.Targets[0].Subscription = "unowned_subscription" }},
		{"wrong slot", func(i *db.Input) { i.Targets[0].Slot = "unowned_slot" }},
		{"escaping owned mount", func(i *db.Input) { i.Targets[0].Leader.DSNFile = db.SecretDirectory + "unowned/dsn" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
			secrets := databaseMaterialFixture(t, inventory, input)

			for index := range secrets {
				if secrets[index].Metadata.Name != "postgres-follower-bootstrap" {
					continue
				}

				manifest, err := db.Decode(secrets[index].Data[db.InputName])
				require.NoError(t, err)
				test.edit(&manifest)
				raw, err := json.Marshal(manifest)
				require.NoError(t, err)

				secrets[index].Data[db.InputName] = raw
			}

			require.Error(t, ValidateDatabaseJobMaterial(inventory, input, secrets))
		})
	}
}
