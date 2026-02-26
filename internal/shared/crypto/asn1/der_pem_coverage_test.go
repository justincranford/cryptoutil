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
	require.NoError(t, os.WriteFile(blockingFile, []byte("block"), 0o600))

	key, err := rsa.GenerateKey(crand.Reader, 2048)
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
	require.NoError(t, os.WriteFile(blockingFile, []byte("block"), 0o600))

	key, err := rsa.GenerateKey(crand.Reader, 2048)
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
	require.NoError(t, os.MkdirAll(roDir, 0o755))
	require.NoError(t, os.Chmod(roDir, 0o444))
	t.Cleanup(func() { os.Chmod(roDir, 0o755) }) //nolint:errcheck,gosec // cleanup: chmod restore on cleanup, error ignored intentionally

	key, err := rsa.GenerateKey(crand.Reader, 2048)
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
	require.NoError(t, os.MkdirAll(roDir, 0o755))
	require.NoError(t, os.Chmod(roDir, 0o444))
	t.Cleanup(func() { os.Chmod(roDir, 0o755) }) //nolint:errcheck,gosec // cleanup: chmod restore on cleanup, error ignored intentionally

	key, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	err = DERWrite(key, filepath.Join(roDir, "key.der"))
	require.Error(t, err)
	require.ErrorContains(t, err, "write failed")
}

// TestDEREncode_PrivateKeyMarshalError covers DEREncode's MarshalPKCS8PrivateKey error path.
// NOTE: Must NOT use t.Parallel() - modifies package-level x509MarshalPKCS8PrivateKeyFn.
func TestDEREncode_PrivateKeyMarshalError(t *testing.T) {
	orig := x509MarshalPKCS8PrivateKeyFn
	x509MarshalPKCS8PrivateKeyFn = func(_ any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	}

	defer func() { x509MarshalPKCS8PrivateKeyFn = orig }()

	key, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	_, _, err = DEREncode(key)
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}

// TestDEREncode_PublicKeyMarshalError covers DEREncode's MarshalPKIXPublicKey error path.
// NOTE: Must NOT use t.Parallel() - modifies package-level x509MarshalPKIXPublicKeyFn.
func TestDEREncode_PublicKeyMarshalError(t *testing.T) {
	orig := x509MarshalPKIXPublicKeyFn
	x509MarshalPKIXPublicKeyFn = func(_ any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	}

	defer func() { x509MarshalPKIXPublicKeyFn = orig }()

	key, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	_, _, err = DEREncode(&key.PublicKey)
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}

// TestDERDecodes_AllDecodersFail covers DERDecodes' "decode failed" path when types list is empty.
// NOTE: Must NOT use t.Parallel() - modifies package-level derDecodesPEMTypes.
func TestDERDecodes_AllDecodersFail(t *testing.T) {
	orig := derDecodesPEMTypes
	derDecodesPEMTypes = []string{}

	defer func() { derDecodesPEMTypes = orig }()

	_, _, err := DERDecodes([]byte("any bytes"))
	require.Error(t, err)
	require.ErrorContains(t, err, "decode failed")
}

// TestDERRead_DecodesAllFail covers DERRead's error path when DERDecodes fails.
// NOTE: Must NOT use t.Parallel() - modifies package-level derDecodesPEMTypes.
func TestDERRead_DecodesAllFail(t *testing.T) {
	orig := derDecodesPEMTypes
	derDecodesPEMTypes = []string{}

	defer func() { derDecodesPEMTypes = orig }()

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.der")
	require.NoError(t, os.WriteFile(filename, []byte("any bytes"), 0o600))

	_, _, err := DERRead(filename)
	require.Error(t, err)
	require.ErrorContains(t, err, "decode failed")
}

func TestPEMEncodes_EncodeError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := pemEncodeInternalFn

	defer func() { pemEncodeInternalFn = originalFn }()

	pemEncodeInternalFn = func(_ any) ([]byte, error) {
		return nil, errors.New("injected PEM encode failure")
	}

	// Create a minimal certificate slice to trigger the loop.
	certs := []*x509.Certificate{{Raw: []byte("test")}}
	_, err := PEMEncodes(certs)
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}

func TestDEREncodes_EncodeError(t *testing.T) {
	// Cannot be parallel: modifies package-level injectable var.
	originalFn := derEncodeInternalFn

	defer func() { derEncodeInternalFn = originalFn }()

	derEncodeInternalFn = func(_ any) ([]byte, string, error) {
		return nil, "", errors.New("injected DER encode failure")
	}

	// Create a minimal certificate slice to trigger the loop.
	certs := []*x509.Certificate{{Raw: []byte("test")}}
	_, err := DEREncodes(certs)
	require.Error(t, err)
	require.ErrorContains(t, err, "encode failed")
}
