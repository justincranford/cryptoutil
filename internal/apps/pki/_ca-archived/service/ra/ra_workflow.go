// Copyright (c) 2025 Justin Cranford

// Package ra provides Registration Authority (RA) workflow services.
// The RA handles certificate request validation, approval workflows,
// and acts as an intermediary between end-entities and the CA.
package ra

import (
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"
)

// RequestStatus represents the current state of a certificate request.

// Check name constants for validation results.
const (
	checkNameCSRSignature    = "csr_signature"
	checkNameKeyStrength     = "key_strength"
	checkNameDomainBlocklist = "domain_blocklist"
)

func (s *RAService) processAction(_ context.Context, requestID googleUuid.UUID, approverID, action, comment string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, exists := s.requests[requestID]
	if !exists {
		return fmt.Errorf("request not found: %s", requestID)
	}

	if request.Status != StatusPending {
		return fmt.Errorf("cannot %s request with status: %s", action, request.Status)
	}

	// Check self-approval.
	if !s.config.Workflow.AllowSelfApproval && request.RequesterID == approverID && action == actionApprove {
		return errors.New("self-approval is not allowed")
	}

	now := time.Now().UTC()

	// Record action.
	request.ApprovalHistory = append(request.ApprovalHistory, ApprovalAction{
		ActionID:   mustNewUUID(),
		RequestID:  requestID,
		ApproverID: approverID,
		Action:     action,
		Comment:    comment,
		Timestamp:  now,
	})

	// Update status based on action.
	switch action {
	case actionApprove:
		approvalCount := countApprovals(request.ApprovalHistory)
		if approvalCount >= s.config.Workflow.MinApprovers {
			request.Status = StatusApproved
		}
	case actionReject:
		request.Status = StatusRejected
	case actionEscalate:
		// Keep pending but log escalation.
	}

	request.UpdatedAt = now

	return nil
}

// validateRequest runs validation checks on a CSR.
func (s *RAService) validateRequest(_ context.Context, csr *x509.CertificateRequest, _ string) []ValidationResult {
	var results []ValidationResult

	now := time.Now().UTC()

	// Validate CSR signature.
	if err := csr.CheckSignature(); err != nil {
		results = append(results, ValidationResult{
			CheckName: checkNameCSRSignature,
			Passed:    false,
			Message:   fmt.Sprintf("Invalid CSR signature: %v", err),
			Timestamp: now,
		})
	} else {
		results = append(results, ValidationResult{
			CheckName: checkNameCSRSignature,
			Passed:    true,
			Message:   "CSR signature is valid",
			Timestamp: now,
		})
	}

	// Validate key strength.
	if s.config.Validation.ValidateKeyStrength {
		keyResult := s.validateKeyStrength(csr.PublicKey, now)
		results = append(results, keyResult)
	}

	// Validate domains against blocklist.
	if len(s.config.Validation.BlocklistedDomains) > 0 {
		domainResult := s.validateDomains(csr, now)
		results = append(results, domainResult)
	}

	return results
}

// validateKeyStrength checks if the public key meets minimum requirements.
func (s *RAService) validateKeyStrength(pubKey any, timestamp time.Time) ValidationResult {
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		bits := key.N.BitLen()
		if bits < s.config.Validation.MinRSAKeySize {
			return ValidationResult{
				CheckName: checkNameKeyStrength,
				Passed:    false,
				Message:   fmt.Sprintf("RSA key size %d bits is below minimum %d bits", bits, s.config.Validation.MinRSAKeySize),
				Timestamp: timestamp,
			}
		}

		return ValidationResult{
			CheckName: checkNameKeyStrength,
			Passed:    true,
			Message:   fmt.Sprintf("RSA key size %d bits meets requirements", bits),
			Timestamp: timestamp,
		}

	case *ecdsa.PublicKey:
		bits := key.Curve.Params().BitSize
		if bits < s.config.Validation.MinECKeySize {
			return ValidationResult{
				CheckName: checkNameKeyStrength,
				Passed:    false,
				Message:   fmt.Sprintf("EC key size %d bits is below minimum %d bits", bits, s.config.Validation.MinECKeySize),
				Timestamp: timestamp,
			}
		}

		return ValidationResult{
			CheckName: checkNameKeyStrength,
			Passed:    true,
			Message:   fmt.Sprintf("EC key size %d bits meets requirements", bits),
			Timestamp: timestamp,
		}

	case ed25519.PublicKey:
		return ValidationResult{
			CheckName: checkNameKeyStrength,
			Passed:    true,
			Message:   "Ed25519 key meets requirements",
			Timestamp: timestamp,
		}

	default:
		return ValidationResult{
			CheckName: checkNameKeyStrength,
			Passed:    false,
			Message:   "Unknown key type",
			Timestamp: timestamp,
		}
	}
}

// validateDomains checks domains against blocklist.
func (s *RAService) validateDomains(csr *x509.CertificateRequest, timestamp time.Time) ValidationResult {
	// Collect all domains from CSR.
	var domains []string

	if csr.Subject.CommonName != "" {
		domains = append(domains, csr.Subject.CommonName)
	}

	domains = append(domains, csr.DNSNames...)

	// Check against blocklist.
	for _, domain := range domains {
		for _, blocked := range s.config.Validation.BlocklistedDomains {
			if strings.HasSuffix(domain, blocked) || domain == blocked {
				return ValidationResult{
					CheckName: checkNameDomainBlocklist,
					Passed:    false,
					Message:   fmt.Sprintf("Domain %s is blocklisted", domain),
					Timestamp: timestamp,
				}
			}
		}
	}

	return ValidationResult{
		CheckName: checkNameDomainBlocklist,
		Passed:    true,
		Message:   "All domains pass blocklist check",
		Timestamp: timestamp,
	}
}

// shouldAutoApprove determines if a request should be auto-approved.
func (s *RAService) shouldAutoApprove(profileID string, results []ValidationResult) bool {
	// Check if profile is in auto-approve list.
	autoApprove := false

	for _, p := range s.config.Workflow.AutoApproveProfiles {
		if p == profileID {
			autoApprove = true

			break
		}
	}

	if !autoApprove {
		return false
	}

	// All validations must pass.
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}

	return true
}

// countApprovals counts the number of approve actions.
func countApprovals(history []ApprovalAction) int {
	count := 0

	for _, action := range history {
		if action.Action == actionApprove {
			count++
		}
	}

	return count
}

// parseCSR parses a PEM or DER encoded CSR.
func parseCSR(data []byte) (*x509.CertificateRequest, error) {
	// Try PEM first.
	block, _ := pem.Decode(data)
	if block != nil {
		if block.Type != cryptoutilSharedMagic.StringPEMTypeCSR && block.Type != "NEW CERTIFICATE REQUEST" {
			return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
		}

		data = block.Bytes
	}

	csr, err := x509.ParseCertificateRequest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR: %w", err)
	}

	return csr, nil
}

// computeCSRHash computes a hash of the CSR for deduplication.
func computeCSRHash(csr []byte) string {
	// Use a simple hash for identification.
	hash := crypto.SHA256.New()
	hash.Write(csr)

	return fmt.Sprintf("%x", hash.Sum(nil))
}

// mustNewUUID generates a new UUIDv7 or panics.
func mustNewUUID() googleUuid.UUID {
	id, err := googleUuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}

	return id
}

// GenerateTestCSR creates a test CSR for development/testing purposes.
func GenerateTestCSR(subject pkix.Name, dnsNames []string) ([]byte, crypto.PrivateKey, error) {
	// Generate a test key using P-256 curve.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	template := &x509.CertificateRequest{
		Subject:  subject,
		DNSNames: dnsNames,
	}

	csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
		Bytes: csrDER,
	})

	return csrPEM, key, nil
}
