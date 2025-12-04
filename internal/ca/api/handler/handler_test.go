// Copyright (c) 2025 Justin Cranford

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		issuer  interface{}
		wantErr bool
	}{
		{
			name:    "nil-issuer-fails",
			issuer:  nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// NewHandler requires an actual *Issuer, so we test nil case.
			_, err := NewHandler(nil, nil)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "issuer is required")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
