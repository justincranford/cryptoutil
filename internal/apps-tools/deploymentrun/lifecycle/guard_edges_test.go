// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	json "encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsToolsDeploymentgen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestContainsDesired(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name            string
		actual, desired any
		want            bool
	}{
		{"map type", nil, map[string]any{testKey: testValue}, false},
		{"map missing key", map[string]any{}, map[string]any{testKey: testValue}, false},
		{"map defaults", map[string]any{testKey: testValue, "default": true}, map[string]any{testKey: testValue}, true},
		{"list type", nil, []any{testOne}, false},
		{"list length", []any{testOne, testTwo}, []any{testOne}, false},
		{"list order", []any{testTwo, testOne}, []any{testOne, testTwo}, false},
		{"list exact", []any{testOne}, []any{testOne}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.want, containsDesired(test.actual, test.desired))
		})
	}
}

func TestExistingPodRejectsAdmissionDrift(t *testing.T) {
	t.Parallel()

	for _, name := range []string{testExact, testInertDefaults, testHostNetwork, testEnvInjection, testExtraContainer, testPrivileged, testPodContext, testChangedDefault} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			desired := map[string]any{
				testKind: jobKind, testSpec: map[string]any{testTemplate: map[string]any{testSpec: map[string]any{
					testRestartPolicy: testNever, testSecurityContext: map[string]any{"runAsNonRoot": true},
					testContainers: []any{map[string]any{testImage: "pinned", testSecurityContext: map[string]any{"allowPrivilegeEscalation": false}}},
				}}},
			}

			var actual map[string]any
			require.NoError(t, json.Unmarshal(encodedDocument(t, desired), &actual))
			pod := objectMap(objectMap(objectMap(actual["spec"])[cryptoutilSharedMagic.SkeletonTemplateServiceName])["spec"])
			containers, ok := pod[testContainers].([]any)
			require.True(t, ok)

			container := objectMap(containers[0])

			switch name {
			case testInertDefaults:
				pod["dnsPolicy"], pod["schedulerName"], pod["enableServiceLinks"] = "ClusterFirst", "default-scheduler", true
				container["terminationMessagePath"], container["terminationMessagePolicy"] = "/dev/termination-log", "File"
			case testHostNetwork:
				pod["hostNetwork"] = true
			case testEnvInjection:
				container[cryptoutilSharedMagic.CICDExcludeDirEnv] = []any{map[string]any{testName: "INJECTED", testValue: "untrusted"}}
			case testExtraContainer:
				pod[testContainers] = append(containers, map[string]any{})
			case testPrivileged:
				objectMap(container["securityContext"])["privileged"] = true
			case testPodContext:
				objectMap(pod["securityContext"])["runAsUser"] = float64(0)
			case testChangedDefault:
				pod["dnsPolicy"] = "None"
			}

			require.Equal(t, name == testExact || name == testInertDefaults, safeExistingPod(actual, desired))
		})
	}
}

func TestPrepareErrorSurfaces(t *testing.T) {
	t.Parallel()

	for _, name := range []string{testTransport, testMalformed, testDeadline, testImage} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := executionFixture(t, testSQLiteSelector)
			values, err := cryptoutilAppsToolsDeploymentgen.BuildLifecycleChartValues(request.Model, request.Inventory, request.Input)
			require.NoError(t, err)
			client, err := NewKubectlClient(testConfig(t), runnerFunc(func(context.Context, string, []string, []byte) ([]byte, error) {
				if name == testTransport {
					return nil, errors.New("private")
				}

				if name == testMalformed {
					return []byte("{"), nil
				}

				raw := renderTestValues(t, values)
				if name == testDeadline {
					// Values are only copied locally by range in renderTestValues.
					values.Jobs[0].ActiveDeadlineSeconds++
					raw = renderTestValues(t, values)
					values.Jobs[0].ActiveDeadlineSeconds--
				}

				if name == testImage {
					original := values.Jobs[0].Image
					values.Jobs[0].Image = testForeign
					raw = renderTestValues(t, values)
					values.Jobs[0].Image = original
				}

				return raw, nil
			}))
			require.NoError(t, err)
			err = client.Prepare(context.Background(), values)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "private")
			require.Nil(t, client.resources)
		})
	}
}

func TestKubectlBoundaryErrors(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		testMissingPrerequisite, testRequireTransport, testBootstrapUnprepared, testBootstrapCollision,
		testUnknownJob, testUnboundImport, testOversize, testCanceledTransport, testEncodeResource,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client, err := NewKubectlClient(testConfig(t), runnerFunc(func(context.Context, string, []string, []byte) ([]byte, error) {
				switch name {
				case testRequireTransport:
					return nil, errors.New("private")
				case testOversize:
					return make([]byte, maxOutputBytes+1), nil
				case testCanceledTransport:
					return nil, context.Canceled
				}

				return nil, nil
			}))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			switch name {
			case testMissingPrerequisite, testRequireTransport:
				err = client.Require(ctx, Identity{Kind: namespaceKind, Name: testOwned, Run: testRun})
			case testBootstrapUnprepared:
				err = client.EnsureBootstrap(ctx)
			case testBootstrapCollision:
				id := Identity{Kind: testRole, Name: bootstrapName, Namespace: testOwned, Run: testRun}
				client.resources = map[string]map[string]any{identityKey(id): testDocument(id)}
				client.bootstrap = []Identity{id}
				err = client.EnsureBootstrap(ctx)
			case testUnknownJob:
				err = client.EnsureJob(ctx, cryptoutilAppsToolsDeploymentgen.LifecycleJobValues{})
			case testUnboundImport:
				err = client.EnsureSecret(ctx, cryptoutilAppsToolsDeploymentgen.LifecycleSecret{})
			case testOversize, testCanceledTransport:
				_, err = client.command(ctx, []string{testGet}, nil)
			case testEncodeResource:
				err = client.ensure(ctx, Identity{}, map[string]any{"unsupported": make(chan int)})
			}

			require.Error(t, err)
			require.NotContains(t, err.Error(), "private")
		})
	}
}

func TestSafeFailureClasses(t *testing.T) {
	t.Parallel()

	for _, cause := range []error{ErrCollision, ErrPrerequisite, ErrJobFailed, ErrReplaced, ErrTransport, context.Canceled, context.DeadlineExceeded} {
		t.Run(cause.Error(), func(t *testing.T) {
			t.Parallel()

			err := safeError("operation", Identity{}, errors.Join(cause, errors.New("private")))
			require.ErrorIs(t, err, cause)
			require.NotContains(t, err.Error(), "private")
		})
	}
}
