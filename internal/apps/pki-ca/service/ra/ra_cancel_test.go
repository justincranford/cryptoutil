// Copyright (c) 2025 Justin Cranford

package ra

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
	time.Sleep(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Millisecond)

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
	require.Equal(t, cryptoutilSharedMagic.SystemInitiatorName, req.ApprovalHistory[0].ApproverID)

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
	cfg.Validation.MinECKeySize = cryptoutilSharedMagic.MaxUnsealSharedSecrets

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

// TestRAService_ValidateKeyStrength_RSA tests RSA key strength validation.
func TestRAService_ValidateKeyStrength_RSA(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Validation.MinRSAKeySize = cryptoutilSharedMagic.DefaultMetricsBatchSize

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	timestamp := time.Now().UTC()

	// Valid RSA 2048-bit key.
	rsaKey2048, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	result := svc.validateKeyStrength(&rsaKey2048.PublicKey, timestamp)
	require.True(t, result.Passed)
	require.Equal(t, "key_strength", result.CheckName)

	// Invalid RSA 1024-bit key (below minimum).
	rsaKey1024, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultLogsBatchSize) //nolint:gosec // G403: intentionally weak RSA key to test minimum size rejection
	require.NoError(t, err)

	result = svc.validateKeyStrength(&rsaKey1024.PublicKey, timestamp)
	require.False(t, result.Passed)
	require.Contains(t, result.Message, "below minimum")
}

// TestRAService_ValidateKeyStrength_ECDSAInvalid tests ECDSA key below minimum size.
func TestRAService_ValidateKeyStrength_ECDSAInvalid(t *testing.T) {
	t.Parallel()

	cfg := DefaultRAConfig()
	cfg.Validation.MinECKeySize = cryptoutilSharedMagic.MaxUnsealSharedSecrets

	svc, err := NewRAService(cfg)
	require.NoError(t, err)

	timestamp := time.Now().UTC()

	// P-224 is only 224 bits, below 256-bit minimum.
	ecKey224, err := ecdsa.GenerateKey(elliptic.P224(), crand.Reader)
	require.NoError(t, err)

	result := svc.validateKeyStrength(&ecKey224.PublicKey, timestamp)
	require.False(t, result.Passed)
	require.Contains(t, result.Message, "below minimum")
}

// TestRAService_ValidateKeyStrength_Ed25519 tests Ed25519 key validation.
func TestRAService_ValidateKeyStrength_Ed25519(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	timestamp := time.Now().UTC()

	// Ed25519 key is always valid.
	ed25519Key, _, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	result := svc.validateKeyStrength(ed25519Key, timestamp)
	require.True(t, result.Passed)
	require.Equal(t, "Ed25519 key meets requirements", result.Message)
}

// TestRAService_ValidateKeyStrength_UnknownKeyType tests unknown key type.
func TestRAService_ValidateKeyStrength_UnknownKeyType(t *testing.T) {
	t.Parallel()

	svc, err := NewRAService(nil)
	require.NoError(t, err)

	timestamp := time.Now().UTC()

	// Pass a non-key type to trigger the default case.
	result := svc.validateKeyStrength("not-a-key", timestamp)
	require.False(t, result.Passed)
	require.Equal(t, "Unknown key type", result.Message)
}
