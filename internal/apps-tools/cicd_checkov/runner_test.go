// Copyright (c) 2025-2026 Justin Cranford.
package cicd_checkov

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testDirectoryFlag     = "-d"
	testArgumentSeparator = "--"
	testLocalOnlyMode     = "local-only mode"
	testCurrentDirectory  = "."
)

func TestDiscoverNestedGitRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		markerIsDir      bool
		relativeCheckout string
	}{
		{name: "nested .git directory", markerIsDir: true, relativeCheckout: filepath.Join("src", cryptoutilSharedMagic.CICDExcludeDirBuild, "foreign")},
		{name: "nested .git marker file", relativeCheckout: filepath.Join("src", cryptoutilSharedMagic.CICDExcludeDirTemp, "foreign")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirGit), 0o700))
			checkout := filepath.Join(root, test.relativeCheckout)
			require.NoError(t, os.MkdirAll(checkout, 0o700))

			marker := filepath.Join(checkout, cryptoutilSharedMagic.CICDExcludeDirGit)
			if test.markerIsDir {
				require.NoError(t, os.Mkdir(marker, 0o700))
			} else {
				require.NoError(t, os.WriteFile(marker, []byte("gitdir: ../metadata"), 0o600))
			}

			roots, err := DiscoverNestedGitRoots(root, nil)

			require.NoError(t, err)
			require.Equal(t, []string{filepath.Clean(test.relativeCheckout)}, roots)
		})
	}
}

func TestDiscoverNestedGitRootsHonorsExplicitDirectorySkipPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	excludedCheckout := filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirTestOutput, "foreign")
	require.NoError(t, os.MkdirAll(filepath.Join(excludedCheckout, cryptoutilSharedMagic.CICDExcludeDirGit), 0o700))
	sourceCheckout := filepath.Join(root, "src", cryptoutilSharedMagic.CICDExcludeDirBuild, "foreign")
	require.NoError(t, os.MkdirAll(filepath.Join(sourceCheckout, cryptoutilSharedMagic.CICDExcludeDirGit), 0o700))

	roots, err := DiscoverNestedGitRoots(root, []string{
		`--skip-path=^(?:\.[\\/])?test-output(?:[\\/]|$)`,
	})

	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join("src", cryptoutilSharedMagic.CICDExcludeDirBuild, "foreign")}, roots)
}

func TestBuildNestedGitSkipPathPattern(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "repository")
	roots := []string{
		filepath.Join("nested", "foreign-checkout"),
		filepath.Join("src", cryptoutilSharedMagic.CICDExcludeDirTemp, "another-checkout"),
	}
	pattern, err := buildNestedGitSkipPathPattern(root, roots)
	require.NoError(t, err)

	goCompatiblePattern := strings.ReplaceAll(pattern, "(?:", "(")
	compiledPattern, err := regexp.Compile(goCompatiblePattern)
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "relative path with dot prefix",
			path: filepath.Join(testCurrentDirectory, "nested", "foreign-checkout", "unsafe.yaml"),
			want: true,
		},
		{
			name: "absolute path",
			path: filepath.Join(root, "src", cryptoutilSharedMagic.CICDExcludeDirTemp, "another-checkout", "unsafe.yaml"),
			want: true,
		},
		{
			name: "similar prefix is not excluded",
			path: filepath.Join(testCurrentDirectory, "nested", "foreign-checkout-copy", "unsafe.yaml"),
			want: false,
		},
		{
			name: "tracked generic source remains eligible",
			path: filepath.Join(testCurrentDirectory, "src", cryptoutilSharedMagic.CICDExcludeDirBuild, "tracked.yaml"),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, compiledPattern.MatchString(test.path))
		})
	}
}

func TestNormalizeCheckovArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "full directory mode", args: []string{testDirectoryFlag, testCurrentDirectory}},
		{name: "long directory mode", args: []string{testCommandDirectoryArgument, testCurrentDirectory}},
		{name: "go run separator", args: []string{testArgumentSeparator, testDirectoryFlag, testCurrentDirectory}},
		{name: "file mode rejected", args: []string{"-f", "manifest.yaml"}, wantErr: "full -d . directory scan"},
		{name: "missing directory rejected", args: []string{"--quiet"}, wantErr: "exactly one -d ."},
		{name: "non-root directory rejected", args: []string{testDirectoryFlag, "deployments"}, wantErr: "must be ."},
		{name: "multiple directory roots rejected", args: []string{testDirectoryFlag, testCurrentDirectory, testDirectoryFlag, testCurrentDirectory}, wantErr: "found 2"},
		{name: "API key argument rejected", args: []string{testDirectoryFlag, testCurrentDirectory, "--bc-api-key=do-not-log"}, wantErr: testLocalOnlyMode},
		{name: "external policy source rejected", args: []string{testDirectoryFlag, testCurrentDirectory, "--external-checks-git=https://example.invalid"}, wantErr: testLocalOnlyMode},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeCheckovArguments(test.args)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				require.NotContains(t, strings.Join(got, " "), "do-not-log")

				return
			}

			require.NoError(t, err)

			want := test.args
			if len(want) > 0 && want[0] == testArgumentSeparator {
				want = want[1:]
			}

			require.Equal(t, want, got)
		})
	}
}

func TestAppendLocalOnlyArgumentsPreservesChecksAndNestedExclusions(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "repository")
	baseArgs := []string{"--skip-check=CKV_DOCKER_2", testDirectoryFlag, testCurrentDirectory}
	roots := []string{filepath.Join("nested", "foreign")}

	got, err := appendLocalOnlyArguments(baseArgs, root, roots)

	require.NoError(t, err)
	require.Contains(t, got, "--skip-check=CKV_DOCKER_2")
	require.Contains(t, got, "--skip-results-upload")
	require.Contains(t, got, "--skip-download")
	require.Contains(t, got, "--include-all-checkov-policies")
	require.Contains(t, got, "--skip-path")
	require.Contains(t, got, testDirectoryFlag)
	require.Contains(t, got, testCurrentDirectory)
}

func TestAppendLocalOnlyArgumentsRejectsDisabledLocalOnlyFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		flag string
	}{
		{name: "results upload", args: []string{testDirectoryFlag, testCurrentDirectory, "--skip-results-upload=false"}, flag: "--skip-results-upload"},
		{name: "platform downloads", args: []string{testDirectoryFlag, testCurrentDirectory, "--skip-download=false"}, flag: "--skip-download"},
		{name: "all policies", args: []string{testDirectoryFlag, testCurrentDirectory, "--include-all-checkov-policies=false"}, flag: "--include-all-checkov-policies"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			normalized, err := normalizeCheckovArguments(test.args)
			require.NoError(t, err)

			_, err = appendLocalOnlyArguments(normalized, t.TempDir(), nil)

			require.ErrorContains(t, err, test.flag)
		})
	}
}

func TestLocalCheckovEnvironmentIsolatesCredentialsAndHome(t *testing.T) {
	t.Parallel()

	homeDirectory := filepath.Join(t.TempDir(), "isolated-home")
	sourceEnvironment := []string{
		"PATH=/pinned/checkov/bin",
		"BC_API_KEY=secret-value",
		"BC_API_URL=https://platform.example.invalid",
		"PRISMA_API_URL=https://prisma.example.invalid",
		"REPO_ID=repo-secret",
		"CHECKOV_CONFIG=C:\\Users\\someone\\.checkov.yml",
		"CHECKOV_EXTERNAL_CHECKS_GIT=https://example.invalid/policies.git",
		"HTTP_PROXY=http://proxy-user:proxy-secret@example.invalid",
		"PYTHONPATH=C:\\Users\\someone\\site-packages",
		"HOME=C:\\Users\\someone",
		"USERPROFILE=C:\\Users\\someone",
		"SYSTEMROOT=C:\\Windows",
		"LANG=en_US.UTF-8",
		"TEMP=C:\\Temp\\inherited",
		"TMP=C:\\Temp\\inherited-tmp",
		"TMPDIR=C:\\tmp\\inherited",
	}

	environment := LocalCheckovEnvironment(sourceEnvironment, homeDirectory)
	joinedEnvironment := strings.Join(environment, "\n")

	require.Contains(t, environment, "PATH=/pinned/checkov/bin")
	require.Contains(t, environment, "HOME="+homeDirectory)
	require.Contains(t, environment, "USERPROFILE="+homeDirectory)
	require.Contains(t, environment, "SYSTEMROOT=C:\\Windows")
	require.Contains(t, environment, "LANG=en_US.UTF-8")
	require.Contains(t, environment, "TEMP=C:\\Temp\\inherited")
	require.Contains(t, environment, "TMP=C:\\Temp\\inherited-tmp")
	require.Contains(t, environment, "TMPDIR=C:\\tmp\\inherited")
	require.NotContains(t, joinedEnvironment, "secret-value")
	require.NotContains(t, joinedEnvironment, "BC_API_KEY")
	require.NotContains(t, joinedEnvironment, "PRISMA_API_URL")
	require.NotContains(t, joinedEnvironment, "REPO_ID")
	require.NotContains(t, joinedEnvironment, "CHECKOV_CONFIG")
	require.NotContains(t, joinedEnvironment, "CHECKOV_EXTERNAL_CHECKS_GIT")
	require.NotContains(t, joinedEnvironment, "proxy-secret")
	require.NotContains(t, joinedEnvironment, "PYTHONPATH")
	require.NotContains(t, joinedEnvironment, "C:\\Users\\someone")
}

func TestRejectRemoteCheckovConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contents    string
		wantErr     string
		secretValue string
	}{
		{
			name:     "local scan settings are permitted",
			contents: "framework: kubernetes\nskip-results-upload: true\n",
		},
		{
			name:        "platform credential is rejected",
			contents:    "integration:\n  bc_api_key: never-print-this\n",
			wantErr:     "bc_api_key",
			secretValue: "never-print-this",
		},
		{
			name:     "contributor reporting setting is rejected",
			contents: "repo-id: project\n",
			wantErr:  "repo_id",
		},
		{
			name:     "remote policy source is rejected",
			contents: "external-checks-git: https://example.invalid/policies.git\n",
			wantErr:  "external_checks_git",
		},
		{
			name:     "external module download is rejected",
			contents: "download_external_modules: true\n",
			wantErr:  "download_external_modules",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			configPath := filepath.Join(root, ".checkov.yaml")
			require.NoError(t, os.WriteFile(configPath, []byte(test.contents), 0o600))

			err := rejectRemoteCheckovConfiguration(root)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)

				if test.secretValue != "" {
					require.NotContains(t, err.Error(), test.secretValue)
				}

				return
			}

			require.NoError(t, err)
		})
	}
}
