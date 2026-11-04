// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	json "encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestHelmRuntimeRenderedSecurityCoversCanonicalAndGeneratedCharts(t *testing.T) {
	t.Parallel()

	helmBinary, releaseName, namespace, valuesPath := runtimeSecurityProjection(t)
	charts := []struct {
		name string
		path string
	}{
		{
			name: "canonical source chart",
			path: filepath.Join(
				"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
				"cryptoutil-runtime",
			),
		},
		{
			name: "generated chart mirror",
			path: filepath.Join("..", "..", "..", "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", "cryptoutil-runtime"),
		},
	}

	for _, chart := range charts {
		t.Run(chart.name, func(t *testing.T) {
			t.Parallel()

			rendered, err := renderHelmDocuments(
				t, helmBinary, chart.path, releaseName, namespace, valuesPath, t.TempDir(), 0,
			)
			require.NoError(t, err)
			require.Empty(t, runtimeRenderedSecurityViolations(rendered))
		})
	}
}

func TestHelmRuntimeRenderedSecurityRejectsUnsafeTemplateMutations(t *testing.T) {
	t.Parallel()

	helmBinary, releaseName, namespace, valuesPath := runtimeSecurityProjection(t)
	mirrorPath := filepath.Join("..", "..", "..", "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", "cryptoutil-runtime")

	tests := []struct {
		name             string
		before           string
		after            string
		replacementCount int
		wantViolation    string
	}{
		{
			name:             "pod allows root identity",
			before:           "      securityContext:\n        runAsNonRoot: true",
			after:            "      securityContext:\n        runAsNonRoot: false",
			replacementCount: 1,
			wantViolation:    testRuntimeNonRootViolation,
		},
		{
			name:             "unconfined pod seccomp",
			before:           "type: RuntimeDefault",
			after:            "type: Unconfined",
			replacementCount: 1,
			wantViolation:    testRuntimeSeccompViolation,
		},
		{
			name:             "pod automounts service-account token",
			before:           "      automountServiceAccountToken: false",
			after:            "      automountServiceAccountToken: true",
			replacementCount: 1,
			wantViolation:    testRuntimeTokenAutomountViolation,
		},
		{
			name:             "privilege escalation in workload containers",
			before:           "allowPrivilegeEscalation: false",
			after:            "allowPrivilegeEscalation: true",
			replacementCount: 2,
			wantViolation:    testRuntimePrivilegeEscalationViolation,
		},
		{
			name:             "containers have writable root filesystems",
			before:           "readOnlyRootFilesystem: true",
			after:            "readOnlyRootFilesystem: false",
			replacementCount: 2,
			wantViolation:    testRuntimeReadOnlyRootFSViolation,
		},
		{
			name: "added Linux capability",
			before: "capabilities:\n" +
				"              drop:\n" +
				"                - ALL",
			after: "capabilities:\n" +
				"              add:\n" +
				"                - " + testRuntimeSecurityCapabilityNetRaw + "\n" +
				"              drop:\n" +
				"                - ALL",
			replacementCount: 2,
			wantViolation:    testRuntimeCapabilityViolation,
		},
		{
			name: "application mounts bootstrap token",
			before: "            - name: pki-work\n" +
				"              mountPath: /run/cryptoutil/pki\n" +
				"              readOnly: true",
			after: "            - name: pki-work\n" +
				"              mountPath: /run/cryptoutil/pki\n" +
				"              readOnly: true\n" +
				"            - name: bootstrap-api-token\n" +
				"              mountPath: /var/run/secrets/kubernetes.io/serviceaccount\n" +
				"              readOnly: true",
			replacementCount: 1,
			wantViolation:    testRuntimeAppTokenMountViolation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mutatedChart := copyRuntimeChartWithTemplateMutation(
				t, mirrorPath, test.before, test.after, test.replacementCount,
			)
			rendered, err := renderHelmDocuments(
				t, helmBinary, mutatedChart, releaseName, namespace, valuesPath, t.TempDir(), 0,
			)
			require.NoError(t, err)

			violations := runtimeRenderedSecurityViolations(rendered)
			require.Contains(t, strings.Join(violations, "\n"), test.wantViolation)
		})
	}
}

func TestHelmRuntimeValuesRejectSecurityOverrides(t *testing.T) {
	t.Parallel()

	helmBinary, releaseName, namespace, valuesPath := runtimeSecurityProjection(t)
	valuesJSON, err := os.ReadFile(valuesPath)
	require.NoError(t, err)

	values := make(map[string]any)
	require.NoError(t, json.Unmarshal(valuesJSON, &values))

	workloads, ok := values["workloads"].([]any)
	require.True(t, ok)
	workload, ok := workloads[0].(map[string]any)
	require.True(t, ok)
	container, ok := workload["container"].(map[string]any)
	require.True(t, ok)

	container["securityContext"] = map[string]any{"privileged": true}

	unsafeValues, err := json.Marshal(values)
	require.NoError(t, err)

	unsafeValuesPath := filepath.Join(t.TempDir(), "unsafe-runtime-values.json")
	require.NoError(t, os.WriteFile(unsafeValuesPath, unsafeValues, 0o600))

	chartPath := filepath.Join("..", "..", "..", "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", "cryptoutil-runtime")
	_, err = renderHelmDocuments(
		t, helmBinary, chartPath, releaseName, namespace, unsafeValuesPath, t.TempDir(), 0,
	)
	require.ErrorContains(t, err, testRuntimeSecurityOverrideField)
}

func runtimeSecurityProjection(t *testing.T) (string, string, string, string) {
	t.Helper()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	inventory, err := build(model, scopes, testServiceSMKMSPGSelector)
	require.NoError(t, err)

	values, err := BuildRuntimeChartValues(
		model, inventory, testServiceSMKMSPGSelector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, testServiceSMKMSPGSelector),
	)
	require.NoError(t, err)

	valuesJSON, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)

	valuesPath := filepath.Join(t.TempDir(), "runtime-values.json")
	require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

	return helmBinary, values.HelmRelease, values.Namespace, valuesPath
}

func runtimeRenderedSecurityViolations(documents []helmRenderedDocument) []string {
	violations := make([]string, 0)
	statefulSetCount := 0

	for _, document := range documents {
		if document.Kind != testStatefulSetKind {
			continue
		}

		statefulSetCount++
		resourceName := document.Metadata.Name
		pod := document.Spec.Template.Spec

		if !pod.SecurityContext.RunAsNonRoot {
			violations = append(violations, fmt.Sprintf("%s: %s", resourceName, testRuntimeNonRootViolation))
		}

		if pod.SecurityContext.RunAsUser != testRuntimeSecurityUserID {
			violations = append(violations, fmt.Sprintf("%s: pod UID must use the runtime non-root identity", resourceName))
		}

		if pod.SecurityContext.RunAsGroup != testRuntimeSecurityGroupID ||
			pod.SecurityContext.FSGroup != testRuntimeSecurityGroupID {
			violations = append(violations, fmt.Sprintf("%s: pod groups must use the runtime non-root identity", resourceName))
		}

		if pod.SecurityContext.SeccompProfile.Type != testRuntimeSecurityProfileDefault {
			violations = append(violations, testRuntimeSeccompViolation)
		}

		if pod.AutomountServiceAccountToken == nil || *pod.AutomountServiceAccountToken {
			violations = append(violations, fmt.Sprintf("%s: %s", resourceName, testRuntimeTokenAutomountViolation))
		}

		if len(pod.Containers) != 1 || pod.Containers[0].Name != cryptoutilSharedMagic.DefaultOTLPServiceDefault {
			violations = append(violations, fmt.Sprintf("%s: expected exactly one application container", resourceName))
		}

		containers := make([]helmRenderedContainer, 0, len(pod.InitContainers)+len(pod.Containers))
		containers = append(containers, pod.InitContainers...)
		containers = append(containers, pod.Containers...)

		materializerHasToken := false

		for _, container := range containers {
			containerName := fmt.Sprintf("%s/%s", resourceName, container.Name)
			security := container.SecurityContext

			if security.AllowPrivilegeEscalation == nil || *security.AllowPrivilegeEscalation {
				violations = append(violations, testRuntimePrivilegeEscalationViolation)
			}

			if security.Privileged != nil && *security.Privileged {
				violations = append(violations, fmt.Sprintf("%s: container must not be privileged", containerName))
			}

			if security.ReadOnlyRootFilesystem == nil || !*security.ReadOnlyRootFilesystem {
				violations = append(violations, fmt.Sprintf("%s: %s", containerName, testRuntimeReadOnlyRootFSViolation))
			}

			if !security.RunAsNonRoot || security.RunAsUser != testRuntimeSecurityUserID ||
				security.RunAsGroup != testRuntimeSecurityGroupID {
				violations = append(violations, fmt.Sprintf("%s: container must use the runtime non-root identity", containerName))
			}

			if !slices.Contains(security.Capabilities.Drop, cryptoutilSharedMagic.TestDefaultLogLevelAll) {
				violations = append(violations, fmt.Sprintf("%s: container must drop all Linux capabilities", containerName))
			}

			if len(security.Capabilities.Add) > 0 {
				violations = append(violations, testRuntimeCapabilityViolation)
			}

			for _, mount := range container.VolumeMounts {
				if container.Name == cryptoutilSharedMagic.DefaultOTLPServiceDefault && mount.Name == testRuntimeBootstrapTokenVolume {
					violations = append(violations, testRuntimeAppTokenMountViolation)
				}

				if container.Name == testRuntimeInitContainerName &&
					mount.Name == testRuntimeBootstrapTokenVolume &&
					mount.MountPath == testRuntimeBootstrapTokenPath {
					materializerHasToken = true
				}
			}
		}

		if !materializerHasToken {
			violations = append(violations, fmt.Sprintf("%s: PKI materializer must receive the bootstrap API token", resourceName))
		}
	}

	if statefulSetCount == 0 {
		violations = append(violations, "rendered runtime chart has no StatefulSet workloads")
	}

	return violations
}

func copyRuntimeChartWithTemplateMutation(
	t *testing.T,
	sourceChart, before, after string,
	replacementCount int,
) string {
	t.Helper()

	destinationChart := filepath.Join(t.TempDir(), "cryptoutil-runtime")
	require.NoError(t, os.MkdirAll(destinationChart, 0o700))

	sourceRoot, err := os.OpenRoot(sourceChart)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sourceRoot.Close())
	})

	err = filepath.WalkDir(sourceChart, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(sourceChart, sourcePath)
		if err != nil {
			return err
		}

		destinationPath := filepath.Join(destinationChart, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(destinationPath, 0o700)
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("runtime chart contains non-regular file %s", relativePath)
		}

		sourceFile, err := sourceRoot.Open(relativePath)
		if err != nil {
			return err
		}

		content, readErr := io.ReadAll(sourceFile)
		closeErr := sourceFile.Close()

		if readErr != nil {
			return readErr
		}

		if closeErr != nil {
			return closeErr
		}

		if filepath.ToSlash(relativePath) == "templates/workloads.yaml" {
			template := string(content)
			if found := strings.Count(template, before); found != replacementCount {
				return fmt.Errorf(
					"runtime template mutation expected %d matches, found %d",
					replacementCount, found,
				)
			}

			content = []byte(strings.Replace(template, before, after, replacementCount))
		}

		if err := os.MkdirAll(filepath.Dir(destinationPath), 0o700); err != nil {
			return err
		}

		return os.WriteFile(destinationPath, content, info.Mode().Perm())
	})
	require.NoError(t, err)

	return destinationChart
}
