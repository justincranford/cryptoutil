// Copyright (c) 2025 Justin Cranford

//go:build !fuzz

package businesslogic

import (
	crand "crypto/rand"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	testify "github.com/stretchr/testify/require"
)

// TestEncryptDecryptProperty_RandomPayloads verifies the encrypt-then-decrypt identity property.
// For any random payload, decrypt(encrypt(payload)) must equal the original payload.
func TestEncryptDecryptProperty_RandomPayloads(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierElasticKeyForCoverage(t, stack, "prop-enc-dec", cryptoutilOpenapiModel.A256GCMDir)

	for _, size := range []int{cryptoutilSharedMagic.TestRandomStringLength16, cryptoutilSharedMagic.TestRandomStringLength64, cryptoutilSharedMagic.TestRandomStringLength256, cryptoutilSharedMagic.TestRandomStringLength1024} {
		payload := make([]byte, size)

		_, err := crand.Read(payload)
		testify.NoError(t, err)

		jweBytes, err := stack.service.PostEncryptByElasticKeyID(
			stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, payload)
		testify.NoError(t, err)
		testify.NotEmpty(t, jweBytes)

		decrypted, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, jweBytes)
		testify.NoError(t, err)
		testify.Equal(t, payload, decrypted, "decrypt(encrypt(x)) must equal x for size %d", size)
	}
}

// TestSignVerifyProperty_RandomPayloads verifies the sign-then-verify identity property.
// For any random payload, verify(sign(payload)) must return the original payload.
func TestSignVerifyProperty_RandomPayloads(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierElasticKeyForCoverage(t, stack, "prop-sign-verify", cryptoutilOpenapiModel.ES256)

	for _, size := range []int{cryptoutilSharedMagic.TestRandomStringLength16, cryptoutilSharedMagic.TestRandomStringLength64, cryptoutilSharedMagic.TestRandomStringLength256, cryptoutilSharedMagic.TestRandomStringLength1024} {
		payload := make([]byte, size)

		_, err := crand.Read(payload)
		testify.NoError(t, err)

		jwsBytes, err := stack.service.PostSignByElasticKeyID(stack.ctx, &ekID, payload)
		testify.NoError(t, err)
		testify.NotEmpty(t, jwsBytes)

		verified, err := stack.service.PostVerifyByElasticKeyID(stack.ctx, &ekID, jwsBytes)
		testify.NoError(t, err)
		testify.Equal(t, payload, verified, "verify(sign(x)) must return original payload for size %d", size)
	}
}

// TestEncryptProperty_OutputDiffersFromInput verifies semantic security: ciphertext must not equal plaintext.
func TestEncryptProperty_OutputDiffersFromInput(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierElasticKeyForCoverage(t, stack, "prop-semantic", cryptoutilOpenapiModel.A256GCMDir)

	payload := make([]byte, cryptoutilSharedMagic.AES256KeySize)

	_, err := crand.Read(payload)
	testify.NoError(t, err)

	jweBytes, err := stack.service.PostEncryptByElasticKeyID(
		stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, payload)
	testify.NoError(t, err)
	testify.NotEqual(t, payload, jweBytes, "encrypted output must differ from plaintext input")
}
