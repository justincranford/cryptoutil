// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	googleUuid "github.com/google/uuid"

	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func (h *Handler) storeIssuedCertificate(
	ctx context.Context,
	issued *cryptoutilCAServiceIssuer.IssuedCertificate,
	profileID string,
	requestID googleUuid.UUID,
) (*cryptoutilCAStorage.StoredCertificate, error) {
	if issued == nil || issued.Certificate == nil {
		return nil, cryptoutilCAStorage.ErrInvalidCertificate
	}

	requesterID := ""
	if tenantContext := cryptoutilSharedDatabase.GetTenantContext(ctx); tenantContext != nil && tenantContext.UserID != googleUuid.Nil {
		requesterID = tenantContext.UserID.String()
	}

	certificate, err := cryptoutilCAStorage.NewStoredCertificateFromX509(issued.Certificate, profileID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare issued certificate for storage: %w", err)
	}

	if h.issuer == nil || h.issuer.GetCAConfig() == nil {
		return nil, fmt.Errorf("issuing authority identity is not configured")
	}

	caConfig := h.issuer.GetCAConfig()

	issuerAuthorityID := strings.TrimSpace(caConfig.AuthorityID)
	if issuerAuthorityID == "" || issuerAuthorityID != caConfig.AuthorityID {
		return nil, fmt.Errorf("issuing authority identity is not configured")
	}

	if caConfig.Certificate == nil {
		return nil, fmt.Errorf("issuing authority certificate is not configured")
	}

	if !bytes.Equal(issued.Certificate.RawIssuer, caConfig.Certificate.RawSubject) {
		return nil, fmt.Errorf("issued certificate issuer does not match configured authority")
	}

	if len(caConfig.Certificate.SubjectKeyId) == 0 ||
		!bytes.Equal(issued.Certificate.AuthorityKeyId, caConfig.Certificate.SubjectKeyId) {
		return nil, fmt.Errorf("issued certificate authority key identifier does not match configured authority")
	}

	if err := issued.Certificate.CheckSignatureFrom(caConfig.Certificate); err != nil {
		return nil, fmt.Errorf("issued certificate signature does not match configured authority: %w", err)
	}

	if issued.SerialNumber != issued.Certificate.SerialNumber.Text(cryptoutilSharedMagic.HexBase) {
		return nil, fmt.Errorf("issued certificate serial number does not match the certificate")
	}

	certificate.SerialNumber = issued.SerialNumber
	certificate.IssuerAuthorityID = issuerAuthorityID
	certificate.CertificatePEM = string(issued.CertificatePEM)
	certificate.PublicKeyHash = issued.Fingerprint
	certificate.RequestID = requestID

	if err := h.storage.Store(ctx, certificate); err != nil {
		return nil, fmt.Errorf("failed to persist issued certificate: %w", err)
	}

	return certificate, nil
}

func validateStoredCertificateIssuer(
	certificate *cryptoutilCAStorage.StoredCertificate,
	issuer *cryptoutilCAServiceIssuer.Issuer,
) error {
	if certificate == nil {
		return fmt.Errorf("stored certificate is required")
	}

	if issuer == nil || issuer.GetCAConfig() == nil {
		return fmt.Errorf("issuing authority identity is not configured")
	}

	caConfig := issuer.GetCAConfig()

	authorityID := strings.TrimSpace(caConfig.AuthorityID)
	if authorityID == "" || authorityID != caConfig.AuthorityID {
		return fmt.Errorf("stored certificate authority ID does not match configured issuer")
	}

	legacyAuthorityID := certificate.IssuerAuthorityID == ""
	if !legacyAuthorityID && certificate.IssuerAuthorityID != authorityID {
		return fmt.Errorf("stored certificate authority ID does not match configured issuer")
	}

	if caConfig.Certificate == nil || len(caConfig.Certificate.SubjectKeyId) == 0 {
		return fmt.Errorf("issuing authority certificate identity is not configured")
	}

	if certificate.IssuerDN != caConfig.Certificate.Subject.String() {
		return fmt.Errorf("stored certificate issuer DN does not match configured issuer")
	}

	expectedAuthorityKeyID := fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId)
	if !strings.EqualFold(certificate.AuthorityKeyID, expectedAuthorityKeyID) {
		return fmt.Errorf("stored certificate authority key ID does not match configured issuer")
	}

	if legacyAuthorityID {
		// Pre-migration records lack the stable authority ID; verify their certificate against this issuer.
		if err := validateLegacyStoredCertificateIssuer(certificate, caConfig.Certificate); err != nil {
			return fmt.Errorf("legacy stored certificate cannot be verified against configured issuer: %w", err)
		}
	}

	return nil
}

func validateLegacyStoredCertificateIssuer(
	certificate *cryptoutilCAStorage.StoredCertificate,
	issuerCertificate *x509.Certificate,
) error {
	block, _ := pem.Decode([]byte(certificate.CertificatePEM))
	if block == nil || block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate {
		return fmt.Errorf("stored certificate PEM is invalid")
	}

	parsedCertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse stored certificate: %w", err)
	}

	if parsedCertificate.SerialNumber.Text(cryptoutilSharedMagic.HexBase) != certificate.SerialNumber {
		return fmt.Errorf("stored certificate serial number does not match its PEM")
	}

	if !bytes.Equal(parsedCertificate.RawIssuer, issuerCertificate.RawSubject) {
		return fmt.Errorf("stored certificate issuer does not match configured issuer")
	}

	if len(parsedCertificate.AuthorityKeyId) == 0 ||
		!bytes.Equal(parsedCertificate.AuthorityKeyId, issuerCertificate.SubjectKeyId) {
		return fmt.Errorf("stored certificate authority key ID does not match configured issuer")
	}

	if err := parsedCertificate.CheckSignatureFrom(issuerCertificate); err != nil {
		return fmt.Errorf("stored certificate signature does not match configured issuer: %w", err)
	}

	return nil
}
