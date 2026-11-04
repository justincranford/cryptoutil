// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration && checkov

package deploymentgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCICDCheckov "cryptoutil/internal/apps-tools/cicd_checkov"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

func requirePinnedCheckov(t *testing.T) string {
	t.Helper()

	checkovBinary, err := exec.LookPath("checkov")
	require.NoError(t, err, "Checkov 3.3.29 must be installed for rendered-manifest security scans")
	versionCommand, err := cryptoutilCICDCheckov.CommandContext(t.Context(), checkovBinary, cryptoutilSharedMagic.CLIVersionFlag)
	require.NoError(t, err)
	versionCommand.Env = localCheckovEnvironment(t)
	checkovVersion, err := versionCommand.CombinedOutput()
	require.NoError(t, err, string(checkovVersion))
	require.Contains(t, string(checkovVersion), "3.3.29")

	return checkovBinary
}

func localCheckovCommand(
	t *testing.T,
	ctx context.Context,
	checkovBinary string,
	args ...string,
) *exec.Cmd {
	t.Helper()

	args = append(args, "--skip-results-upload", "--skip-download", "--include-all-checkov-policies")
	command, err := cryptoutilCICDCheckov.CommandContext(ctx, checkovBinary, args...)
	require.NoError(t, err)
	command.Env = localCheckovEnvironment(t)

	return command
}

func localCheckovEnvironment(t *testing.T) []string {
	t.Helper()

	return cryptoutilCICDCheckov.LocalCheckovEnvironment(os.Environ(), t.TempDir())
}

func requireCheckovReportIncludesFilePath(
	t *testing.T,
	report checkovJSONReport,
	manifestPath string,
	workingDirectory string,
) {
	t.Helper()

	expectedPath, err := filepath.Abs(manifestPath)
	require.NoError(t, err)
	expectedPath = filepath.Clean(expectedPath)
	expectedInfo, err := os.Stat(expectedPath)
	require.NoError(t, err)

	reportPaths := make([]string, 0,
		len(report.Results.PassedChecks)+len(report.Results.FailedChecks)+len(report.Results.SkippedChecks))
	found := false
	for _, results := range [][]checkovJSONCheck{
		report.Results.PassedChecks,
		report.Results.FailedChecks,
		report.Results.SkippedChecks,
	} {
		for _, result := range results {
			if result.FilePath == "" {
				continue
			}

			absoluteResultPath, pathErr := checkovResultFilePath(result.FilePath, workingDirectory)
			require.NoError(t, pathErr)
			reportPaths = append(reportPaths, absoluteResultPath)
			resultInfo, statErr := os.Stat(absoluteResultPath)
			if statErr == nil && os.SameFile(expectedInfo, resultInfo) {
				found = true
			}
		}
	}

	require.True(
		t,
		found,
		"Checkov results must identify rendered manifest %q in file_path; result paths: %v",
		expectedPath,
		reportPaths,
	)
}

func testCheckovDirectoryScanScopeNegativeControl(t *testing.T, checkovBinary string) {
	t.Helper()

	root := t.TempDir()
	genericDirectories := []string{cryptoutilSharedMagic.CICDExcludeDirBuild, "lib", cryptoutilSharedMagic.CICDExcludeDirEnv, cryptoutilSharedMagic.CICDExcludeDirOutput, cryptoutilSharedMagic.CICDExcludeDirTarget, "coverage", cryptoutilSharedMagic.CICDExcludeDirTemp}
	trackedPaths := make([]string, 0, len(genericDirectories)*2)
	manifestPaths := make([]string, 0, len(genericDirectories)*2)

	for _, directory := range genericDirectories {
		for _, parent := range []string{"", "src"} {
			relativePath := filepath.Join(parent, directory, "unsafe.yaml")
			manifestPath := filepath.Join(root, relativePath)
			nameSuffix := "root"
			if parent != "" {
				nameSuffix = "nested"
			}
			require.NoError(t, os.MkdirAll(filepath.Dir(manifestPath), 0o700))
			require.NoError(t, os.WriteFile(
				manifestPath,
				privilegedKubernetesDeployment(
					"unsafe-"+directory+"-"+nameSuffix,
					"scanner-scope-negative-control",
				),
				0o600,
			))

			trackedPaths = append(trackedPaths, filepath.ToSlash(relativePath))
			manifestPaths = append(manifestPaths, manifestPath)
		}
	}

	gitInit := exec.Command("git", "init", "--quiet")
	gitInit.Dir = root
	initOutput, initErr := gitInit.CombinedOutput()
	require.NoError(t, initErr, string(initOutput))

	gitAddArgs := append([]string{"add", "--force", "--"}, trackedPaths...)
	gitAdd := exec.Command("git", gitAddArgs...)
	gitAdd.Dir = root
	addOutput, addErr := gitAdd.CombinedOutput()
	require.NoError(t, addErr, string(addOutput))

	gitListArgs := append([]string{"ls-files", "--error-unmatch", "--"}, trackedPaths...)
	gitList := exec.Command("git", gitListArgs...)
	gitList.Dir = root
	trackedOutput, trackedErr := gitList.CombinedOutput()
	require.NoError(t, trackedErr, string(trackedOutput))
	require.ElementsMatch(t, trackedPaths, strings.Fields(string(trackedOutput)),
		"generic nested IaC negative controls must be tracked in the temporary repository")

	foreignCheckout := filepath.Join(root, "nested", "third_party", "foreign-checkout")
	foreignGit := filepath.Join(foreignCheckout, cryptoutilSharedMagic.CICDExcludeDirGit)
	require.NoError(t, os.MkdirAll(foreignGit, 0o700))
	gitMarkerPath := filepath.Join(foreignGit, "HEAD")
	gitMarker := []byte("ref: refs/heads/main\n")
	require.NoError(t, os.WriteFile(gitMarkerPath, gitMarker, 0o600))

	foreignManifestPath := filepath.Join(foreignCheckout, "unsafe.yaml")
	foreignManifest := privilegedKubernetesDeployment("foreign-unsafe", "foreign-checkout")
	require.NoError(t, os.WriteFile(foreignManifestPath, foreignManifest, 0o600))
	malformedManifestPath := filepath.Join(foreignCheckout, "malformed.yaml")
	malformedManifest := []byte("apiVersion: [\nkind: Deployment\n")
	require.NoError(t, os.WriteFile(malformedManifestPath, malformedManifest, 0o600))

	skipPathArgs := checkovDirectorySkipPathArgs(t)
	args := []string{"--framework", "kubernetes", "--output", "json", "-d", "."}
	args = append(args, skipPathArgs...)
	nestedGitRoots, discoverErr := cryptoutilCICDCheckov.DiscoverNestedGitRoots(root, args)
	require.NoError(t, discoverErr)
	require.Equal(t, []string{filepath.Join("nested", "third_party", "foreign-checkout")}, nestedGitRoots,
		"the Go wrapper must discover an arbitrary nested .git checkout outside statically ignored roots")

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	var output bytes.Buffer
	var stderr bytes.Buffer
	runner := cryptoutilCICDCheckov.Runner{
		Root:        root,
		CheckovPath: checkovBinary,
		Args:        args,
		Stdout:      &output,
		Stderr:      &stderr,
	}
	scanErr := runner.Run(ctx)
	outputBytes := output.Bytes()
	exitErr := &exec.ExitError{}
	require.True(
		t,
		errors.As(scanErr, &exitErr),
		"tracked privileged IaC must fail Checkov while the foreign worktree stays excluded: %s",
		outputBytes,
	)
	require.Equal(t, 1, exitErr.ExitCode(), stderr.String())

	report := requireCheckovJSONReport(t, outputBytes)
	require.Equal(t, "kubernetes", report.CheckType)
	require.Greater(t, report.Summary.ResourceCount, 0)
	require.Empty(t, report.Results.ParsingErrors,
		"the excluded foreign worktree's malformed manifest must not be parsed")
	require.Contains(t, checkovFailedCheckIDs(report), "CKV_K8S_20")

	for index, manifestPath := range manifestPaths {
		requireCheckovReportIncludesFilePath(t, report, manifestPath, root)
		requireCheckovFailureIncludesFilePath(
			t, report, "CKV_K8S_20", manifestPath, root,
		)
		require.Contains(t, strings.Fields(string(trackedOutput)), trackedPaths[index],
			"each failing generic-path source manifest must be tracked")
	}

	requireCheckovReportExcludesDirectory(t, report, foreignCheckout, root)

	afterGitMarker, readErr := os.ReadFile(gitMarkerPath)
	require.NoError(t, readErr)
	require.Equal(t, gitMarker, afterGitMarker, "excluded foreign Git metadata must remain untouched")
	afterForeignManifest, readErr := os.ReadFile(foreignManifestPath)
	require.NoError(t, readErr)
	require.Equal(t, foreignManifest, afterForeignManifest, "excluded foreign worktree content must remain untouched")
	afterMalformedManifest, readErr := os.ReadFile(malformedManifestPath)
	require.NoError(t, readErr)
	require.Equal(t, malformedManifest, afterMalformedManifest,
		"excluded foreign malformed content must remain untouched")
}

func privilegedKubernetesDeployment(name, namespace string) []byte {
	return []byte(fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
        - name: %s
          image: nginx:1.27.4
          securityContext:
            privileged: true
`, name, namespace, name, name, name))
}

func checkovDirectorySkipPathArgs(t *testing.T) []string {
	t.Helper()

	configPath := filepath.Join("..", "..", "..", ".pre-commit-config.yaml")
	configContents, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config struct {
		Repos []struct {
			Hooks []struct {
				ID            string   `yaml:"id"`
				Entry         string   `yaml:"entry"`
				PassFilenames bool     `yaml:"pass_filenames"`
				Args          []string `yaml:"args"`
			} `yaml:"hooks"`
		} `yaml:"repos"`
	}
	require.NoError(t, yaml.Unmarshal(configContents, &config))

	var hookEntry string
	var hookArgs []string
	var hookPassFilenames bool
	for _, repository := range config.Repos {
		for _, hook := range repository.Hooks {
			if hook.ID == "checkov" {
				hookEntry = hook.Entry
				hookPassFilenames = hook.PassFilenames
				hookArgs = hook.Args
				break
			}
		}
	}
	require.NotEmpty(t, hookArgs, "the full-repository Checkov hook must define scan arguments")
	require.Equal(t, "go", hookEntry, "the local wrapper must use the pinned Checkov hook environment")
	require.False(t, hookPassFilenames,
		"the full-repository Checkov hook must not append changed filenames to directory mode")
	require.GreaterOrEqual(t, len(hookArgs), 3)
	require.Equal(t, []string{cryptoutilSharedMagic.WorkflowSubCmdRun, "./cmd/cicd-checkov", "--"}, hookArgs[:3])
	require.Contains(t, hookArgs, "--skip-results-upload")
	require.Contains(t, hookArgs, "--skip-download")
	require.Contains(t, hookArgs, "--include-all-checkov-policies")
	require.Contains(
		t,
		hookArgs,
		"--skip-check=CKV_DOCKER_2,CKV_DOCKER_3,CKV2_GHA_1,CKV_OPENAPI_3,CKV_OPENAPI_4,CKV_OPENAPI_5,CKV_OPENAPI_21,CKV_SECRET_4",
	)
	foundFullDirectoryScan := false
	for index := 0; index+1 < len(hookArgs); index++ {
		if hookArgs[index] == "-d" && hookArgs[index+1] == "." {
			foundFullDirectoryScan = true

			break
		}
	}
	require.True(t, foundFullDirectoryScan, "the full Checkov hook must preserve -d . directory mode")

	skipPathArgs := make([]string, 0)
	for index := 0; index < len(hookArgs); index++ {
		if strings.HasPrefix(hookArgs[index], "--skip-path=") {
			skipPathArgs = append(skipPathArgs, hookArgs[index])
			continue
		}

		if hookArgs[index] == "--skip-path" && index+1 < len(hookArgs) {
			skipPathArgs = append(skipPathArgs, hookArgs[index], hookArgs[index+1])
			index++
		}
	}
	require.NotEmpty(t, skipPathArgs, "the full-repository Checkov hook must define path exclusions")
	joinedSkipPathArgs := strings.Join(skipPathArgs, "\n")
	for _, renderedHelmTemplatePath := range []string{
		`api[\\/]cryptosuite-registry[\\/]templates[\\/]kubernetes[\\/]cryptoutil-infra[\\/]templates`,
		`api[\\/]cryptosuite-registry[\\/]templates[\\/]kubernetes[\\/]cryptoutil-runtime[\\/]templates`,
		`deployments[\\/]k8s[\\/]charts[\\/]cryptoutil-infra[\\/]templates`,
		`deployments[\\/]k8s[\\/]charts[\\/]cryptoutil-runtime[\\/]templates`,
	} {
		require.Contains(
			t,
			joinedSkipPathArgs,
			renderedHelmTemplatePath,
			"raw non-renderable Helm paths must be excluded only because their rendered manifests are scanned",
		)
	}

	return skipPathArgs
}

func checkovResultFilePath(resultPath, workingDirectory string) (string, error) {
	if !filepath.IsAbs(resultPath) {
		resultPath = filepath.Join(workingDirectory, resultPath)
	}

	absolutePath, err := filepath.Abs(resultPath)
	if err != nil {
		return "", err
	}

	return filepath.Clean(absolutePath), nil
}

func requireCheckovFailureIncludesFilePath(
	t *testing.T,
	report checkovJSONReport,
	checkID string,
	manifestPath string,
	workingDirectory string,
) {
	t.Helper()

	expectedPath, err := filepath.Abs(manifestPath)
	require.NoError(t, err)
	expectedInfo, err := os.Stat(expectedPath)
	require.NoError(t, err)

	var found bool
	resultPaths := make([]string, 0, len(report.Results.FailedChecks))
	for _, result := range report.Results.FailedChecks {
		if result.CheckID != checkID || result.FilePath == "" {
			continue
		}

		resultPath, pathErr := checkovResultFilePath(result.FilePath, workingDirectory)
		require.NoError(t, pathErr)
		resultPaths = append(resultPaths, resultPath)
		resultInfo, statErr := os.Stat(resultPath)
		if statErr == nil && os.SameFile(expectedInfo, resultInfo) {
			found = true
		}
	}

	require.True(
		t,
		found,
		"Checkov must report %s for %q; failure result paths: %v",
		checkID,
		expectedPath,
		resultPaths,
	)
}

func requireCheckovReportExcludesDirectory(
	t *testing.T,
	report checkovJSONReport,
	excludedDirectory string,
	workingDirectory string,
) {
	t.Helper()

	excludedPath, err := filepath.Abs(excludedDirectory)
	require.NoError(t, err)
	for _, results := range [][]checkovJSONCheck{
		report.Results.PassedChecks,
		report.Results.FailedChecks,
		report.Results.SkippedChecks,
	} {
		for _, result := range results {
			if result.FilePath == "" {
				continue
			}

			resultPath, pathErr := checkovResultFilePath(result.FilePath, workingDirectory)
			require.NoError(t, pathErr)
			relativePath, relErr := filepath.Rel(excludedPath, resultPath)
			require.NoError(t, relErr)
			require.False(
				t,
				relativePath == "." ||
					(relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(os.PathSeparator))),
				"Checkov must not report excluded foreign worktree path %q",
				resultPath,
			)
		}
	}
}
