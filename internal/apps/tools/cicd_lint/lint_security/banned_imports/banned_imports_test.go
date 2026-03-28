// Copyright (c) 2025 Justin Cranford

package banned_imports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCheck_NoFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := Check(logger, map[string][]string{})

	require.NoError(t, err)
}

func TestCheck_CleanFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "clean.go")
	content := `package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func main() {
	fmt.Println(rand.Reader, sha256.New())
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.NoError(t, err)
}

func TestCheck_BannedMathRand(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad.go")
	content := `package main

import "math/rand"

func main() {
	_ = rand.Intn(10)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_BannedArgon2(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad_argon2.go")
	content := `package main

import (
	"golang.org/x/crypto/argon2"
)

func main() {
	_ = argon2.IDKey(nil, nil, 1, 64*1024, 4, 32)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_BannedBcrypt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad_bcrypt.go")
	content := `package main

import "golang.org/x/crypto/bcrypt"

func main() {
	_, _ = bcrypt.GenerateFromPassword(nil, 12)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_BannedDES(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad_des.go")
	content := `package main

import "crypto/des"

func main() {
	_, _ = des.NewCipher(nil)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_BannedMD5(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad_md5.go")
	content := `package main

import "crypto/md5"

func main() {
	_ = md5.New()
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_BannedRC4(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad_rc4.go")
	content := `package main

import "crypto/rc4"

func main() {
	_, _ = rc4.NewCipher(nil)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_BannedScrypt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "bad_scrypt.go")
	content := `package main

import "golang.org/x/crypto/scrypt"

func main() {
	_, _ = scrypt.Key(nil, nil, 32768, 8, 1, 32)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheck_TestFileExcluded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "crypto_test.go")
	content := `package main

import "math/rand"

func TestSomething(t *testing.T) {
	t.Parallel()
	_ = rand.Intn(10)
}
`

	err := os.WriteFile(testFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {testFile}})

	require.NoError(t, err, "test files should be excluded from banned import checks")
}

func TestCheck_MultipleBannedImports(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "multi_bad.go")
	content := `package main

import (
	"crypto/des"
	"crypto/md5"
	"math/rand"
)

func main() {
	_, _ = des.NewCipher(nil)
	_ = md5.New()
	_ = rand.Intn(10)
}
`

	err := os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger, map[string][]string{"go": {goFile}})

	require.Error(t, err)
	require.Contains(t, err.Error(), "banned-imports")
}

func TestCheckFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := checkFile("/nonexistent/path/file.go")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open file")
}
