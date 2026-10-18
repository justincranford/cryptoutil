// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAuthzAdminServer_LoadTLSConfigStaticCertificate(t *testing.T) {
	t.Parallel()

	generated, err := generateSelfSignedTLSConfig()
	require.NoError(t, err)
	require.Len(t, generated.Certificates, 1)

	privateKey, ok := generated.Certificates[0].PrivateKey.(*ecdsa.PrivateKey)
	require.True(t, ok)

	certPath := filepath.Join(t.TempDir(), "server.crt")
	keyPath := filepath.Join(t.TempDir(), "server.key")
	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: generated.Certificates[0].Certificate[0],
	})
	privateKeyDER, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeECPrivateKey,
		Bytes: privateKeyDER,
	})

	require.NoError(t, os.WriteFile(certPath, certificatePEM, cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(keyPath, privateKeyPEM, cryptoutilSharedMagic.CacheFilePermissions))

	cfg := cryptoutilIdentityConfig.RequireNewForTest("authz_admin_static_tls")
	cfg.AuthZ.TLSCertFile = certPath
	cfg.AuthZ.TLSKeyFile = keyPath
	admin := &AdminServer{config: cfg}

	tlsConfig, err := admin.loadTLSConfig(context.Background())
	require.NoError(t, err)
	require.Len(t, tlsConfig.Certificates, 1)
	require.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MinVersion)
}
