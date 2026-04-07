// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// encodeCertChainPEM encodes a certificate chain (leaf first, root last) as concatenated PEM blocks.
func encodeCertChainPEM(subject *cryptoutilSharedCryptoCertificate.Subject) []byte {
	var out []byte

	for _, cert := range subject.KeyMaterial.CertificateChain {
		block := &pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: cert.Raw}
		out = append(out, pem.EncodeToMemory(block)...)
	}

	return out
}

// encodePrivateKeyPEM encodes a private key as a PKCS#8 PEM block.
func encodePrivateKeyPEM(key any) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PKCS8 private key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: der}), nil
}
