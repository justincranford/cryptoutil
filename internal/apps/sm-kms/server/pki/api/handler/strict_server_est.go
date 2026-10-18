// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

func (s *StrictServer) ESTCACerts(_ context.Context, _ cryptoutilApiCaServer.ESTCACertsRequestObject) (cryptoutilApiCaServer.ESTCACertsResponseObject, error) {
	h := s.handler
	if h.issuer == nil {
		return strictError(fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured"), nil
	}

	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil {
		return strictError(fiber.StatusServiceUnavailable, "service_unavailable", "CA not configured"), nil
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: caConfig.Certificate.Raw})

	return strictBinary(fiber.StatusOK, cryptoutilSharedMagic.PKIContentTypePKCS7CertsOnly, certPEM, [2]string{
		cryptoutilSharedMagic.PKIHeaderContentTransferEncoding,
		cryptoutilSharedMagic.PKIContentTransferEncodingBase64,
	}), nil
}

func (s *StrictServer) ESTCSRAttrs(_ context.Context, _ cryptoutilApiCaServer.ESTCSRAttrsRequestObject) (cryptoutilApiCaServer.ESTCSRAttrsResponseObject, error) {
	return strictStatus(fiber.StatusNoContent, ""), nil
}

func (s *StrictServer) ESTServerKeyGen(_ context.Context, _ cryptoutilApiCaServer.ESTServerKeyGenRequestObject) (cryptoutilApiCaServer.ESTServerKeyGenResponseObject, error) {
	return strictError(fiber.StatusServiceUnavailable, "service_unavailable", "secure server-side key generation is unavailable"), nil
}

func (s *StrictServer) ESTSimpleEnroll(ctx context.Context, request cryptoutilApiCaServer.ESTSimpleEnrollRequestObject) (cryptoutilApiCaServer.ESTSimpleEnrollResponseObject, error) {
	return s.estSimpleEnroll(ctx, request.Body)
}

func (s *StrictServer) ESTSimpleReenroll(ctx context.Context, request cryptoutilApiCaServer.ESTSimpleReenrollRequestObject) (cryptoutilApiCaServer.ESTSimpleReenrollResponseObject, error) {
	return s.estSimpleEnroll(ctx, request.Body)
}

func (s *StrictServer) estSimpleEnroll(ctx context.Context, body io.Reader) (strictResponse, error) {
	if body == nil {
		return strictError(fiber.StatusBadRequest, "bad_request", "empty CSR"), nil
	}

	csrBytes, err := io.ReadAll(body)
	if err != nil {
		return strictError(fiber.StatusBadRequest, "invalid_csr", err.Error()), nil
	}

	if len(csrBytes) == 0 {
		return strictError(fiber.StatusBadRequest, "bad_request", "empty CSR"), nil
	}

	csr, err := s.handler.parseESTCSR(csrBytes)
	if err != nil {
		return strictError(fiber.StatusBadRequest, "invalid_csr", err.Error()), nil
	}

	if err := csr.CheckSignature(); err != nil {
		return strictHandledResponse(strictError(
			fiber.StatusBadRequest,
			"invalid_signature",
			"CSR signature verification failed",
		))
	}

	s.handler.mu.RLock()

	var profile *ProfileConfig
	for _, candidate := range s.handler.profiles {
		profile = candidate

		break
	}

	s.handler.mu.RUnlock()

	if profile == nil {
		return strictError(fiber.StatusServiceUnavailable, "no_profile", "no certificate profiles configured"), nil
	}

	if s.handler.issuer == nil {
		return strictError(fiber.StatusServiceUnavailable, "service_unavailable", "CA issuer not configured"), nil
	}

	issued, _, err := s.handler.issuer.Issue(s.handler.buildESTIssueRequest(csr, profile))
	if err != nil {
		return strictError(fiber.StatusInternalServerError, "issuance_error", err.Error()), nil
	}

	requestID, err := googleUuid.NewV7()
	if err != nil {
		return strictHandledResponse(strictError(
			fiber.StatusInternalServerError,
			"request_id_error",
			"failed to generate request ID",
		))
	}

	if _, err := s.handler.storeIssuedCertificate(ctx, issued, profile.ID, requestID); err != nil {
		return strictError(fiber.StatusInternalServerError, "storage_error", err.Error()), nil
	}

	return strictBinary(fiber.StatusOK, cryptoutilSharedMagic.PKIContentTypePKCS7CertsOnly, issued.CertificatePEM, [2]string{
		cryptoutilSharedMagic.PKIHeaderContentTransferEncoding,
		cryptoutilSharedMagic.PKIContentTransferEncodingBase64,
	}), nil
}

func (s *StrictServer) HandleOCSP(ctx context.Context, request cryptoutilApiCaServer.HandleOCSPRequestObject) (cryptoutilApiCaServer.HandleOCSPResponseObject, error) {
	h := s.handler
	h.mu.RLock()
	ocspService := h.ocspService
	h.mu.RUnlock()

	if ocspService == nil {
		return strictStatus(fiber.StatusServiceUnavailable, cryptoutilSharedMagic.PKIContentTypeOCSPResponse), nil
	}

	if request.Body == nil {
		return strictStatus(fiber.StatusBadRequest, cryptoutilSharedMagic.PKIContentTypeOCSPResponse), nil
	}

	requestBody, err := io.ReadAll(request.Body)
	if err != nil || len(requestBody) == 0 {
		return strictResponseResult[cryptoutilApiCaServer.HandleOCSPResponseObject](
			strictStatus(fiber.StatusBadRequest, cryptoutilSharedMagic.PKIContentTypeOCSPResponse),
		)
	}

	if h.issuer == nil || h.issuer.GetCAConfig() == nil || h.issuer.GetCAConfig().Certificate == nil {
		return strictStatus(fiber.StatusServiceUnavailable, cryptoutilSharedMagic.PKIContentTypeOCSPResponse), nil
	}

	issuerKeyID := fmt.Sprintf("%x", h.issuer.GetCAConfig().Certificate.SubjectKeyId)
	if issuerKeyID == "" {
		return strictStatus(fiber.StatusServiceUnavailable, cryptoutilSharedMagic.PKIContentTypeOCSPResponse), nil
	}

	lookup := func(serialNumber *big.Int) (*x509.Certificate, *cryptoutilCAServiceRevocation.RevokedCertificate, error) {
		return h.lookupCertificateStatusBySerial(ctx, serialNumber, issuerKeyID)
	}

	responseBytes, err := ocspService.RespondToRequestWithStatus(requestBody, lookup)
	if err != nil {
		return strictResponseResult[cryptoutilApiCaServer.HandleOCSPResponseObject](
			strictStatus(fiber.StatusInternalServerError, cryptoutilSharedMagic.PKIContentTypeOCSPResponse),
		)
	}

	return strictBinary(fiber.StatusOK, cryptoutilSharedMagic.PKIContentTypeOCSPResponse, responseBytes), nil
}
