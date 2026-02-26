// Copyright (c) 2025 Justin Cranford

package no_hardcoded_passwords

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestEnforceHardcodedPasswords_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, []string{})

	require.NoError(t, err)
}

func TestEnforceHardcodedPasswords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
	}{
		{
			name: "valid_dynamic_password",
			fileContent: `package test

import (
	"testing"
	googleUuid "github.com/google/uuid"
)

func TestSomething(t *testing.T) {
	t.Parallel()
	password := googleUuid.NewV7().String()
	_ = password
}
`,
			wantErr: false,
		},
		{
			name: "invalid_hardcoded_password",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	password := "test123"
	_ = password
}
`,
			wantErr: true,
		},
		{
			name: "invalid_hardcoded_secret",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	secret := "secret"
	_ = secret
}
`,
			wantErr: true,
		},
		{
			name: "no_passwords",
			fileContent: `package test

import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()
	value := "some value"
	_ = value
}
`,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			testFiles := []string{testFile}

			err = Check(logger, testFiles)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckHardcodedPasswords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantIssues  bool
	}{
		{
			name:        "hardcoded_password",
			fileContent: `password := "test123"`,
			wantIssues:  true,
		},
		{
			name:        "hardcoded_password_alt",
			fileContent: `password := "password"`,
			wantIssues:  true,
		},
		{
			name:        "hardcoded_secret",
			fileContent: `secret := "secret"`,
			wantIssues:  true,
		},
		{
			name:        "no_hardcoded",
			fileContent: `value := "other value"`,
			wantIssues:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test_test.go")
			err := os.WriteFile(testFile, []byte(tc.fileContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			issues := CheckHardcodedPasswords(testFile)

			if tc.wantIssues {
				require.NotEmpty(t, issues)
			} else {
				require.Empty(t, issues)
			}
		})
	}
}

func TestCheckHardcodedPasswords_ReadFileError(t *testing.T) {
	t.Parallel()

	issues := CheckHardcodedPasswords("/nonexistent/path/that/does/not/exist_test.go")
	require.NotEmpty(t, issues)
	require.Contains(t, issues[0], "Error reading file")
}
