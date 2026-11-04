// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsToolsDeploymentgen "cryptoutil/internal/apps-tools/deploymentgen"
	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func executionFixture(t *testing.T, selector string) Request {
	t.Helper()

	raw, err := os.ReadFile("../../../../api/cryptosuite-registry/k8s/k8s-model.json")
	require.NoError(t, err)
	model, err := cryptoutilAppsToolsDeploymentgen.DecodeModel(raw)
	require.NoError(t, err)
	inventory, err := cryptoutilAppsToolsDeploymentgen.Build(model, selector)
	require.NoError(t, err)
	identities, err := cryptoutilAppsToolsDeploymentgen.BuildLifecycleInventory(inventory)
	require.NoError(t, err)

	input := cryptoutilAppsToolsDeploymentgen.LifecycleInput{Version: 1}

	for _, id := range identities.Identities {
		switch id.Kind {
		case secretKind:
			files := map[string]string{}
			for _, key := range []string{testTLSCert, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey, "ca.crt", "username", "password"} {
				files[key] = id.Namespace + "/" + id.Name + "/" + key
			}

			files[db.InputName] = id.Namespace + "/" + id.Name + "/" + db.InputName
			files["pki-verification.json"] = id.Namespace + "/" + id.Name + "/pki-verification.json"

			input.Secrets = append(input.Secrets, cryptoutilAppsToolsDeploymentgen.LifecycleSecretSource{
				Namespace: id.Namespace, Name: id.Name, Type: cryptoutilSharedMagic.KubernetesSecretTypeOpaque, Files: files,
			})
		case jobKind:
			name := strings.TrimSuffix(id.Name, "-"+model.Revision)

			args, manifestSecret, _ := db.JobArguments(name)
			secretNames := []string{manifestSecret}

			if name == infraPKIName || name == testPKIInit {
				for _, source := range identities.Identities {
					if source.Namespace == id.Namespace && source.Kind == secretKind {
						args = []string{"pki-verify", "--input=/run/cryptoutil/secrets/" + source.Name + "/pki-verification.json"}
						secretNames = []string{source.Name}

						break
					}
				}
			}

			input.Jobs = append(input.Jobs, cryptoutilAppsToolsDeploymentgen.LifecycleJobInput{
				Namespace: id.Namespace, Name: name, Image: "registry.example.invalid/cryptoutil@sha256:" + strings.Repeat("a", 64),
				Command: []string{cryptoutilSharedMagic.KubernetesSuiteImageBinary}, Args: args, ActiveDeadlineSeconds: 10,
				SecretNames: secretNames,
				Resources: cryptoutilAppsToolsDeploymentgen.ResourceRequirements{
					Requests: cryptoutilAppsToolsDeploymentgen.ResourceSet{CPU: "10m", Memory: "16Mi"},
					Limits:   cryptoutilAppsToolsDeploymentgen.ResourceSet{CPU: "100m", Memory: "64Mi"},
				},
			})
		}
	}

	model.Revision, err = cryptoutilAppsToolsDeploymentgen.LifecycleRevision(input)
	require.NoError(t, err)
	inventory, err = cryptoutilAppsToolsDeploymentgen.Build(model, selector)
	require.NoError(t, err)

	request := Request{
		Model: model, Inventory: inventory, Input: input,
		VerifyPKI: func(context.Context, cryptoutilAppsToolsDeploymentgen.LifecycleJobValues, Client) error { return nil },
	}
	for _, source := range input.Secrets {
		data := map[string][]byte{}
		for key := range source.Files {
			data[key] = []byte("protected-test-material")
		}

		request.Secrets = append(request.Secrets, cryptoutilAppsToolsDeploymentgen.LifecycleSecret{
			APIVersion: apiVersion, Kind: secretKind, Type: source.Type,
			Metadata: cryptoutilAppsToolsDeploymentgen.LifecycleMetadata{
				Namespace: source.Namespace, Name: source.Name,
				Labels: map[string]string{runLabel: model.Run, "cu.io/owner": importOwner},
			},
			Data: data,
		})
	}

	return request
}

type recordingClient struct {
	calls []string
	hook  func(context.Context, string) error
}

func (c *recordingClient) record(ctx context.Context, call string) error {
	c.calls = append(c.calls, call)
	if c.hook != nil {
		return c.hook(ctx, call)
	}

	return nil
}

func (c *recordingClient) Prepare(ctx context.Context, _ cryptoutilAppsToolsDeploymentgen.LifecycleChartValues) error {
	return c.record(ctx, testPrepare)
}

func (c *recordingClient) Require(ctx context.Context, id Identity) error {
	return c.record(ctx, "require:"+id.Name)
}

func (c *recordingClient) EnsureSecret(ctx context.Context, s cryptoutilAppsToolsDeploymentgen.LifecycleSecret) error {
	return c.record(ctx, "secret:"+s.Metadata.Name)
}

func (c *recordingClient) EnsureBootstrap(ctx context.Context) error {
	return c.record(ctx, bootstrapName)
}

func (c *recordingClient) EnsureJob(ctx context.Context, job cryptoutilAppsToolsDeploymentgen.LifecycleJobValues) error {
	return c.record(ctx, testApplyPrefix+job.Name)
}

func (c *recordingClient) WaitJob(ctx context.Context, id Identity) error {
	return c.record(ctx, testWaitPrefix+id.Name)
}

func (c *recordingClient) WaitDatabase(ctx context.Context, id Identity) error {
	return c.record(ctx, "ready:"+id.Name)
}

func TestExecuteFailuresAreBoundedAndRedacted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, prefix string
		timeout      bool
		cancel       bool
	}{
		{testPrepare, testPrepare, false, false},
		{"prerequisite", "require:", false, false},
		{"import", "secret:", false, false},
		{bootstrapName, bootstrapName, false, false},
		{"apply", testApplyPrefix, false, false},
		{testWait, testWaitPrefix, false, false},
		{cryptoutilSharedMagic.RealmStorageTypeDatabase, "ready:", false, false},
		{"operation deadline", testPrepare, true, false},
		{"job deadline", testWaitPrefix, true, false},
		{"cancel between operations", testPrepare, false, true},
		{"cancel apply", testApplyPrefix, false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testPGSelector)
			policy := Policy{MaxDuration: time.Minute, OperationTimeout: 10 * time.Second}

			if test.timeout {
				if test.prefix == testPrepare {
					policy.OperationTimeout = time.Second
				} else {
					// Exercise the real Job deadline, not an overall timeout
					// expiring during unrelated validation or earlier stages.
					for index := range request.Input.Jobs {
						request.Input.Jobs[index].ActiveDeadlineSeconds = 1
					}

					var err error

					request.Model.Revision, err = cryptoutilAppsToolsDeploymentgen.LifecycleRevision(request.Input)
					require.NoError(t, err)
					request.Inventory, err = cryptoutilAppsToolsDeploymentgen.Build(request.Model, testPGSelector)
					require.NoError(t, err)
				}
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			secretError := errors.New("private payload must not appear in error")
			client := &recordingClient{}
			client.hook = func(ctx context.Context, call string) error {
				require.NotZero(t, len(client.calls))

				_, bounded := ctx.Deadline()
				require.True(t, bounded)

				if !strings.HasPrefix(call, test.prefix) {
					return nil
				}

				if test.cancel {
					cancel()

					return nil
				}

				if test.timeout {
					<-ctx.Done()

					return ctx.Err()
				}

				return secretError
			}
			_, err := Execute(ctx, request, client, policy)
			require.Error(t, err)
			require.NotContains(t, err.Error(), secretError.Error())
			require.False(t, errors.Is(err, secretError))

			var executionErr *ExecutionError
			require.ErrorAs(t, err, &executionErr)

			if test.timeout {
				require.ErrorIs(t, err, context.DeadlineExceeded)
			}

			if test.cancel {
				require.ErrorIs(t, err, context.Canceled)
			}

			require.NotEmpty(t, client.calls, "execution must reach the injected boundary before its deadline")
			require.True(t, strings.HasPrefix(client.calls[len(client.calls)-1], test.prefix))
		})
	}
}

func TestExecuteRejectsBeforeTransport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*Request, *Policy)
	}{
		{"missing duration", func(_ *Request, p *Policy) { p.MaxDuration = 0 }},
		{"excess duration", func(_ *Request, p *Policy) { p.MaxDuration = 25 * time.Hour }},
		{"missing operation bound", func(_ *Request, p *Policy) { p.OperationTimeout = 0 }},
		{"excess operation bound", func(_ *Request, p *Policy) { p.OperationTimeout = 2 * time.Hour }},
		{"revision mismatch", func(r *Request, _ *Policy) { r.Model.Revision = testOther }},
		{"missing imports", func(r *Request, _ *Policy) { r.Secrets = nil }},
		{"duplicate import", func(r *Request, _ *Policy) { r.Secrets[1] = r.Secrets[0] }},
		{"unowned import", func(r *Request, _ *Policy) { r.Secrets[0].Metadata.Labels[runLabel] = testOther }},
		{"wrong owner", func(r *Request, _ *Policy) { r.Secrets[0].Metadata.Labels["cu.io/owner"] = testOther }},
		{"wrong kind", func(r *Request, _ *Policy) { r.Secrets[0].Kind = configMapKind }},
		{"wrong type", func(r *Request, _ *Policy) { r.Secrets[0].Type = "kubernetes.io/tls" }},
		{"wrong version", func(r *Request, _ *Policy) { r.Secrets[0].APIVersion = testOther }},
		{testExtraData, func(r *Request, _ *Policy) { r.Secrets[0].Data["extra"] = []byte("test") }},
		{"missing key", func(r *Request, _ *Policy) { delete(r.Secrets[0].Data, testTLSCert) }},
		{"empty data", func(r *Request, _ *Policy) { r.Secrets[0].Data[testTLSCert] = nil }},
		{"aggregate bound", func(r *Request, _ *Policy) { r.Secrets[0].Data[testTLSCert] = make([]byte, 1<<20) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testSQLiteSelector)
			policy := Policy{MaxDuration: time.Minute, OperationTimeout: time.Second}
			test.edit(&request, &policy)

			client := &recordingClient{}
			_, err := Execute(context.Background(), request, client, policy)
			require.Error(t, err)
			require.Empty(t, client.calls)
		})
	}
}

func TestExecuteCanceledAndNil(t *testing.T) {
	t.Parallel()
	request := executionFixture(t, testSQLiteSelector)
	policy := Policy{MaxDuration: time.Minute, OperationTimeout: time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &recordingClient{}
	_, err := Execute(ctx, request, client, policy)
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, client.calls)

	var nilContext context.Context

	_, err = Execute(nilContext, request, client, policy)
	require.Error(t, err)
	_, err = Execute(context.Background(), request, nil, policy)
	require.Error(t, err)
}

func TestJobStagesRejectsUnknownOrIncomplete(t *testing.T) {
	t.Parallel()

	for _, name := range []string{testUnknown, testMissingInfra, testMissingReleases, testMissingReplication, testMissingMigration, testMissingDB} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testPGSelector)
			values, err := cryptoutilAppsToolsDeploymentgen.BuildLifecycleChartValues(request.Model, request.Inventory, request.Input)
			require.NoError(t, err)

			jobs := values.Jobs[:0]
			for _, job := range values.Jobs {
				base := strings.TrimSuffix(job.Name, "-"+values.Revision)
				if (name == testMissingInfra && base == infraPKIName) ||
					(name == testMissingReleases && base == testPKIInit) ||
					(name == testMissingReplication && base == testReplication) ||
					(name == testMissingMigration && strings.HasSuffix(base, "-migrate")) ||
					(name == testMissingDB && base == databaseInitName) {
					continue
				}

				jobs = append(jobs, job)
			}

			values.Jobs = jobs
			if name == testUnknown {
				values.Jobs[0].Name = "unknown-" + values.Revision
			}

			_, err = jobStages(values, request.Inventory.InfraNamespace)
			require.Error(t, err)
		})
	}
}

func TestExecuteOrdering(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{testPGSelector, testSQLiteSelector, ""} {
		t.Run(selector, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, selector)
			client := &recordingClient{}
			result, err := Execute(context.Background(), request, client, Policy{MaxDuration: time.Minute, OperationTimeout: time.Second})
			require.NoError(t, err)
			require.Equal(t, len(request.Secrets), result.Secrets)
			require.Equal(t, len(request.Input.Jobs), result.Jobs)

			stage := 0

			for index, call := range client.calls {
				if !strings.HasPrefix(call, testApplyPrefix) {
					continue
				}

				require.Equal(t, strings.Replace(call, testApplyPrefix, testWaitPrefix, 1), client.calls[index+1])
				base := strings.TrimSuffix(strings.TrimPrefix(call, testApplyPrefix), "-"+request.Model.Revision)
				next := 0

				switch {
				case base == testPKIInit:
					next = 1
				case base == databaseInitName:
					next = 2

					require.Contains(t, client.calls[:index], "ready:postgres-leader")
					require.Contains(t, client.calls[:index], "ready:postgres-follower")
				case strings.HasSuffix(base, "-migrate"):
					next = 3
				case base == testReplication:
					next = 4
				}

				require.GreaterOrEqual(t, next, stage)
				stage = next
			}

			if selector == testSQLiteSelector {
				require.Equal(t, 1, stage)
			} else {
				require.Equal(t, 4, stage)
			}
		})
	}
}
