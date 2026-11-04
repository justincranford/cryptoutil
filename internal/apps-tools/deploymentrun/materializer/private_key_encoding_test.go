// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	ecdsa "crypto/ecdsa"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestSupportedPrivateKeyPEMEncodings(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	tests := make([]struct {
		name string
		pem  []byte
	}, 0, testPrivateKeyEncodingCaseCount)
	tests = append(tests, struct {
		name string
		pem  []byte
	}{name: "PKCS8 ECDSA", pem: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]})
	privateKey, err := parsePrivateKeyPEM(fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey])
	require.NoError(t, err)

	ecKey, ok := privateKey.(*ecdsa.PrivateKey)
	require.True(t, ok)

	ecDER, err := x509.MarshalECPrivateKey(ecKey)
	require.NoError(t, err)

	tests = append(tests, struct {
		name string
		pem  []byte
	}{name: "SEC1 ECDSA", pem: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeECPrivateKey, Bytes: ecDER})})

	rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	tests = append(tests, struct {
		name string
		pem  []byte
	}{name: "PKCS1 RSA", pem: pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey, Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)})})

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parsed, parseErr := parsePrivateKeyPEM(testCase.pem)
			require.NoError(t, parseErr)
			require.NotNil(t, parsed)
		})
	}
}
