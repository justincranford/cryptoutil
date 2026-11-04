// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

type namespacedObject struct {
	Kind      string
	Namespace string
	Name      string
}

func expectedRuntimeControllerPVCs(release Release) []namespacedObject {
	objects := make([]namespacedObject, 0)

	for _, workload := range release.Workloads {
		for _, pod := range workload.Pods {
			if pod.Active && pod.PVC != "" {
				objects = append(objects, namespacedObject{
					Kind: testPersistentVolumeClaimKind, Namespace: release.Namespace, Name: pod.PVC,
				})
			}
		}
	}

	return objects
}

func expectedInfrastructureControllerPVCs(inventory Inventory) []namespacedObject {
	objects := make([]namespacedObject, 0)

	for _, object := range inventory.Infra {
		if object.Kind == testPersistentVolumeClaimKind {
			objects = append(objects, namespacedObject{
				Kind: object.Kind, Namespace: inventory.InfraNamespace, Name: object.Name,
			})
		}
	}

	return objects
}

func controllerGeneratedPVCs(
	documents []helmRenderedDocument,
	expectedNamespace string,
) ([]namespacedObject, error) {
	objects := make([]namespacedObject, 0)

	for _, document := range documents {
		if document.Kind != testStatefulSetKind {
			continue
		}

		if document.Metadata.Namespace != expectedNamespace {
			return nil, fmt.Errorf(
				"controller StatefulSet has namespace %q, expected %q",
				document.Metadata.Namespace, expectedNamespace,
			)
		}

		for _, claimTemplate := range document.Spec.VolumeClaimTemplates {
			for ordinal := range document.Spec.Replicas {
				objects = append(objects, namespacedObject{
					Kind: testPersistentVolumeClaimKind, Namespace: document.Metadata.Namespace,
					Name: fmt.Sprintf("%s-%s-%d", claimTemplate.Metadata.Name, document.Metadata.Name, ordinal),
				})
			}
		}
	}

	return objects, nil
}

func namespacedObjectKeys(objects []namespacedObject) []string {
	result := make([]string, len(objects))
	for index, object := range objects {
		result[index] = object.Kind + "/" + object.Namespace + "/" + object.Name
	}

	slices.Sort(result)

	return result
}

func TestRenderedObjectIdentityRequiresExpectedNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		renderedNamespace string
		wantErr           bool
		wantKeys          []string
	}{
		{
			name:              testSelectedNamespaceCaseName,
			renderedNamespace: testReleaseNamespace,
			wantKeys:          []string{"ConfigMap/" + testReleaseNamespace + "/settings"},
		},
		{
			name:              testWrongNamespaceCaseName,
			renderedNamespace: testOtherNamespace,
			wantErr:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			document := helmRenderedDocument{Kind: testConfigMapKind}
			document.Metadata.Name = "settings"
			document.Metadata.Namespace = test.renderedNamespace

			keys, err := objectKeysFromRendered([]helmRenderedDocument{document}, testReleaseNamespace)
			if test.wantErr {
				require.ErrorContains(t, err, "rendered resource has invalid identity")
				require.Nil(t, keys)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantKeys, keys)
		})
	}
}

func TestControllerGeneratedPVCIdentityUsesStatefulSetNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		renderedNamespace string
		wantErr           bool
		wantKeys          []string
	}{
		{
			name:              testSelectedNamespaceCaseName,
			renderedNamespace: testReleaseNamespace,
			wantKeys: []string{
				"PersistentVolumeClaim/" + testReleaseNamespace + "/data-worker-0",
				"PersistentVolumeClaim/" + testReleaseNamespace + "/data-worker-1",
			},
		},
		{
			name:              testWrongNamespaceCaseName,
			renderedNamespace: testOtherNamespace,
			wantErr:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			document := helmRenderedDocument{Kind: testStatefulSetKind}
			document.Metadata.Name = "worker"
			document.Metadata.Namespace = test.renderedNamespace
			document.Spec.Replicas = 2
			document.Spec.VolumeClaimTemplates = append(document.Spec.VolumeClaimTemplates, struct {
				Metadata struct {
					Name string `yaml:"name"`
				} `yaml:"metadata"`
			}{})
			document.Spec.VolumeClaimTemplates[0].Metadata.Name = "data"

			objects, err := controllerGeneratedPVCs([]helmRenderedDocument{document}, testReleaseNamespace)
			if test.wantErr {
				require.ErrorContains(t, err, "controller StatefulSet has namespace")
				require.Nil(t, objects)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantKeys, namespacedObjectKeys(objects))
		})
	}
}
