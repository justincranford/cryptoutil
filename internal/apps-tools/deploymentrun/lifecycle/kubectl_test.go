// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	json "encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsToolsDeploymentgen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type runnerFunc func(context.Context, string, []string, []byte) ([]byte, error)

func (f runnerFunc) Run(ctx context.Context, command string, args []string, input []byte) ([]byte, error) {
	return f(ctx, command, args, input)
}

func testConfig(t *testing.T) KubectlConfig {
	t.Helper()

	root, err := filepath.Abs("../../../..")
	require.NoError(t, err)

	return KubectlConfig{
		Kubectl: filepath.Join(root, "kubectl"), Helm: filepath.Join(root, "helm"), RepositoryRoot: root,
		Kubeconfig: filepath.Join(root, "operator-kubeconfig"), Context: "owned-context", PollInterval: time.Millisecond,
	}
}

func testDocument(id Identity) map[string]any {
	return map[string]any{
		"apiVersion": apiVersion, testKind: id.Kind,
		"metadata": map[string]any{
			testName: id.Name, "namespace": id.Namespace, "uid": testUID, "generation": float64(1),
			"labels":      map[string]any{runLabel: id.Run, cryptoutilSharedMagic.KubernetesLabelRelease: id.Release},
			"annotations": map[string]any{"cu.io/revision": id.Revision},
		},
	}
}

func encodedDocument(t *testing.T, doc any) []byte {
	t.Helper()

	raw, err := json.Marshal(doc)
	require.NoError(t, err)

	return raw
}

func TestNewKubectlClientValidation(t *testing.T) {
	t.Parallel()

	for _, field := range []string{testKubectl, testHelm, testRoot, testKubeconfig, testContext, testOptionContext, testPoll, testLargePoll, testRunner, testValid} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			config := testConfig(t)

			var runner Runner = runnerFunc(func(context.Context, string, []string, []byte) ([]byte, error) { return nil, nil })

			switch field {
			case testKubectl:
				config.Kubectl = testKubectl
			case testHelm:
				config.Helm = testHelm
			case testRoot:
				config.RepositoryRoot = "."
			case testKubeconfig:
				config.Kubeconfig = testConfigCase
			case testContext:
				config.Context = ""
			case testOptionContext:
				config.Context = "--all"
			case testPoll:
				config.PollInterval = 0
			case testLargePoll:
				config.PollInterval = time.Hour
			case testRunner:
				runner = nil
			}

			client, err := NewKubectlClient(config, runner)
			if field == testValid {
				require.NoError(t, err)
				require.NotContains(t, client.String(), config.Kubeconfig)
			} else {
				require.Error(t, err)
				require.Nil(t, client)
			}
		})
	}
}

func TestKubectlEnsureSecretGuard(t *testing.T) {
	t.Parallel()

	tests := []string{
		testAbsent, testIdentical, testForeignRun, testForeignOwner, testChangedData, testExtraData, testStringData, testDeleted,
		testMissingUID, testGetFailure, testCreateRace, testCreateMalformed, testGetMalformed,
	}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testSQLiteSelector)
			secret := request.Secrets[0]

			var desired map[string]any
			require.NoError(t, json.Unmarshal(encodedDocument(t, secret), &desired))
			existing := encodedDocument(t, secret)

			var doc map[string]any
			require.NoError(t, json.Unmarshal(existing, &doc))
			meta := objectMap(doc["metadata"])
			meta["uid"] = testUID

			switch name {
			case testForeignRun:
				objectMap(meta["labels"])[runLabel] = testAnother
			case testForeignOwner:
				objectMap(meta["labels"])["cu.io/owner"] = testAnother
			case testChangedData:
				objectMap(doc["data"])[testTLSCert] = "changed"
			case testExtraData:
				objectMap(doc["data"])["extra"] = "extra"
			case testStringData:
				doc["stringData"] = map[string]any{testTLSCert: "private"}
			case testDeleted:
				meta["deletionTimestamp"] = "timestamp"
			case testMissingUID:
				delete(meta, "uid")
			}

			creates := 0
			config := testConfig(t)
			client, err := NewKubectlClient(config, runnerFunc(func(_ context.Context, executable string, args []string, input []byte) ([]byte, error) {
				require.Equal(t, config.Kubectl, executable)
				require.Equal(t, []string{"--kubeconfig", config.Kubeconfig, "--context", config.Context}, args[:4])
				require.NotContains(t, strings.Join(args, " "), "protected-test-material")

				switch args[4] {
				case testGet:
					require.Empty(t, input)

					if name == testGetFailure {
						return nil, errors.New("private transport payload")
					}

					if name == testGetMalformed {
						return []byte("{"), nil
					}

					if name == testAbsent || name == testCreateRace || name == testCreateMalformed {
						return nil, nil
					}

					return encodedDocument(t, doc), nil
				case testCreate:
					creates++

					require.Equal(t, []string{testCreate, "-f", "-", "-o", "json"}, args[4:])

					var created map[string]any
					require.NoError(t, json.Unmarshal(input, &created))
					require.Equal(t, desired, created)

					if name == testCreateRace {
						return nil, errors.New("AlreadyExists private payload")
					}

					if name == testCreateMalformed {
						return []byte("{"), nil
					}

					objectMap(created["metadata"])["uid"] = testUID

					return encodedDocument(t, created), nil
				default:
					require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected mutation %s", args[4])

					return nil, nil
				}
			}))
			require.NoError(t, err)

			client.imports = map[string]cryptoutilAppsToolsDeploymentgen.LifecycleSecretSource{
				secret.Metadata.Namespace + "/" + secret.Metadata.Name: request.Input.Secrets[0],
			}
			client.run = request.Model.Run

			err = client.EnsureSecret(context.Background(), secret)
			if name == testAbsent || name == testIdentical {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "private")
			}

			if strings.HasPrefix(name, testCreate) || name == testAbsent {
				require.Equal(t, 1, creates)
			} else {
				require.Zero(t, creates)
			}
		})
	}
}

func TestKubectlWaitGuards(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		testCompleteCase, testFailed, testBothConditions, testWaitThenComplete, testTimeout, testCanceled,
		testReplaced, testMissing, testForeign, testMissingUID, testTransport, testUnensured,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id := Identity{Namespace: testOwned, Kind: jobKind, Name: "job-revision", Run: testRun, Release: testRelease, Revision: testRevision}
			doc := testDocument(id)
			condition := func(value string) map[string]any { return map[string]any{testType: value, testStatus: testTrue} }
			doc[testStatus] = map[string]any{testConditions: []any{condition(testComplete)}}

			switch name {
			case testFailed:
				doc[cryptoutilSharedMagic.StringStatus] = map[string]any{testConditions: []any{condition("Failed")}}
			case testBothConditions:
				doc[cryptoutilSharedMagic.StringStatus] = map[string]any{testConditions: []any{condition(testComplete), condition("Failed")}}
			case testTimeout, testWaitThenComplete:
				doc[cryptoutilSharedMagic.StringStatus] = map[string]any{testConditions: []any{map[string]any{"type": testComplete, cryptoutilSharedMagic.StringStatus: "False"}}}
			case testReplaced:
				objectMap(doc["metadata"])["uid"] = "replacement"
			case testForeign:
				objectMap(objectMap(doc["metadata"])["labels"])[runLabel] = testOther
			case testMissingUID:
				delete(objectMap(doc["metadata"]), "uid")
			}

			calls := 0
			client, err := NewKubectlClient(testConfig(t), runnerFunc(func(context.Context, string, []string, []byte) ([]byte, error) {
				calls++

				if name == testTransport {
					return nil, errors.New("private")
				}

				if name == testMissing {
					return nil, nil
				}

				if name == testWaitThenComplete && calls > 1 {
					doc[cryptoutilSharedMagic.StringStatus] = map[string]any{testConditions: []any{condition(testComplete)}}
				}

				return encodedDocument(t, doc), nil
			}))
			require.NoError(t, err)

			if name != testUnensured {
				client.uids[identityKey(id)] = testUID
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			if name == testCanceled {
				cancel()
			}

			err = client.WaitJob(ctx, id)

			switch name {
			case testCompleteCase, testWaitThenComplete:
				require.NoError(t, err)
			case testFailed, testBothConditions:
				require.ErrorIs(t, err, ErrJobFailed)
			case testTimeout:
				require.ErrorIs(t, err, context.DeadlineExceeded)
			case testCanceled:
				require.ErrorIs(t, err, context.Canceled)
				require.Zero(t, calls)
			case testReplaced, testMissingUID:
				require.ErrorIs(t, err, ErrReplaced)
			default:
				require.Error(t, err)
			}
		})
	}
}

func TestDatabaseReadyAndPrerequisites(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		testReady, testNoReplicas, testStaleGeneration, testPartialReady, testPartialUpdate, testOldRevision,
		testMissingRevision, "namespace active", testNamespaceTerminating, testConfigCase, testForeign,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id := Identity{Namespace: testOwned, Kind: statefulSetKind, Name: testPostgresLeader, Run: testRun, Release: testRelease}
			if strings.HasPrefix(name, "namespace") {
				id.Kind, id.Namespace = namespaceKind, ""
			}

			if name == testConfigCase {
				id.Kind = configMapKind
			}

			doc := testDocument(id)
			doc[testSpec] = map[string]any{testReplicas: float64(1)}
			doc[cryptoutilSharedMagic.StringStatus] = map[string]any{
				testObservedGeneration: float64(1), testReadyReplicas: float64(1), testUpdatedReplicas: float64(1),
				testCurrentRevision: testRevision, testUpdateRevision: testRevision, testPhase: testActive,
			}

			switch name {
			case testNoReplicas:
				objectMap(doc["spec"])["replicas"] = float64(0)
			case testStaleGeneration:
				objectMap(doc["metadata"])["generation"] = float64(2)
			case testPartialReady:
				objectMap(doc[cryptoutilSharedMagic.StringStatus])["readyReplicas"] = float64(0)
			case testPartialUpdate:
				objectMap(doc[cryptoutilSharedMagic.StringStatus])["updatedReplicas"] = float64(0)
			case testOldRevision:
				objectMap(doc[cryptoutilSharedMagic.StringStatus])["updateRevision"] = "new"
			case testMissingRevision:
				delete(objectMap(doc[cryptoutilSharedMagic.StringStatus]), "currentRevision")
			case testNamespaceTerminating:
				objectMap(doc[cryptoutilSharedMagic.StringStatus])["phase"] = "Terminating"
			case testForeign:
				objectMap(objectMap(doc["metadata"])["labels"])[cryptoutilSharedMagic.KubernetesLabelRelease] = testAnother
			}

			client, err := NewKubectlClient(testConfig(t), runnerFunc(func(context.Context, string, []string, []byte) ([]byte, error) { return encodedDocument(t, doc), nil }))
			require.NoError(t, err)

			if id.Kind != statefulSetKind || name == testForeign {
				err = client.Require(context.Background(), id)
				if name == testNamespaceTerminating || name == testForeign {
					require.ErrorIs(t, err, ErrPrerequisite)
				} else {
					require.NoError(t, err)
				}

				return
			}

			if name != testReady {
				require.False(t, databaseReady(doc))

				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			require.NoError(t, client.WaitDatabase(ctx, id))
		})
	}
}

func renderTestValues(t *testing.T, values cryptoutilAppsToolsDeploymentgen.LifecycleChartValues) []byte {
	t.Helper()

	var result []byte

	add := func(id Identity, job *cryptoutilAppsToolsDeploymentgen.LifecycleJobValues) {
		doc := testDocument(id)
		delete(objectMap(doc["metadata"]), "uid")
		delete(objectMap(doc["metadata"]), "generation")

		if job != nil {
			doc["apiVersion"] = "batch/v1"
			doc[testSpec] = map[string]any{
				"activeDeadlineSeconds": job.ActiveDeadlineSeconds, "backoffLimit": job.BackoffLimit, "completions": 1, "parallelism": 1,
				testTemplate: map[string]any{testSpec: map[string]any{
					"serviceAccountName": bootstrapName, "automountServiceAccountToken": false, testRestartPolicy: testNever,
					testContainers: []any{map[string]any{testImage: job.Image, "command": job.Command, "args": job.Args, "resources": job.Resources}},
				}},
			}
		}

		result = append(result, []byte("\n---\n")...)
		result = append(result, encodedDocument(t, doc)...)
	}

	for _, boundary := range values.Namespaces {
		for _, kind := range []string{"ServiceAccount", testRole, "RoleBinding"} {
			add(Identity{Namespace: boundary.Name, Kind: kind, Name: bootstrapName, Run: values.Run, Release: boundary.Release}, nil)
		}
	}

	for _, job := range values.Jobs {
		add(Identity{Namespace: job.Namespace, Kind: jobKind, Name: job.Name, Run: values.Run, Release: job.Release, Revision: values.Revision}, &job)
	}

	return result
}

func TestKubectlPrepareAndOrderedWiring(t *testing.T) {
	t.Parallel()
	request := executionFixture(t, testPGSelector)
	config := testConfig(t)
	state := map[string]map[string]any{}

	var mutations []string

	client, err := NewKubectlClient(config, runnerFunc(func(_ context.Context, command string, args []string, input []byte) ([]byte, error) {
		if command == config.Helm {
			require.Contains(t, args, "--values")
			require.Contains(t, args, "-")
			require.Contains(t, args, filepath.Join(config.RepositoryRoot, "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-lifecycle"))
			require.NotContains(t, string(input), "protected-test-material")

			var values cryptoutilAppsToolsDeploymentgen.LifecycleChartValues
			require.NoError(t, json.Unmarshal(input, &values))

			for _, boundary := range values.Namespaces {
				id := Identity{Kind: namespaceKind, Name: boundary.Name, Run: values.Run, Release: boundary.Release}
				doc := testDocument(id)
				doc[cryptoutilSharedMagic.StringStatus] = map[string]any{testPhase: testActive}
				state[identityKey(id)] = doc
			}

			for _, name := range []string{cryptoutilSharedMagic.PKIInitPostgresLeaderService, cryptoutilSharedMagic.PKIInitPostgresFollowerService} {
				id := Identity{Namespace: request.Inventory.InfraNamespace, Kind: statefulSetKind, Name: name, Run: values.Run, Release: "cu-" + values.Run + "-infra"}
				doc := testDocument(id)
				doc[testSpec] = map[string]any{testReplicas: float64(1)}
				doc[testStatus] = map[string]any{testObservedGeneration: float64(1), testReadyReplicas: float64(1), testUpdatedReplicas: float64(1), testCurrentRevision: testRev, testUpdateRevision: testRev}
				state[identityKey(id)] = doc
			}

			return renderTestValues(t, values), nil
		}

		require.Equal(t, config.Kubectl, command)

		switch args[4] {
		case testGet:
			id := Identity{Kind: args[5], Name: args[6]}
			if len(args) > 10 {
				id.Namespace = args[len(args)-1]
			}

			if doc := state[identityKey(id)]; doc != nil {
				return encodedDocument(t, doc), nil
			}

			return nil, nil
		case testCreate:
			var doc map[string]any
			require.NoError(t, json.Unmarshal(input, &doc))
			id := documentIdentity(doc)
			mutations = append(mutations, identityKey(id))
			require.Nil(t, state[identityKey(id)])

			objectMap(doc["metadata"])["uid"] = testUID
			if id.Kind == jobKind {
				doc[testStatus] = map[string]any{testConditions: []any{map[string]any{testType: testComplete, testStatus: testTrue}}}
			}

			state[identityKey(id)] = doc

			return encodedDocument(t, doc), nil
		default:
			require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "unexpected destructive command %v", args)

			return nil, nil
		}
	}))
	require.NoError(t, err)
	result, err := Execute(context.Background(), request, client, Policy{MaxDuration: cryptoutilSharedMagic.TestTimeout, OperationTimeout: cryptoutilSharedMagic.TestTimeout})
	require.NoError(t, err)
	require.Equal(t, len(request.Secrets), result.Secrets)
	require.Equal(t, len(request.Input.Jobs), result.Jobs)
	require.Len(t, mutations, len(request.Secrets)+len(request.Input.Jobs)+3*(len(request.Inventory.Releases)+1))
	require.Error(t, client.Prepare(context.Background(), cryptoutilAppsToolsDeploymentgen.LifecycleChartValues{}))
}

func TestDecodeRenderedRejectsCoverageAndOwnership(t *testing.T) {
	t.Parallel()

	id := Identity{Namespace: testOwned, Kind: jobKind, Name: "job", Run: testRun, Release: testRelease, Revision: testRevision}
	for _, name := range []string{testValid, testEmpty, testDuplicate, testUnknown, testForeign, testRevision, testInvalidYAML, testDuplicateKey, testNonJSON, testOversized} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc := testDocument(id)

			var raw []byte

			switch name {
			case testUnknown:
				doc["kind"] = secretKind
			case testForeign:
				objectMap(objectMap(doc["metadata"])["labels"])[runLabel] = testOther
			case testRevision:
				objectMap(objectMap(doc["metadata"])["annotations"])["cu.io/revision"] = testOther
			}

			raw = encodedDocument(t, doc)

			switch name {
			case testEmpty:
				raw = []byte("---\n")
			case testDuplicate:
				raw = append(append(raw, []byte("\n---\n")...), raw...)
			case testInvalidYAML:
				raw = []byte("{")
			case testDuplicateKey:
				raw = []byte("kind: Job\nkind: Secret\n")
			case testNonJSON:
				raw = []byte("value: .inf\n")
			case testOversized:
				raw = make([]byte, maxOutputBytes+1)
			}

			result, err := decodeRendered(raw, map[string]Identity{identityKey(id): id})
			if name == testValid {
				require.NoError(t, err)
				require.Len(t, result, 1)
			} else {
				require.Error(t, err)
			}
		})
	}
}
