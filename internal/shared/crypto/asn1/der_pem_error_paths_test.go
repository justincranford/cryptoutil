// Copyright (c) 2025-2026 Justin Cranford.
package asn1

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"testing"

	crand "crypto/rand"
	rsa "crypto/rsa"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

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
