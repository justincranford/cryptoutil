// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// StrictServer adapts PKI operations to the generated typed OpenAPI contract.
type StrictServer struct {
	handler *Handler
}

func NewStrictServer(handler *Handler) *StrictServer {
	return &StrictServer{handler: handler}
}

var _ cryptoutilApiCaServer.StrictServerInterface = (*StrictServer)(nil)

func (s *StrictServer) ListCAs(ctx context.Context, _ cryptoutilApiCaServer.ListCAsRequestObject) (cryptoutilApiCaServer.ListCAsResponseObject, error) {
	h := s.handler
	if h.issuer == nil {
		return strictError(fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured"), nil
	}

	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil {
		return strictError(fiber.StatusInternalServerError, "ca_config_error", "CA configuration not available"), nil
	}

	caCert := caConfig.Certificate

	caType := cryptoutilApiCaServer.CASummaryTypeIntermediate
	if caCert.Issuer.String() == caCert.Subject.String() {
		caType = cryptoutilApiCaServer.CASummaryTypeRoot
	}

	validUntil := caCert.NotAfter
	response := cryptoutilApiCaServer.CAListResponse{
		Authorities: []cryptoutilApiCaServer.CASummary{{
			ID:         caConfig.Name,
			Name:       caConfig.Name,
			Type:       caType,
			Status:     cryptoutilApiCaServer.CASummaryStatusActive,
			SubjectCN:  &caCert.Subject.CommonName,
			ValidUntil: &validUntil,
		}},
	}

	return strictJSON(fiber.StatusOK, response), nil
}

func (s *StrictServer) GetCA(_ context.Context, request cryptoutilApiCaServer.GetCARequestObject) (cryptoutilApiCaServer.GetCAResponseObject, error) {
	h := s.handler
	if h.issuer == nil {
		return strictError(fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured"), nil
	}

	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil {
		return strictError(fiber.StatusInternalServerError, "ca_config_error", "CA configuration not available"), nil
	}

	if request.CAID != caConfig.Name {
		return strictError(fiber.StatusNotFound, "not_found", "CA not found"), nil
	}

	caCert := caConfig.Certificate

	caType := cryptoutilApiCaServer.CAResponseTypeIntermediate
	if caCert.Issuer.String() == caCert.Subject.String() {
		caType = cryptoutilApiCaServer.CAResponseTypeRoot
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: caCert.Raw})
	notBefore, notAfter := caCert.NotBefore, caCert.NotAfter
	serialNumber := fmt.Sprintf("%X", caCert.SerialNumber)
	keyAlgo, keySize := getKeyInfo(caCert)

	response := cryptoutilApiCaServer.CAResponse{
		ID:                    caConfig.Name,
		Name:                  caConfig.Name,
		Type:                  caType,
		Status:                cryptoutilApiCaServer.CAResponseStatusActive,
		Subject:               buildCertificateSubject(caCert.Subject.String()),
		Issuer:                buildCertificateSubject(caCert.Issuer.String()),
		SerialNumber:          ptrString(serialNumber),
		NotBefore:             &notBefore,
		NotAfter:              &notAfter,
		CertificatePEM:        string(certPEM),
		KeyAlgorithm:          &keyAlgo,
		KeySize:               &keySize,
		SignatureAlgorithm:    ptrString(caCert.SignatureAlgorithm.String()),
		CRLDistributionPoints: ptrStringSlice(caCert.CRLDistributionPoints),
		OCSPUrls:              ptrStringSlice(caCert.OCSPServer),
		IssuingUrls:           ptrStringSlice(caCert.IssuingCertificateURL),
	}
	if caCert.BasicConstraintsValid && caCert.IsCA {
		pathLen := caCert.MaxPathLen
		response.PathLength = &pathLen
	}

	return strictJSON(fiber.StatusOK, response), nil
}

func (s *StrictServer) GetCRL(ctx context.Context, request cryptoutilApiCaServer.GetCRLRequestObject) (cryptoutilApiCaServer.GetCRLResponseObject, error) {
	h := s.handler
	h.mu.RLock()
	crlService := h.crlService
	h.mu.RUnlock()

	if crlService == nil {
		return strictError(fiber.StatusServiceUnavailable, "service_unavailable", "CRL service not configured"), nil
	}

	if h.issuer == nil {
		return strictError(fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured"), nil
	}

	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil || caConfig.Name != request.CAID {
		return strictError(fiber.StatusNotFound, "not_found", "CA not found"), nil
	}

	if len(caConfig.Certificate.SubjectKeyId) == 0 {
		return strictError(fiber.StatusInternalServerError, "ca_config_error", "CA subject key identifier is not available"), nil
	}

	issuerKeyID := fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId)

	storedRevocations, err := h.storage.GetRevokedByIssuerKeyID(ctx, issuerKeyID)
	if err != nil {
		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	revokedCertificates := make([]cryptoutilCAServiceRevocation.RevokedCertificate, 0, len(storedRevocations))
	for _, storedRevocation := range storedRevocations {
		if storedRevocation.RevocationTime == nil || storedRevocation.RevocationReason == nil {
			return strictError(fiber.StatusInternalServerError, "storage_error", "revoked certificate is missing revocation metadata"), nil
		}

		serialNumber, ok := new(big.Int).SetString(storedRevocation.SerialNumber, cryptoutilSharedMagic.HexBase)
		if !ok {
			return strictError(fiber.StatusInternalServerError, "storage_error", "revoked certificate has an invalid serial number"), nil
		}

		revokedCertificates = append(revokedCertificates, cryptoutilCAServiceRevocation.RevokedCertificate{
			SerialNumber:   serialNumber,
			RevocationTime: *storedRevocation.RevocationTime,
			Reason:         cryptoutilCAServiceRevocation.RevocationReason(*storedRevocation.RevocationReason),
		})
	}

	format := cryptoutilSharedMagic.PKICRLFormatDER
	if request.Params.Format != nil {
		format = string(*request.Params.Format)
	}

	switch format {
	case cryptoutilSharedMagic.PKICRLFormatPEM:
		body, err := crlService.GenerateCRLPEMWithRevokedCertificatesContext(ctx, revokedCertificates)
		if err != nil {
			return strictError(fiber.StatusInternalServerError, "crl_error", err.Error()), nil
		}

		return strictBinary(fiber.StatusOK, cryptoutilSharedMagic.PKIContentTypeCRLPEM, body, [2]string{cryptoutilSharedMagic.PKIHeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s.crl.pem\"", request.CAID)}), nil
	default:
		body, err := crlService.GenerateCRLWithRevokedCertificatesContext(ctx, revokedCertificates)
		if err != nil {
			return strictError(fiber.StatusInternalServerError, "crl_error", err.Error()), nil
		}

		return strictBinary(fiber.StatusOK, cryptoutilSharedMagic.PKIContentTypeCRLDER, body, [2]string{cryptoutilSharedMagic.PKIHeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s.crl\"", request.CAID)}), nil
	}
}

func (s *StrictServer) ListCertificates(ctx context.Context, request cryptoutilApiCaServer.ListCertificatesRequestObject) (cryptoutilApiCaServer.ListCertificatesResponseObject, error) {
	filter := &cryptoutilCAStorage.ListFilter{Limit: cryptoutilSharedMagic.DefaultPageLimit}

	params := request.Params
	if params.PageSize != nil {
		filter.Limit = *params.PageSize
	}

	if params.Page != nil && *params.Page > 1 {
		filter.Offset = (*params.Page - 1) * filter.Limit
	}

	if params.Profile != nil {
		filter.ProfileID = params.Profile
	}

	if params.Status != nil {
		status := cryptoutilCAStorage.CertificateStatus(*params.Status)
		filter.Status = &status
	}

	certs, total, err := s.handler.storage.List(ctx, filter)
	if err != nil {
		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	certResponses := make([]cryptoutilApiCaServer.CertificateSummary, 0, len(certs))
	for _, cert := range certs {
		notBefore, notAfter, profileID := cert.NotBefore, cert.NotAfter, cert.ProfileID
		certResponses = append(certResponses, cryptoutilApiCaServer.CertificateSummary{
			SerialNumber: cert.SerialNumber,
			SubjectCN:    extractCommonName(cert.SubjectDN),
			NotBefore:    &notBefore,
			NotAfter:     &notAfter,
			Profile:      &profileID,
			Status:       cryptoutilApiCaServer.CertificateStatus(cert.Status),
		})
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	return strictJSON(fiber.StatusOK, cryptoutilApiCaServer.CertificateListResponse{
		Certificates: certResponses,
		Total:        total,
		Page:         page,
		PageSize:     filter.Limit,
	}), nil
}

func (s *StrictServer) GetCertificate(ctx context.Context, request cryptoutilApiCaServer.GetCertificateRequestObject) (cryptoutilApiCaServer.GetCertificateResponseObject, error) {
	if request.SerialNumber == "" {
		return strictError(fiber.StatusBadRequest, "invalid_serial", "serial number is required"), nil
	}

	cert, err := s.handler.storage.GetBySerialNumber(ctx, request.SerialNumber)
	if err != nil {
		if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
			return strictError(fiber.StatusNotFound, "not_found", "certificate not found"), nil
		}

		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	notBefore, notAfter, profileID := cert.NotBefore, cert.NotAfter, cert.ProfileID

	return strictJSON(fiber.StatusOK, cryptoutilApiCaServer.CertificateResponse{
		SerialNumber:   cert.SerialNumber,
		Subject:        buildCertificateSubject(cert.SubjectDN),
		Issuer:         buildCertificateSubject(cert.IssuerDN),
		NotBefore:      &notBefore,
		NotAfter:       &notAfter,
		Profile:        &profileID,
		Status:         cryptoutilApiCaServer.CertificateStatus(cert.Status),
		CertificatePEM: cert.CertificatePEM,
	}), nil
}

func (s *StrictServer) GetCertificateChain(ctx context.Context, request cryptoutilApiCaServer.GetCertificateChainRequestObject) (cryptoutilApiCaServer.GetCertificateChainResponseObject, error) {
	if request.SerialNumber == "" {
		return strictError(fiber.StatusBadRequest, "invalid_serial", "serial number is required"), nil
	}

	cert, err := s.handler.storage.GetBySerialNumber(ctx, request.SerialNumber)
	if err != nil {
		if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
			return strictError(fiber.StatusNotFound, "not_found", "certificate not found"), nil
		}

		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	return strictJSON(fiber.StatusOK, cryptoutilApiCaServer.CertificateChainResponse{
		Certificates: []cryptoutilApiCaServer.ChainCertificate{{
			CertificatePEM: cert.CertificatePEM,
			Subject:        buildCertificateSubjectValue(cert.SubjectDN),
			Issuer:         buildCertificateSubject(cert.IssuerDN),
		}},
	}), nil
}

func (s *StrictServer) RevokeCertificate(ctx context.Context, request cryptoutilApiCaServer.RevokeCertificateRequestObject) (cryptoutilApiCaServer.RevokeCertificateResponseObject, error) {
	if request.SerialNumber == "" {
		return strictError(fiber.StatusBadRequest, "invalid_serial", "serial number is required"), nil
	}

	if request.Body == nil {
		return strictError(fiber.StatusBadRequest, cryptoutilSharedMagic.ErrorInvalidRequest, cryptoutilSharedMagic.PKIMessageFailedToParseRequestBody), nil
	}

	req := *request.Body
	if req.Reason == cryptoutilApiCaServer.RemoveFromCRL {
		return strictError(fiber.StatusUnprocessableEntity, "unsupported_revocation_reason", cryptoutilSharedMagic.PKIErrorRemoveFromCRLUnsupported), nil
	}

	cert, err := s.handler.storage.GetBySerialNumber(ctx, request.SerialNumber)
	if err != nil {
		if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
			return strictError(fiber.StatusNotFound, "not_found", "certificate not found"), nil
		}

		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	if cert.Status == cryptoutilCAStorage.StatusRevoked {
		return strictError(fiber.StatusConflict, "already_revoked", "certificate is already revoked"), nil
	}

	if err := validateStoredCertificateIssuer(cert, s.handler.issuer); err != nil {
		return strictError(fiber.StatusInternalServerError, "issuer_mismatch", err.Error()), nil
	}

	if err := s.handler.storage.Revoke(ctx, cert.ID, mapAPIRevocationReasonToStorage(req.Reason)); err != nil {
		return strictError(fiber.StatusInternalServerError, "revocation_failed", err.Error()), nil
	}

	now := time.Now().UTC()
	message := fmt.Sprintf("Certificate %s has been revoked", request.SerialNumber)

	return strictJSON(fiber.StatusOK, cryptoutilApiCaServer.RevocationResponse{
		SerialNumber: request.SerialNumber,
		Status:       cryptoutilApiCaServer.RevocationResponseStatusRevoked,
		RevokedAt:    now,
		Reason:       req.Reason,
		Message:      &message,
	}), nil
}

func (s *StrictServer) SubmitEnrollment(ctx context.Context, request cryptoutilApiCaServer.SubmitEnrollmentRequestObject) (cryptoutilApiCaServer.SubmitEnrollmentResponseObject, error) {
	if request.Body == nil {
		return strictError(fiber.StatusBadRequest, cryptoutilSharedMagic.ErrorInvalidRequest, cryptoutilSharedMagic.PKIMessageFailedToParseRequestBody), nil
	}

	req := *request.Body
	if req.CSR == "" {
		return strictError(fiber.StatusBadRequest, "missing_csr", "CSR is required"), nil
	}

	if req.Profile == "" {
		return strictError(fiber.StatusBadRequest, "missing_profile", "profile is required"), nil
	}

	csr, err := s.handler.parseCSR(req.CSR)
	if err != nil {
		return strictError(fiber.StatusUnprocessableEntity, "invalid_csr", err.Error()), nil
	}

	s.handler.mu.RLock()
	profile, exists := s.handler.profiles[req.Profile]
	s.handler.mu.RUnlock()

	if !exists {
		return strictError(fiber.StatusBadRequest, "unknown_profile", "certificate profile not found"), nil
	}

	issueReq, err := s.handler.buildIssueRequest(csr, profile, &req)
	if err != nil {
		return strictError(fiber.StatusUnprocessableEntity, "validation_error", err.Error()), nil
	}

	requestID, err := googleUuid.NewV7()
	if err != nil {
		return strictResponseResult[cryptoutilApiCaServer.SubmitEnrollmentResponseObject](strictError(
			fiber.StatusInternalServerError,
			"request_id_error",
			"failed to generate enrollment request ID",
		))
	}

	issued, _, err := s.handler.issuer.Issue(issueReq)
	if err != nil {
		return strictError(fiber.StatusInternalServerError, "issuance_error", err.Error()), nil
	}

	response := s.handler.buildEnrollmentResponse(issued, requestID)
	if _, err := s.handler.storeIssuedCertificate(ctx, issued, req.Profile, response.RequestID); err != nil {
		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	return strictJSON(fiber.StatusCreated, response), nil
}

func (s *StrictServer) GetEnrollmentStatus(ctx context.Context, request cryptoutilApiCaServer.GetEnrollmentStatusRequestObject) (cryptoutilApiCaServer.GetEnrollmentStatusResponseObject, error) {
	requestID := googleUuid.UUID(request.RequestID)

	cert, err := s.handler.storage.GetByRequestID(ctx, requestID)
	if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
		return strictError(fiber.StatusNotFound, "not_found", "enrollment request not found"), nil
	}

	if err != nil {
		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	submittedAt, updatedAt := cert.CreatedAt, cert.UpdatedAt
	response := cryptoutilApiCaServer.EnrollmentStatusResponse{
		RequestID:   request.RequestID,
		Status:      cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued,
		SubmittedAt: &submittedAt,
		UpdatedAt:   &updatedAt,
	}
	response.Certificate = &cryptoutilApiCaServer.IssuedCertificate{
		SerialNumber:   cert.SerialNumber,
		CertificatePEM: cert.CertificatePEM,
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
	}

	return strictJSON(fiber.StatusOK, response), nil
}
