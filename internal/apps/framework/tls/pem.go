// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"fmt"

	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
)

// encodeCertChainPEM encodes a certificate chain (leaf first, root last) as concatenated PEM blocks
// using the shared asn1.PEMEncode utility.
func encodeCertChainPEM(subject *cryptoutilSharedCryptoCertificate.Subject) ([]byte, error) {
	var out []byte

	for i, cert := range subject.KeyMaterial.CertificateChain {
		pemBytes, err := cryptoutilSharedCryptoAsn1.PEMEncode(cert)
		if err != nil {
			return nil, fmt.Errorf("failed to PEM-encode certificate %d: %w", i, err)
		}

		out = append(out, pemBytes...)
	}

	return out, nil
}
