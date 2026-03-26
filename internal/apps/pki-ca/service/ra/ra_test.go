// Copyright (c) 2025 Justin Cranford

package ra

import (
	"context"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

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

	for i := 0; i < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		_, err = svc.SubmitRequest(ctx, csrPEM, "tls-server", "user-123")
		require.NoError(t, err)
	}

	// List all requests.
	results, total, err := svc.ListRequests(ctx, nil, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, total)
	require.Len(t, results, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	// List with pagination.
	results, total, err = svc.ListRequests(ctx, nil, 2, 0)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, total)
	require.Len(t, results, 2)

	// Filter by status.
	pending := StatusPending
	results, total, err = svc.ListRequests(ctx, &pending, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, 0)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, total)
	require.Len(t, results, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
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
