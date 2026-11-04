// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testResumeApplicationNamespace = "application-namespace"
	testResumeApplicationJob       = "application-job"
	testResumeApplicationName      = "app"
	testResumeMalformedInventory   = "malformed-inventory"
	testResumeDuplicateInventory   = "duplicate-inventory"
	testResumeEmptyInventory       = cryptoutilSharedMagic.EMPTY
	testResumeApplicationPod       = "application-pod"
	testResumeMissingSystem        = "missing-system"
	testResumeRunLabel             = "run-label"
	testResumeUnknownKind          = "unknown-kind"
	testResumeUnnamed              = "unnamed"
)

func TestDevelopmentBootstrapInventoryRejectsApplicationAdoption(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{
		"", testResumeMalformedInventory, testResumeEmptyInventory, testResumeDuplicateInventory, testResumeMissingSystem, testResumeApplicationNamespace,
		testResumeApplicationPod, testResumeApplicationJob, testResumeRunLabel, testResumeUnknownKind, testResumeUnnamed,
	} {
		t.Run("inventory="+failure, func(t *testing.T) {
			t.Parallel()

			var objects []map[string]any
			for _, name := range []string{developmentDefaultContext, developmentKubeSystem, developmentKubePublic, developmentKubeNodeLease} {
				objects = append(objects, map[string]any{developmentKind: developmentNamespaceKind, developmentMetadata: map[string]any{developmentName: name}})
			}

			switch failure {
			case testResumeDuplicateInventory:
				objects = append(objects, objects[0])
			case testResumeMissingSystem:
				objects = objects[1:]
			case testResumeApplicationNamespace:
				objects = append(objects, map[string]any{developmentKind: developmentNamespaceKind, developmentMetadata: map[string]any{developmentName: testResumeApplicationName}})
			case testResumeApplicationPod, testResumeApplicationJob:
				kind := developmentPodKind
				if failure == testResumeApplicationJob {
					kind = developmentJobKind
				}

				objects = append(objects, map[string]any{developmentKind: kind, developmentMetadata: map[string]any{developmentName: testResumeApplicationName, testImageNamespace: developmentDefaultContext}})
			case testResumeRunLabel:
				objects[0][developmentMetadata] = map[string]any{developmentName: developmentDefaultContext, "labels": map[string]string{testPKIRunLabel: testDiagnosticPrivate}}
			case testResumeUnknownKind:
				objects[0][developmentKind] = cryptoutilSharedMagic.KubernetesSecretKind
			case testResumeUnnamed:
				objects[0][developmentMetadata] = map[string]any{}
			case testResumeEmptyInventory:
				objects = nil
			}

			raw, err := json.Marshal(map[string]any{developmentKind: developmentListKind, developmentItems: objects})
			require.NoError(t, err)

			if failure == testResumeMalformedInventory {
				raw = []byte(testDiagnosticPrivate)
			}

			err = validateDevelopmentBootstrapInventory(raw)
			if failure == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), testDiagnosticPrivate)
			}
		})
	}
}
