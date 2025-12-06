// Copyright (c) 2025 Justin Cranford

package handler

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	cryptoutilCAMagic "cryptoutil/internal/ca/magic"
	cryptoutilCAProfileCertificate "cryptoutil/internal/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/ca/service/revocation"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/ca/service/timestamp"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
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
	requests   map[uuid.UUID]*enrollmentEntry
	maxEntries int
}

// enrollmentEntry represents a tracked enrollment request.
type enrollmentEntry struct {
	RequestID    uuid.UUID
	Status       cryptoutilCAServer.EnrollmentStatusResponseStatus
	SerialNumber string
	CreatedAt    time.Time
	CompletedAt  time.Time
}

// newEnrollmentTracker creates a new enrollment tracker with max entry limit.
func newEnrollmentTracker(maxEntries int) *enrollmentTracker {
	return &enrollmentTracker{
		requests:   make(map[uuid.UUID]*enrollmentEntry),
		maxEntries: maxEntries,
	}
}

// track records an enrollment.
func (t *enrollmentTracker) track(requestID uuid.UUID, status cryptoutilCAServer.EnrollmentStatusResponseStatus, serialNumber string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Enforce max entries by removing oldest if needed.
	if len(t.requests) >= t.maxEntries {
		var oldestID uuid.UUID

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
func (t *enrollmentTracker) get(requestID uuid.UUID) (*enrollmentEntry, bool) {
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
func (h *Handler) ListCertificates(c *fiber.Ctx, params cryptoutilCAServer.ListCertificatesParams) error {
	// Build filter from params.
	filter := &cryptoutilCAStorage.ListFilter{
		Limit:  cryptoutilCAMagic.DefaultPageLimit,
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
	certResponses := make([]cryptoutilCAServer.CertificateSummary, 0, len(certs))

	for _, cert := range certs {
		status := cryptoutilCAServer.CertificateStatus(cert.Status)
		notBefore := cert.NotBefore
		notAfter := cert.NotAfter
		profileID := cert.ProfileID

		certResponses = append(certResponses, cryptoutilCAServer.CertificateSummary{
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

	if err := c.JSON(cryptoutilCAServer.CertificateListResponse{
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

	status := cryptoutilCAServer.CertificateStatus(cert.Status)
	notBefore := cert.NotBefore
	notAfter := cert.NotAfter
	profileID := cert.ProfileID

	response := cryptoutilCAServer.CertificateResponse{
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
	chainCerts := make([]cryptoutilCAServer.ChainCertificate, 0, 1)

	// Add the certificate itself.
	chainCerts = append(chainCerts, cryptoutilCAServer.ChainCertificate{
		CertificatePEM: cert.CertificatePEM,
		Subject:        buildCertificateSubjectValue(cert.SubjectDN),
		Issuer:         buildCertificateSubject(cert.IssuerDN),
	})

	if err := c.JSON(cryptoutilCAServer.CertificateChainResponse{
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
	var req cryptoutilCAServer.RevocationRequest
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

	response := cryptoutilCAServer.RevocationResponse{
		SerialNumber: serialNumber,
		Status:       cryptoutilCAServer.RevocationResponseStatusRevoked,
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
func mapAPIRevocationReasonToStorage(reason cryptoutilCAServer.RevocationReason) cryptoutilCAStorage.RevocationReason {
	switch reason {
	case cryptoutilCAServer.KeyCompromise:
		return cryptoutilCAStorage.ReasonKeyCompromise
	case cryptoutilCAServer.CACompromise:
		return cryptoutilCAStorage.ReasonCACompromise
	case cryptoutilCAServer.AffiliationChanged:
		return cryptoutilCAStorage.ReasonAffiliationChanged
	case cryptoutilCAServer.Superseded:
		return cryptoutilCAStorage.ReasonSuperseded
	case cryptoutilCAServer.CessationOfOperation:
		return cryptoutilCAStorage.ReasonCessationOfOperation
	case cryptoutilCAServer.CertificateHold:
		return cryptoutilCAStorage.ReasonCertificateHold
	case cryptoutilCAServer.RemoveFromCRL:
		return cryptoutilCAStorage.ReasonRemoveFromCRL
	case cryptoutilCAServer.PrivilegeWithdrawn:
		return cryptoutilCAStorage.ReasonPrivilegeWithdrawn
	case cryptoutilCAServer.AaCompromise:
		return cryptoutilCAStorage.ReasonAACompromise
	default:
		return cryptoutilCAStorage.ReasonUnspecified
	}
}

// SubmitEnrollment handles POST /enroll.
func (h *Handler) SubmitEnrollment(c *fiber.Ctx) error {
	var req cryptoutilCAServer.EnrollmentRequest
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
	statusForTracking := cryptoutilCAServer.EnrollmentStatusResponseStatus(resp.Status)
	h.enrollmentTracker.track(resp.RequestID, statusForTracking, issued.SerialNumber)

	if err := c.Status(fiber.StatusCreated).JSON(resp); err != nil {
		return fmt.Errorf("failed to send enrollment response: %w", err)
	}

	return nil
}

// GetEnrollmentStatus handles GET /enroll/{requestId}.
func (h *Handler) GetEnrollmentStatus(c *fiber.Ctx, requestID uuid.UUID) error {
	// Look up the enrollment in the tracker.
	entry, found := h.enrollmentTracker.get(requestID)
	if !found {
		return h.errorResponse(c, fiber.StatusNotFound, "not_found", "enrollment request not found")
	}

	// Build response based on tracked status.
	submittedAt := entry.CreatedAt
	updatedAt := entry.CompletedAt

	resp := cryptoutilCAServer.EnrollmentStatusResponse{
		RequestID:   entry.RequestID,
		Status:      entry.Status,
		SubmittedAt: &submittedAt,
		UpdatedAt:   &updatedAt,
	}

	// If issued, try to get the certificate from storage.
	if entry.Status == cryptoutilCAServer.EnrollmentStatusResponseStatusIssued && entry.SerialNumber != "" {
		cert, err := h.storage.GetBySerialNumber(c.Context(), entry.SerialNumber)
		if err == nil {
			notBefore := cert.NotBefore
			notAfter := cert.NotAfter

			resp.Certificate = &cryptoutilCAServer.IssuedCertificate{
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
func (h *Handler) ListCAs(c *fiber.Ctx) error {
	// Check if issuer is configured.
	if h.issuer == nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured")
	}

	// Get CA info from issuer.
	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "ca_config_error", "CA configuration not available")
	}

	// Build summary from the issuer's CA certificate.
	caCert := caConfig.Certificate
	caType := cryptoutilCAServer.CASummaryTypeIntermediate

	// Check if this is a self-signed (root) CA.
	if caCert.Issuer.String() == caCert.Subject.String() {
		caType = cryptoutilCAServer.CASummaryTypeRoot
	}

	validUntil := caCert.NotAfter
	summary := cryptoutilCAServer.CASummary{
		ID:         caConfig.Name,
		Name:       caConfig.Name,
		Type:       caType,
		Status:     cryptoutilCAServer.CASummaryStatusActive,
		SubjectCN:  &caCert.Subject.CommonName,
		ValidUntil: &validUntil,
	}

	response := cryptoutilCAServer.CAListResponse{
		Authorities: []cryptoutilCAServer.CASummary{summary},
	}

	if err := c.JSON(response); err != nil {
		return fmt.Errorf("failed to send CA list response: %w", err)
	}

	return nil
}

// GetCA handles GET /ca/{caId}.
func (h *Handler) GetCA(c *fiber.Ctx, caID string) error {
	// Check if issuer is configured.
	if h.issuer == nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured")
	}

	// Get CA info from issuer.
	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "ca_config_error", "CA configuration not available")
	}

	// Check if requested CA matches.
	if caID != caConfig.Name {
		return h.errorResponse(c, fiber.StatusNotFound, "not_found", "CA not found")
	}

	caCert := caConfig.Certificate
	caType := cryptoutilCAServer.CAResponseTypeIntermediate

	// Check if this is a self-signed (root) CA.
	if caCert.Issuer.String() == caCert.Subject.String() {
		caType = cryptoutilCAServer.CAResponseTypeRoot
	}

	// Encode certificate to PEM.
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	notBefore := caCert.NotBefore
	notAfter := caCert.NotAfter
	serialNumber := fmt.Sprintf("%X", caCert.SerialNumber)

	// Determine key algorithm and size.
	keyAlgo, keySize := getKeyInfo(caCert)

	response := cryptoutilCAServer.CAResponse{
		ID:                    caConfig.Name,
		Name:                  caConfig.Name,
		Type:                  caType,
		Status:                cryptoutilCAServer.CAResponseStatusActive,
		Subject:               buildCertificateSubject(caCert.Subject.String()),
		Issuer:                buildCertificateSubject(caCert.Issuer.String()),
		SerialNumber:          &serialNumber,
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

	// Add path length if basic constraints apply.
	if caCert.BasicConstraintsValid && caCert.IsCA {
		pathLen := caCert.MaxPathLen
		response.PathLength = &pathLen
	}

	if err := c.JSON(response); err != nil {
		return fmt.Errorf("failed to send CA response: %w", err)
	}

	return nil
}

// GetCRL handles GET /ca/{caId}/crl.
func (h *Handler) GetCRL(c *fiber.Ctx, caID string, params cryptoutilCAServer.GetCRLParams) error {
	// Check if CRL service is configured.
	h.mu.RLock()
	crlService := h.crlService
	h.mu.RUnlock()

	if crlService == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "service_unavailable", "CRL service not configured")
	}

	// Verify the CA ID matches the configured issuer.
	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil || caConfig.Name != caID {
		return h.errorResponse(c, fiber.StatusNotFound, "not_found", "CA not found")
	}

	// Determine output format (default to DER).
	format := "der"
	if params.Format != nil {
		format = string(*params.Format)
	}

	// Generate the CRL based on format.
	switch format {
	case "pem":
		crlPEM, err := crlService.GenerateCRLPEM()
		if err != nil {
			return h.errorResponse(c, fiber.StatusInternalServerError, "crl_error", err.Error())
		}

		c.Set("Content-Type", "application/x-pem-file")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.crl.pem\"", caID))

		if err := c.Send(crlPEM); err != nil {
			return fmt.Errorf("failed to send CRL PEM: %w", err)
		}

		return nil
	default:
		// DER format.
		crlDER, err := crlService.GenerateCRL()
		if err != nil {
			return h.errorResponse(c, fiber.StatusInternalServerError, "crl_error", err.Error())
		}

		c.Set("Content-Type", "application/pkix-crl")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.crl\"", caID))

		if err := c.Send(crlDER); err != nil {
			return fmt.Errorf("failed to send CRL DER: %w", err)
		}

		return nil
	}
}

// getKeyInfo extracts key algorithm and size from a certificate.
func getKeyInfo(cert *x509.Certificate) (string, int) {
	switch pub := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		return "ECDSA", pub.Curve.Params().BitSize
	case *rsa.PublicKey:
		return "RSA", pub.N.BitLen()
	case ed25519.PublicKey:
		return "EdDSA", ed25519.PublicKeySize * cryptoutilCAMagic.BitsPerByte
	default:
		return "unknown", 0
	}
}

// ptrString returns a pointer to a string, or nil if empty.
func ptrString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// ptrStringSlice returns a pointer to a string slice, or nil if empty.
func ptrStringSlice(s []string) *[]string {
	if len(s) == 0 {
		return nil
	}

	return &s
}

// ListProfiles handles GET /profiles.
func (h *Handler) ListProfiles(c *fiber.Ctx) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	profiles := make([]cryptoutilCAServer.ProfileSummary, 0, len(h.profiles))

	for _, p := range h.profiles {
		category := h.mapCategory(p.Category)
		profiles = append(profiles, cryptoutilCAServer.ProfileSummary{
			ID:          p.ID,
			Name:        p.Name,
			Description: &p.Description,
			Category:    &category,
		})
	}

	if err := c.JSON(cryptoutilCAServer.ProfileListResponse{Profiles: profiles}); err != nil {
		return fmt.Errorf("failed to send profile list response: %w", err)
	}

	return nil
}

// GetProfile handles GET /profiles/{profileId}.
func (h *Handler) GetProfile(c *fiber.Ctx, profileID string) error {
	h.mu.RLock()
	profile, exists := h.profiles[profileID]
	h.mu.RUnlock()

	if !exists {
		return h.errorResponse(c, fiber.StatusNotFound, "not_found", "profile not found")
	}

	if err := c.JSON(h.buildProfileResponse(profile)); err != nil {
		return fmt.Errorf("failed to send profile response: %w", err)
	}

	return nil
}

// EstCACerts handles GET /est/cacerts - RFC 7030 Section 4.1.
// Returns the CA certificates in PKCS#7 format for EST clients.
// Note: Full PKCS#7 degenerate format requires a CMS library.
// This implementation returns Base64-encoded PEM for compatibility.
func (h *Handler) EstCACerts(c *fiber.Ctx) error {
	// Check if issuer is configured.
	if h.issuer == nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured")
	}

	// Get CA configuration.
	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "service_unavailable", "CA not configured")
	}

	// Encode CA certificate to PEM.
	caCert := caConfig.Certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// Per RFC 7030, the response should be Base64-encoded PKCS#7.
	// Since we don't have a PKCS#7 library, return Base64-encoded PEM
	// which many EST clients can handle.
	c.Set("Content-Type", "application/pkcs7-mime; smime-type=certs-only")
	c.Set("Content-Transfer-Encoding", "base64")

	// Return the PEM-encoded certificate.
	if err := c.Send(certPEM); err != nil {
		return fmt.Errorf("failed to send CA certificates: %w", err)
	}

	return nil
}

// EstCSRAttrs handles GET /est/csrattrs - RFC 7030 Section 4.5.
// Returns the CSR attributes required or recommended by the CA.
func (h *Handler) EstCSRAttrs(c *fiber.Ctx) error {
	// Most CAs don't require specific CSR attributes.
	// Return 204 No Content to indicate no attributes required.
	if err := c.SendStatus(fiber.StatusNoContent); err != nil {
		return fmt.Errorf("failed to send no content status: %w", err)
	}

	return nil
}

// EstSimpleEnroll handles POST /est/simpleenroll - RFC 7030 Section 4.2.
// Processes a PKCS#10 CSR and returns the issued certificate in PKCS#7 format.
// Note: Full mTLS authentication requires TLS middleware configuration.
// This implementation accepts CSR in DER or Base64 format.
func (h *Handler) EstSimpleEnroll(c *fiber.Ctx) error {
	// Read the CSR from request body.
	csrBytes := c.Body()
	if len(csrBytes) == 0 {
		return h.errorResponse(c, fiber.StatusBadRequest, "bad_request", "empty CSR")
	}

	// Parse the CSR (accept DER or Base64-encoded DER).
	csr, err := h.parseESTCSR(csrBytes)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_csr", err.Error())
	}

	// Verify CSR signature.
	if err := csr.CheckSignature(); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_signature", "CSR signature verification failed")
	}

	// Use default profile for EST enrollment.
	// In production, this would be determined by mTLS client certificate or request path.
	h.mu.RLock()

	var profile *ProfileConfig

	for _, p := range h.profiles {
		profile = p

		break // Use first available profile.
	}

	h.mu.RUnlock()

	if profile == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "no_profile", "no certificate profiles configured")
	}

	// Build the issuance request using CSR subject.
	issueReq := h.buildESTIssueRequest(csr, profile)

	// Issue the certificate.
	issued, _, err := h.issuer.Issue(issueReq)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuance_error", err.Error())
	}

	// Return the certificate in PEM format (wrapped as PKCS#7 would be in production).
	c.Set("Content-Type", "application/pkcs7-mime; smime-type=certs-only")
	c.Set("Content-Transfer-Encoding", "base64")

	if err := c.Send(issued.CertificatePEM); err != nil {
		return fmt.Errorf("failed to send certificate: %w", err)
	}

	return nil
}

// parseESTCSR parses a CSR in DER or Base64 format for EST.
func (h *Handler) parseESTCSR(data []byte) (*x509.CertificateRequest, error) {
	// Try to parse as DER first.
	csr, err := x509.ParseCertificateRequest(data)
	if err == nil {
		return csr, nil
	}

	// Try Base64-decoding.
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))

	n, decodeErr := base64.StdEncoding.Decode(decoded, data)
	if decodeErr == nil {
		csr, err = x509.ParseCertificateRequest(decoded[:n])
		if err == nil {
			return csr, nil
		}
	}

	// Try PEM format as fallback.
	block, _ := pem.Decode(data)
	if block != nil && block.Type == pemTypeCertificateReq {
		csr, err = x509.ParseCertificateRequest(block.Bytes)
		if err == nil {
			return csr, nil
		}
	}

	return nil, fmt.Errorf("failed to parse CSR: invalid format")
}

// buildESTIssueRequest constructs an issuance request for EST enrollment.
func (h *Handler) buildESTIssueRequest(
	csr *x509.CertificateRequest,
	profile *ProfileConfig,
) *cryptoutilCAServiceIssuer.CertificateRequest {
	_ = profile // Profile would be used in production to configure extensions.

	// Build subject request from CSR.
	subjectReq := &cryptoutilCAProfileSubject.Request{
		CommonName:         csr.Subject.CommonName,
		Organization:       csr.Subject.Organization,
		OrganizationalUnit: csr.Subject.OrganizationalUnit,
		Country:            csr.Subject.Country,
		State:              csr.Subject.Province,
		Locality:           csr.Subject.Locality,
		DNSNames:           csr.DNSNames,
		IPAddresses:        h.ipsToStrings(csr.IPAddresses),
		EmailAddresses:     csr.EmailAddresses,
		URIs:               h.urisToStrings(csr.URIs),
	}

	return &cryptoutilCAServiceIssuer.CertificateRequest{
		PublicKey:        csr.PublicKey,
		SubjectRequest:   subjectReq,
		ValidityDuration: time.Duration(defaultValidityDays) * hoursPerDay * time.Hour,
	}
}

// EstSimpleReenroll handles POST /est/simplereenroll - RFC 7030 Section 4.2.2.
// Processes a PKCS#10 CSR to renew an existing certificate.
// Note: Full mTLS authentication requires the client to authenticate with
// the certificate being renewed.
func (h *Handler) EstSimpleReenroll(c *fiber.Ctx) error {
	// Re-enrollment uses the same logic as simple enrollment.
	// In production, mTLS would verify the client certificate being renewed.
	return h.EstSimpleEnroll(c)
}

// EstServerKeyGen handles POST /est/serverkeygen - RFC 7030 Section 4.4.
// Generates a key pair on the server and returns the certificate and encrypted private key.
func (h *Handler) EstServerKeyGen(c *fiber.Ctx) error {
	// Server key generation requires mTLS and secure key transport.
	// For now, return not implemented.
	return h.errorResponse(c, fiber.StatusNotImplemented, "not_implemented", "EST serverkeygen requires mTLS authentication")
}

// TsaTimestamp handles POST /tsa/timestamp - RFC 3161 timestamp request.
func (h *Handler) TsaTimestamp(c *fiber.Ctx) error {
	// Check if TSA service is configured.
	h.mu.RLock()
	tsaService := h.tsaService
	h.mu.RUnlock()

	if tsaService == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "service_unavailable", "TSA service not configured")
	}

	// Read the timestamp request body.
	requestBody := c.Body()
	if len(requestBody) == 0 {
		return h.errorResponse(c, fiber.StatusBadRequest, "bad_request", "empty timestamp request")
	}

	// Parse the DER-encoded TimeStampReq.
	tsReq, err := cryptoutilCAServiceTimestamp.ParseTimestampRequest(requestBody)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_request", fmt.Sprintf("failed to parse timestamp request: %v", err))
	}

	// Process the timestamp request.
	tsResp, err := tsaService.CreateTimestamp(tsReq)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "timestamp_error", err.Error())
	}

	// Serialize the response to DER format.
	respDER, err := cryptoutilCAServiceTimestamp.SerializeTimestampResponse(tsResp)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "serialization_error", err.Error())
	}

	// Return the timestamp response.
	c.Set("Content-Type", "application/timestamp-reply")

	if err := c.Send(respDER); err != nil {
		return fmt.Errorf("failed to send timestamp response: %w", err)
	}

	return nil
}

// SetTSAService configures the TSA service for the handler.
// This is optional - if not set, TSA requests will return service unavailable.
func (h *Handler) SetTSAService(tsaService *cryptoutilCAServiceTimestamp.TSAService) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.tsaService = tsaService
}

// SetOCSPService configures the OCSP service for the handler.
// This is optional - if not set, OCSP requests will return service unavailable.
func (h *Handler) SetOCSPService(ocspService *cryptoutilCAServiceRevocation.OCSPService) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ocspService = ocspService
}

// SetCRLService configures the CRL service for the handler.
// This is optional - if not set, CRL requests will return service unavailable.
func (h *Handler) SetCRLService(crlService *cryptoutilCAServiceRevocation.CRLService) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.crlService = crlService
}

// HandleOCSP handles POST /ocsp - RFC 6960 OCSP responder.
func (h *Handler) HandleOCSP(c *fiber.Ctx) error {
	// Check if OCSP service is configured.
	h.mu.RLock()
	ocspService := h.ocspService
	h.mu.RUnlock()

	if ocspService == nil {
		return h.ocspErrorResponse(c, fiber.StatusServiceUnavailable)
	}

	// Read the OCSP request body.
	requestBody, err := io.ReadAll(c.Request().BodyStream())
	if err != nil {
		return h.ocspErrorResponse(c, fiber.StatusBadRequest)
	}

	if len(requestBody) == 0 {
		requestBody = c.Body()
	}

	if len(requestBody) == 0 {
		return h.ocspErrorResponse(c, fiber.StatusBadRequest)
	}

	// Create a certificate lookup function that captures the context.
	ctx := c.Context()
	lookupFunc := func(serialNumber *big.Int) *x509.Certificate {
		return h.lookupCertificateBySerial(ctx, serialNumber)
	}

	// Process the OCSP request using a certificate lookup function.
	responseBytes, err := ocspService.RespondToRequest(requestBody, lookupFunc)
	if err != nil {
		return h.ocspErrorResponse(c, fiber.StatusInternalServerError)
	}

	// Set content type for OCSP response.
	c.Set("Content-Type", "application/ocsp-response")

	if err := c.Send(responseBytes); err != nil {
		return fmt.Errorf("failed to send OCSP response: %w", err)
	}

	return nil
}

// lookupCertificateBySerial finds a certificate by serial number for OCSP processing.
func (h *Handler) lookupCertificateBySerial(ctx context.Context, serialNumber *big.Int) *x509.Certificate {
	if serialNumber == nil {
		return nil
	}

	// Look up in storage using hex-encoded serial.
	serialHex := serialNumber.Text(cryptoutilCAMagic.HexBase)

	storedCert, err := h.storage.GetBySerialNumber(ctx, serialHex)
	if err != nil {
		return nil
	}

	// Parse the stored certificate.
	block, _ := pem.Decode([]byte(storedCert.CertificatePEM))
	if block == nil {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	return cert
}

// ocspErrorResponse sends an OCSP error response with appropriate content type.
func (h *Handler) ocspErrorResponse(c *fiber.Ctx, statusCode int) error {
	c.Set("Content-Type", "application/ocsp-response")

	if err := c.SendStatus(statusCode); err != nil {
		return fmt.Errorf("failed to send OCSP error response: %w", err)
	}

	return nil
}

// parseCSR parses a PEM-encoded CSR.
func (h *Handler) parseCSR(csrPEM string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != pemTypeCertificateReq {
		return nil, fmt.Errorf("expected %s, got %s", pemTypeCertificateReq, block.Type)
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR: %w", err)
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("CSR signature verification failed: %w", err)
	}

	return csr, nil
}

// buildIssueRequest constructs an issuance request from the enrollment request.
func (h *Handler) buildIssueRequest(
	csr *x509.CertificateRequest,
	profile *ProfileConfig,
	req *cryptoutilCAServer.EnrollmentRequest,
) (*cryptoutilCAServiceIssuer.CertificateRequest, error) {
	// Build subject request from CSR.
	subjectReq := &cryptoutilCAProfileSubject.Request{
		CommonName:         csr.Subject.CommonName,
		Organization:       csr.Subject.Organization,
		OrganizationalUnit: csr.Subject.OrganizationalUnit,
		Country:            csr.Subject.Country,
		State:              csr.Subject.Province,
		Locality:           csr.Subject.Locality,
	}

	// Apply subject overrides.
	if req.SubjectOverride != nil {
		h.applySubjectOverrides(subjectReq, req.SubjectOverride)
	}

	// Apply SAN values from CSR.
	subjectReq.DNSNames = csr.DNSNames
	subjectReq.IPAddresses = h.ipsToStrings(csr.IPAddresses)
	subjectReq.EmailAddresses = csr.EmailAddresses
	subjectReq.URIs = h.urisToStrings(csr.URIs)

	// Apply SAN overrides.
	if req.SANOverride != nil {
		h.applySANOverrides(subjectReq, req.SANOverride)
	}

	// Determine validity.
	validityDays := defaultValidityDays
	if req.ValidityDays != nil {
		validityDays = *req.ValidityDays
	}

	validity := time.Duration(validityDays) * hoursPerDay * time.Hour

	return &cryptoutilCAServiceIssuer.CertificateRequest{
		PublicKey:        csr.PublicKey,
		SubjectRequest:   subjectReq,
		ValidityDuration: validity,
	}, nil
}

// applySubjectOverrides applies subject field overrides from the request.
func (h *Handler) applySubjectOverrides(
	subjectReq *cryptoutilCAProfileSubject.Request,
	override *cryptoutilCAServer.SubjectOverride,
) {
	if override.Organization != nil && len(*override.Organization) > 0 {
		subjectReq.Organization = *override.Organization
	}

	if override.OrganizationalUnit != nil && len(*override.OrganizationalUnit) > 0 {
		subjectReq.OrganizationalUnit = *override.OrganizationalUnit
	}

	if override.Country != nil && len(*override.Country) > 0 {
		subjectReq.Country = *override.Country
	}

	if override.State != nil && len(*override.State) > 0 {
		subjectReq.State = *override.State
	}

	if override.Locality != nil && len(*override.Locality) > 0 {
		subjectReq.Locality = *override.Locality
	}
}

// applySANOverrides applies SAN overrides from the request.
func (h *Handler) applySANOverrides(
	subjectReq *cryptoutilCAProfileSubject.Request,
	override *cryptoutilCAServer.SANOverride,
) {
	if override.DNSNames != nil && len(*override.DNSNames) > 0 {
		subjectReq.DNSNames = *override.DNSNames
	}

	if override.IPAddresses != nil && len(*override.IPAddresses) > 0 {
		subjectReq.IPAddresses = *override.IPAddresses
	}

	if override.EmailAddresses != nil && len(*override.EmailAddresses) > 0 {
		subjectReq.EmailAddresses = *override.EmailAddresses
	}

	if override.Uris != nil && len(*override.Uris) > 0 {
		subjectReq.URIs = *override.Uris
	}
}

// ipsToStrings converts IP addresses to strings.
func (h *Handler) ipsToStrings(ips []net.IP) []string {
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}

	return result
}

// urisToStrings converts URIs to strings.
func (h *Handler) urisToStrings(uris []*url.URL) []string {
	result := make([]string, len(uris))
	for i, u := range uris {
		result[i] = u.String()
	}

	return result
}

// buildEnrollmentResponse constructs the enrollment response.
func (h *Handler) buildEnrollmentResponse(issued *cryptoutilCAServiceIssuer.IssuedCertificate) *cryptoutilCAServer.EnrollmentResponse {
	certPEM := string(issued.CertificatePEM)
	chainPEM := string(issued.ChainPEM)

	notBefore := issued.Certificate.NotBefore
	notAfter := issued.Certificate.NotAfter
	serialNumber := issued.SerialNumber

	subject := h.certSubjectToAPI(issued.Certificate)

	return &cryptoutilCAServer.EnrollmentResponse{
		RequestID: uuid.New(),
		Status:    cryptoutilCAServer.Issued,
		Certificate: cryptoutilCAServer.IssuedCertificate{
			SerialNumber:      serialNumber,
			CertificatePEM:    certPEM,
			ChainPEM:          &chainPEM,
			NotBefore:         notBefore,
			NotAfter:          notAfter,
			Subject:           &subject,
			FingerprintSha256: &issued.Fingerprint,
		},
	}
}

// certSubjectToAPI converts certificate subject to API format.
func (h *Handler) certSubjectToAPI(cert *x509.Certificate) cryptoutilCAServer.CertificateSubject {
	return cryptoutilCAServer.CertificateSubject{
		CommonName:         &cert.Subject.CommonName,
		Organization:       &cert.Subject.Organization,
		OrganizationalUnit: &cert.Subject.OrganizationalUnit,
		Country:            &cert.Subject.Country,
		State:              &cert.Subject.Province,
		Locality:           &cert.Subject.Locality,
		DNSNames:           &cert.DNSNames,
		IPAddresses:        h.ptrSlice(h.ipsToStrings(cert.IPAddresses)),
		EmailAddresses:     &cert.EmailAddresses,
	}
}

// ptrSlice returns a pointer to a slice.
func (h *Handler) ptrSlice(s []string) *[]string {
	return &s
}

// mapCategory maps category string to API enum.
func (h *Handler) mapCategory(category string) cryptoutilCAServer.ProfileSummaryCategory {
	categoryMap := map[string]cryptoutilCAServer.ProfileSummaryCategory{
		"tls":              cryptoutilCAServer.TLS,
		"email":            cryptoutilCAServer.Email,
		"code_signing":     cryptoutilCAServer.CodeSigning,
		"document_signing": cryptoutilCAServer.DocumentSigning,
		"ca":               cryptoutilCAServer.CA,
	}

	if cat, ok := categoryMap[category]; ok {
		return cat
	}

	return cryptoutilCAServer.Other
}

// buildProfileResponse constructs a profile response.
func (h *Handler) buildProfileResponse(profile *ProfileConfig) *cryptoutilCAServer.ProfileResponse {
	category := profile.Category

	keyUsage := h.mapKeyUsage(profile.CertificateProfile)
	extKeyUsage := h.mapExtKeyUsage(profile.CertificateProfile)

	var maxValidityDays *int

	if profile.CertificateProfile != nil && profile.CertificateProfile.Validity.MaxDuration != "" {
		duration, err := time.ParseDuration(profile.CertificateProfile.Validity.MaxDuration)
		if err == nil {
			days := int(duration.Hours() / hoursPerDay)
			maxValidityDays = &days
		}
	}

	return &cryptoutilCAServer.ProfileResponse{
		ID:                  profile.ID,
		Name:                profile.Name,
		Description:         &profile.Description,
		Category:            &category,
		KeyUsage:            &keyUsage,
		ExtendedKeyUsage:    &extKeyUsage,
		MaxValidityDays:     maxValidityDays,
		SubjectRequirements: h.mapSubjectRequirements(profile.SubjectProfile),
		SANRequirements:     h.mapSANRequirements(profile.SubjectProfile),
	}
}

// mapKeyUsage maps certificate profile key usage to strings.
func (h *Handler) mapKeyUsage(profile *cryptoutilCAProfileCertificate.Profile) []string {
	if profile == nil {
		return nil
	}

	var usages []string
	if profile.KeyUsage.DigitalSignature {
		usages = append(usages, "digitalSignature")
	}

	if profile.KeyUsage.ContentCommitment {
		usages = append(usages, "contentCommitment")
	}

	if profile.KeyUsage.KeyEncipherment {
		usages = append(usages, "keyEncipherment")
	}

	if profile.KeyUsage.DataEncipherment {
		usages = append(usages, "dataEncipherment")
	}

	if profile.KeyUsage.KeyAgreement {
		usages = append(usages, "keyAgreement")
	}

	if profile.KeyUsage.CertSign {
		usages = append(usages, "keyCertSign")
	}

	if profile.KeyUsage.CRLSign {
		usages = append(usages, "cRLSign")
	}

	return usages
}

// mapExtKeyUsage maps certificate profile extended key usage to strings.
func (h *Handler) mapExtKeyUsage(profile *cryptoutilCAProfileCertificate.Profile) []string {
	if profile == nil {
		return nil
	}

	var usages []string
	if profile.ExtendedKeyUsage.ServerAuth {
		usages = append(usages, "serverAuth")
	}

	if profile.ExtendedKeyUsage.ClientAuth {
		usages = append(usages, "clientAuth")
	}

	if profile.ExtendedKeyUsage.CodeSigning {
		usages = append(usages, "codeSigning")
	}

	if profile.ExtendedKeyUsage.EmailProtection {
		usages = append(usages, "emailProtection")
	}

	if profile.ExtendedKeyUsage.TimeStamping {
		usages = append(usages, "timeStamping")
	}

	if profile.ExtendedKeyUsage.OCSPSigning {
		usages = append(usages, "ocspSigning")
	}

	return usages
}

// mapSubjectRequirements maps subject profile to API requirements.
func (h *Handler) mapSubjectRequirements(profile *cryptoutilCAProfileSubject.Profile) *cryptoutilCAServer.SubjectRequirements {
	if profile == nil {
		return nil
	}

	return &cryptoutilCAServer.SubjectRequirements{
		RequireCommonName:   &profile.Constraints.RequireCommonName,
		RequireOrganization: &profile.Constraints.RequireOrganization,
		RequireCountry:      &profile.Constraints.RequireCountry,
		AllowedCountries:    &profile.Constraints.ValidCountries,
	}
}

// mapSANRequirements maps subject profile SAN config to API requirements.
func (h *Handler) mapSANRequirements(profile *cryptoutilCAProfileSubject.Profile) *cryptoutilCAServer.SANRequirements {
	if profile == nil {
		return nil
	}

	return &cryptoutilCAServer.SANRequirements{
		DNSNamesAllowed:       &profile.SubjectAltNames.DNSNames.Allowed,
		DNSNamesRequired:      &profile.SubjectAltNames.DNSNames.Required,
		IPAddressesAllowed:    &profile.SubjectAltNames.IPAddresses.Allowed,
		EmailAddressesAllowed: &profile.SubjectAltNames.EmailAddresses.Allowed,
		UrisAllowed:           &profile.SubjectAltNames.URIs.Allowed,
	}
}

// extractCommonName extracts the common name from a distinguished name string.
func extractCommonName(dn string) string {
	// Simple extraction - look for CN= prefix.
	const cnPrefix = "CN="

	start := 0

	for i := 0; i < len(dn); i++ {
		if i+len(cnPrefix) <= len(dn) && dn[i:i+len(cnPrefix)] == cnPrefix {
			start = i + len(cnPrefix)

			break
		}
	}

	if start == 0 {
		return dn
	}

	end := len(dn)

	for i := start; i < len(dn); i++ {
		if dn[i] == ',' {
			end = i

			break
		}
	}

	return dn[start:end]
}

// buildCertificateSubject builds a CertificateSubject pointer from a DN string.
func buildCertificateSubject(dn string) *cryptoutilCAServer.CertificateSubject {
	cn := extractCommonName(dn)

	return &cryptoutilCAServer.CertificateSubject{
		CommonName: &cn,
	}
}

// buildCertificateSubjectValue builds a CertificateSubject value from a DN string.
func buildCertificateSubjectValue(dn string) cryptoutilCAServer.CertificateSubject {
	cn := extractCommonName(dn)

	return cryptoutilCAServer.CertificateSubject{
		CommonName: &cn,
	}
}

// errorResponse sends an error response.
func (h *Handler) errorResponse(c *fiber.Ctx, status int, errorCode, message string) error {
	if err := c.Status(status).JSON(cryptoutilCAServer.ErrorResponse{
		Error:   errorCode,
		Message: &message,
	}); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	return nil
}

// Verify Handler implements ServerInterface.
var _ cryptoutilCAServer.ServerInterface = (*Handler)(nil)

// Constants.
const (
	defaultValidityDays   = 365
	hoursPerDay           = 24
	maxTrackedEnrollments = 1000
	pemTypeCertificateReq = "CERTIFICATE REQUEST"
)
