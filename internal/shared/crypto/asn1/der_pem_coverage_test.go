// Copyright (c) 2025 Justin Cranford

package asn1

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	crand "crypto/rand"
	rsa "crypto/rsa"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestPEMWrite_MkdirFailure covers PEMWrite's mkdir error path.
func TestPEMWrite_MkdirFailure(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("file-as-dir trick not reliable on Windows")
	}

	tmpDir := t.TempDir()
	// Create a regular FILE where PEMWrite would need to create a DIRECTORY.
	blockingFile := filepath.Join(tmpDir, "notadir")
	require.NoError(t, os.WriteFile(blockingFile, []byte("block"), cryptoutilSharedMagic.CacheFilePermissions))

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	// PEMWrite tries to MkdirAll(notadir), but notadir is a file -> fails.
	err = PEMWrite(key, filepath.Join(blockingFile, "key.pem"))
	require.Error(t, err)
	require.ErrorContains(t, err, "mkdir failed")
}

// TestDERWrite_MkdirFailure covers DERWrite's mkdir error path.
func TestDERWrite_MkdirFailure(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("file-as-dir trick not reliable on Windows")
	}

	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "notadir")
	require.NoError(t, os.WriteFile(blockingFile, []byte("block"), cryptoutilSharedMagic.CacheFilePermissions))

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	err = DERWrite(key, filepath.Join(blockingFile, "key.der"))
	require.Error(t, err)
	require.ErrorContains(t, err, "mkdir failed")
}

// TestPEMWrite_WriteFileFailure covers PEMWrite's write error path via read-only dir.
func TestPEMWrite_WriteFileFailure(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("read-only dirs behave differently on Windows")
	}

	tmpDir := t.TempDir()
	roDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.MkdirAll(roDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Chmod(roDir, cryptoutilSharedMagic.FilePermOwnerReadOnlyGroupOtherReadOnly))
	t.Cleanup(func() { os.Chmod(roDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute) }) //nolint:errcheck,gosec // cleanup: chmod restore on cleanup, error ignored intentionally

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	err = PEMWrite(key, filepath.Join(roDir, "key.pem"))
	require.Error(t, err)
	require.ErrorContains(t, err, "write failed")
}

// TestDERWrite_WriteFileFailure covers DERWrite's write error path via read-only dir.
func TestDERWrite_WriteFileFailure(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("read-only dirs behave differently on Windows")
	}

	tmpDir := t.TempDir()
	roDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.MkdirAll(roDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Chmod(roDir, cryptoutilSharedMagic.FilePermOwnerReadOnlyGroupOtherReadOnly))
	t.Cleanup(func() { os.Chmod(roDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute) }) //nolint:errcheck,gosec // cleanup: chmod restore on cleanup, error ignored intentionally

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	err = DERWrite(key, filepath.Join(roDir, "key.der"))
	require.Error(t, err)
	require.ErrorContains(t, err, "write failed")
}

// TestDEREncode_PrivateKeyMarshalError covers DEREncode's MarshalPKCS8PrivateKey error path.
func TestDEREncode_PrivateKeyMarshalError(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	_, _, err = derEncodeWithFns(key, func(_ any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	}, x509.MarshalPKIXPublicKey)
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}

// TestDEREncode_PublicKeyMarshalError covers DEREncode's MarshalPKIXPublicKey error path.
func TestDEREncode_PublicKeyMarshalError(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	_, _, err = derEncodeWithFns(&key.PublicKey, x509.MarshalPKCS8PrivateKey, func(_ any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}

// TestDERDecodes_AllDecodersFail covers DERDecodes' "decode failed" path when types list is empty.
func TestDERDecodes_AllDecodersFail(t *testing.T) {
	t.Parallel()

	_, _, err := derDecodesWithTypes([]byte("any bytes"), []string{})
	require.Error(t, err)
	require.ErrorContains(t, err, "decode failed")
}

// TestDERRead_DecodesAllFail covers DERRead's error path when DERDecodes fails.
func TestDERRead_DecodesAllFail(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.der")
	require.NoError(t, os.WriteFile(filename, []byte("any bytes"), cryptoutilSharedMagic.CacheFilePermissions))

	_, _, err := derReadWithTypes(filename, []string{})
	require.Error(t, err)
	require.ErrorContains(t, err, "decode failed")
}

// TestPEMEncodes_EncodeError covers PEMEncodes error path.
func TestPEMEncodes_EncodeError(t *testing.T) {
	t.Parallel()

	// Create a minimal certificate slice to trigger the loop.
	certs := []*x509.Certificate{{Raw: []byte("test")}}
	_, err := pemEncodesWithFn(certs, func(_ any) ([]byte, error) {
		return nil, errors.New("injected PEM encode failure")
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}

// TestDEREncodes_EncodeError covers DEREncodes error path.
func TestDEREncodes_EncodeError(t *testing.T) {
	t.Parallel()

	// Create a minimal certificate slice to trigger the loop.
	certs := []*x509.Certificate{{Raw: []byte("test")}}
	_, err := derEncodesWithFn(certs, func(_ any) ([]byte, string, error) {
		return nil, "", errors.New("injected DER encode failure")
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}
