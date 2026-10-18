// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testServerRepositoryMigrationsDirectoryName = "migrations"
	testServerRepositoryMigrationsFilename      = "migrations.go"
	testMismatchSameLine                        = "same"
)

func TestServerDirectoryChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		check     func(string, cryptoutilFitnessRegistry.ProductService, psIDManifest, psIDExclusions) []string
		configure func(*testing.T, string, *psIDManifest, *psIDExclusions)
		want      string
	}{
		{
			name:  "missing server directory",
			check: checkServerDirs,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, _ *psIDExclusions) {
				manifest.RequiredServerDirs = []string{testServerAPIsDirectoryName}
			},
			want: "missing required server subdirectory",
		},
		{
			name:  "server directory exclusion",
			check: checkServerDirs,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, exclusions *psIDExclusions) {
				manifest.RequiredServerDirs = []string{testServerAPIsDirectoryName}
				exclusions.serverDirs = map[string]map[string]bool{testServerAPIsDirectoryName: {cryptoutilSharedMagic.OTLPServiceSMKMS: true}}
			},
		},
		{
			name:  "missing config file",
			check: checkServerConfigFiles,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, _ *psIDExclusions) {
				manifest.RequiredServerConfigFiles = []string{testServerConfigGoFilename}
			},
			want: "missing required server config file",
		},
		{
			name:  "config file exclusion",
			check: checkServerConfigFiles,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, exclusions *psIDExclusions) {
				manifest.RequiredServerConfigFiles = []string{testServerConfigGoFilename}
				exclusions.configFiles = map[string]map[string]bool{testServerConfigGoFilename: {cryptoutilSharedMagic.OTLPServiceSMKMS: true}}
			},
		},
		{
			name:  "missing repository file",
			check: checkServerRepositoryFiles,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, _ *psIDExclusions) {
				manifest.RequiredServerRepositoryFiles = []string{testServerRepositoryMigrationsFilename}
			},
			want: "missing required server repository file",
		},
		{
			name:  "repository file exclusion",
			check: checkServerRepositoryFiles,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, exclusions *psIDExclusions) {
				manifest.RequiredServerRepositoryFiles = []string{testServerRepositoryMigrationsFilename}
				exclusions.repoFiles = map[string]map[string]bool{testServerRepositoryMigrationsFilename: {cryptoutilSharedMagic.OTLPServiceSMKMS: true}}
			},
		},
		{
			name:  "missing repository directory",
			check: checkServerRepositoryDirs,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, _ *psIDExclusions) {
				manifest.RequiredServerRepositoryDirs = []string{testServerRepositoryMigrationsDirectoryName}
			},
			want: "missing required server repository subdirectory",
		},
		{
			name:  "repository directory exclusion",
			check: checkServerRepositoryDirs,
			configure: func(_ *testing.T, _ string, manifest *psIDManifest, exclusions *psIDExclusions) {
				manifest.RequiredServerRepositoryDirs = []string{testServerRepositoryMigrationsDirectoryName}
				exclusions.repoDirs = map[string]map[string]bool{testServerRepositoryMigrationsDirectoryName: {cryptoutilSharedMagic.OTLPServiceSMKMS: true}}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			psDir := t.TempDir()
			manifest := psIDManifest{}
			exclusions := psIDExclusions{}
			tc.configure(t, psDir, &manifest, &exclusions)

			violations := tc.check(psDir, cryptoutilFitnessRegistry.ProductService{
				PSID: cryptoutilSharedMagic.OTLPServiceSMKMS,
			}, manifest, exclusions)

			if tc.want == "" {
				require.Empty(t, violations)

				return
			}

			require.Len(t, violations, 1)
			require.Contains(t, violations[0], tc.want)
		})
	}
}

func TestCheckE2EFiles(t *testing.T) {
	t.Parallel()

	ps := cryptoutilFitnessRegistry.ProductService{
		PSID:    cryptoutilSharedMagic.OTLPServiceSMKMS,
		Service: cryptoutilSharedMagic.KMSServiceName,
	}
	manifest := psIDManifest{RequiredE2EFiles: []string{testRequiredServiceE2ETestFilenameTemplate}}

	t.Run("absent e2e directory is skipped", func(t *testing.T) {
		t.Parallel()

		require.Empty(t, checkE2EFiles(t.TempDir(), ps, manifest, psIDExclusions{}))
	})

	t.Run("missing e2e file is reported", func(t *testing.T) {
		t.Parallel()

		psDir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(psDir, "e2e"), cryptoutilSharedMagic.DirPermissions))

		violations := checkE2EFiles(psDir, ps, manifest, psIDExclusions{})

		require.Len(t, violations, 1)
		require.Contains(t, violations[0], "missing required e2e file: kms_e2e_test.go")
	})

	t.Run("present and excluded files pass", func(t *testing.T) {
		t.Parallel()

		psDir := t.TempDir()
		e2eDir := filepath.Join(psDir, "e2e")
		require.NoError(t, os.Mkdir(e2eDir, cryptoutilSharedMagic.DirPermissions))
		require.NoError(t, os.WriteFile(filepath.Join(e2eDir, "kms_e2e_test.go"), nil, cryptoutilSharedMagic.CacheFilePermissions))

		require.Empty(t, checkE2EFiles(psDir, ps, manifest, psIDExclusions{}))
		require.Empty(t, checkE2EFiles(psDir, ps, manifest, psIDExclusions{
			e2eFiles: map[string]map[string]bool{testRequiredServiceE2ETestFilenameTemplate: {ps.PSID: true}},
		}))
	})
}

func TestServiceTemplateUnsupportedPSID(t *testing.T) {
	t.Parallel()

	ps := cryptoutilFitnessRegistry.ProductService{PSID: cryptoutilSharedMagic.UNKNOWN, Service: cryptoutilSharedMagic.UNKNOWN}

	_, err := serviceRootTemplateValues(ps)

	require.ErrorContains(t, err, "unsupported PS-ID")
	require.Len(t, checkRootTemplates(t.TempDir(), t.TempDir(), ps), 1)
}

func TestCheckPackageTestmainConformance(t *testing.T) {
	t.Parallel()

	ps := cryptoutilFitnessRegistry.ProductService{PSID: cryptoutilSharedMagic.OTLPServiceSMKMS}

	t.Run("required testmain and split variants", func(t *testing.T) {
		t.Parallel()

		psDir := t.TempDir()
		serverDir := filepath.Join(psDir, cryptoutilSharedMagic.CMD_SERVER)
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.DirPermissions))
		require.NoError(t, os.Mkdir(filepath.Join(psDir, "client"), cryptoutilSharedMagic.DirPermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_split_test.go"), nil, cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.Mkdir(filepath.Join(serverDir, "testmain_directory_test.go"), cryptoutilSharedMagic.DirPermissions))

		violations := checkPackageTestmainConformance(psDir, ps, cryptoutilSharedMagic.CMD_SERVER, true)

		require.Len(t, violations, 2)
		require.Contains(t, violations[0], "split TestMain file")
		require.Contains(t, violations[1], "missing required testmain_test.go")
	})

	t.Run("testmain build tags are rejected", func(t *testing.T) {
		t.Parallel()

		psDir := t.TempDir()
		serverDir := filepath.Join(psDir, cryptoutilSharedMagic.CMD_SERVER)
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.DirPermissions))
		require.NoError(t, os.WriteFile(
			filepath.Join(serverDir, "testmain_test.go"),
			[]byte("//go:build integration\n// +build integration\n\npackage server\n"),
			cryptoutilSharedMagic.CacheFilePermissions,
		))

		violations := checkPackageTestmainConformance(psDir, ps, cryptoutilSharedMagic.CMD_SERVER, true)

		require.Len(t, violations, 2)
		require.Contains(t, violations[0], "must not include build tags")
		require.Contains(t, violations[1], "must not include build tags")
	})

	t.Run("optional client package may omit testmain", func(t *testing.T) {
		t.Parallel()

		psDir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(psDir, "client"), cryptoutilSharedMagic.DirPermissions))

		require.Empty(t, checkPackageTestmainConformance(psDir, ps, "client", false))
	})
}

func TestCheckSingleTemplateFilePaths(t *testing.T) {
	t.Parallel()

	ps := cryptoutilFitnessRegistry.ProductService{PSID: cryptoutilSharedMagic.OTLPServiceSMKMS}
	spec := templateFileSpec{
		actualPath:  "server/service.go",
		displayName: "service file",
	}

	t.Run("template read error", func(t *testing.T) {
		t.Parallel()

		violations := checkSingleTemplateFile(filepath.Join(t.TempDir(), "missing.tmpl"), t.TempDir(), spec, ps, nil)

		require.Len(t, violations, 1)
		require.Contains(t, violations[0], "failed to read service file template")
	})

	t.Run("actual file read error", func(t *testing.T) {
		t.Parallel()

		templatePath := filepath.Join(t.TempDir(), "service.tmpl")
		require.NoError(t, os.WriteFile(templatePath, []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		violations := checkSingleTemplateFile(templatePath, t.TempDir(), spec, ps, nil)

		require.Len(t, violations, 1)
		require.Contains(t, violations[0], "failed to read service file")
	})

	t.Run("canonical match", func(t *testing.T) {
		t.Parallel()

		templatePath := filepath.Join(t.TempDir(), "service.tmpl")
		psDir := t.TempDir()
		require.NoError(t, os.WriteFile(templatePath, []byte("//go:build ignore\n\npackage server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		actualPath := filepath.Join(psDir, filepath.FromSlash(spec.actualPath))
		require.NoError(t, os.MkdirAll(filepath.Dir(actualPath), cryptoutilSharedMagic.DirPermissions))
		require.NoError(t, os.WriteFile(actualPath, []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		require.Empty(t, checkSingleTemplateFile(templatePath, psDir, spec, ps, nil))
	})

	t.Run("mismatch is reported", func(t *testing.T) {
		t.Parallel()

		templatePath := filepath.Join(t.TempDir(), "service.tmpl")
		psDir := t.TempDir()
		require.NoError(t, os.WriteFile(templatePath, []byte("package expected\n"), cryptoutilSharedMagic.CacheFilePermissions))

		actualPath := filepath.Join(psDir, filepath.FromSlash(spec.actualPath))
		require.NoError(t, os.MkdirAll(filepath.Dir(actualPath), cryptoutilSharedMagic.DirPermissions))
		require.NoError(t, os.WriteFile(actualPath, []byte("package actual\n"), cryptoutilSharedMagic.CacheFilePermissions))

		violations := checkSingleTemplateFile(templatePath, psDir, spec, ps, nil)

		require.Len(t, violations, 1)
		require.Contains(t, violations[0], "line 1")
	})
}

func TestFirstMismatchLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expected   string
		actual     string
		wantLine   int
		wantExpect string
		wantActual string
	}{
		{name: "first differing line", expected: testMismatchSameLine + "\nexpected", actual: testMismatchSameLine + "\nactual", wantLine: 2, wantExpect: "expected", wantActual: "actual"},
		{name: "actual has extra line", expected: testMismatchSameLine, actual: testMismatchSameLine + "\n" + testMismatchExtraLine, wantLine: 2, wantActual: testMismatchExtraLine},
		{name: "expected has extra line", expected: testMismatchSameLine + "\n" + testMismatchExtraLine, actual: testMismatchSameLine, wantLine: 2, wantExpect: testMismatchExtraLine},
		{name: "identical content", expected: testMismatchSameLine, actual: testMismatchSameLine},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			line, expected, actual := firstMismatchLine(tc.expected, tc.actual)

			require.Equal(t, tc.wantLine, line)
			require.Equal(t, tc.wantExpect, expected)
			require.Equal(t, tc.wantActual, actual)
		})
	}
}
