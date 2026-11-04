// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type helmRenderedDocument struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Namespace   string            `yaml:"namespace"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
	Data map[string]string `yaml:"data"`
	Spec struct {
		PodSelector          helmLabelSelector       `yaml:"podSelector"`
		Replicas             int                     `yaml:"replicas"`
		MinReplicas          int                     `yaml:"minReplicas"`
		MaxReplicas          int                     `yaml:"maxReplicas"`
		Template             helmRenderedPodTemplate `yaml:"template"`
		Ingress              []helmNetworkPolicyRule `yaml:"ingress"`
		Egress               []helmNetworkPolicyRule `yaml:"egress"`
		VolumeClaimTemplates []struct {
			Metadata struct {
				Name string `yaml:"name"`
			} `yaml:"metadata"`
		} `yaml:"volumeClaimTemplates"`
	} `yaml:"spec"`
	Rules []struct {
		Resources     []string `yaml:"resources"`
		ResourceNames []string `yaml:"resourceNames"`
		Verbs         []string `yaml:"verbs"`
	} `yaml:"rules"`
}

type helmRenderedPodTemplate struct {
	Spec struct {
		AutomountServiceAccountToken *bool                          `yaml:"automountServiceAccountToken"`
		SecurityContext              helmRenderedPodSecurityContext `yaml:"securityContext"`
		InitContainers               []helmRenderedContainer        `yaml:"initContainers"`
		Containers                   []helmRenderedContainer        `yaml:"containers"`
		Volumes                      []struct {
			Name   string `yaml:"name"`
			Secret struct {
				SecretName string `yaml:"secretName"`
			} `yaml:"secret"`
			Projected struct {
				Sources []struct {
					ServiceAccountToken struct {
						Path string `yaml:"path"`
					} `yaml:"serviceAccountToken"`
					ConfigMap struct {
						Name  string `yaml:"name"`
						Items []struct {
							Key  string `yaml:"key"`
							Path string `yaml:"path"`
						} `yaml:"items"`
					} `yaml:"configMap"`
					Secret struct {
						Name string `yaml:"name"`
					} `yaml:"secret"`
				} `yaml:"sources"`
			} `yaml:"projected"`
		} `yaml:"volumes"`
	} `yaml:"spec"`
}

type helmRenderedPodSecurityContext struct {
	RunAsNonRoot   bool `yaml:"runAsNonRoot"`
	RunAsUser      int  `yaml:"runAsUser"`
	RunAsGroup     int  `yaml:"runAsGroup"`
	FSGroup        int  `yaml:"fsGroup"`
	SeccompProfile struct {
		Type string `yaml:"type"`
	} `yaml:"seccompProfile"`
}

type helmRenderedContainer struct {
	Command []string `yaml:"command"`
	Args    []string `yaml:"args"`
	Name    string   `yaml:"name"`
	Env     []struct {
		Name      string `yaml:"name"`
		Value     string `yaml:"value"`
		ValueFrom struct {
			FieldRef struct {
				FieldPath string `yaml:"fieldPath"`
			} `yaml:"fieldRef"`
			SecretKeyRef struct {
				Name string `yaml:"name"`
				Key  string `yaml:"key"`
			} `yaml:"secretKeyRef"`
		} `yaml:"valueFrom"`
	} `yaml:"env"`
	EnvFrom []struct {
		SecretRef struct {
			Name string `yaml:"name"`
		} `yaml:"secretRef"`
	} `yaml:"envFrom"`
	SecurityContext helmRenderedContainerSecurity `yaml:"securityContext"`
	VolumeMounts    []helmRenderedVolumeMount     `yaml:"volumeMounts"`
}

type helmRenderedContainerSecurity struct {
	AllowPrivilegeEscalation *bool `yaml:"allowPrivilegeEscalation"`
	Privileged               *bool `yaml:"privileged"`
	ReadOnlyRootFilesystem   *bool `yaml:"readOnlyRootFilesystem"`
	RunAsNonRoot             bool  `yaml:"runAsNonRoot"`
	RunAsUser                int   `yaml:"runAsUser"`
	RunAsGroup               int   `yaml:"runAsGroup"`
	Capabilities             struct {
		Add  []string `yaml:"add"`
		Drop []string `yaml:"drop"`
	} `yaml:"capabilities"`
}

type helmRenderedVolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly"`
}

type helmLabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

type helmNetworkPolicyPeer struct {
	NamespaceSelector helmLabelSelector `yaml:"namespaceSelector"`
	PodSelector       helmLabelSelector `yaml:"podSelector"`
}

type helmNetworkPolicyPort struct {
	Protocol string `yaml:"protocol"`
	Port     int    `yaml:"port"`
}

type helmNetworkPolicyRule struct {
	From  []helmNetworkPolicyPeer `yaml:"from"`
	To    []helmNetworkPolicyPeer `yaml:"to"`
	Ports []helmNetworkPolicyPort `yaml:"ports"`
}

func requirePinnedHelm(t *testing.T) string {
	t.Helper()

	helmBinary, err := exec.LookPath("helm")
	if err != nil {
		localHelm := filepath.Join(
			"..", "..", "..", cryptoutilSharedMagic.CICDExcludeDirTestOutput, "framework-v26-toolchain-helm-v4.3.1",
			"windows-amd64", "helm.exe",
		)
		if _, localErr := os.Stat(localHelm); localErr == nil {
			helmBinary = localHelm
		} else {
			require.NoError(t, err, "Helm 4.3.1 must be installed for integration parity checks")
		}
	}

	output, err := exec.Command(helmBinary, cryptoutilSharedMagic.CLIVersionCommand, "--short").CombinedOutput()
	require.NoError(t, err, string(output))
	require.Contains(t, string(output), "v4.3.1")

	return helmBinary
}

func renderHelmDocuments(
	t *testing.T,
	helmBinary, chartPath, releaseName, namespace string,
	valuesPath, outputDirectory string,
	index int,
) ([]helmRenderedDocument, error) {
	t.Helper()

	return renderHelmChart(
		t, helmBinary, chartPath, releaseName, namespace, valuesPath, outputDirectory, index,
	)
}

func renderHelmChart(
	t *testing.T,
	helmBinary, chartPath, releaseName, namespace, valuesPath, outputDirectory string,
	index int,
) ([]helmRenderedDocument, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	command := exec.CommandContext(
		ctx, helmBinary, cryptoutilSharedMagic.SkeletonTemplateServiceName, releaseName, chartPath,
		"--namespace", namespace, "--kube-version", "1.37.1", "--values", valuesPath,
	)

	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("helm template failed: %w: %s", err, output)
	}

	manifestPath := filepath.Join(outputDirectory, fmt.Sprintf("rendered-%02d.yaml", index))
	if err := os.WriteFile(manifestPath, output, 0o600); err != nil {
		return nil, fmt.Errorf("write rendered manifest: %w", err)
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(output)))
	result := make([]helmRenderedDocument, 0)

	for {
		var document helmRenderedDocument

		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decode rendered manifest: %w", err)
		}

		if document.Kind == "" {
			continue
		}

		if document.Metadata.Name == "" || document.Metadata.Namespace != namespace {
			return nil, fmt.Errorf(
				"rendered resource has invalid identity: %s/%s namespace=%q expected=%q",
				document.Kind, document.Metadata.Name, document.Metadata.Namespace, namespace,
			)
		}

		result = append(result, document)
	}

	return result, nil
}

func expectedRuntimeChartObjects(release Release) []Object {
	objects := make([]Object, 0)

	for _, object := range release.Objects {
		if object.Kind == testConfigMapKind || object.Kind == testNetworkPolicyKind ||
			object.Kind == testResourceQuotaKind {
			objects = append(objects, object)
		}
	}

	for _, workload := range release.Workloads {
		for _, object := range workload.Objects {
			if object.Kind != testLifecycleSecretKind {
				objects = append(objects, object)
			}
		}

		for _, pod := range workload.Pods {
			objects = append(objects, Object{Kind: testServiceKind, Name: pod.Service})
		}
	}

	return objects
}

func expectedInfrastructureChartObjects(inventory Inventory) []Object {
	objects := make([]Object, 0)
	chartOwnedKinds := map[string]bool{
		testConfigMapKind: true, testNetworkPolicyKind: true, "PodDisruptionBudget": true,
		"ResourceQuota": true, testServiceKind: true, testStatefulSetKind: true,
	}

	for _, object := range inventory.Infra {
		if chartOwnedKinds[object.Kind] {
			objects = append(objects, object)
		}
	}

	return objects
}

// objectKeys applies the caller's known namespace to chart-owned, namespaced objects.
// Cluster-scoped resources need a separate identity path and must not use this helper.
func objectKeys(objects []Object, namespace string) []string {
	namespaced := make([]namespacedObject, len(objects))
	for index, object := range objects {
		namespaced[index] = namespacedObject{
			Kind: object.Kind, Namespace: namespace, Name: object.Name,
		}
	}

	return namespacedObjectKeys(namespaced)
}

func objectKeysFromRendered(documents []helmRenderedDocument, expectedNamespace string) ([]string, error) {
	objects := make([]namespacedObject, len(documents))
	for index, document := range documents {
		if document.Metadata.Name == "" || document.Metadata.Namespace != expectedNamespace {
			return nil, fmt.Errorf(
				"rendered resource has invalid identity: %s/%s namespace=%q expected=%q",
				document.Kind, document.Metadata.Name, document.Metadata.Namespace, expectedNamespace,
			)
		}

		objects[index] = namespacedObject{
			Kind: document.Kind, Namespace: document.Metadata.Namespace, Name: document.Metadata.Name,
		}
	}

	return namespacedObjectKeys(objects), nil
}
