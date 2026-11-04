// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	json "encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testVerifyChangedBytes    = "changed bytes"
	testVerifyWrongOwner      = "postJob wrong owner"
	testVerifyEncodedValue    = "Zm9yZ2Vk"
	testVerifyForged          = "forged"
	testVerifyGeneration      = "unsupported generation"
	testVerifyMissingVerifier = "missing verifier"
	testVerifyExtraKey        = "extra key"
	testVerifyFailure         = "verification failure"
	testVerifyWrongRun        = "wrong run"
	testVerifyReplacedUID     = "replaced UID"
	testVerifyUnensured       = "unensured"
	testVerifyUnknownImport   = "unknown import"
	testVerifyTransport       = "transport"
	testVerifyInvalidImport   = "invalid desired key set"
)

func TestVerifySecretIsReadOnlyAndRejectsChangedMaterial(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testExact, testVerifyChangedBytes, testVerifyExtraKey, testVerifyWrongOwner, testVerifyWrongRun, testVerifyReplacedUID, testMissing, testVerifyUnensured, testVerifyUnknownImport, testVerifyTransport, testMalformed, testVerifyInvalidImport} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testSQLiteSelector)
			secret := request.Secrets[0]
			raw, err := json.Marshal(secret)
			require.NoError(t, err)

			var doc map[string]any
			require.NoError(t, json.Unmarshal(raw, &doc))
			objectMap(doc["metadata"])["uid"] = testUID
			wantErr := ErrCollision

			switch test {
			case testVerifyChangedBytes:
				objectMap(doc["data"])[testTLSCert] = testVerifyEncodedValue
			case testVerifyExtraKey:
				objectMap(doc["data"])["extra"] = testVerifyEncodedValue
			case testVerifyWrongOwner:
				objectMap(objectMap(doc["metadata"])["labels"])["cu.io/owner"] = testVerifyForged
			case testVerifyWrongRun:
				objectMap(objectMap(doc["metadata"])["labels"])[runLabel] = testVerifyForged
			case testVerifyReplacedUID:
				objectMap(doc["metadata"])["uid"] = testReplaced
				wantErr = ErrReplaced
			case testMissing, testVerifyUnensured:
				wantErr = ErrPrerequisite
			}

			calls := 0
			client, err := NewKubectlClient(testConfig(t), runnerFunc(func(_ context.Context, _ string, args []string, _ []byte) ([]byte, error) {
				calls++

				require.Contains(t, args, "get")
				require.NotContains(t, args, "create")

				if test == testMissing {
					return nil, nil
				}

				if test == testVerifyTransport {
					return nil, errors.New("DO-NOT-LOG")
				}

				if test == testMalformed {
					return []byte(cryptoutilSharedMagic.DevSetupTestInvalidName), nil
				}

				return encodedDocument(t, doc), nil
			}))
			require.NoError(t, err)

			client.run = request.Model.Run
			client.imports = map[string]gen.LifecycleSecretSource{secret.Metadata.Namespace + "/" + secret.Metadata.Name: request.Input.Secrets[0]}
			id := Identity{Kind: secretKind, Namespace: secret.Metadata.Namespace, Name: secret.Metadata.Name}

			client.uids[identityKey(id)] = testUID
			if test == testVerifyUnensured {
				delete(client.uids, identityKey(id))
			}

			if test == testVerifyUnknownImport {
				client.imports = nil
			}

			if test == testVerifyInvalidImport {
				for key := range secret.Data {
					delete(secret.Data, key)

					break
				}
			}

			err = client.VerifySecret(context.Background(), secret)
			if test == testExact {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")

				if test != testVerifyTransport && test != testMalformed {
					require.ErrorIs(t, err, wantErr)
				}
			}

			if test == testVerifyUnensured || test == testVerifyUnknownImport || test == testVerifyInvalidImport {
				require.Zero(t, calls)
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}

func TestPKICompleteAloneDoesNotAdvanceLifecycle(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testVerifyMissingVerifier, testVerifyGeneration, testVerifyFailure, "success"} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testPGSelector)
			verified := 0

			request.VerifyPKI = func(context.Context, gen.LifecycleJobValues, Client) error {
				verified++

				if test == testVerifyFailure {
					return ErrCollision
				}

				return nil
			}
			if test == testVerifyMissingVerifier {
				request.VerifyPKI = nil
			}

			if test == testVerifyGeneration {
				for index := range request.Input.Jobs {
					if request.Input.Jobs[index].Name == infraPKIName {
						request.Input.Jobs[index].Args = []string{cryptoutilSharedMagic.PSIDPKIInit, "--domain=cryptoutil", "--output-dir=/tmp/pki"}
					}
				}

				revision, err := gen.LifecycleRevision(request.Input)
				require.NoError(t, err)

				request.Model.Revision = revision

				var errBuild error

				request.Inventory, errBuild = gen.Build(request.Model, testPGSelector)
				require.NoError(t, errBuild)
			}

			client := &recordingClient{}
			result, err := Execute(context.Background(), request, client, Policy{MaxDuration: time.Minute, OperationTimeout: 10 * time.Second})

			switch test {
			case testVerifyMissingVerifier, testVerifyGeneration:
				require.ErrorIs(t, err, ErrUnsupportedPKI)
				require.Empty(t, client.calls, "unsupported generation must fail before Prepare/import")
			case testVerifyFailure:
				require.ErrorIs(t, err, ErrCollision)
				require.Zero(t, result.Jobs, "Complete without read-only verification is not counted")
				require.Equal(t, 1, verified)

				for _, call := range client.calls {
					require.False(t, strings.Contains(call, "database-ready"))
				}
			default:
				require.NoError(t, err)
				require.Positive(t, result.Jobs)
				require.Positive(t, verified)
			}
		})
	}
}
