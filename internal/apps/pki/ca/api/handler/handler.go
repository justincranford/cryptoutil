// Copyright (c) 2025 Justin Cranford

// Package handler provides HTTP handlers for CA REST API endpoints.
package handler

import (
	"errors"
	"fmt"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/pki/ca/service/revocation"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/apps/pki/ca/service/timestamp"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Handler implements the CA enrollment ServerInterface.
type Handler struct {
	issuer            *cryptoutilCAServiceIssuer.Issuer
	storage           cryptoutilCAStorage.Store
	ocspService       *cryptoutilCAServiceRevocation.OCSPService
	crlService        *cryptoutilCAServiceRevocation.CRLService
	tsaService        *cryptoutilCAServiceTimestamp.TSAService
	profiles          map[string]*ProfileConfig
	enrollmentTracker *enrollmentTracker
	mu                sync.RWMutex
}

// enrollmentTracker tracks enrollment request status.
type enrollmentTracker struct {
	mu         sync.RWMutex
	requests   map[googleUuid.UUID]*enrollmentEntry
	maxEntries int
}

// enrollmentEntry represents a tracked enrollment request.
type enrollmentEntry struct {
	RequestID    googleUuid.UUID
	Status       cryptoutilApiCaServer.EnrollmentStatusResponseStatus
	SerialNumber string
	CreatedAt    time.Time
	CompletedAt  time.Time
}

// newEnrollmentTracker creates a new enrollment tracker with max entry limit.
func newEnrollmentTracker(maxEntries int) *enrollmentTracker {
	return &enrollmentTracker{
		requests:   make(map[googleUuid.UUID]*enrollmentEntry),
		maxEntries: maxEntries,
	}
}

// track records an enrollment.
func (t *enrollmentTracker) track(requestID googleUuid.UUID, status cryptoutilApiCaServer.EnrollmentStatusResponseStatus, serialNumber string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Enforce max entries by removing oldest if needed.
	if len(t.requests) >= t.maxEntries {
		var oldestID googleUuid.UUID

		var oldestTime time.Time

		for id, entry := range t.requests {
			if oldestTime.IsZero() || entry.CreatedAt.Before(oldestTime) {
				oldestID = id
				oldestTime = entry.CreatedAt
			}
		}

		delete(t.requests, oldestID)
	}

	now := time.Now().UTC()

	t.requests[requestID] = &enrollmentEntry{
		RequestID:    requestID,
		Status:       status,
		SerialNumber: serialNumber,
		CreatedAt:    now,
		CompletedAt:  now,
	}
}

// get retrieves an enrollment entry.
func (t *enrollmentTracker) get(requestID googleUuid.UUID) (*enrollmentEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entry, ok := t.requests[requestID]

	return entry, ok
}

// ProfileConfig holds combined profile configuration.
type ProfileConfig struct {
	ID                 string
	Name               string
	Description        string
	Category           string
	SubjectProfile     *cryptoutilCAProfileSubject.Profile
	CertificateProfile *cryptoutilCAProfileCertificate.Profile
}

// NewHandler creates a new enrollment handler.
func NewHandler(issuer *cryptoutilCAServiceIssuer.Issuer, storage cryptoutilCAStorage.Store, profiles map[string]*ProfileConfig) (*Handler, error) {
	if issuer == nil {
		return nil, fmt.Errorf("issuer is required")
	}

	if storage == nil {
		return nil, fmt.Errorf("storage is required")
	}

	if profiles == nil {
		profiles = make(map[string]*ProfileConfig)
	}

	return &Handler{
		issuer:            issuer,
		storage:           storage,
		profiles:          profiles,
		enrollmentTracker: newEnrollmentTracker(maxTrackedEnrollments),
	}, nil
}

// ListCertificates handles GET /certificates.
func (h *Handler) ListCertificates(c *fiber.Ctx, params cryptoutilApiCaServer.ListCertificatesParams) error {
	// Build filter from params.
	filter := &cryptoutilCAStorage.ListFilter{
		Limit:  cryptoutilSharedMagic.DefaultPageLimit,
		Offset: 0,
	}

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

	// List certificates from storage.
	certs, total, err := h.storage.List(c.Context(), filter)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", err.Error())
	}

	// Build response.
	certResponses := make([]cryptoutilApiCaServer.CertificateSummary, 0, len(certs))

	for _, cert := range certs {
		status := cryptoutilApiCaServer.CertificateStatus(cert.Status)
		notBefore := cert.NotBefore
		notAfter := cert.NotAfter
		profileID := cert.ProfileID

		certResponses = append(certResponses, cryptoutilApiCaServer.CertificateSummary{
			SerialNumber: cert.SerialNumber,
			SubjectCN:    extractCommonName(cert.SubjectDN),
			NotBefore:    &notBefore,
			NotAfter:     &notAfter,
			Profile:      &profileID,
			Status:       status,
		})
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	pageSize := filter.Limit

	if err := c.JSON(cryptoutilApiCaServer.CertificateListResponse{
		Certificates: certResponses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}); err != nil {
		return fmt.Errorf("failed to send certificate list response: %w", err)
	}

	return nil
}

// GetCertificate handles GET /certificates/{serialNumber}.
func (h *Handler) GetCertificate(c *fiber.Ctx, serialNumber string) error {
	if serialNumber == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_serial", "serial number is required")
	}

	cert, err := h.storage.GetBySerialNumber(c.Context(), serialNumber)
	if err != nil {
		if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "not_found", "certificate not found")
		}

		return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", err.Error())
	}

	status := cryptoutilApiCaServer.CertificateStatus(cert.Status)
	notBefore := cert.NotBefore
	notAfter := cert.NotAfter
	profileID := cert.ProfileID

	response := cryptoutilApiCaServer.CertificateResponse{
		SerialNumber:   cert.SerialNumber,
		Subject:        buildCertificateSubject(cert.SubjectDN),
		Issuer:         buildCertificateSubject(cert.IssuerDN),
		NotBefore:      &notBefore,
		NotAfter:       &notAfter,
		Profile:        &profileID,
		Status:         status,
		CertificatePEM: cert.CertificatePEM,
	}

	if err := c.JSON(response); err != nil {
		return fmt.Errorf("failed to send certificate response: %w", err)
	}

	return nil
}

// GetCertificateChain handles GET /certificates/{serialNumber}/chain.
func (h *Handler) GetCertificateChain(c *fiber.Ctx, serialNumber string) error {
	if serialNumber == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_serial", "serial number is required")
	}

	cert, err := h.storage.GetBySerialNumber(c.Context(), serialNumber)
	if err != nil {
		if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "not_found", "certificate not found")
		}

		return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", err.Error())
	}

	// Parse the certificate to build chain response.
	chainCerts := make([]cryptoutilApiCaServer.ChainCertificate, 0, 1)

	// Add the certificate itself.
	chainCerts = append(chainCerts, cryptoutilApiCaServer.ChainCertificate{
		CertificatePEM: cert.CertificatePEM,
		Subject:        buildCertificateSubjectValue(cert.SubjectDN),
		Issuer:         buildCertificateSubject(cert.IssuerDN),
	})

	if err := c.JSON(cryptoutilApiCaServer.CertificateChainResponse{
		Certificates: chainCerts,
	}); err != nil {
		return fmt.Errorf("failed to send certificate chain response: %w", err)
	}

	return nil
}

// RevokeCertificate handles POST /certificates/{serialNumber}/revoke.
func (h *Handler) RevokeCertificate(c *fiber.Ctx, serialNumber string) error {
	if serialNumber == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_serial", "serial number is required")
	}

	// Parse request body.
	var req cryptoutilApiCaServer.RevocationRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_request", "failed to parse request body")
	}

	// Look up certificate by serial number.
	cert, err := h.storage.GetBySerialNumber(c.Context(), serialNumber)
	if err != nil {
		if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
			return h.errorResponse(c, fiber.StatusNotFound, "not_found", "certificate not found")
		}

		return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", err.Error())
	}

	// Check if already revoked.
	if cert.Status == cryptoutilCAStorage.StatusRevoked {
		return h.errorResponse(c, fiber.StatusConflict, "already_revoked", "certificate is already revoked")
	}

	// Convert API reason to storage reason.
	storageReason := mapAPIRevocationReasonToStorage(req.Reason)

	// Revoke the certificate.
	if err := h.storage.Revoke(c.Context(), cert.ID, storageReason); err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "revocation_failed", err.Error())
	}

	// Build response.
	now := time.Now().UTC()
	message := fmt.Sprintf("Certificate %s has been revoked", serialNumber)

	response := cryptoutilApiCaServer.RevocationResponse{
		SerialNumber: serialNumber,
		Status:       cryptoutilApiCaServer.RevocationResponseStatusRevoked,
		RevokedAt:    now,
		Reason:       req.Reason,
		Message:      &message,
	}

	if err := c.JSON(response); err != nil {
		return fmt.Errorf("failed to send revocation response: %w", err)
	}

	return nil
}

// mapAPIRevocationReasonToStorage converts an API RevocationReason to storage RevocationReason.
func mapAPIRevocationReasonToStorage(reason cryptoutilApiCaServer.RevocationReason) cryptoutilCAStorage.RevocationReason {
	switch reason {
	case cryptoutilApiCaServer.KeyCompromise:
		return cryptoutilCAStorage.ReasonKeyCompromise
	case cryptoutilApiCaServer.CACompromise:
		return cryptoutilCAStorage.ReasonCACompromise
	case cryptoutilApiCaServer.AffiliationChanged:
		return cryptoutilCAStorage.ReasonAffiliationChanged
	case cryptoutilApiCaServer.Superseded:
		return cryptoutilCAStorage.ReasonSuperseded
	case cryptoutilApiCaServer.CessationOfOperation:
		return cryptoutilCAStorage.ReasonCessationOfOperation
	case cryptoutilApiCaServer.CertificateHold:
		return cryptoutilCAStorage.ReasonCertificateHold
	case cryptoutilApiCaServer.RemoveFromCRL:
		return cryptoutilCAStorage.ReasonRemoveFromCRL
	case cryptoutilApiCaServer.PrivilegeWithdrawn:
		return cryptoutilCAStorage.ReasonPrivilegeWithdrawn
	case cryptoutilApiCaServer.AaCompromise:
		return cryptoutilCAStorage.ReasonAACompromise
	default:
		return cryptoutilCAStorage.ReasonUnspecified
	}
}

// SubmitEnrollment handles POST /enroll.
func (h *Handler) SubmitEnrollment(c *fiber.Ctx) error {
	var req cryptoutilApiCaServer.EnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_request", "failed to parse request body")
	}

	// Validate required fields.
	if req.CSR == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "missing_csr", "CSR is required")
	}

	if req.Profile == "" {
		return h.errorResponse(c, fiber.StatusBadRequest, "missing_profile", "profile is required")
	}

	// Parse the CSR.
	csrPEM, err := h.parseCSR(req.CSR)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnprocessableEntity, "invalid_csr", err.Error())
	}

	// Get the profile.
	h.mu.RLock()
	profile, exists := h.profiles[req.Profile]
	h.mu.RUnlock()

	if !exists {
		return h.errorResponse(c, fiber.StatusBadRequest, "unknown_profile", "certificate profile not found")
	}

	// Build the issuance request.
	issueReq, err := h.buildIssueRequest(csrPEM, profile, &req)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnprocessableEntity, "validation_error", err.Error())
	}

	// Issue the certificate.
	issued, _, err := h.issuer.Issue(issueReq)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuance_error", err.Error())
	}

	// Build the response.
	resp := h.buildEnrollmentResponse(issued)

	// Track the enrollment - convert status type.
	statusForTracking := cryptoutilApiCaServer.EnrollmentStatusResponseStatus(resp.Status)
	h.enrollmentTracker.track(resp.RequestID, statusForTracking, issued.SerialNumber)

	if err := c.Status(fiber.StatusCreated).JSON(resp); err != nil {
		return fmt.Errorf("failed to send enrollment response: %w", err)
	}

	return nil
}

// GetEnrollmentStatus handles GET /enroll/{requestId}.
func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx, requestID googleUuid.UUID) error {
	// Look up the enrollment in the tracker.
	entry, found := h.enrollmentTracker.get(requestID)
	if !found {
		return h.errorResponse(c, fiber.StatusNotFound, "not_found", "enrollment request not found")
	}

	// Build response based on tracked status.
	submittedAt := entry.CreatedAt
	updatedAt := entry.CompletedAt

	resp := cryptoutilApiCaServer.EnrollmentStatusResponse{
		RequestID:   entry.RequestID,
		Status:      entry.Status,
		SubmittedAt: &submittedAt,
		UpdatedAt:   &updatedAt,
	}

	// If issued, try to get the certificate from storage.
	if entry.Status == cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued && entry.SerialNumber != "" {
		cert, err := h.storage.GetBySerialNumber(c.Context(), entry.SerialNumber)
		if err == nil {
			notBefore := cert.NotBefore
			notAfter := cert.NotAfter

			resp.Certificate = &cryptoutilApiCaServer.IssuedCertificate{
				SerialNumber:   cert.SerialNumber,
				CertificatePEM: cert.CertificatePEM,
				NotBefore:      notBefore,
				NotAfter:       notAfter,
			}
		}
	}

	if err := c.JSON(resp); err != nil {
		return fmt.Errorf("failed to send enrollment status response: %w", err)
	}

	return nil
}

// ListCAs handles GET /ca.
