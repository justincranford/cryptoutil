// Copyright (c) 2025 Justin Cranford

package banned_product_names

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	err := os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions)
	require.NoError(t, err)

	path := filepath.Join(dir, name)
	err = os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	return path
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_NoBannedPhrases_Passes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, tmp, "types.go", "package sm\n\ntype IMService struct{}\n")
	writeFile(t, tmp, "config.yml", "service-name: sm-im\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_DetectsBannedPhrase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		phrase  string
		file    string
		content string
	}{
		{
			name:    cryptoutilSharedMagic.OTLPServiceSMIM,
			phrase:  "cipher-im",
			file:    "config.yml",
			content: "service-name: cipher-im\n",
		},
		{
			name:    "CipherIM",
			phrase:  "CipherIM",
			file:    "types.go",
			content: "package foo\n\ntype CipherIM struct{}\n",
		},
		{
			name:    "cipher_im",
			phrase:  "cipher_im",
			file:    "schema.sql",
			content: "CREATE TABLE cipher_im_messages (id TEXT PRIMARY KEY);\n",
		},
		{
			name:    "Cipher IM",
			phrase:  "Cipher IM",
			file:    "README.md",
			content: "# Cipher IM Service\n",
		},
		{
			name:    "cryptoutilCmdCipher",
			phrase:  "cryptoutilCmdCipher",
			file:    "main.go",
			content: "package main\n\nfunc main() { cryptoutilCmdCipher.Run() }\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			writeFile(t, tmp, tc.file, tc.content)

			err := CheckInDir(newTestLogger(), tmp)
			require.Error(t, err)
			require.Contains(t, err.Error(), "banned product/service name violations")
		})
	}
}

func TestCheckInDir_CipherSubstringNotBanned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		file    string
		content string
	}{
		{
			name:    "cipher_block",
			file:    "crypto.go",
			content: "package crypto\n\nvar _ cipher.Block\n",
		},
		{
			name:    "ciphertext",
			file:    "aes.go",
			content: "package aes\n\nfunc Encrypt(plaintext []byte) (ciphertext []byte) { return nil }\n",
		},
		{
			name:    "decipher",
			file:    "decode.go",
			content: "package decode\n\n// decipher decodes the input.\nfunc decipher(b []byte) {}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			writeFile(t, tmp, tc.file, tc.content)

			err := CheckInDir(newTestLogger(), tmp)
			require.NoError(t, err, "substring 'cipher' must not trigger banned phrase detection")
		})
	}
}

func TestCheckInDir_SkipsExcludedDirs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		subDir string
	}{
		{name: "git dir skipped", subDir: cryptoutilSharedMagic.CICDExcludeDirGit},
		{name: "vendor dir skipped", subDir: cryptoutilSharedMagic.CICDExcludeDirVendor},
		{name: "docs dir skipped", subDir: cryptoutilSharedMagic.CICDExcludeDirDocs},
		{name: "test-output dir skipped", subDir: cryptoutilSharedMagic.CICDExcludeDirTestOutput},
		{name: "banned_product_names dir skipped", subDir: cryptoutilSharedMagic.CICDExcludeDirBannedProductNamesCheck},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			writeFile(t, filepath.Join(tmp, tc.subDir), "old.go", "package old\n\ntype CipherIM struct{}\n")

			err := CheckInDir(newTestLogger(), tmp)
			require.NoError(t, err)
		})
	}
}

func TestCheckInDir_SkipsNonScannedExtensions(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeFile(t, tmp, "ignore.txt", "CipherIM\n")
	writeFile(t, tmp, "ignore.json", `{"name": "CipherIM"}`)
	writeFile(t, tmp, "ignore.sh", "#!/bin/sh\necho CipherIM\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipsTestGoFiles(t *testing.T) {
	t.Parallel()

	// _test.go files may reference banned phrases as negative test data
	// (e.g., verifying that a validator rejects the old name). They must not
	// be treated as production drift.
	tmp := t.TempDir()
	writeFile(t, tmp, "old_checker_test.go", "package foo\n\nconst badName = \"CipherIM\"\n")

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestFindViolationsInFile_CorrectLineNumber(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := writeFile(t, tmp, "types.go", "package foo\n\n// CipherIM was the old name.\ntype SMService struct{}\n")

	violations, err := FindViolationsInFile(path)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, 3, violations[0].Line)
	require.Equal(t, "CipherIM", violations[0].Phrase)
}

func TestFindViolationsInFile_NonExistentFile(t *testing.T) {
	t.Parallel()

	violations, err := FindViolationsInFile("/nonexistent/path/missing.go")
	require.Error(t, err)
	require.Nil(t, violations)
}

func TestFindViolationsInDir_NonExistentRoot(t *testing.T) {
	t.Parallel()

	violations, err := FindViolationsInDir("/nonexistent/root/dir")
	require.Error(t, err)
	require.Nil(t, violations)
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	err = Check(newTestLogger())
	require.NoError(t, err)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}
