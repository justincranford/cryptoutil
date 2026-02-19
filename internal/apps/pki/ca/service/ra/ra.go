// Copyright (c) 2025 Justin Cranford

// Package ra provides Registration Authority (RA) workflow services.
// The RA handles certificate request validation, approval workflows,
// and acts as an intermediary between end-entities and the CA.
package ra

import (
	"context"
	"errors"
	"fmt"
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
