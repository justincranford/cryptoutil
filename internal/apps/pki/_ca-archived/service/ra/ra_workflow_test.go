// Copyright (c) 2025 Justin Cranford

package ra

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestParseCSR_InvalidPEMBlockType tests parseCSR with wrong PEM block type.
func TestParseCSR_InvalidPEMBlockType(t *testing.T) {
	t.Parallel()

	// Create PEM block with wrong type.
	block := &pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: []byte("dummy")}
	data := pem.EncodeToMemory(block)

	_, err := parseCSR(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected PEM block type")
}

// TestRAService_ValidateDomains_Blocklist tests that blocklisted domains fail.
func TestRAService_ValidateDomains_Blocklist(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Validation.BlocklistedDomains = []string{"blocked.com"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	// Generate a valid CSR with a blocklisted domain.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "evil.blocked.com"}, []string{"evil.blocked.com"})
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Request should be pending with failed domain_blocklist validation.
	found := false

	for _, v := range req.ValidationResults {
		if v.CheckName == checkNameDomainBlocklist {
			require.False(t, v.Passed)
			require.Contains(t, v.Message, "blocklisted")

			found = true
		}
	}

	require.True(t, found, "expected domain_blocklist validation result")
}

// TestRAService_AutoApprove tests that requests on auto-approve profiles are auto-approved.
func TestRAService_AutoApprove(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Workflow.AutoApproveProfiles = []string{"tls-server"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "auto.example.com"}, []string{"auto.example.com"})
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-123")
	require.NoError(t, err)
	require.Equal(t, StatusApproved, req.Status)
}

// TestRAService_EscalateRequest tests the escalate action path.
func TestRAService_EscalateRequest_ActionCoverage(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "escalate.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	err = svc.EscalateRequest(context.Background(), req.RequestID, "reviewer-456", "needs senior review")
	require.NoError(t, err)

	// Status should still be pending after escalation.
	fetched, err := svc.GetRequest(context.Background(), req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusPending, fetched.Status)
	require.Len(t, fetched.ApprovalHistory, 1)
}

// TestRAService_CancelRequest_Authorization tests that only requester can cancel.
func TestRAService_CancelRequest_Authorization(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "cancel-auth.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-A")
	require.NoError(t, err)

	// User-B tries to cancel user-A's request.
	err = svc.CancelRequest(context.Background(), req.RequestID, "user-B", "unauthorized cancel")
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the original requester")
}

// TestRAService_CancelRequest_NonPending tests that non-pending requests cannot be cancelled.
func TestRAService_CancelRequest_NonPending(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Workflow.AutoApproveProfiles = []string{"tls-server"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "cancel-status.example.com"}, nil)
	require.NoError(t, err)

	// Auto-approve makes status = Approved.
	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-A")
	require.NoError(t, err)
	require.Equal(t, StatusApproved, req.Status)

	// Cannot cancel an approved request.
	err = svc.CancelRequest(context.Background(), req.RequestID, "user-A", "too late")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot cancel")
}

// TestRAService_MarkIssued_NonApproved tests that non-approved requests cannot be marked issued.
func TestRAService_MarkIssued_NonApproved(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "issue-status.example.com"}, nil)
	require.NoError(t, err)

	// Submit without auto-approve, status = Pending.
	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-A")
	require.NoError(t, err)
	require.Equal(t, StatusPending, req.Status)

	// Cannot mark a pending request as issued.
	err = svc.MarkIssued(context.Background(), req.RequestID, "cert-xyz")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot mark as issued")
}

// TestRAService_ListRequests_BeyondOffset tests ListRequests with offset beyond total.
func TestRAService_ListRequests_BeyondOffset(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	// Empty service: list with offset=100 returns empty slice, total=0.
	items, total, err := svc.ListRequests(context.Background(), nil, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.JoseJAMaxMaterials)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, items)
}

// TestRAService_RequestNotFound tests operations on non-existent request IDs.
func TestRAService_RequestNotFound(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	noSuchID, err := googleUuid.NewV7()
	require.NoError(t, err)

	// ApproveRequest with non-existent ID.
	err = svc.ApproveRequest(context.Background(), noSuchID, "approver", "ok")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// CancelRequest with non-existent ID.
	err = svc.CancelRequest(context.Background(), noSuchID, "user", "reason")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// MarkIssued with non-existent ID.
	err = svc.MarkIssued(context.Background(), noSuchID, "cert-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestRAService_ProcessAction_NonPendingStatus tests approving an already-rejected request.
func TestRAService_ProcessAction_NonPendingStatus(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "action-status.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-A")
	require.NoError(t, err)

	// Reject it first.
	err = svc.RejectRequest(context.Background(), req.RequestID, "approver-1", "not approved")
	require.NoError(t, err)

	// Now try to approve the already-rejected request.
	err = svc.ApproveRequest(context.Background(), req.RequestID, "approver-2", "too late")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot approve")
}

// TestRAService_ValidateDomains_PassCase tests that non-blocklisted domains pass.
func TestRAService_ValidateDomains_PassCase(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Validation.BlocklistedDomains = []string{"blocked.com"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	// Submit CSR with a domain NOT in the blocklist.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "safe.example.com"}, []string{"safe.example.com"})
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Find domain_blocklist result and assert Passed=true.
	found := false

	for _, v := range req.ValidationResults {
		if v.CheckName == checkNameDomainBlocklist {
			require.True(t, v.Passed)

			found = true
		}
	}

	require.True(t, found, "expected domain_blocklist validation result")
}

// TestRAService_AutoApprove_ValidationFails tests shouldAutoApprove returns false on failed validation.
func TestRAService_AutoApprove_ValidationFails(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Workflow.AutoApproveProfiles = []string{"tls-server"}
	cfg.Validation.BlocklistedDomains = []string{"blocked.com"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	// Blocklisted domain blocks auto-approve even for auto-approve profile.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "evil.blocked.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(context.Background(), csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Should still be Pending (not auto-approved) because domain validation failed.
	require.Equal(t, StatusPending, req.Status)
}

// TestRAService_CSRSignatureInvalid tests csr_signature validation failure path.
func TestRAService_CSRSignatureInvalid(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	// Generate a valid CSR then corrupt its signature bytes.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "tampered.example.com"}, nil)
	require.NoError(t, err)

	block, _ := pem.Decode(csrPEM)
	require.NotNil(t, block)

	// Flip the last 8 bytes (signature portion) to corrupt signature.
	der := make([]byte, len(block.Bytes))
	copy(der, block.Bytes)

	for i := len(der) - cryptoutilSharedMagic.IMMinPasswordLength; i < len(der); i++ {
		der[i] ^= 0xFF
	}

	tamperedPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCSR, Bytes: der})

	req, err := svc.SubmitRequest(context.Background(), tamperedPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Find csr_signature result and assert Passed=false.
	found := false

	for _, v := range req.ValidationResults {
		if v.CheckName == checkNameCSRSignature {
			require.False(t, v.Passed, "expected csr_signature to fail for tampered CSR")

			found = true
		}
	}

	require.True(t, found, "expected csr_signature validation result")
}
