// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewStoredCertificateFromX509 creates a StoredCertificate from an x509.Certificate.
func NewStoredCertificateFromX509(cert *x509.Certificate, profileID, requesterID string) (*StoredCertificate, error) {
	if cert == nil {
		return nil, ErrInvalidCertificate
	}

	id, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	return &StoredCertificate{
		ID:             id,
		SerialNumber:   cert.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		SubjectDN:      cert.Subject.String(),
		IssuerDN:       cert.Issuer.String(),
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
		Status:         StatusActive,
		ProfileID:      profileID,
		RequesterID:    requesterID,
		CertificatePEM: string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: cert.Raw})),
		CertificateDER: cert.Raw,
		SubjectKeyID:   fmt.Sprintf("%x", cert.SubjectKeyId),
		AuthorityKeyID: fmt.Sprintf("%x", cert.AuthorityKeyId),
		Metadata:       make(map[string]string),
	}, nil
}
