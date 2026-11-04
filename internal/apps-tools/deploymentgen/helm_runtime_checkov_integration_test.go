// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration && checkov

package deploymentgen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type checkovJSONReport struct {
	CheckType string `json:"check_type"`
	Results   struct {
		PassedChecks  []checkovJSONCheck `json:"passed_checks"`
		FailedChecks  []checkovJSONCheck `json:"failed_checks"`
		SkippedChecks []checkovJSONCheck `json:"skipped_checks"`
		ParsingErrors []string           `json:"parsing_errors"`
	} `json:"results"`
	Summary struct {
		Passed        int `json:"passed"`
		Failed        int `json:"failed"`
		Skipped       int `json:"skipped"`
		ParsingErrors int `json:"parsing_errors"`
		ResourceCount int `json:"resource_count"`
	} `json:"summary"`
}

type checkovJSONCheck struct {
	CheckID  string `json:"check_id"`
	FilePath string `json:"file_path"`
	Resource string `json:"resource"`
}

type checkovManifestScan struct {
	subject      string
	directory    string
	manifestPath string
}

const (
	expectedInfrastructureCheckovPassedCheckCount = 1102
	testCheckovDiagnosticSubject                  = "release 12"
)

func TestHelmRuntimeCheckovScansProductionRenderedFullMatrix(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	checkovBinary := requirePinnedCheckov(t)

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)
	require.Len(t, inventory.Releases, 55)

	root := t.TempDir()
	valuesOutput := t.TempDir()

	images := chartImageFixture(scopes)
	canonicalChart := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-runtime",
	)
	mirrorChart := filepath.Join("..", "..", "..", "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", "cryptoutil-runtime")
	manifestScans := make([]checkovManifestScan, 0, len(inventory.Releases)*2)
	renderedManifestCount := 0

	for index, release := range inventory.Releases {
		releaseOutput := filepath.Join(root, fmt.Sprintf("release-%02d", index))
		canonicalOutput := filepath.Join(releaseOutput, "canonical")
		mirrorOutput := filepath.Join(releaseOutput, "mirror")
		for _, directory := range []string{canonicalOutput, mirrorOutput} {
			require.NoError(t, os.MkdirAll(directory, 0o700))
		}
		releaseInventory := inventoryWithRelease(inventory, release.Key)
		values, projectionErr := BuildRuntimeChartValues(
			model, releaseInventory, release.Key, images, chartMaterializerFixture(),
			ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
			chartRuntimeInputFixture(releaseInventory, release.Key),
		)
		require.NoError(t, projectionErr)

		valuesJSON, marshalErr := MarshalRuntimeChartValues(values)
		require.NoError(t, marshalErr)

		valuesPath := filepath.Join(valuesOutput, fmt.Sprintf("runtime-%02d.json", index))
		require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

		for _, chart := range []struct {
			name      string
			path      string
			outputDir string
		}{
			{name: "canonical", path: canonicalChart, outputDir: canonicalOutput},
			{name: "mirror", path: mirrorChart, outputDir: mirrorOutput},
		} {
			rendered, renderErr := renderHelmDocuments(
				t, helmBinary, chart.path, values.HelmRelease, values.Namespace, valuesPath, chart.outputDir, index,
			)
			require.NoError(t, renderErr)
			renderedManifestCount++
			require.NotEmpty(t, rendered)
			require.Empty(t, runtimeRenderedSecurityViolations(rendered))

			expected := expectedRuntimeChartObjects(release)
			renderedKeys, identityErr := objectKeysFromRendered(rendered, release.Namespace)
			require.NoError(t, identityErr)
			require.Equal(t, objectKeys(expected, release.Namespace), renderedKeys)

			actualPVCs, pvcErr := controllerGeneratedPVCs(rendered, release.Namespace)
			require.NoError(t, pvcErr)
			require.Equal(
				t, namespacedObjectKeys(expectedRuntimeControllerPVCs(release)),
				namespacedObjectKeys(actualPVCs),
			)

			manifestScans = append(manifestScans, checkovManifestScan{
				subject:      fmt.Sprintf("release %d %s", index, chart.name),
				directory:    chart.outputDir,
				manifestPath: filepath.Join(chart.outputDir, fmt.Sprintf("rendered-%02d.yaml", index)),
			})
		}
	}
	require.Equal(t, len(inventory.Releases)*2, renderedManifestCount,
		"Checkov must scan canonical and mirror renders for every runtime release")
	require.Len(t, manifestScans, len(inventory.Releases)*2)

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Minute)
	defer cancel()

	for _, manifestScan := range manifestScans {
		command := localCheckovCommand(
			t,
			ctx,
			checkovBinary,
			"--framework",
			"kubernetes",
			"--output",
			"json",
			"--directory",
			".",
		)
		command.Dir = manifestScan.directory

		output, scanErr := command.Output()
		if scanErr != nil {
			t.Fatal(checkovScanFailureMessage(manifestScan.subject, scanErr, output))
		}

		report := requireCheckovJSONReport(t, output)
		require.Equal(t, "kubernetes", report.CheckType)
		require.Greater(t, report.Summary.ResourceCount, 0,
			"Checkov must analyze rendered Kubernetes resources for %s", manifestScan.subject)
		require.Greater(t, len(report.Results.PassedChecks)+len(report.Results.FailedChecks), 0,
			"Checkov must execute applicable Kubernetes checks for %s", manifestScan.subject)
		require.Empty(t, report.Results.FailedChecks,
			"rendered runtime charts must pass Checkov checks for %s", manifestScan.subject)
		require.Empty(t, report.Results.SkippedChecks,
			"runtime manifests must not hide Checkov skips for %s", manifestScan.subject)
		require.Empty(t, report.Results.ParsingErrors,
			"Checkov must parse every rendered runtime manifest for %s", manifestScan.subject)
		require.Zero(t, report.Summary.Failed)
		require.Zero(t, report.Summary.Skipped)
		require.Zero(t, report.Summary.ParsingErrors)
		require.Equal(t, len(report.Results.PassedChecks), report.Summary.Passed)
		requireCheckovReportIncludesFilePath(t, report, manifestScan.manifestPath, manifestScan.directory)
	}

	t.Run("unsafe manifest fails its expected Checkov check", func(t *testing.T) {
		t.Parallel()

		unsafeDirectory := t.TempDir()
		unsafeManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: unsafe
  namespace: scanner-negative-control
spec:
  replicas: 1
  selector:
    matchLabels:
      app: unsafe
  template:
    metadata:
      labels:
        app: unsafe
    spec:
      containers:
        - name: unsafe
          image: nginx:1.27.4
          securityContext:
            privileged: true
`
		manifestPath := filepath.Join(unsafeDirectory, "unsafe.yaml")
		require.NoError(t, os.WriteFile(manifestPath, []byte(unsafeManifest), 0o600))

		negativeContext, negativeCancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer negativeCancel()
		negativeCommand := localCheckovCommand(
			t,
			negativeContext,
			checkovBinary,
			"--framework",
			"kubernetes",
			"--output",
			"json",
			"--directory",
			".",
		)
		negativeCommand.Dir = unsafeDirectory
		negativeOutput, negativeErr := negativeCommand.Output()
		exitErr := &exec.ExitError{}
		require.True(
			t, errors.As(negativeErr, &exitErr),
			"unsafe directory scan must produce a Checkov failure, output: %s", negativeOutput,
		)
		require.Equal(t, 1, exitErr.ExitCode(), string(exitErr.Stderr))

		negativeReport := requireCheckovJSONReport(t, negativeOutput)
		require.Equal(t, "kubernetes", negativeReport.CheckType)
		require.Greater(t, negativeReport.Summary.ResourceCount, 0)
		require.Greater(t, len(negativeReport.Results.FailedChecks), 0)
		require.Zero(t, negativeReport.Summary.ParsingErrors)
		requireCheckovReportIncludesFilePath(t, negativeReport, manifestPath, unsafeDirectory)

		require.Contains(t, checkovFailedCheckIDs(negativeReport), "CKV_K8S_20",
			"Checkov must reject privileged containers")
	})

	t.Run("repository exclusions preserve nested generic IaC and prune ignored worktrees", func(t *testing.T) {
		t.Parallel()

		testCheckovDirectoryScanScopeNegativeControl(t, checkovBinary)
	})
}

func TestHelmInfrastructureCheckovScansProductionRenderedSelections(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	checkovBinary := requirePinnedCheckov(t)
	model, scopes := fixture(t)

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	root := t.TempDir()
	infrastructureDirectory := filepath.Join(root, "infrastructure")
	require.NoError(t, os.MkdirAll(infrastructureDirectory, 0o700))
	valuesOutput := t.TempDir()
	manifestScans := make([]checkovManifestScan, 0, 4)

	charts := []struct {
		name string
		path string
	}{
		{
			name: "canonical",
			path: filepath.Join(
				"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-infra",
			),
		},
		{
			name: "mirror",
			path: filepath.Join("..", "..", "..", "deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "charts", "cryptoutil-infra"),
		},
	}
	selectors := []struct {
		name     string
		selector string
	}{
		{name: "full registry"},
		{name: "SQLite-only", selector: testServiceSMKMSMemSelector},
	}

	expectedResourceCount := 0
	projectedResourceCount := 0
	projectedGraphVertexCount := 0
	expectedSkippedChecks := make([]string, 0, 4)
	renderedChartCount := 0

	for selectionIndex, selection := range selectors {
		inventory, buildErr := build(model, scopes, selection.selector)
		require.NoError(t, buildErr)

		values, projectionErr := BuildInfrastructureChartValues(model, inventory, defaults)
		require.NoError(t, projectionErr)

		valuesJSON, marshalErr := MarshalInfrastructureChartValues(values)
		require.NoError(t, marshalErr)

		valuesPath := filepath.Join(valuesOutput, fmt.Sprintf("infra-%02d.json", selectionIndex))
		require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

		expected := expectedInfrastructureChartObjects(inventory)
		projectedResourceCount += len(expected) * len(charts)
		projectedGraphVertexCount += len(expected) * len(charts)
		// Checkov includes a separate Pod-template vertex for each StatefulSet.
		for _, object := range expected {
			if object.Kind == testStatefulSetKind {
				projectedGraphVertexCount += len(charts)
			}
		}

		for chartIndex, chart := range charts {
			outputDirectory := filepath.Join(
				infrastructureDirectory, fmt.Sprintf("selection-%02d", selectionIndex), chart.name,
			)
			require.NoError(t, os.MkdirAll(outputDirectory, 0o700))

			rendered, renderErr := renderHelmDocuments(
				t, helmBinary, chart.path, values.HelmRelease, values.Namespace, valuesPath, outputDirectory,
				selectionIndex*len(charts)+chartIndex,
			)
			require.NoError(t, renderErr)
			require.NotEmpty(t, rendered, "%s infrastructure selection, %s chart", selection.name, chart.name)

			renderedKeys, identityErr := objectKeysFromRendered(rendered, inventory.InfraNamespace)
			require.NoError(t, identityErr)
			require.Equal(t, objectKeys(expected, inventory.InfraNamespace), renderedKeys,
				"%s infrastructure selection, %s chart", selection.name, chart.name)
			assertRenderedDatabaseCatalog(t, rendered, inventory)
			assertRenderedInfrastructureSecurity(t, rendered, inventory)

			actualPVCs, pvcErr := controllerGeneratedPVCs(rendered, inventory.InfraNamespace)
			require.NoError(t, pvcErr)
			require.Equal(
				t, namespacedObjectKeys(expectedInfrastructureControllerPVCs(inventory)),
				namespacedObjectKeys(actualPVCs),
			)

			expectedResourceCount += len(rendered)
			renderedChartCount++
			manifestScans = append(manifestScans, checkovManifestScan{
				subject:      fmt.Sprintf("%s infrastructure selection, %s chart", selection.name, chart.name),
				directory:    outputDirectory,
				manifestPath: filepath.Join(outputDirectory, fmt.Sprintf("rendered-%02d.yaml", selectionIndex*len(charts)+chartIndex)),
			})

			for _, name := range []string{
				cryptoutilSharedMagic.KMSPostgresLeaderContainer,
				cryptoutilSharedMagic.PKIInitPostgresFollowerService,
			} {
				if inventoryContainsObject(inventory.Infra, testStatefulSetKind, name) {
					expectedSkippedChecks = append(
						expectedSkippedChecks,
						fmt.Sprintf("CKV_K8S_40|%s.%s.%s", testStatefulSetKind, inventory.InfraNamespace, name),
					)
				}
			}
		}
	}

	require.Equal(t, len(selectors)*len(charts), renderedChartCount)
	require.Len(t, manifestScans, len(selectors)*len(charts))
	require.Equal(t, projectedResourceCount, expectedResourceCount,
		"the full and SQLite-only canonical-and-mirror selections must render the exact projected resource count")
	require.Greater(t, expectedResourceCount, 0)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()

	command := localCheckovCommand(
		t, ctx, checkovBinary, "--framework", "kubernetes", "--output", "json", "--directory", ".",
	)
	command.Dir = infrastructureDirectory

	output, scanErr := command.Output()
	if scanErr != nil {
		t.Fatal(checkovScanFailureMessage("infrastructure selections", scanErr, output))
	}

	report := requireCheckovJSONReport(t, output)
	require.Equal(t, "kubernetes", report.CheckType)
	require.Greater(t, report.Summary.ResourceCount, 0)
	t.Logf(
		"infrastructure Checkov report: expected_manifest_resources=%d graph_vertices=%d passed_checks=%d failed_checks=%d skipped_checks=%d parsing_errors=%d",
		expectedResourceCount, report.Summary.ResourceCount, report.Summary.Passed,
		len(report.Results.FailedChecks), len(report.Results.SkippedChecks), report.Summary.ParsingErrors,
	)
	require.Equal(
		t, projectedGraphVertexCount, report.Summary.ResourceCount,
		"Checkov must analyze every graph vertex across canonical and mirror infrastructure full and SQLite-only selections",
	)
	require.Greater(t, len(report.Results.PassedChecks), 0, "Checkov must execute applicable Kubernetes checks")
	require.Empty(t, report.Results.FailedChecks, "rendered infrastructure charts must pass applicable Checkov checks")
	require.Empty(t, report.Results.ParsingErrors, "Checkov must parse every rendered infrastructure manifest")
	require.Zero(t, report.Summary.Failed)
	require.Zero(t, report.Summary.ParsingErrors)
	require.Equal(t, expectedInfrastructureCheckovPassedCheckCount, report.Summary.Passed)
	require.Equal(t, len(report.Results.PassedChecks), report.Summary.Passed)

	actualSkippedChecks := make([]string, 0, len(report.Results.SkippedChecks))
	for _, skippedCheck := range report.Results.SkippedChecks {
		actualSkippedChecks = append(actualSkippedChecks, skippedCheck.CheckID+"|"+skippedCheck.Resource)
	}

	require.Len(t, report.Results.SkippedChecks, 4)
	require.Equal(t, len(report.Results.SkippedChecks), report.Summary.Skipped)
	require.ElementsMatch(t, expectedSkippedChecks, actualSkippedChecks,
		"only per-resource CKV_K8S_40 PostgreSQL UID exceptions may be skipped")

	for _, manifestScan := range manifestScans {
		manifestContext, manifestCancel := context.WithTimeout(ctx, time.Minute)
		manifestCommand := localCheckovCommand(
			t,
			manifestContext,
			checkovBinary,
			"--framework",
			"kubernetes",
			"--output",
			"json",
			"--directory",
			".",
		)
		manifestCommand.Dir = manifestScan.directory
		manifestOutput, manifestErr := manifestCommand.Output()
		manifestCancel()
		if manifestErr != nil {
			t.Fatal(checkovScanFailureMessage(manifestScan.subject, manifestErr, manifestOutput))
		}

		manifestReport := requireCheckovJSONReport(t, manifestOutput)
		require.Equal(t, "kubernetes", manifestReport.CheckType)
		require.Greater(t, manifestReport.Summary.ResourceCount, 0, "%s", manifestScan.subject)
		require.Greater(
			t,
			len(manifestReport.Results.PassedChecks)+len(manifestReport.Results.SkippedChecks),
			0,
			"Checkov must execute applicable checks for %s",
			manifestScan.subject,
		)
		require.Empty(t, manifestReport.Results.FailedChecks, "%s", manifestScan.subject)
		require.Empty(t, manifestReport.Results.ParsingErrors, "%s", manifestScan.subject)
		require.Zero(t, manifestReport.Summary.Failed, "%s", manifestScan.subject)
		require.Zero(t, manifestReport.Summary.ParsingErrors, "%s", manifestScan.subject)
		requireCheckovReportIncludesFilePath(
			t, manifestReport, manifestScan.manifestPath, manifestScan.directory,
		)
	}

	t.Logf(
		"infrastructure Checkov selections: full_and_sqlite_only canonical_and_mirror_renders=%d resources=%d passed_checks=%d skipped_checks=%d parsing_errors=%d",
		renderedChartCount, report.Summary.ResourceCount, report.Summary.Passed,
		len(report.Results.SkippedChecks), report.Summary.ParsingErrors,
	)
}
