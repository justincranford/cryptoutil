// Copyright (c) 2025 Justin Cranford

package config_rules

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Synthetic tests for error paths ---

func TestCheckKeyNaming_NonKebabCase(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		"# sm-kms Common Configuration\nbad_key: value\n")

	logger := newTestLogger()

	err := checkKeyNamingInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-kebab-case key")
	assert.Contains(t, err.Error(), "bad_key")
}

func TestCheckKeyNaming_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		":\ninvalid: [yaml\n")

	logger := newTestLogger()

	err := checkKeyNamingInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCheckHeaderIdentity_WrongPSID(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		"# jose-ja Common Configuration\nbind-public-address: 0.0.0.0\n")

	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not reference PS-ID")
}

func TestCheckHeaderIdentity_NoComment(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		"bind-public-address: 0.0.0.0\n")

	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a comment")
}

func TestCheckHeaderIdentity_EmptyFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"), "")

	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestCheckInstanceMinimal_ExtraKey(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-sqlite-1.yml"),
		"# sm-kms SQLite Instance 1 Configuration\ncors-origins: []\notlp-service: test\nextra-key: bad\n")

	logger := newTestLogger()

	err := checkInstanceMinimalInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected key")
	assert.Contains(t, err.Error(), "extra-key")
}

func TestCheckInstanceMinimal_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-sqlite-1.yml"),
		":\n[invalid\n")

	logger := newTestLogger()

	err := checkInstanceMinimalInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCheckCommonComplete_MissingKeys(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		"# sm-kms Common Configuration\nbind-public-address: 0.0.0.0\n")

	logger := newTestLogger()

	err := checkCommonCompleteInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required key")
	assert.Contains(t, err.Error(), "tls-cert-file")
}

func TestCheckCommonComplete_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		":\n[invalid\n")

	logger := newTestLogger()

	err := checkCommonCompleteInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCheckKeyNaming_NestedNonKebab(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	writeFile(t, filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml"),
		"# sm-kms Common Configuration\nouter:\n  inner_bad: value\n")

	logger := newTestLogger()

	err := checkKeyNamingInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inner_bad")
}

func TestCheckInstanceMinimal_ReadError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	badPath := filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-sqlite-1.yml")

	require.NoError(t, os.MkdirAll(badPath, cryptoutilSharedMagic.CICDTempDirPermissions))

	logger := newTestLogger()

	err := checkInstanceMinimalInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}

func TestCheckCommonComplete_ReadError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	badPath := filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml")
	_ = os.Remove(badPath)

	require.NoError(t, os.MkdirAll(badPath, cryptoutilSharedMagic.CICDTempDirPermissions))

	logger := newTestLogger()

	err := checkCommonCompleteInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}

func TestCheckKeyNaming_GlobError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	badRoot := filepath.Join(tmp, "bad[dir")
	setupMinimalStructure(t, badRoot)

	logger := newTestLogger()

	err := checkKeyNamingInDir(logger, badRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "glob error")
}

func TestCheckHeaderIdentity_GlobError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	badRoot := filepath.Join(tmp, "bad[dir")
	setupMinimalStructure(t, badRoot)

	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, badRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "glob error")
}

func TestCheckInstanceMinimal_GlobError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	badRoot := filepath.Join(tmp, "bad[dir")
	setupMinimalStructure(t, badRoot)

	logger := newTestLogger()

	err := checkInstanceMinimalInDir(logger, badRoot)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "glob error")
}

func TestCheckKeyNaming_MissingFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	badPath := filepath.Join(tmp, "deployments", cryptoutilSharedMagic.OTLPServiceSMKMS,
		"config", cryptoutilSharedMagic.OTLPServiceSMKMS+"-app-common.yml")
	_ = os.Remove(badPath)

	require.NoError(t, os.MkdirAll(badPath, cryptoutilSharedMagic.CICDTempDirPermissions))

	logger := newTestLogger()

	err := checkKeyNamingInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}

func TestCheckHeaderIdentity_MissingFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	standaloneFile := filepath.Join(tmp, cryptoutilSharedMagic.CICDConfigsDir,
		cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMKMS+"-framework.yml")
	_ = os.Remove(standaloneFile)

	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}

func TestCheckHeaderIdentity_StandaloneWrongPSID(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	setupMinimalStructure(t, tmp)

	standaloneFile := filepath.Join(tmp, cryptoutilSharedMagic.CICDConfigsDir,
		cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMKMS+"-framework.yml")
	writeFile(t, standaloneFile, "# Wrong Service Configuration\n# This is not the right service.\n")

	logger := newTestLogger()

	err := checkHeaderIdentityInDir(logger, tmp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not reference PS-ID")
}
