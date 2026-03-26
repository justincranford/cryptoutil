// Copyright (c) 2025 Justin Cranford

package certificate

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadProfile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadProfile("/nonexistent/path/profile.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to read certificate profile")
}

func TestLoadProfile_Success(t *testing.T) {
	t.Parallel()

	content := `name: test-profile
type: tls-server
validity:
  duration: 8760h
key_usage:
  digital_signature: true
`

	dir := t.TempDir()
	path := filepath.Join(dir, "profile.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	profile, err := LoadProfile(path)
	require.NoError(t, err)
	require.Equal(t, "test-profile", profile.Name)
}

func TestLoadProfile_InvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: valid: yaml: :\n  - broken"), cryptoutilSharedMagic.CacheFilePermissions))

	_, err := LoadProfile(path)
	require.Error(t, err)
}

func TestGetDuration_Empty(t *testing.T) {
	t.Parallel()

	v := &ValidityConfig{Duration: ""}
	_, err := v.GetDuration()
	require.Error(t, err)
	require.ErrorContains(t, err, "duration not specified")
}

func TestGetDuration_Invalid(t *testing.T) {
	t.Parallel()

	v := &ValidityConfig{Duration: "not-a-duration"}
	_, err := v.GetDuration()
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid duration")
}

func TestGetMaxDuration_Empty(t *testing.T) {
	t.Parallel()

	v := &ValidityConfig{MaxDuration: ""}
	_, err := v.GetMaxDuration()
	require.Error(t, err)
	require.ErrorContains(t, err, "max duration not specified")
}

func TestGetMaxDuration_Invalid(t *testing.T) {
	t.Parallel()

	v := &ValidityConfig{MaxDuration: "not-a-duration"}
	_, err := v.GetMaxDuration()
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid max duration")
}

func TestGetBackdateBuffer_Invalid(t *testing.T) {
	t.Parallel()

	v := &ValidityConfig{BackdateBuffer: "not-a-duration"}
	_, err := v.GetBackdateBuffer()
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid backdate buffer")
}

func TestValidateDuration_GetDurationError(t *testing.T) {
	t.Parallel()

	// AllowCustom=false with empty Duration triggers GetDuration error
	v := &ValidityConfig{AllowCustom: false, Duration: ""}
	err := v.ValidateDuration(time.Hour)
	require.Error(t, err)
	require.ErrorContains(t, err, "duration not specified")
}
