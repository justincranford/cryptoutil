// Copyright (c) 2025-2026 Justin Cranford.
// Package handler provides HTTP handlers for CA REST API endpoints.
package handler

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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
	caType := cryptoutilApiCaServer.CASummaryTypeIntermediate

	// Check if this is a self-signed (root) CA.
	if caCert.Issuer.String() == caCert.Subject.String() {
		caType = cryptoutilApiCaServer.CASummaryTypeRoot
	}

	validUntil := caCert.NotAfter
	summary := cryptoutilApiCaServer.CASummary{
		ID:         caConfig.Name,
		Name:       caConfig.Name,
		Type:       caType,
		Status:     cryptoutilApiCaServer.CASummaryStatusActive,
		SubjectCN:  &caCert.Subject.CommonName,
		ValidUntil: &validUntil,
	}

	response := cryptoutilApiCaServer.CAListResponse{
		Authorities: []cryptoutilApiCaServer.CASummary{summary},
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
	caType := cryptoutilApiCaServer.CAResponseTypeIntermediate

	// Check if this is a self-signed (root) CA.
	if caCert.Issuer.String() == caCert.Subject.String() {
		caType = cryptoutilApiCaServer.CAResponseTypeRoot
	}

	// Encode certificate to PEM.
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: caCert.Raw,
	})

	notBefore := caCert.NotBefore
	notAfter := caCert.NotAfter
	serialNumber := fmt.Sprintf("%X", caCert.SerialNumber)

	// Determine key algorithm and size.
	keyAlgo, keySize := getKeyInfo(caCert)

	response := cryptoutilApiCaServer.CAResponse{
		ID:                    caConfig.Name,
		Name:                  caConfig.Name,
		Type:                  caType,
		Status:                cryptoutilApiCaServer.CAResponseStatusActive,
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
func (h *Handler) GetCRL(c *fiber.Ctx, caID string, params cryptoutilApiCaServer.GetCRLParams) error {
	// Check if CRL service is configured.
	h.mu.RLock()
	crlService := h.crlService
	h.mu.RUnlock()

	if crlService == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "service_unavailable", "CRL service not configured")
	}

	// Verify the CA ID matches the configured issuer.
	if h.issuer == nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuer_not_configured", "CA issuer not configured")
	}

	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil || caConfig.Name != caID {
		return h.errorResponse(c, fiber.StatusNotFound, "not_found", "CA not found")
	}

	storedRevocations, err := h.storage.GetRevoked(c.UserContext(), caConfig.Certificate.Subject.String())
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", err.Error())
	}

	revokedCertificates := make([]cryptoutilCAServiceRevocation.RevokedCertificate, 0, len(storedRevocations))
	for _, storedRevocation := range storedRevocations {
		if storedRevocation.RevocationTime == nil || storedRevocation.RevocationReason == nil {
			return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", "revoked certificate is missing revocation metadata")
		}

		serialNumber, ok := new(big.Int).SetString(storedRevocation.SerialNumber, cryptoutilSharedMagic.HexBase)
		if !ok {
			return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", "revoked certificate has an invalid serial number")
		}

		revokedCertificates = append(revokedCertificates, cryptoutilCAServiceRevocation.RevokedCertificate{
			SerialNumber:   serialNumber,
			RevocationTime: *storedRevocation.RevocationTime,
			Reason:         cryptoutilCAServiceRevocation.RevocationReason(*storedRevocation.RevocationReason),
		})
	}

	// Determine output format (default to DER).
	format := "der"
	if params.Format != nil {
		format = string(*params.Format)
	}

	// Generate the CRL based on format.
	switch format {
	case "pem":
		crlPEM, err := crlService.GenerateCRLPEMWithRevokedCertificates(revokedCertificates)
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
		crlDER, err := crlService.GenerateCRLWithRevokedCertificates(revokedCertificates)
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
		return cryptoutilSharedMagic.ECDSA, pub.Curve.Params().BitSize
	case *rsa.PublicKey:
		return cryptoutilSharedMagic.KeyTypeRSA, pub.N.BitLen()
	case ed25519.PublicKey:
		return cryptoutilSharedMagic.JoseAlgEdDSA, ed25519.PublicKeySize * cryptoutilSharedMagic.BitsPerByte
	default:
		return cryptoutilSharedMagic.UNKNOWN, 0
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

	profiles := make([]cryptoutilApiCaServer.ProfileSummary, 0, len(h.profiles))

	for _, p := range h.profiles {
		category := h.mapCategory(p.Category)
		profiles = append(profiles, cryptoutilApiCaServer.ProfileSummary{
			ID:          p.ID,
			Name:        p.Name,
			Description: &p.Description,
			Category:    &category,
		})
	}

	if err := c.JSON(cryptoutilApiCaServer.ProfileListResponse{Profiles: profiles}); err != nil {
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

// ESTCACerts handles GET /est/cacerts - RFC 7030 Section 4.1.
// Returns the CA certificates in PKCS#7 format for EST clients.
// Note: Full PKCS#7 degenerate format requires a CMS library.
// This implementation returns Base64-encoded PEM for compatibility.
func (h *Handler) ESTCACerts(c *fiber.Ctx) error {
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
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
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

// ESTCSRAttrs handles GET /est/csrattrs - RFC 7030 Section 4.5.
// Returns the CSR attributes required or recommended by the CA.
func (h *Handler) ESTCSRAttrs(c *fiber.Ctx) error {
	// Most CAs don't require specific CSR attributes.
	// Return 204 No Content to indicate no attributes required.
	if err := c.SendStatus(fiber.StatusNoContent); err != nil {
		return fmt.Errorf("failed to send no content status: %w", err)
	}

	return nil
}

// ESTSimpleEnroll handles POST /est/simpleenroll - RFC 7030 Section 4.2.
// Processes a PKCS#10 CSR and returns the issued certificate in PKCS#7 format.
// Note: Full mTLS authentication requires TLS middleware configuration.
// This implementation accepts CSR in DER or Base64 format.
func (h *Handler) ESTSimpleEnroll(c *fiber.Ctx) error {
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

	if h.issuer == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "service_unavailable", "CA issuer not configured")
	}
	// Build the issuance request using CSR subject.
	issueReq := h.buildESTIssueRequest(csr, profile)

	// Issue the certificate.
	issued, _, err := h.issuer.Issue(issueReq)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuance_error", err.Error())
	}

	if _, err := h.storeIssuedCertificate(c.UserContext(), issued, profile.ID, googleUuid.Nil); err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "storage_error", err.Error())
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
	if block != nil && block.Type == cryptoutilSharedMagic.StringPEMTypeCSR {
		csr, err = x509.ParseCertificateRequest(block.Bytes)
		if err == nil {
			return csr, nil
		}
	}

	return nil, fmt.Errorf("failed to parse CSR: invalid format")
}

// buildESTIssueRequest constructs an issuance request for EST enrollment.
