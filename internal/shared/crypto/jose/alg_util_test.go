// Copyright (c) 2025 Justin Cranford

// Package crypto provides JOSE cryptographic utilities for key generation and validation.
package crypto

import (
	"testing"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

func TestIsJWEAlg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     *joseJwa.KeyAlgorithm
		want    bool
		wantErr bool
	}{
		{
			name: "valid JWE key encryption algorithm - RSA-OAEP",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.RSA_OAEP()

				return &alg
			}(),
			want:    true,
			wantErr: false,
		},
		{
			name: "valid JWE key encryption algorithm - A128KW",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.A128KW()

				return &alg
			}(),
			want:    true,
			wantErr: false,
		},
		{
			name: "not a JWE algorithm - signature algorithm RS256",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.RS256()

				return &alg
			}(),
			want:    false,
			wantErr: false,
		},
		{
			name: "not a JWE algorithm - signature algorithm EdDSA",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.EdDSA()

				return &alg
			}(),
			want:    false,
			wantErr: false,
		},
		{
			name:    "nil algorithm pointer returns error",
			alg:     nil,
			want:    false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := IsJWEAlg(tc.alg, 0)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "alg 0 invalid")
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestIsJWSAlg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     *joseJwa.KeyAlgorithm
		want    bool
		wantErr bool
	}{
		{
			name: "valid JWS signature algorithm - RS256",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.RS256()

				return &alg
			}(),
			want:    true,
			wantErr: false,
		},
		{
			name: "valid JWS signature algorithm - EdDSA",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.EdDSA()

				return &alg
			}(),
			want:    true,
			wantErr: false,
		},
		{
			name: "valid JWS signature algorithm - HS256",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.HS256()

				return &alg
			}(),
			want:    true,
			wantErr: false,
		},
		{
			name: "not a JWS algorithm - key encryption algorithm RSA-OAEP",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.RSA_OAEP()

				return &alg
			}(),
			want:    false,
			wantErr: false,
		},
		{
			name: "not a JWS algorithm - key encryption algorithm A128KW",
			alg: func() *joseJwa.KeyAlgorithm {
				var alg joseJwa.KeyAlgorithm = joseJwa.A128KW()

				return &alg
			}(),
			want:    false,
			wantErr: false,
		},
		{
			name:    "nil algorithm pointer returns error",
			alg:     nil,
			want:    false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := IsJWSAlg(tc.alg, 0)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "alg 0 invalid")
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
