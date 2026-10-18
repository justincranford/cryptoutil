// Copyright (c) 2025-2026 Justin Cranford.

package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testSecretDirectoryReadError = "injected directory read failure"
	testSecretFileReadError      = "injected secret read failure"
	testHighEntropySecretFile    = "hash-pepper-v3.secret"
	testDatabaseUsernameFile     = "postgres-username.secret"
	testDirectoryMode            = 0o700
)

const testDBUsernameSecretValue = "dbuser"

func TestValidateSecretsCommand_InjectedFilesystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		files        map[string]string
		directories  []string
		readDirErr   bool
		readFileErr  bool
		wantSuccess  bool
		wantMessages []string
	}{
		{name: "empty directory", wantSuccess: true},
		{
			name:        "valid secret and non-secret files skip directories",
			files:       map[string]string{testHighEntropySecretFile: strings.Repeat("x", cryptoutilSharedMagic.DockerSecretMinLength), testDatabaseUsernameFile: testDBUsernameSecretValue},
			directories: []string{"nested"},
			wantSuccess: true,
		},
		{
			name: "aggregates suffix empty and short-secret errors",
			files: map[string]string{
				"invalid-name.txt":        "ignored",
				testHighEntropySecretFile: "",
				"api-key.secret":          "short",
				testDatabaseUsernameFile:  "dbuser",
			},
			wantMessages: []string{"required '.secret' suffix", "is empty", "minimum:"},
		},
		{
			name:         "directory read error",
			readDirErr:   true,
			wantMessages: []string{"cannot read", testSecretDirectoryReadError},
		},
		{
			name:         "secret read error",
			files:        map[string]string{testHighEntropySecretFile: strings.Repeat("x", cryptoutilSharedMagic.DockerSecretMinLength)},
			readFileErr:  true,
			wantMessages: []string{"cannot read secret file", testSecretFileReadError},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			secretsDir := t.TempDir()
			for name, contents := range testCase.files {
				err := os.WriteFile(filepath.Join(secretsDir, name), []byte(contents), cryptoutilSharedMagic.CacheFilePermissions)
				require.NoError(t, err)
			}

			for _, name := range testCase.directories {
				err := os.Mkdir(filepath.Join(secretsDir, name), testDirectoryMode)
				require.NoError(t, err)
			}

			readDirFn := os.ReadDir
			if testCase.readDirErr {
				readDirFn = func(string) ([]os.DirEntry, error) {
					return nil, errors.New(testSecretDirectoryReadError)
				}
			}

			readFileFn := os.ReadFile
			if testCase.readFileErr {
				readFileFn = func(string) ([]byte, error) {
					return nil, errors.New(testSecretFileReadError)
				}
			}

			var stdout, stderr bytes.Buffer

			exitCode := validateSecretsCommand(nil, &stdout, &stderr, secretsDir, readDirFn, readFileFn)

			if testCase.wantSuccess {
				require.Zero(t, exitCode)
				require.Contains(t, stdout.String(), "All secrets validated successfully")
				require.Empty(t, stderr.String())

				return
			}

			require.Equal(t, 1, exitCode)

			for _, message := range testCase.wantMessages {
				require.Contains(t, stderr.String(), message)
			}

			require.Empty(t, stdout.String())
		})
	}
}

func TestIsHighEntropySecret_Patterns(t *testing.T) {
	t.Parallel()

	tests := make([]struct {
		name string
		want bool
	}, 0, len(highEntropySecretPatterns)+1)
	for _, pattern := range highEntropySecretPatterns {
		tests = append(tests, struct {
			name string
			want bool
		}{name: strings.ToUpper(pattern), want: true})
	}

	tests = append(tests, struct {
		name string
		want bool
	}{name: "postgres-username", want: false})

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, isHighEntropySecret(testCase.name))
		})
	}
}
