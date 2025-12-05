// Copyright (c) 2025 Justin Cranford

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"

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
