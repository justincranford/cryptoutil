// Copyright (c) 2025 Justin Cranford

package handler

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	cryptoutilCAProfileCertificate "cryptoutil/internal/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/ca/service/issuer"
)

// Handler implements the CA enrollment ServerInterface.
type Handler struct {
	issuer   *cryptoutilCAServiceIssuer.Issuer
	profiles map[string]*ProfileConfig
	mu       sync.RWMutex
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
func NewHandler(issuer *cryptoutilCAServiceIssuer.Issuer, profiles map[string]*ProfileConfig) (*Handler, error) {
	if issuer == nil {
		return nil, fmt.Errorf("issuer is required")
	}

	if profiles == nil {
		profiles = make(map[string]*ProfileConfig)
	}

	return &Handler{
		issuer:   issuer,
		profiles: profiles,
	}, nil
}

// ListCertificates handles GET /certificates.
func (h *Handler) ListCertificates(_ *fiber.Ctx, _ cryptoutilCAServer.ListCertificatesParams) error {
	// TODO: Implement certificate listing with storage backend.
	return fiber.NewError(fiber.StatusNotImplemented, "certificate listing not yet implemented")
}

// GetCertificate handles GET /certificates/{serialNumber}.
func (h *Handler) GetCertificate(_ *fiber.Ctx, _ string) error {
	// TODO: Implement certificate retrieval with storage backend.
	return fiber.NewError(fiber.StatusNotImplemented, "certificate retrieval not yet implemented")
}

// GetCertificateChain handles GET /certificates/{serialNumber}/chain.
func (h *Handler) GetCertificateChain(_ *fiber.Ctx, _ string) error {
	// TODO: Implement chain retrieval with storage backend.
	return fiber.NewError(fiber.StatusNotImplemented, "chain retrieval not yet implemented")
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

	if err := c.Status(fiber.StatusCreated).JSON(resp); err != nil {
		return fmt.Errorf("failed to send enrollment response: %w", err)
	}

	return nil
}

// GetEnrollmentStatus handles GET /enroll/{requestId}.
func (h *Handler) GetEnrollmentStatus(_ *fiber.Ctx, _ uuid.UUID) error {
	// TODO: Implement enrollment status tracking with storage backend.
	return fiber.NewError(fiber.StatusNotImplemented, "enrollment status not yet implemented")
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

// parseCSR parses a PEM-encoded CSR.
func (h *Handler) parseCSR(csrPEM string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("expected CERTIFICATE REQUEST, got %s", block.Type)
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
	defaultValidityDays = 365
	hoursPerDay         = 24
)
