// Copyright (c) 2025 Justin Cranford

package ra

import (
	"context"
	"crypto/x509/pkix"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewRAService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *RAConfig
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "custom config",
			config:  DefaultRAConfig(),
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewRAService(tc.config)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, svc)
		})
	}
}

func TestRAService_SubmitRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc, err := NewRAService(nil)
	require.NoError(t, err)

	// Generate test CSR.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, []string{"test.example.com"})
	require.NoError(t, err)

	tests := []struct {
		name        string
		csr         []byte
		profileID   string
		requesterID string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid request",
			csr:         csrPEM,
			profileID:   "tls-server",
			requesterID: "user-123",
			wantErr:     false,
		},
		{
			name:        "empty CSR",
			csr:         []byte{},
			profileID:   "tls-server",
			requesterID: "user-123",
			wantErr:     true,
			errContains: "CSR is required",
		},
		{
			name:        "empty profile ID",
			csr:         csrPEM,
			profileID:   "",
			requesterID: "user-123",
			wantErr:     true,
			errContains: "profile ID is required",
		},
		{
			name:        "empty requester ID",
			csr:         csrPEM,
			profileID:   "tls-server",
			requesterID: "",
			wantErr:     true,
			errContains: "requester ID is required",
		},
		{
			name:        "invalid CSR",
			csr:         []byte("not a valid CSR"),
			profileID:   "tls-server",
			requesterID: "user-123",
			wantErr:     true,
			errContains: "invalid CSR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := svc.SubmitRequest(ctx, tc.csr, tc.profileID, tc.requesterID)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, req)
			require.Equal(t, tc.profileID, req.ProfileID)
			require.Equal(t, tc.requesterID, req.RequesterID)
			require.Equal(t, StatusPending, req.Status)
			require.NotEmpty(t, req.CSRHash)
		})
	}
}

func TestRAService_GetRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc, err := NewRAService(nil)
	require.NoError(t, err)

	// Generate test CSR and submit request.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	tests := []struct {
		name      string
		requestID googleUuid.UUID
		wantErr   bool
	}{
		{
			name:      "existing request",
			requestID: req.RequestID,
			wantErr:   false,
		},
		{
			name:      "non-existent request",
			requestID: googleUuid.Must(googleUuid.NewV7()),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := svc.GetRequest(ctx, tc.requestID)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.requestID, result.RequestID)
		})
	}
}

func TestRAService_ListRequests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc, err := NewRAService(nil)
	require.NoError(t, err)

	// Generate test CSR and submit multiple requests.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err = svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
		require.NoError(t, err)
	}

	// List all requests.
	results, total, err := svc.ListRequests(ctx, nil, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 5, total)
	require.Len(t, results, 5)

	// List with pagination.
	results, total, err = svc.ListRequests(ctx, nil, 2, 0)
	require.NoError(t, err)
	require.Equal(t, 5, total)
	require.Len(t, results, 2)

	// Filter by status.
	pending := StatusPending
	results, total, err = svc.ListRequests(ctx, &pending, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 5, total)
	require.Len(t, results, 5)
}

func TestRAService_ApproveRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		setupConfig  func() *RAConfig
		approverID   string
		requesterID  string
		wantErr      bool
		errContains  string
		expectStatus RequestStatus
	}{
		{
			name: "successful approval",
			setupConfig: func() *RAConfig {
				cfg := DefaultRAConfig()
				cfg.Workflow.MinApprovers = 1

				return cfg
			},
			approverID:   "approver-1",
			requesterID:  "user-123",
			wantErr:      false,
			expectStatus: StatusApproved,
		},
		{
			name: "self-approval denied",
			setupConfig: func() *RAConfig {
				cfg := DefaultRAConfig()
				cfg.Workflow.AllowSelfApproval = false

				return cfg
			},
			approverID:  "user-123",
			requesterID: "user-123",
			wantErr:     true,
			errContains: "self-approval is not allowed",
		},
		{
			name: "self-approval allowed",
			setupConfig: func() *RAConfig {
				cfg := DefaultRAConfig()
				cfg.Workflow.AllowSelfApproval = true
				cfg.Workflow.MinApprovers = 1

				return cfg
			},
			approverID:   "user-123",
			requesterID:  "user-123",
			wantErr:      false,
			expectStatus: StatusApproved,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewRAService(tc.setupConfig())
			require.NoError(t, err)

			csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
			require.NoError(t, err)

			req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", tc.requesterID)
			require.NoError(t, err)

			err = svc.ApproveRequest(ctx, req.RequestID, tc.approverID, "Approved for testing")
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)

				return
			}

			require.NoError(t, err)

			// Verify status changed.
			updated, err := svc.GetRequest(ctx, req.RequestID)
			require.NoError(t, err)
			require.Equal(t, tc.expectStatus, updated.Status)
		})
	}
}

func TestRAService_RejectRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc, err := NewRAService(nil)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Reject without reason should fail.
	err = svc.RejectRequest(ctx, req.RequestID, "approver-1", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rejection reason is required")

	// Reject with reason should succeed.
	err = svc.RejectRequest(ctx, req.RequestID, "approver-1", "Policy violation")
	require.NoError(t, err)

	updated, err := svc.GetRequest(ctx, req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusRejected, updated.Status)
}

func TestRAService_CancelRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc, err := NewRAService(nil)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Cancel by different user should fail.
	err = svc.CancelRequest(ctx, req.RequestID, "other-user", "Changed my mind")
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the original requester can cancel")

	// Cancel by original requester should succeed.
	err = svc.CancelRequest(ctx, req.RequestID, "user-123", "Changed my mind")
	require.NoError(t, err)

	updated, err := svc.GetRequest(ctx, req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusCancelled, updated.Status)
}

const escalateAction = "escalate"

func TestRAService_EscalateRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name              string
		escalationEnabled bool
		wantErr           bool
		errContains       string
	}{
		{
			name:              "escalation enabled",
			escalationEnabled: true,
			wantErr:           false,
		},
		{
			name:              "escalation disabled",
			escalationEnabled: false,
			wantErr:           true,
			errContains:       "escalation is not enabled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := DefaultRAConfig()
			cfg.Workflow.EscalationEnabled = tc.escalationEnabled

			svc, err := NewRAService(cfg)
			require.NoError(t, err)

			csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
			require.NoError(t, err)

			req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
			require.NoError(t, err)

			err = svc.EscalateRequest(ctx, req.RequestID, "approver-1", "Needs senior review")
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)

				return
			}

			require.NoError(t, err)

			// Request should still be pending but with escalation in history.
			updated, err := svc.GetRequest(ctx, req.RequestID)
			require.NoError(t, err)
			require.Equal(t, StatusPending, updated.Status)
			require.NotEmpty(t, updated.ApprovalHistory)

			found := false

			for _, action := range updated.ApprovalHistory {
				if action.Action == escalateAction {
					found = true

					break
				}
			}

			require.True(t, found, "expected escalation action in history")
		})
	}
}

func TestRAService_MarkIssued(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultRAConfig()
	cfg.Workflow.MinApprovers = 1

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Cannot mark pending request as issued.
	err = svc.MarkIssued(ctx, req.RequestID, "cert-123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected approved")

	// Approve request first.
	err = svc.ApproveRequest(ctx, req.RequestID, "approver-1", "Approved")
	require.NoError(t, err)

	// Now can mark as issued.
	err = svc.MarkIssued(ctx, req.RequestID, "cert-123")
	require.NoError(t, err)

	updated, err := svc.GetRequest(ctx, req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusIssued, updated.Status)
	require.Equal(t, "cert-123", updated.IssuedCertID)
}

func TestRAService_CleanupExpired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultRAConfig()
	cfg.Workflow.RequestTTL = 1 * time.Millisecond // Very short TTL for testing.

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Wait for expiration.
	time.Sleep(5 * time.Millisecond)

	count, err := svc.CleanupExpired(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	updated, err := svc.GetRequest(ctx, req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusExpired, updated.Status)
}

func TestRAService_AutoApproval(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultRAConfig()
	cfg.Workflow.AutoApproveProfiles = []string{"auto-tls"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	// Request with auto-approve profile should be approved automatically.
	req, err := svc.SubmitRequest(ctx, csrPEM, "auto-tls", "user-123")
	require.NoError(t, err)
	require.Equal(t, StatusApproved, req.Status)
	require.NotEmpty(t, req.ApprovalHistory)
	require.Equal(t, "system", req.ApprovalHistory[0].ApproverID)

	// Request with non-auto-approve profile should remain pending.
	req2, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)
	require.Equal(t, StatusPending, req2.Status)
}

func TestRAService_MultipleApprovers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultRAConfig()
	cfg.Workflow.MinApprovers = 2

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// First approval - should still be pending.
	err = svc.ApproveRequest(ctx, req.RequestID, "approver-1", "First approval")
	require.NoError(t, err)

	updated, err := svc.GetRequest(ctx, req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusPending, updated.Status)

	// Second approval - should be approved.
	err = svc.ApproveRequest(ctx, req.RequestID, "approver-2", "Second approval")
	require.NoError(t, err)

	updated, err = svc.GetRequest(ctx, req.RequestID)
	require.NoError(t, err)
	require.Equal(t, StatusApproved, updated.Status)
}

func TestRAService_KeyStrengthValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultRAConfig()
	cfg.Validation.ValidateKeyStrength = true
	cfg.Validation.MinECKeySize = 256

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	// Valid key size.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.example.com"}, nil)
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Check validation results.
	var keyStrengthResult *ValidationResult

	for i := range req.ValidationResults {
		if req.ValidationResults[i].CheckName == "key_strength" {
			keyStrengthResult = &req.ValidationResults[i]

			break
		}
	}

	require.NotNil(t, keyStrengthResult)
	require.True(t, keyStrengthResult.Passed)
}

func TestRAService_DomainBlocklist(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultRAConfig()
	cfg.Validation.BlocklistedDomains = []string{"blocked.com"}

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	// Request with blocklisted domain.
	csrPEM, _, err := GenerateTestCSR(pkix.Name{CommonName: "test.blocked.com"}, []string{"test.blocked.com"})
	require.NoError(t, err)

	req, err := svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
	require.NoError(t, err)

	// Check validation results.
	var blocklistResult *ValidationResult

	for i := range req.ValidationResults {
		if req.ValidationResults[i].CheckName == "domain_blocklist" {
			blocklistResult = &req.ValidationResults[i]

			break
		}
	}

	require.NotNil(t, blocklistResult)
	require.False(t, blocklistResult.Passed)
	require.Contains(t, blocklistResult.Message, "blocklisted")
}

func TestRequestStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   RequestStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusApproved, "approved"},
		{StatusRejected, "rejected"},
		{StatusIssued, "issued"},
		{StatusExpired, "expired"},
		{StatusCancelled, "cancelled"},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestDefaultRAConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	require.NotNil(t, cfg)
	require.True(t, cfg.Workflow.RequireApproval)
	require.Equal(t, 1, cfg.Workflow.MinApprovers)
	require.True(t, cfg.Validation.ValidateKeyStrength)
	require.Equal(t, minRSAKeyBits, cfg.Validation.MinRSAKeySize)
}

func TestGenerateTestCSR(t *testing.T) {
	t.Parallel()

	subject := pkix.Name{
		CommonName:   "test.example.com",
		Organization: []string{"Test Org"},
	}
	dnsNames := []string{"test.example.com", "www.test.example.com"}

	csr, key, err := GenerateTestCSR(subject, dnsNames)
	require.NoError(t, err)
	require.NotEmpty(t, csr)
	require.NotNil(t, key)
}
