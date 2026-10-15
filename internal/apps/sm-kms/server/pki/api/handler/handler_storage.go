// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"fmt"
	"strings"

	googleUuid "github.com/google/uuid"

	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
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

	issuerAuthorityID := strings.TrimSpace(h.issuer.GetCAConfig().AuthorityID)
	if issuerAuthorityID == "" {
		return nil, fmt.Errorf("issuing authority identity is not configured")
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
