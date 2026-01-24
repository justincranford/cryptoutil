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
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"
)

// RequestStatus represents the current state of a certificate request.
type RequestStatus string

// Request status constants.
const (
	StatusPending   RequestStatus = "pending"
	StatusApproved  RequestStatus = "approved"
	StatusRejected  RequestStatus = "rejected"
	StatusIssued    RequestStatus = "issued"
	StatusExpired   RequestStatus = "expired"
	StatusCancelled RequestStatus = "cancelled"
)

// ValidationResult represents the outcome of a validation check.
type ValidationResult struct {
	CheckName string    `json:"check_name"`
	Passed    bool      `json:"passed"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ApprovalAction represents an approval workflow action.
type ApprovalAction struct {
	ActionID   googleUuid.UUID `json:"action_id"`
	RequestID  googleUuid.UUID `json:"request_id"`
	ApproverID string          `json:"approver_id"`
	Action     string          `json:"action"` // approve, reject, escalate, comment
	Comment    string          `json:"comment,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	IPAddress  string          `json:"ip_address,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
}

// CertificateRequest represents a certificate signing request in the RA workflow.
type CertificateRequest struct {
	RequestID         googleUuid.UUID    `json:"request_id"`
	CSR               []byte             `json:"csr"`
	CSRHash           string             `json:"csr_hash"`
	ProfileID         string             `json:"profile_id"`
	RequesterID       string             `json:"requester_id"`
	RequesterEmail    string             `json:"requester_email,omitempty"`
	OrganizationID    string             `json:"organization_id,omitempty"`
	Status            RequestStatus      `json:"status"`
	ValidationResults []ValidationResult `json:"validation_results"`
	ApprovalHistory   []ApprovalAction   `json:"approval_history"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	ExpiresAt         time.Time          `json:"expires_at"`
	IssuedCertID      string             `json:"issued_cert_id,omitempty"`
	Metadata          map[string]string  `json:"metadata,omitempty"`
}

// WorkflowConfig defines the approval workflow configuration.
type WorkflowConfig struct {
	RequireApproval      bool          `json:"require_approval"`
	MinApprovers         int           `json:"min_approvers"`
	AutoApproveProfiles  []string      `json:"auto_approve_profiles"`
	RequestTTL           time.Duration `json:"request_ttl"`
	AllowSelfApproval    bool          `json:"allow_self_approval"`
	EscalationEnabled    bool          `json:"escalation_enabled"`
	EscalationTimeout    time.Duration `json:"escalation_timeout"`
	NotificationsEnabled bool          `json:"notifications_enabled"`
}

// ValidationConfig defines validation check configuration.
type ValidationConfig struct {
	ValidateDomainOwnership bool     `json:"validate_domain_ownership"`
	ValidateEmailOwnership  bool     `json:"validate_email_ownership"`
	ValidateOrganization    bool     `json:"validate_organization"`
	ValidateKeyStrength     bool     `json:"validate_key_strength"`
	AllowedKeyAlgorithms    []string `json:"allowed_key_algorithms"`
	MinRSAKeySize           int      `json:"min_rsa_key_size"`
	MinECKeySize            int      `json:"min_ec_key_size"`
	BlocklistedDomains      []string `json:"blocklisted_domains"`
	RequiredDNSRecords      bool     `json:"required_dns_records"`
}

// RAConfig holds the complete RA configuration.
type RAConfig struct {
	Workflow   WorkflowConfig   `json:"workflow"`
	Validation ValidationConfig `json:"validation"`
}

// DefaultRAConfig returns sensible default RA configuration.
func DefaultRAConfig() *RAConfig {
	return &RAConfig{
		Workflow: WorkflowConfig{
			RequireApproval:      true,
			MinApprovers:         1,
			AutoApproveProfiles:  []string{},
			RequestTTL:           sevenDays,
			AllowSelfApproval:    false,
			EscalationEnabled:    true,
			EscalationTimeout:    oneDay,
			NotificationsEnabled: true,
		},
		Validation: ValidationConfig{
			ValidateDomainOwnership: true,
			ValidateEmailOwnership:  true,
			ValidateOrganization:    false,
			ValidateKeyStrength:     true,
			AllowedKeyAlgorithms:    []string{"RSA", "ECDSA", "Ed25519"},
			MinRSAKeySize:           minRSAKeyBits,
			MinECKeySize:            minECKeyBits,
			BlocklistedDomains:      []string{},
			RequiredDNSRecords:      false,
		},
	}
}

// Time duration constants.
const (
	sevenDays = 7 * 24 * time.Hour
	oneDay    = 24 * time.Hour
)

// Key size constants.
const (
	minRSAKeyBits = 2048
	minECKeyBits  = 256
)

// Action constants.
const (
	actionApprove  = "approve"
	actionReject   = "reject"
	actionEscalate = "escalate"
	actionCancel   = "cancel"
)

// RAService provides Registration Authority workflow services.
type RAService struct {
	config   *RAConfig
	requests map[googleUuid.UUID]*CertificateRequest
	mu       sync.RWMutex
}

// NewRAService creates a new RA service instance.
func NewRAService(config *RAConfig) (*RAService, error) {
	if config == nil {
		config = DefaultRAConfig()
	}

	return &RAService{
		config:   config,
		requests: make(map[googleUuid.UUID]*CertificateRequest),
	}, nil
}

// SubmitRequest creates a new certificate request.
func (s *RAService) SubmitRequest(ctx context.Context, csr []byte, profileID, requesterID string) (*CertificateRequest, error) {
	if len(csr) == 0 {
		return nil, errors.New("CSR is required")
	}

	if profileID == "" {
		return nil, errors.New("profile ID is required")
	}

	if requesterID == "" {
		return nil, errors.New("requester ID is required")
	}

	// Parse and validate CSR.
	parsedCSR, err := parseCSR(csr)
	if err != nil {
		return nil, fmt.Errorf("invalid CSR: %w", err)
	}

	// Generate request ID.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate request ID: %w", err)
	}

	now := time.Now().UTC()

	// Create request.
	request := &CertificateRequest{
		RequestID:         requestID,
		CSR:               csr,
		CSRHash:           computeCSRHash(csr),
		ProfileID:         profileID,
		RequesterID:       requesterID,
		Status:            StatusPending,
		ValidationResults: []ValidationResult{},
		ApprovalHistory:   []ApprovalAction{},
		CreatedAt:         now,
		UpdatedAt:         now,
		ExpiresAt:         now.Add(s.config.Workflow.RequestTTL),
		Metadata:          make(map[string]string),
	}

	// Run validation checks.
	validationResults := s.validateRequest(ctx, parsedCSR, profileID)
	request.ValidationResults = validationResults

	// Check if auto-approval applies.
	if s.shouldAutoApprove(profileID, validationResults) {
		request.Status = StatusApproved
		request.ApprovalHistory = append(request.ApprovalHistory, ApprovalAction{
			ActionID:   mustNewUUID(),
			RequestID:  requestID,
			ApproverID: "system",
			Action:     "approve",
			Comment:    "Auto-approved based on profile configuration",
			Timestamp:  now,
		})
	}

	// Store request.
	s.mu.Lock()
	s.requests[requestID] = request
	s.mu.Unlock()

	return request, nil
}

// GetRequest retrieves a certificate request by ID.
func (s *RAService) GetRequest(_ context.Context, requestID googleUuid.UUID) (*CertificateRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	request, exists := s.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("request not found: %s", requestID)
	}

	return request, nil
}

// ListRequests returns requests filtered by status.
func (s *RAService) ListRequests(_ context.Context, status *RequestStatus, limit, offset int) ([]*CertificateRequest, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*CertificateRequest

	for _, req := range s.requests {
		if status == nil || req.Status == *status {
			filtered = append(filtered, req)
		}
	}

	total := len(filtered)

	// Apply pagination.
	if offset >= len(filtered) {
		return []*CertificateRequest{}, total, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total, nil
}

// ApproveRequest approves a pending certificate request.
func (s *RAService) ApproveRequest(ctx context.Context, requestID googleUuid.UUID, approverID, comment string) error {
	return s.processAction(ctx, requestID, approverID, "approve", comment)
}

// RejectRequest rejects a pending certificate request.
func (s *RAService) RejectRequest(ctx context.Context, requestID googleUuid.UUID, approverID, reason string) error {
	if reason == "" {
		return errors.New("rejection reason is required")
	}

	return s.processAction(ctx, requestID, approverID, "reject", reason)
}

// EscalateRequest escalates a request to higher approval level.
func (s *RAService) EscalateRequest(ctx context.Context, requestID googleUuid.UUID, approverID, reason string) error {
	if !s.config.Workflow.EscalationEnabled {
		return errors.New("escalation is not enabled")
	}

	return s.processAction(ctx, requestID, approverID, "escalate", reason)
}

// CancelRequest cancels a pending request.
func (s *RAService) CancelRequest(_ context.Context, requestID googleUuid.UUID, requesterID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, exists := s.requests[requestID]
	if !exists {
		return fmt.Errorf("request not found: %s", requestID)
	}

	// Only the original requester can cancel.
	if request.RequesterID != requesterID {
		return errors.New("only the original requester can cancel the request")
	}

	if request.Status != StatusPending {
		return fmt.Errorf("cannot cancel request with status: %s", request.Status)
	}

	now := time.Now().UTC()
	request.Status = StatusCancelled
	request.UpdatedAt = now

	request.ApprovalHistory = append(request.ApprovalHistory, ApprovalAction{
		ActionID:   mustNewUUID(),
		RequestID:  requestID,
		ApproverID: requesterID,
		Action:     "cancel",
		Comment:    reason,
		Timestamp:  now,
	})

	return nil
}

// MarkIssued marks a request as issued after certificate generation.
func (s *RAService) MarkIssued(_ context.Context, requestID googleUuid.UUID, certID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, exists := s.requests[requestID]
	if !exists {
		return fmt.Errorf("request not found: %s", requestID)
	}

	if request.Status != StatusApproved {
		return fmt.Errorf("cannot mark as issued: request status is %s, expected approved", request.Status)
	}

	request.Status = StatusIssued
	request.IssuedCertID = certID
	request.UpdatedAt = time.Now().UTC()

	return nil
}

// CleanupExpired removes expired pending requests.
func (s *RAService) CleanupExpired(_ context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	var count int

	for id, req := range s.requests {
		if req.Status == StatusPending && now.After(req.ExpiresAt) {
			req.Status = StatusExpired
			req.UpdatedAt = now
			s.requests[id] = req
			count++
		}
	}

	return count, nil
}

// processAction handles approval workflow actions.
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
			CheckName: "csr_signature",
			Passed:    false,
			Message:   fmt.Sprintf("Invalid CSR signature: %v", err),
			Timestamp: now,
		})
	} else {
		results = append(results, ValidationResult{
			CheckName: "csr_signature",
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
				CheckName: "key_strength",
				Passed:    false,
				Message:   fmt.Sprintf("RSA key size %d bits is below minimum %d bits", bits, s.config.Validation.MinRSAKeySize),
				Timestamp: timestamp,
			}
		}

		return ValidationResult{
			CheckName: "key_strength",
			Passed:    true,
			Message:   fmt.Sprintf("RSA key size %d bits meets requirements", bits),
			Timestamp: timestamp,
		}

	case *ecdsa.PublicKey:
		bits := key.Curve.Params().BitSize
		if bits < s.config.Validation.MinECKeySize {
			return ValidationResult{
				CheckName: "key_strength",
				Passed:    false,
				Message:   fmt.Sprintf("EC key size %d bits is below minimum %d bits", bits, s.config.Validation.MinECKeySize),
				Timestamp: timestamp,
			}
		}

		return ValidationResult{
			CheckName: "key_strength",
			Passed:    true,
			Message:   fmt.Sprintf("EC key size %d bits meets requirements", bits),
			Timestamp: timestamp,
		}

	case ed25519.PublicKey:
		return ValidationResult{
			CheckName: "key_strength",
			Passed:    true,
			Message:   "Ed25519 key meets requirements",
			Timestamp: timestamp,
		}

	default:
		return ValidationResult{
			CheckName: "key_strength",
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
					CheckName: "domain_blocklist",
					Passed:    false,
					Message:   fmt.Sprintf("Domain %s is blocklisted", domain),
					Timestamp: timestamp,
				}
			}
		}
	}

	return ValidationResult{
		CheckName: "domain_blocklist",
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
		if block.Type != "CERTIFICATE REQUEST" && block.Type != "NEW CERTIFICATE REQUEST" {
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
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	return csrPEM, key, nil
}
