// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGenerateKeyPairFromCSRAndEncodePrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm x509.PublicKeyAlgorithm
		wantError string
	}{
		{name: cryptoutilSharedMagic.KeyTypeRSA, algorithm: x509.RSA},
		{name: cryptoutilSharedMagic.ECDSA, algorithm: x509.ECDSA},
		{name: cryptoutilSharedMagic.EdCurveEd25519, algorithm: x509.Ed25519},
		{name: "unsupported algorithm", algorithm: x509.UnknownPublicKeyAlgorithm, wantError: "unsupported public key algorithm"},
	}

	handler := &Handler{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			privateKey, publicKey, err := handler.generateKeyPairFromCSR(&x509.CertificateRequest{
				PublicKeyAlgorithm: tc.algorithm,
			})
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				require.Nil(t, privateKey)
				require.Nil(t, publicKey)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, privateKey)
			require.NotNil(t, publicKey)

			encodedKey, err := handler.encodePrivateKeyPEM(privateKey)
			require.NoError(t, err)

			block, _ := pem.Decode(encodedKey)
			require.NotNil(t, block)
			require.NotEmpty(t, block.Bytes)
		})
	}
}

func TestCreateCSRWithKey_SuccessAndUnsupportedKey(t *testing.T) {
	t.Parallel()

	handler := &Handler{}
	templateCSR, err := handler.parseCSR(createValidCSRPEM(t))
	require.NoError(t, err)

	tests := []struct {
		name          string
		privateKey    any
		wantErrorText string
	}{
		{
			name: "valid ECDSA key",
			privateKey: func() any {
				key, keyErr := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
				require.NoError(t, keyErr)

				return key
			}(),
		},
		{
			name:          "unsupported key",
			privateKey:    struct{}{},
			wantErrorText: "failed to create CSR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			csr, err := handler.createCSRWithKey(templateCSR, tc.privateKey)
			if tc.wantErrorText != "" {
				require.ErrorContains(t, err, tc.wantErrorText)

				return
			}

			require.NoError(t, err)
			require.Equal(t, templateCSR.Subject, csr.Subject)
			require.NoError(t, csr.CheckSignature())
		})
	}
}

func TestEncodePrivateKeyPEM_UnsupportedType(t *testing.T) {
	t.Parallel()

	_, err := (&Handler{}).encodePrivateKeyPEM(struct{}{})
	require.ErrorContains(t, err, "unsupported private key type")

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)
	require.NotNil(t, publicKey)

	encodedKey, err := (&Handler{}).encodePrivateKeyPEM(privateKey)
	require.NoError(t, err)

	block, _ := pem.Decode(encodedKey)
	require.NotNil(t, block)
}

func TestParseCSR_InvalidSignature(t *testing.T) {
	t.Parallel()

	csrDER := createValidCSRDER(t)
	csrDER[len(csrDER)-1] ^= 0xff
	csrPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
		Bytes: csrDER,
	}))

	_, err := (&Handler{}).parseCSR(csrPEM)
	require.ErrorContains(t, err, "CSR signature verification failed")
}
