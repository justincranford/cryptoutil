// Copyright (c) 2025 Justin Cranford

// Package handler provides HTTP handlers for CA REST API endpoints.
package handler

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func (h *Handler) lookupCertificateBySerial(ctx context.Context, serialNumber *big.Int) *x509.Certificate {
	if serialNumber == nil {
		return nil
	}

	// Look up in storage using hex-encoded serial.
	serialHex := serialNumber.Text(cryptoutilSharedMagic.HexBase)

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
	_ *ProfileConfig,
	req *cryptoutilApiCaServer.EnrollmentRequest,
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

	validity := time.Duration(validityDays) * cryptoutilSharedMagic.HoursPerDay * time.Hour

	return &cryptoutilCAServiceIssuer.CertificateRequest{
		PublicKey:        csr.PublicKey,
		SubjectRequest:   subjectReq,
		ValidityDuration: validity,
	}, nil
}

// applySubjectOverrides applies subject field overrides from the request.
func (h *Handler) applySubjectOverrides(
	subjectReq *cryptoutilCAProfileSubject.Request,
	override *cryptoutilApiCaServer.SubjectOverride,
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
	override *cryptoutilApiCaServer.SANOverride,
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
func (h *Handler) buildEnrollmentResponse(issued *cryptoutilCAServiceIssuer.IssuedCertificate) *cryptoutilApiCaServer.EnrollmentResponse {
	certPEM := string(issued.CertificatePEM)
	chainPEM := string(issued.ChainPEM)

	notBefore := issued.Certificate.NotBefore
	notAfter := issued.Certificate.NotAfter
	serialNumber := issued.SerialNumber

	subject := h.certSubjectToAPI(issued.Certificate)

	return &cryptoutilApiCaServer.EnrollmentResponse{
		RequestID: googleUuid.New(),
		Status:    cryptoutilApiCaServer.Issued,
		Certificate: cryptoutilApiCaServer.IssuedCertificate{
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
func (h *Handler) certSubjectToAPI(cert *x509.Certificate) cryptoutilApiCaServer.CertificateSubject {
	return cryptoutilApiCaServer.CertificateSubject{
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
func (h *Handler) mapCategory(category string) cryptoutilApiCaServer.ProfileSummaryCategory {
	categoryMap := map[string]cryptoutilApiCaServer.ProfileSummaryCategory{
		"tls":              cryptoutilApiCaServer.TLS,
		cryptoutilSharedMagic.ClaimEmail:            cryptoutilApiCaServer.Email,
		"code_signing":     cryptoutilApiCaServer.CodeSigning,
		"document_signing": cryptoutilApiCaServer.DocumentSigning,
		"ca":               cryptoutilApiCaServer.CA,
	}

	if cat, ok := categoryMap[category]; ok {
		return cat
	}

	return cryptoutilApiCaServer.Other
}

// buildProfileResponse constructs a profile response.
func (h *Handler) buildProfileResponse(profile *ProfileConfig) *cryptoutilApiCaServer.ProfileResponse {
	category := profile.Category

	keyUsage := h.mapKeyUsage(profile.CertificateProfile)
	extKeyUsage := h.mapExtKeyUsage(profile.CertificateProfile)

	var maxValidityDays *int

	if profile.CertificateProfile != nil && profile.CertificateProfile.Validity.MaxDuration != "" {
		duration, err := time.ParseDuration(profile.CertificateProfile.Validity.MaxDuration)
		if err == nil {
			days := int(duration.Hours() / cryptoutilSharedMagic.HoursPerDay)
			maxValidityDays = &days
		}
	}

	return &cryptoutilApiCaServer.ProfileResponse{
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
func (h *Handler) mapSubjectRequirements(profile *cryptoutilCAProfileSubject.Profile) *cryptoutilApiCaServer.SubjectRequirements {
	if profile == nil {
		return nil
	}

	return &cryptoutilApiCaServer.SubjectRequirements{
		RequireCommonName:   &profile.Constraints.RequireCommonName,
		RequireOrganization: &profile.Constraints.RequireOrganization,
		RequireCountry:      &profile.Constraints.RequireCountry,
		AllowedCountries:    &profile.Constraints.ValidCountries,
	}
}

// mapSANRequirements maps subject profile SAN config to API requirements.
func (h *Handler) mapSANRequirements(profile *cryptoutilCAProfileSubject.Profile) *cryptoutilApiCaServer.SANRequirements {
	if profile == nil {
		return nil
	}

	return &cryptoutilApiCaServer.SANRequirements{
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
func buildCertificateSubject(dn string) *cryptoutilApiCaServer.CertificateSubject {
	cn := extractCommonName(dn)

	return &cryptoutilApiCaServer.CertificateSubject{
		CommonName: &cn,
	}
}

// buildCertificateSubjectValue builds a CertificateSubject value from a DN string.
func buildCertificateSubjectValue(dn string) cryptoutilApiCaServer.CertificateSubject {
	cn := extractCommonName(dn)

	return cryptoutilApiCaServer.CertificateSubject{
		CommonName: &cn,
	}
}

// errorResponse sends an error response.
func (h *Handler) errorResponse(c *fiber.Ctx, status int, errorCode, message string) error {
	if err := c.Status(status).JSON(cryptoutilApiCaServer.ErrorResponse{
		Error:   errorCode,
		Message: &message,
	}); err != nil {
		return fmt.Errorf("failed to send error response: %w", err)
	}

	return nil
}

// Verify Handler implements ServerInterface.
var _ cryptoutilApiCaServer.ServerInterface = (*Handler)(nil)

// Constants.
const (
	defaultValidityDays   = 365
	maxTrackedEnrollments = 1000
	pemTypeCertificateReq = "CERTIFICATE REQUEST"
)
