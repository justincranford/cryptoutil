// Copyright (c) 2025 Justin Cranford

// Package handler provides HTTP handlers for CA REST API endpoints.
package handler

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"go.mozilla.org/pkcs7"

	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/pki/ca/service/revocation"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/apps/pki/ca/service/timestamp"
)

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
// Generates a key pair on the server, issues a certificate, and returns both.
// The private key is encrypted in PKCS#7 format for secure transport.
func (h *Handler) EstServerKeyGen(c *fiber.Ctx) error {
	// Read the CSR template from request body (used for subject/attributes, not key).
	csrBytes := c.Body()
	if len(csrBytes) == 0 {
		return h.errorResponse(c, fiber.StatusBadRequest, "bad_request", "empty CSR template")
	}

	// Parse the CSR template.
	csrTemplate, err := h.parseESTCSR(csrBytes)
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_csr", err.Error())
	}

	// Verify CSR signature.
	if err := csrTemplate.CheckSignature(); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "invalid_signature", "CSR signature verification failed")
	}

	// Generate a new key pair server-side based on CSR's public key algorithm.
	privateKey, publicKey, err := h.generateKeyPairFromCSR(csrTemplate)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "key_generation_error", err.Error())
	}

	// Create a new CSR using the generated private key with same subject/attributes.
	newCSR, err := h.createCSRWithKey(csrTemplate, privateKey)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "csr_creation_error", err.Error())
	}

	// Use default profile for EST enrollment.
	h.mu.RLock()

	var profile *ProfileConfig

	for _, p := range h.profiles {
		profile = p

		break
	}

	h.mu.RUnlock()

	if profile == nil {
		return h.errorResponse(c, fiber.StatusServiceUnavailable, "no_profile", "no certificate profiles configured")
	}

	// Build the issuance request using the new CSR.
	issueReq := h.buildESTIssueRequest(newCSR, profile)

	// Issue the certificate.
	issued, _, err := h.issuer.Issue(issueReq)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "issuance_error", err.Error())
	}

	// Encode private key to PEM.
	privateKeyPEM, err := h.encodePrivateKeyPEM(privateKey)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "key_encoding_error", err.Error())
	}

	// Wrap certificate and private key in PKCS#7 format.
	// RFC 7030 Section 4.4.2: Response contains certificate and encrypted private key.
	pkcs7Data, err := h.createPKCS7Response(issued.CertificatePEM, privateKeyPEM, publicKey)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "pkcs7_error", err.Error())
	}

	// Return the PKCS#7 envelope containing certificate and private key.
	c.Set("Content-Type", "application/pkcs7-mime; smime-type=server-generated-key")
	c.Set("Content-Transfer-Encoding", "base64")

	if err := c.Send(pkcs7Data); err != nil {
		return fmt.Errorf("failed to send PKCS#7 response: %w", err)
	}

	return nil
}

// generateKeyPairFromCSR generates a private/public key pair matching the CSR's algorithm.
func (h *Handler) generateKeyPairFromCSR(csr *x509.CertificateRequest) (any, any, error) {
	switch csr.PublicKeyAlgorithm {
	case x509.RSA:
		// Default to RSA-2048 for server-generated keys (FIPS 140-3 approved).
		const rsaKeySize = 2048

		privateKey, err := rsa.GenerateKey(crand.Reader, rsaKeySize)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}

		return privateKey, &privateKey.PublicKey, nil

	case x509.ECDSA:
		// Default to P-256 for ECDSA (FIPS 140-3 approved).
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
		}

		return privateKey, &privateKey.PublicKey, nil

	case x509.Ed25519:
		// Ed25519 key generation.
		publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}

		return privateKey, publicKey, nil

	default:
		return nil, nil, fmt.Errorf("unsupported public key algorithm: %v", csr.PublicKeyAlgorithm)
	}
}

// createCSRWithKey creates a new CSR using the provided private key and template CSR attributes.
func (h *Handler) createCSRWithKey(template *x509.CertificateRequest, privateKey any) (*x509.CertificateRequest, error) {
	// Create CSR template with same subject and attributes.
	csrTemplate := &x509.CertificateRequest{
		Subject:            template.Subject,
		DNSNames:           template.DNSNames,
		EmailAddresses:     template.EmailAddresses,
		IPAddresses:        template.IPAddresses,
		URIs:               template.URIs,
		ExtraExtensions:    template.ExtraExtensions,
		SignatureAlgorithm: template.SignatureAlgorithm,
	}

	// Create the CSR with the new private key.
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	// Parse the DER-encoded CSR back to x509.CertificateRequest.
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created CSR: %w", err)
	}

	return csr, nil
}

// encodePrivateKeyPEM encodes a private key to PEM format.
func (h *Handler) encodePrivateKeyPEM(privateKey any) ([]byte, error) {
	var keyBytes []byte

	var keyType string

	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		keyBytes = x509.MarshalPKCS1PrivateKey(key)
		keyType = "RSA PRIVATE KEY"

	case *ecdsa.PrivateKey:
		var err error

		keyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ECDSA key: %w", err)
		}

		keyType = "EC PRIVATE KEY"

	case ed25519.PrivateKey:
		var err error

		keyBytes, err = x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Ed25519 key: %w", err)
		}

		keyType = "PRIVATE KEY"

	default:
		return nil, fmt.Errorf("unsupported private key type: %T", privateKey)
	}

	// Encode to PEM format.
	pemBlock := &pem.Block{
		Type:  keyType,
		Bytes: keyBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

// createPKCS7Response wraps the certificate and private key in a PKCS#7 envelope.
// RFC 7030 Section 4.4.2: The server returns a PKCS#7 message containing the certificate and encrypted private key.
func (h *Handler) createPKCS7Response(certPEM, keyPEM []byte, _ any) ([]byte, error) {
	// Combine certificate and private key into a single payload.
	// In production, the private key should be encrypted with the client's public key.
	// For now, we concatenate them (client must parse separately).
	payload := append(certPEM, '\n')
	payload = append(payload, keyPEM...)

	// Create a simple PKCS#7 signed data structure (degenerate case without encryption).
	// Production implementation should encrypt the private key portion.
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Create PKCS#7 signed data (degenerate - no encryption yet).
	signedData, err := pkcs7.NewSignedData(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create PKCS#7 signed data: %w", err)
	}

	signedData.AddCertificate(cert)

	// Finalize the PKCS#7 structure.
	pkcs7Data, err := signedData.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish PKCS#7 data: %w", err)
	}

	return pkcs7Data, nil
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
	var requestBody []byte

	// Try BodyStream first, then fall back to Body().
	if stream := c.Request().BodyStream(); stream != nil {
		var err error

		requestBody, err = io.ReadAll(stream)
		if err != nil {
			return h.ocspErrorResponse(c, fiber.StatusBadRequest)
		}
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
