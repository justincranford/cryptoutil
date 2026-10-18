// Copyright (c) 2025-2026 Justin Cranford.
package asn1

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	pemWriteDirectoryFailure = "mkdir failed"
	pemWriteFileFailure      = "write failed"
)

func TestPEMWriteAndDERWrite_SecretKey(t *testing.T) {
	t.Parallel()

	key := []byte("test secret key bytes")
	baseDir := t.TempDir()

	pemPath := filepath.Join(baseDir, "nested", "key.pem")
	require.NoError(t, PEMWrite(key, pemPath))

	pemKey, err := PEMRead(pemPath)
	require.NoError(t, err)
	require.Equal(t, key, pemKey)

	derPath := filepath.Join(baseDir, "nested", "key.der")
	require.NoError(t, DERWrite(key, derPath))

	derKey, derType, err := DERRead(derPath)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypeSecretKey, derType)
	require.Equal(t, key, derKey)

	_, statErr := os.Stat(pemPath)
	require.NoError(t, statErr)
}

func TestPEMWriteAndDERWrite_FilesystemErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		write   func(string) error
		path    func(*testing.T) string
		wantErr string
	}{
		{
			name: "PEM create directory",
			write: func(filename string) error {
				return PEMWrite([]byte("secret"), filename)
			},
			path: func(t *testing.T) string {
				t.Helper()

				parent := filepath.Join(t.TempDir(), "file-parent")
				require.NoError(t, os.WriteFile(parent, []byte("not a directory"), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly))

				return filepath.Join(parent, "key.pem")
			},
			wantErr: pemWriteDirectoryFailure,
		},
		{
			name: "PEM write file",
			write: func(filename string) error {
				return PEMWrite([]byte("secret"), filename)
			},
			path: func(t *testing.T) string {
				t.Helper()

				dir := filepath.Join(t.TempDir(), "directory")
				require.NoError(t, os.Mkdir(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

				return dir
			},
			wantErr: pemWriteFileFailure,
		},
		{
			name: "DER create directory",
			write: func(filename string) error {
				return DERWrite([]byte("secret"), filename)
			},
			path: func(t *testing.T) string {
				t.Helper()

				parent := filepath.Join(t.TempDir(), "file-parent")
				require.NoError(t, os.WriteFile(parent, []byte("not a directory"), cryptoutilSharedMagic.FilePermOwnerReadWriteOnly))

				return filepath.Join(parent, "key.der")
			},
			wantErr: pemWriteDirectoryFailure,
		},
		{
			name: "DER write file",
			write: func(filename string) error {
				return DERWrite([]byte("secret"), filename)
			},
			path: func(t *testing.T) string {
				t.Helper()

				dir := filepath.Join(t.TempDir(), "directory")
				require.NoError(t, os.Mkdir(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

				return dir
			},
			wantErr: pemWriteFileFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.ErrorContains(t, tc.write(tc.path(t)), tc.wantErr)
		})
	}
}

func TestPEMEncodeCertChain_ConcatenatesCertificates(t *testing.T) {
	t.Parallel()

	certificates := []*x509.Certificate{
		{Raw: []byte("leaf certificate")},
		{Raw: []byte("root certificate")},
	}

	encoded, err := PEMEncodeCertChain(certificates)
	require.NoError(t, err)

	first, rest := pem.Decode(encoded)
	require.NotNil(t, first)
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypeCertificate, first.Type)
	require.Equal(t, certificates[0].Raw, first.Bytes)

	second, rest := pem.Decode(rest)
	require.NotNil(t, second)
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypeCertificate, second.Type)
	require.Equal(t, certificates[1].Raw, second.Bytes)
	require.Empty(t, rest)
}

func TestPEMEncodeCertChain_NilCertificate(t *testing.T) {
	t.Parallel()

	_, err := PEMEncodeCertChain([]*x509.Certificate{nil})
	require.ErrorContains(t, err, "certificate 0: certificate is nil")
}

func TestPEMEncodeCertChain_EncodeError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("injected certificate encoding error")
	certificates := []*x509.Certificate{{Raw: []byte("leaf certificate")}}

	_, err := pemEncodeCertChainWithFn(certificates, func(any) ([]byte, error) {
		return nil, wantErr
	})
	require.ErrorIs(t, err, wantErr)
	require.ErrorContains(t, err, "failed to PEM-encode certificate 0")
}
