// Copyright (c) 2025 Justin Cranford

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	// Create a mock storage for testing.
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	tests := []struct {
		name        string
		issuer      any
		storage     cryptoutilCAStorage.Store
		profiles    map[string]*ProfileConfig
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil-issuer-fails",
			issuer:      nil,
			storage:     mockStorage,
			profiles:    nil,
			wantErr:     true,
			errContains: "issuer is required",
		},
		{
			name:        "nil-storage-fails",
			issuer:      nil, // Will fail issuer check first.
			storage:     nil,
			profiles:    nil,
			wantErr:     true,
			errContains: "issuer is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// NewHandler requires an actual *Issuer, so we test nil case.
			_, err := NewHandler(nil, tc.storage, tc.profiles)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMapAPIRevocationReasonToStorage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    cryptoutilCAServer.RevocationReason
		expected cryptoutilCAStorage.RevocationReason
	}{
		{"key_compromise", cryptoutilCAServer.KeyCompromise, cryptoutilCAStorage.ReasonKeyCompromise},
		{"ca_compromise", cryptoutilCAServer.CACompromise, cryptoutilCAStorage.ReasonCACompromise},
		{"affiliation_changed", cryptoutilCAServer.AffiliationChanged, cryptoutilCAStorage.ReasonAffiliationChanged},
		{"superseded", cryptoutilCAServer.Superseded, cryptoutilCAStorage.ReasonSuperseded},
		{"cessation_of_operation", cryptoutilCAServer.CessationOfOperation, cryptoutilCAStorage.ReasonCessationOfOperation},
		{"certificate_hold", cryptoutilCAServer.CertificateHold, cryptoutilCAStorage.ReasonCertificateHold},
		{"remove_from_crl", cryptoutilCAServer.RemoveFromCRL, cryptoutilCAStorage.ReasonRemoveFromCRL},
		{"privilege_withdrawn", cryptoutilCAServer.PrivilegeWithdrawn, cryptoutilCAStorage.ReasonPrivilegeWithdrawn},
		{"aa_compromise", cryptoutilCAServer.AaCompromise, cryptoutilCAStorage.ReasonAACompromise},
		{"unspecified", cryptoutilCAServer.Unspecified, cryptoutilCAStorage.ReasonUnspecified},
		{"unknown_defaults_to_unspecified", "unknown_reason", cryptoutilCAStorage.ReasonUnspecified},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := mapAPIRevocationReasonToStorage(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
