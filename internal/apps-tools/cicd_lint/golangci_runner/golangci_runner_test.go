// Copyright (c) 2025-2026 Justin Cranford.
package golangci_runner

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testCombinedBuildTags  = "--build-tags=e2e,integration"
	testUnsupportedFlag    = "unsupported golangci-lint flag"
	testDuplicateFlag      = "duplicate"
	testRootPackage        = "."
	testLinuxOS            = cryptoutilSharedMagic.MockRuntimeGoOS
	testE2EBuildTags       = "--build-tags=e2e"
	testExecutable         = "test-golangci-lint"
	testRunCommand         = cryptoutilSharedMagic.WorkflowSubCmdRun
	testPackageA           = "./a"
	testPackageB           = "./b"
	testE2EPackage         = "./e2e"
	testFixFlag            = "--fix"
	testTimeout30m         = "--timeout=30m"
	testRevisionHEAD       = "--new-from-rev=HEAD~1"
	testRootFile           = "root.go"
	testRootSource         = "package root\n"
	testWindowsOS          = cryptoutilSharedMagic.OSNameWindows
	testE2ETag             = "e2e"
	testE2ESource          = "//go:build e2e\n\npackage e2e\n"
	testIntegrationTag     = "integration"
	testIntegrationPackage = "./integration"
	testLinuxPackage       = "./linux"
	testOnlyPackage        = "./testonly"
	testWindowsPackage     = "./windows"
	testPackageC           = "./c"
	testPackageD           = "./d"
	testPackageE           = "./e"
	testModuleFile         = "go.mod"
	testForeignDir         = "foreign"
	testGitMarker          = cryptoutilSharedMagic.CICDExcludeDirGit
	testNestedDir          = "nested"
)

func TestValidateArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr string
	}{
		{
			name: "supported flags",
			args: []string{testFixFlag, testTimeout30m, testRevisionHEAD, testCombinedBuildTags},
			want: []string{testFixFlag, testTimeout30m, testRevisionHEAD, testCombinedBuildTags},
		},
		{
			name:    "rejects recursive package pattern",
			args:    []string{"./..."},
			wantErr: "package argument",
		},
		{
			name:    "rejects option terminator",
			args:    []string{"--", testFixFlag},
			wantErr: testUnsupportedFlag,
		},
		{
			name:    "rejects unknown flag",
			args:    []string{"--config=foreign.yml"},
			wantErr: testUnsupportedFlag,
		},
		{
			name:    "rejects duplicate fix",
			args:    []string{testFixFlag, testFixFlag},
			wantErr: testDuplicateFlag,
		},
		{
			name:    "rejects duplicate timeout forms",
			args:    []string{"--timeout", "30m", "--timeout=1m"},
			wantErr: testDuplicateFlag,
		},
		{
			name:    "rejects nonpositive timeout",
			args:    []string{"--timeout=-1s"},
			wantErr: "must be positive",
		},
		{
			name:    "rejects missing separated value",
			args:    []string{"--new-from-rev"},
			wantErr: "requires a value",
		},
		{
			name:    "rejects revision flag injection",
			args:    []string{"--new-from-rev=--fix"},
			wantErr: "invalid golangci-lint revision",
		},
		{
			name:    "rejects build-tag flag injection",
			args:    []string{"--build-tags=e2e,--fix"},
			wantErr: "invalid build tag",
		},
		{
			name:    "rejects empty build tag",
			args:    []string{"--build-tags=e2e,"},
			wantErr: "invalid empty build tag",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := validateArgs(test.args)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestFilterEligiblePackages_UsesGOOSBuildTagsAndTestFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeRunnerFile(t, rootDir, testRootFile, testRootSource)
	writeRunnerFile(t, rootDir, filepath.Join(testLinuxOS, "only_linux.go"), "package linux\n")
	writeRunnerFile(t, rootDir, filepath.Join(testWindowsOS, "only_windows.go"), "package windows\n")
	writeRunnerFile(t, rootDir, filepath.Join(testE2ETag, "only_e2e_test.go"), testE2ESource)
	writeRunnerFile(t, rootDir, filepath.Join(testIntegrationTag, "only_integration.go"), "//go:build integration\n\npackage integration\n")
	writeRunnerFile(t, rootDir, filepath.Join("testonly", "only_test.go"), "package testonly\n")

	packagePatterns := []string{testRootPackage, testE2EPackage, testIntegrationPackage, testLinuxPackage, testOnlyPackage, testWindowsPackage}

	tests := []struct {
		name     string
		goos     string
		buildTag []string
		want     []string
	}{
		{
			name: "linux defaults exclude tagged and windows-only packages",
			goos: testLinuxOS,
			want: []string{testRootPackage, testLinuxPackage, testOnlyPackage},
		},
		{
			name:     "linux explicit tags include test-only e2e and integration packages",
			goos:     testLinuxOS,
			buildTag: []string{testE2ETag, testIntegrationTag},
			want:     []string{testRootPackage, testE2EPackage, testIntegrationPackage, testLinuxPackage, testOnlyPackage},
		},
		{
			name: "windows defaults exclude linux-only package",
			goos: testWindowsOS,
			want: []string{testRootPackage, testOnlyPackage, testWindowsPackage},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			buildContext := build.Default
			buildContext.GOOS = test.goos
			buildContext.GOARCH = cryptoutilSharedMagic.MockRuntimeGoArch

			buildContext.BuildTags = append([]string(nil), test.buildTag...)

			got, err := filterEligiblePackages(rootDir, packagePatterns, buildContext, os.ReadDir)

			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestFilterEligiblePackages_FailsClosedOnReaderAndEscapingPatterns(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeRunnerFile(t, rootDir, testRootFile, testRootSource)

	buildContext := build.Default

	injectedErr := errors.New("injected ReadDir failure")
	_, err := filterEligiblePackages(rootDir, []string{testRootPackage}, buildContext, func(string) ([]os.DirEntry, error) {
		return nil, injectedErr
	})
	require.ErrorIs(t, err, injectedErr)

	_, err = filterEligiblePackages(rootDir, []string{"./../outside"}, buildContext, os.ReadDir)
	require.ErrorContains(t, err, "escapes repository root")
}

func TestRunWithFunctions_BuildTagsActivateTaggedPackages(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeRunnerFile(t, rootDir, testRootFile, testRootSource)
	writeRunnerFile(t, rootDir, filepath.Join(testE2ETag, "e2e.go"), testE2ESource)

	var captured [][]string

	err := runWithFunctions(
		context.Background(),
		rootDir,
		[]string{testE2EBuildTags},
		os.Lstat,
		os.ReadDir,
		build.Default,
		func(string) (string, error) { return testExecutable, nil },
		func(_ context.Context, _ string, args []string, _ string) error {
			captured = append(captured, append([]string(nil), args...))

			return nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, [][]string{{testRunCommand, testE2EBuildTags, testRootPackage, testE2EPackage}}, captured)
}

func TestPackageBatches_AreBoundedAndDeterministic(t *testing.T) {
	t.Parallel()

	packagePatterns := []string{testPackageA, testPackageB, testPackageC, testPackageD, testPackageE}
	limits := packageBatchLimits{maxPackages: 2, maxCommandLineUTF16: 500}

	batches, err := packageBatches(golangciExecutable, []string{testRunCommand, testFixFlag}, packagePatterns, limits)

	require.NoError(t, err)
	require.Equal(t, [][]string{{testPackageA, testPackageB}, {testPackageC, testPackageD}, {testPackageE}}, batches)

	longPackage := "./" + strings.Repeat("x", 300)
	_, err = packageBatches(
		golangciExecutable,
		[]string{testRunCommand},
		[]string{longPackage},
		packageBatchLimits{maxPackages: 2, maxCommandLineUTF16: 200},
	)
	require.ErrorContains(t, err, "exceeds the safe command-line length limit")
}

func TestRunWithFunctions_UsesExplicitPackagesAndSkipsNestedCheckout(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeRunnerFile(t, rootDir, testModuleFile, "module fixture.test/root\ngo 1.27.3\n")
	writeRunnerFile(t, rootDir, testRootFile, testRootSource)
	writeRunnerFile(t, rootDir, filepath.Join("pkg", "pkg.go"), "package pkg\n")
	writeRunnerFile(t, rootDir, filepath.Join(testForeignDir, testGitMarker), "gitdir: ../.git/worktrees/foreign\n")
	writeRunnerFile(t, rootDir, filepath.Join(testForeignDir, "foreign.go"), "package foreign\n")

	var captured [][]string

	err := runWithFunctions(
		context.Background(),
		rootDir,
		[]string{testFixFlag},
		os.Lstat,
		os.ReadDir,
		build.Default,
		func(name string) (string, error) {
			require.Equal(t, golangciExecutable, name)

			return testExecutable, nil
		},
		func(_ context.Context, _ string, args []string, workingDirectory string) error {
			require.Equal(t, rootDir, workingDirectory)

			captured = append(captured, append([]string(nil), args...))

			return nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, [][]string{{testRunCommand, testFixFlag, testRootPackage, "./pkg"}}, captured)
	require.NotContains(t, strings.Join(captured[0], " "), testForeignDir)
}

func TestRunWithFunctions_EmptyPackageSetAndDiscoveryErrorsFailClosed(t *testing.T) {
	t.Parallel()

	t.Run("empty package set", func(t *testing.T) {
		t.Parallel()

		lookPathCalled := false
		err := runWithFunctions(
			context.Background(),
			t.TempDir(),
			nil,
			os.Lstat,
			os.ReadDir,
			build.Default,
			func(string) (string, error) {
				lookPathCalled = true

				return testExecutable, nil
			},
			func(context.Context, string, []string, string) error {
				t.Fatal("runner must not execute for an empty package set")

				return nil
			},
		)

		require.ErrorContains(t, err, "no eligible repository Go packages")
		require.False(t, lookPathCalled)
	})

	t.Run("Lstat marker error", func(t *testing.T) {
		t.Parallel()

		rootDir := t.TempDir()
		writeRunnerFile(t, rootDir, filepath.Join(testNestedDir, "nested.go"), "package nested\n")

		injectedErr := errors.New("injected marker Lstat failure")

		err := runWithFunctions(
			context.Background(),
			rootDir,
			nil,
			func(path string) (os.FileInfo, error) {
				if filepath.Clean(path) == filepath.Join(rootDir, testNestedDir, testGitMarker) {
					return nil, injectedErr
				}

				return os.Lstat(path)
			},
			os.ReadDir,
			build.Default,
			func(string) (string, error) {
				t.Fatal("executable lookup must not run after package discovery fails")

				return "", nil
			},
			func(context.Context, string, []string, string) error {
				t.Fatal("runner must not execute after package discovery fails")

				return nil
			},
		)

		require.ErrorIs(t, err, injectedErr)
	})
}

func TestRunWithFunctions_AggregatesBatchFailuresDeterministically(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	const packageCount = 257
	for i := 0; i < packageCount; i++ {
		packageDir := fmt.Sprintf("pkg-%03d", i)
		writeRunnerFile(t, rootDir, filepath.Join(packageDir, "package.go"), "package fixture\n")
	}

	var batchArgs [][]string

	err := runWithFunctions(
		context.Background(),
		rootDir,
		nil,
		os.Lstat,
		os.ReadDir,
		build.Default,
		func(string) (string, error) { return testExecutable, nil },
		func(_ context.Context, _ string, args []string, _ string) error {
			batchArgs = append(batchArgs, append([]string(nil), args...))
			switch len(batchArgs) {
			case 1:
				return errors.New("first analyzer failed")
			case 3:
				return errors.New("third analyzer failed")
			default:
				return nil
			}
		},
	)

	require.Error(t, err)
	require.Len(t, batchArgs, 5)
	require.Contains(t, err.Error(), "golangci-lint failed in 2 of 5 package batches")
	require.Contains(t, err.Error(), "batch 1/5: first analyzer failed")
	require.Contains(t, err.Error(), "batch 3/5: third analyzer failed")
	require.NotContains(t, err.Error(), "batch 2/5")
	require.Less(t, strings.Index(err.Error(), "batch 1/5"), strings.Index(err.Error(), "batch 3/5"))
	require.Len(t, batchArgs[0], 1+maxPackagesPerBatch)
	require.Len(t, batchArgs[4], 2)
}

func TestRunExecutable_FormatsFirstPartyAndLeavesNestedCheckoutUntouched(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath(golangciExecutable); err != nil {
		t.Skipf("%s is unavailable: %v", golangciExecutable, err)
	}

	rootDir := t.TempDir()
	writeRunnerFile(t, rootDir, testModuleFile, "module fixture.test/runner\ngo 1.27.3\n")
	writeRunnerFile(t, rootDir, ".golangci.yml", "version: \"2\"\nlinters:\n  default: none\nformatters:\n  enable:\n    - gofumpt\n")
	firstPartyPath := filepath.Join(rootDir, "format_me.go")
	firstPartyBefore := []byte("package fixture\nfunc Value( ) int{return  1}\n")
	require.NoError(t, os.WriteFile(firstPartyPath, firstPartyBefore, 0o600))

	foreignDir := filepath.Join(rootDir, testForeignDir)
	require.NoError(t, os.MkdirAll(foreignDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(foreignDir, testGitMarker), []byte("nested checkout marker"), 0o600))
	foreignPath := filepath.Join(foreignDir, "broken.go")
	foreignBefore := []byte("package foreign\nfunc broken( {\n")
	require.NoError(t, os.WriteFile(foreignPath, foreignBefore, 0o600))

	err := Run(context.Background(), rootDir, []string{testFixFlag, "--timeout=2m"})

	require.NoError(t, err)

	firstPartyAfter, err := os.ReadFile(firstPartyPath)
	require.NoError(t, err)
	require.NotEqual(t, firstPartyBefore, firstPartyAfter)
	require.Equal(t, "package fixture\n\nfunc Value() int { return 1 }\n", string(firstPartyAfter))

	foreignAfter, err := os.ReadFile(foreignPath)
	require.NoError(t, err)
	require.True(t, bytes.Equal(foreignBefore, foreignAfter), "foreign nested checkout bytes changed")
}

func writeRunnerFile(t *testing.T, rootDir, relativePath, content string) {
	t.Helper()

	filePath := filepath.Join(rootDir, relativePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o700))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))
}
