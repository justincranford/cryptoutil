// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testVerificationChangedCommand  = "changed command"
	testVerificationChangedRevision = "changed revision"
	testVerificationExecutableField = "extra executable field"
	testVerificationNotPrepared     = "not prepared"
	testVerificationVerified        = "verified"
)

func TestVerifyJobRechecksRevisionSpecAndSecurityAfterComplete(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testVerificationVerified, testVerificationNotPrepared, testVerificationChangedRevision, testVerificationChangedCommand, testVerificationExecutableField, testReplaced, testMissing} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testSQLiteSelector)
			values, err := gen.BuildLifecycleChartValues(request.Model, request.Inventory, request.Input)
			require.NoError(t, err)
			rendered := renderTestValues(t, values)
			job := values.Jobs[0]
			id := Identity{Kind: jobKind, Name: job.Name, Namespace: job.Namespace, Run: values.Run, Release: job.Release, Revision: values.Revision}

			var actual map[string]any

			config := testConfig(t)
			client, err := NewKubectlClient(config, runnerFunc(func(_ context.Context, command string, args []string, _ []byte) ([]byte, error) {
				if command == config.Helm {
					return rendered, nil
				}

				require.Contains(t, args, "get")
				require.NotContains(t, args, "create")

				if test == testMissing {
					return nil, nil
				}

				return encodedDocument(t, actual), nil
			}))
			require.NoError(t, err)
			require.NoError(t, client.Prepare(context.Background(), values))
			require.NoError(t, jsonDecodeVerifyCopy(client.resources[identityKey(id)], &actual))
			objectMap(actual["metadata"])["uid"] = testUID

			switch test {
			case testVerificationChangedRevision:
				objectMap(objectMap(actual["metadata"])["annotations"])["cu.io/revision"] = "r-forged"
			case testVerificationChangedCommand:
				objectMap(objectMap(objectMap(actual["spec"])[cryptoutilSharedMagic.SkeletonTemplateServiceName])["spec"])["restartPolicy"] = "Always"
			case testVerificationExecutableField:
				objectMap(objectMap(objectMap(actual["spec"])[cryptoutilSharedMagic.SkeletonTemplateServiceName])["spec"])["hostPID"] = true
			case testReplaced:
				objectMap(actual["metadata"])["uid"] = "forged"
			}

			client.uids[identityKey(id)] = testUID
			if test == testVerificationNotPrepared {
				client.resources = nil
			}

			err = client.VerifyJob(context.Background(), id)
			if test == testVerificationVerified {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
