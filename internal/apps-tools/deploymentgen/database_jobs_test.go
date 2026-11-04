// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	json "encoding/json"
	"strings"
	"testing"

	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func databaseMaterialFixture(t *testing.T, inventory Inventory, input LifecycleInput) []LifecycleSecret {
	t.Helper()

	secrets := make([]LifecycleSecret, 0, len(input.Secrets))

	for _, source := range input.Secrets {
		data := map[string][]byte{}
		for key := range source.Files {
			data[key] = []byte("protected-fixture")
		}

		data["dsn"] = []byte("postgres://explicit-host:5432/explicit_database")
		secrets = append(secrets, LifecycleSecret{Metadata: LifecycleMetadata{Name: source.Name, Namespace: source.Namespace}, Data: data})
	}

	for _, job := range input.Jobs {
		if job.Name == lifecycleInfraPKIName || strings.HasPrefix(job.Name, lifecyclePKICommand) {
			continue
		}

		args, secret, err := db.JobArguments(job.Name)
		require.NoError(t, err)

		operation := strings.TrimPrefix(args[1], "--operation=")
		connection := db.Connection{
			DSNFile: db.SecretDirectory + secret + "/dsn", UsernameFile: db.SecretDirectory + secret + "/username",
			PasswordFile: db.SecretDirectory + secret + "/password", CAFile: db.SecretDirectory + secret + "/ca.crt",
			CertificateFile: db.SecretDirectory + secret + "/tls.crt", KeyFile: db.SecretDirectory + secret + "/tls.key",
		}
		manifest := db.Input{Backend: db.Postgres}

		if operation == db.Initialize {
			manifest.LeaderAdmin = connection
			manifest.FollowerAdmin = connection
		}

		for _, release := range inventory.Releases {
			if operation == db.Migrate && release.Namespace != job.Namespace {
				continue
			}

			for _, database := range release.Databases {
				if operation == db.Migrate && database.PSID != strings.TrimSuffix(job.Name, "-migrate") {
					continue
				}

				target := db.Target{PSID: database.PSID, Database: database.Name, Role: database.Role, Leader: connection}
				if operation != db.Migrate {
					target.Publication = database.Publication
					target.Subscription = database.Subscription
					target.Slot = "slot_" + strings.TrimPrefix(database.Subscription, "sub_")
					target.Follower = connection
				}

				if operation == db.Replicate {
					target.SubscriberTLS = db.SubscriberTLS{CA: "/server/ca.crt", Certificate: "/server/tls.crt", Key: "/server/tls.key"}
				}

				manifest.Targets = append(manifest.Targets, target)
			}
		}

		raw, err := json.Marshal(manifest)
		require.NoError(t, err)

		for index := range secrets {
			if secrets[index].Metadata.Namespace == job.Namespace && secrets[index].Metadata.Name == secret {
				secrets[index].Data[db.InputName] = raw

				break
			}
		}
	}

	return secrets
}

func TestDatabaseProtectedManifestGraphAndMountBinding(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		selector string
		mutate   func(*LifecycleInput, []LifecycleSecret)
	}{
		{"service PG", testServiceSMKMSPGSelector, nil},
		{"full selected catalog", "", nil},
		{"SQLite omits shared jobs", "s:sm-kms:mem", nil},
		{"unknown job", testServiceSMKMSPGSelector, func(i *LifecycleInput, _ []LifecycleSecret) {
			i.Jobs = append(i.Jobs, LifecycleJobInput{Name: testLifecycleUnknownName})
		}},
		{"missing manifest", testServiceSMKMSPGSelector, func(_ *LifecycleInput, s []LifecycleSecret) {
			for index := range s {
				delete(s[index].Data, db.InputName)
			}
		}},
		{"missing protected key", testServiceSMKMSPGSelector, func(_ *LifecycleInput, s []LifecycleSecret) {
			for index := range s {
				delete(s[index].Data, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey)
			}
		}},
		{"unmounted source", testServiceSMKMSPGSelector, func(i *LifecycleInput, _ []LifecycleSecret) {
			for index := range i.Jobs {
				i.Jobs[index].SecretNames = nil
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, inventory, input := lifecycleDeployFixture(t, test.selector)
			secrets := databaseMaterialFixture(t, inventory, input)

			if test.mutate != nil {
				test.mutate(&input, secrets)
			}

			err := ValidateDatabaseJobMaterial(inventory, input, secrets)
			if test.mutate != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseJobRequiresSupportedSuiteArguments(t *testing.T) {
	t.Parallel()

	const (
		absentMount    = "absent mount"
		absentSource   = "absent source"
		fixtureCommand = "fixture command"
		validCommand   = "valid"
	)
	for _, test := range []string{fixtureCommand, absentMount, absentSource, validCommand} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()

			args, secret, err := db.JobArguments("sm-kms-migrate")
			require.NoError(t, err)

			job := LifecycleJobInput{Name: "sm-kms-migrate", Namespace: testReleaseNamespace, Args: args, SecretNames: []string{secret}}
			sources := []LifecycleSecretSource{{Name: secret, Namespace: job.Namespace, Files: map[string]string{db.InputName: "protected/lifecycle.json"}}}

			switch test {
			case fixtureCommand:
				job.Args = []string{testLifecycleLegacyProduct, testLifecycleLegacyService, testLifecycleInitCommand}
			case absentMount:
				job.SecretNames = nil
			case absentSource:
				sources = nil
			}

			err = validateDatabaseJobCommand(job, sources)
			if test == validCommand {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
