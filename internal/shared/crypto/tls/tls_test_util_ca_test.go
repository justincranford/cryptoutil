// Copyright (c) 2025-2026 Justin Cranford.
package tls

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	stdTLS "crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	http "net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewClientForTestWithCA(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	certPEM := newTestCAPEM(t)
	caPath := filepath.Join(dir, "ca.crt")
	require.NoError(t, os.WriteFile(caPath, certPEM, 0o600))

	client := NewClientForTestWithCA(caPath)
	require.NotNil(t, client)
	require.Equal(t, cryptoutilSharedMagic.IMDefaultTimeout, client.Timeout)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.True(t, transport.DisableKeepAlives)
	require.Equal(t, uint16(stdTLS.VersionTLS13), transport.TLSClientConfig.MinVersion)
	require.NotNil(t, transport.TLSClientConfig.RootCAs)

	tests := []struct {
		name string
		path string
	}{
		{name: "missing CA file", path: filepath.Join(dir, "missing.crt")},
		{name: "invalid CA PEM", path: filepath.Join(dir, "invalid.crt")},
	}
	require.NoError(t, os.WriteFile(tests[1].path, []byte("not a certificate"), 0o600))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Panics(t, func() { NewClientForTestWithCA(tc.path) })
		})
	}
}

func newTestCAPEM(t *testing.T) []byte {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	serialNumber, err := crand.Int(crand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}

	certificateDER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: certificateDER})

	return certificatePEM
}
