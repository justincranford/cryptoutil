// Copyright (c) 2025 Justin Cranford

package userauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStubHSM(t *testing.T) {
	t.Parallel()

	hsm := NewStubHSM()
	require.NotNil(t, hsm)

	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "GenerateKey",
			fn: func() error {
				_, err := hsm.GenerateKey(ctx, cryptoutilSharedMagic.KeyTypeRSA, cryptoutilSharedMagic.DefaultMetricsBatchSize)

				return err
			},
		},
		{
			name: "SignData",
			fn: func() error {
				_, err := hsm.SignData(ctx, "key-id", []byte("data"))

				return err
			},
		},
		{
			name: "EncryptData",
			fn: func() error {
				_, err := hsm.EncryptData(ctx, "key-id", []byte("plaintext"))

				return err
			},
		},
		{
			name: "DecryptData",
			fn: func() error {
				_, err := hsm.DecryptData(ctx, "key-id", []byte("ciphertext"))

				return err
			},
		},
		{
			name: "DeleteKey",
			fn: func() error {
				return hsm.DeleteKey(ctx, "key-id")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			require.Contains(t, err.Error(), "stub HSM")
			require.Contains(t, err.Error(), "not implemented")
		})
	}
}

func TestStubHSMVerifySignature(t *testing.T) {
	t.Parallel()

	hsm := NewStubHSM()
	ctx := context.Background()

	result := hsm.VerifySignature(ctx, "key-id", []byte("data"), []byte("signature"))
	require.False(t, result)
}

func TestStubTPM(t *testing.T) {
	t.Parallel()

	tpm := NewStubTPM()
	require.NotNil(t, tpm)

	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "SealData",
			fn: func() error {
				_, err := tpm.SealData(ctx, []byte("data"), []uint32{0, 1, 2})

				return err
			},
		},
		{
			name: "UnsealData",
			fn: func() error {
				_, err := tpm.UnsealData(ctx, []byte("sealed"))

				return err
			},
		},
		{
			name: "GenerateKey",
			fn: func() error {
				_, err := tpm.GenerateKey(ctx, cryptoutilSharedMagic.KeyTypeRSA)

				return err
			},
		},
		{
			name: "Sign",
			fn: func() error {
				_, err := tpm.Sign(ctx, "key-id", []byte("data"))

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			require.Contains(t, err.Error(), "stub TPM")
			require.Contains(t, err.Error(), "not implemented")
		})
	}
}

func TestStubTPMVerify(t *testing.T) {
	t.Parallel()

	tpm := NewStubTPM()
	ctx := context.Background()

	result := tpm.Verify(ctx, "key-id", []byte("data"), []byte("signature"))
	require.False(t, result)
}

func TestStubSecureElement(t *testing.T) {
	t.Parallel()

	se := NewStubSecureElement()
	require.NotNil(t, se)

	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "StoreCredential",
			fn: func() error {
				return se.StoreCredential(ctx, "cred-id", []byte("data"))
			},
		},
		{
			name: "RetrieveCredential",
			fn: func() error {
				_, err := se.RetrieveCredential(ctx, "cred-id")

				return err
			},
		},
		{
			name: "DeleteCredential",
			fn: func() error {
				return se.DeleteCredential(ctx, "cred-id")
			},
		},
		{
			name: "GenerateKey",
			fn: func() error {
				_, err := se.GenerateKey(ctx, "EC")

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			require.Error(t, err)
			require.Contains(t, err.Error(), "stub secure element")
			require.Contains(t, err.Error(), "not implemented")
		})
	}
}
