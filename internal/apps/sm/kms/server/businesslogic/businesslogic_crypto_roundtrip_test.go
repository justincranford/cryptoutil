package businesslogic

import (
	"testing"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

const testCryptoDescription = "test description"

// setupCryptoTestStack creates a testStack with MaxOpenConns increased to allow barrier
// nested read transactions inside ORM ReadOnly transactions. SQLite with MaxOpenConns=1
// deadlocks when barrier DecryptContentWithContext is called inside an ORM transaction
// because both need a connection from the same pool.
func setupCryptoTestStack(t *testing.T) *testStack {
	t.Helper()

	stack := setupTestStack(t)

	sqlDB, err := stack.core.DB.DB()
	testify.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)

	return stack
}

// seedBarrierEncryptedElasticKey creates an elastic key with a properly barrier-encrypted
// material key. This bypasses AddElasticKey which deadlocks on SQLite due to nested
// write transactions (ORM ReadWrite + barrier ReadWrite on same connection pool).
func seedBarrierEncryptedElasticKey(
	t *testing.T,
	stack *testStack,
	name string,
	alg cryptoutilOpenapiModel.ElasticKeyAlgorithm,
) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID

	// Create elastic key directly via GORM (no transaction nesting).
	ekID := stack.service.jwkGenService.GenerateUUIDv7()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                *ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       testCryptoDescription,
		ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
		ElasticKeyAlgorithm:         alg,
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            cryptoutilKmsServer.Active,
	}

	err := stack.core.DB.Create(ek).Error
	testify.NoError(t, err)

	// Generate JWK for the algorithm.
	materialKeyID, _, _, clearNonPublicBytes, clearPublicBytes, err := stack.service.generateJWK(&alg)
	testify.NoError(t, err)

	// Encrypt private material using barrier (standalone call — no outer transaction).
	encryptedNonPublicBytes, err := stack.service.barrierService.EncryptContentWithContext(stack.ctx, clearNonPublicBytes)
	testify.NoError(t, err)

	// Insert material key directly via GORM.
	now := time.Now().UTC().UnixMilli()
	mk := &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  *ekID,
		MaterialKeyID:                 *materialKeyID,
		MaterialKeyClearPublic:        clearPublicBytes,
		MaterialKeyEncryptedNonPublic: encryptedNonPublicBytes,
		MaterialKeyGenerateDate:       &now,
	}

	err = stack.core.DB.Create(mk).Error
	testify.NoError(t, err)

	return *ekID
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "enc-dec-roundtrip", cryptoutilOpenapiModel.A256GCMDir)

	plaintext := []byte("hello world, this is a test payload for encrypt/decrypt round trip")

	// Encrypt.
	jweBytes, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, plaintext)
	testify.NoError(t, err)
	testify.NotEmpty(t, jweBytes)

	// Decrypt.
	decrypted, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, jweBytes)
	testify.NoError(t, err)
	testify.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_WithContext(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "enc-dec-ctx", cryptoutilOpenapiModel.A256GCMDir)

	plaintext := []byte("payload with context")
	encCtx := "encryption-context-value"

	// Encrypt with context.
	jweBytes, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{Context: &encCtx}, plaintext)
	testify.NoError(t, err)
	testify.NotEmpty(t, jweBytes)

	// Decrypt (context is embedded in JWE).
	decrypted, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, jweBytes)
	testify.NoError(t, err)
	testify.Equal(t, plaintext, decrypted)
}

func TestSignVerify_RoundTrip(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "sign-verify", cryptoutilOpenapiModel.ES256)

	payload := []byte("message to sign and verify")

	// Sign.
	jwsBytes, err := stack.service.PostSignByElasticKeyID(stack.ctx, &ekID, payload)
	testify.NoError(t, err)
	testify.NotEmpty(t, jwsBytes)

	// Verify.
	verified, err := stack.service.PostVerifyByElasticKeyID(stack.ctx, &ekID, jwsBytes)
	testify.NoError(t, err)
	testify.Equal(t, payload, verified)
}

func TestSignVerify_HMAC(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "sign-verify-hmac", cryptoutilOpenapiModel.HS256)

	payload := []byte("HMAC symmetric signing payload")

	// Sign with HMAC.
	jwsBytes, err := stack.service.PostSignByElasticKeyID(stack.ctx, &ekID, payload)
	testify.NoError(t, err)
	testify.NotEmpty(t, jwsBytes)

	// Verify with HMAC.
	verified, err := stack.service.PostVerifyByElasticKeyID(stack.ctx, &ekID, jwsBytes)
	testify.NoError(t, err)
	testify.Equal(t, payload, verified)
}

func TestPostGenerate_SymmetricKey(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "postgen-sym", cryptoutilOpenapiModel.A256GCMDir)

	genAlg := cryptoutilOpenapiModel.Oct256

	encryptedPrivate, clearPrivate, clearPublic, err := stack.service.PostGenerateByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.GenerateParams{Alg: &genAlg})
	testify.NoError(t, err)
	testify.NotEmpty(t, encryptedPrivate)
	testify.NotEmpty(t, clearPrivate)
	// oct keys have no public part.
	testify.Nil(t, clearPublic)
}

func TestPostGenerate_AsymmetricKey(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "postgen-asym", cryptoutilOpenapiModel.A256GCMDir)

	genAlg := cryptoutilOpenapiModel.ECP256

	encryptedPrivate, clearPrivate, clearPublic, err := stack.service.PostGenerateByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.GenerateParams{Alg: &genAlg})
	testify.NoError(t, err)
	testify.NotEmpty(t, encryptedPrivate)
	testify.NotEmpty(t, clearPrivate)
	testify.NotEmpty(t, clearPublic)
}

func TestEncryptDecrypt_AsymmetricKey(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Use ECDH-ES+A256KW which is asymmetric JWE.
	ekID := seedBarrierEncryptedElasticKey(t, stack, "enc-dec-asym", cryptoutilOpenapiModel.A256GCMECDHESA256KW)

	plaintext := []byte("asymmetric encrypt/decrypt payload")

	// Encrypt with public key.
	jweBytes, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, plaintext)
	testify.NoError(t, err)
	testify.NotEmpty(t, jweBytes)

	// Decrypt with private key.
	decrypted, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, jweBytes)
	testify.NoError(t, err)
	testify.Equal(t, plaintext, decrypted)
}

func TestDecrypt_WrongElasticKey(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID1 := seedBarrierEncryptedElasticKey(t, stack, "dec-wrong-ek1", cryptoutilOpenapiModel.A256GCMDir)
	ekID2 := seedBarrierEncryptedElasticKey(t, stack, "dec-wrong-ek2", cryptoutilOpenapiModel.A256GCMDir)

	plaintext := []byte("decrypt with wrong elastic key")

	// Encrypt with first elastic key.
	jweBytes, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID1, &cryptoutilOpenapiModel.EncryptParams{}, plaintext)
	testify.NoError(t, err)

	// Attempt decrypt with second elastic key — kid mismatch should fail.
	_, err = stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID2, jweBytes)
	testify.Error(t, err)
}

func TestDecrypt_InvalidJWE(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "dec-invalid-jwe", cryptoutilOpenapiModel.A256GCMDir)

	// Attempt decrypt with invalid JWE bytes.
	_, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, []byte("not-a-jwe"))
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to parse JWE message bytes")
}

func TestVerify_InvalidJWS(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "verify-invalid-jws", cryptoutilOpenapiModel.ES256)

	// Attempt verify with invalid JWS bytes.
	_, err := stack.service.PostVerifyByElasticKeyID(stack.ctx, &ekID, []byte("not-a-jws"))
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to parse JWS message bytes")
}

func TestEncrypt_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// ES256 is JWS, not JWE — encrypt should use it but decrypt should reject.
	ekID := seedBarrierEncryptedElasticKey(t, stack, "enc-jws-alg", cryptoutilOpenapiModel.ES256)

	plaintext := []byte("try to decrypt with JWS algorithm")

	// PostEncryptByElasticKeyID attempts to encrypt — for JWS key it will try to use
	// the non-public key for JWE which should fail because ES256 is not a JWE algorithm.
	_, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, plaintext)
	testify.Error(t, err)
}

func TestImportMaterialKey_NotAllowed(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "import-denied", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)

	_, err := stack.service.ImportMaterialKey(stack.ctx, &ekID, &cryptoutilKmsServer.MaterialKeyImport{JWK: `{"kty":"oct","k":"dGVzdA"}`})
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "import not allowed for ElasticKey")
}

func TestEncryptDecrypt_MultipleVersions(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)
	ekID := seedBarrierEncryptedElasticKey(t, stack, "multi-ver", cryptoutilOpenapiModel.A256GCMDir)

	plaintext1 := []byte("message encrypted with key v1")

	// Encrypt with first material key (v1).
	jwe1, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, plaintext1)
	testify.NoError(t, err)

	// Add a second material key directly (bypass AddElasticKey deadlock).
	alg := cryptoutilOpenapiModel.A256GCMDir
	mk2ID, _, _, clearNonPublic2, clearPublic2, err := stack.service.generateJWK(&alg)
	testify.NoError(t, err)

	encrypted2, err := stack.service.barrierService.EncryptContentWithContext(stack.ctx, clearNonPublic2)
	testify.NoError(t, err)

	now2 := time.Now().UTC().UnixMilli()
	mk2 := &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  ekID,
		MaterialKeyID:                 *mk2ID,
		MaterialKeyClearPublic:        clearPublic2,
		MaterialKeyEncryptedNonPublic: encrypted2,
		MaterialKeyGenerateDate:       &now2,
	}

	err = stack.core.DB.Create(mk2).Error
	testify.NoError(t, err)

	plaintext2 := []byte("message encrypted with key v2")

	// Encrypt with latest material key (v2).
	jwe2, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekID, &cryptoutilOpenapiModel.EncryptParams{}, plaintext2)
	testify.NoError(t, err)

	// Decrypt both — key ID embedded in JWE header selects correct version.
	dec1, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, jwe1)
	testify.NoError(t, err)
	testify.Equal(t, plaintext1, dec1)

	dec2, err := stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekID, jwe2)
	testify.NoError(t, err)
	testify.Equal(t, plaintext2, dec2)
}

func TestDecryptByElasticKeyID_NonJWEAlgorithm(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Create a JWS elastic key (ES256).
	ekIDJWS := seedBarrierEncryptedElasticKey(t, stack, "dec-jws-reject", cryptoutilOpenapiModel.ES256)

	// Create a JWE elastic key to get valid JWE ciphertext.
	ekIDJWE := seedBarrierEncryptedElasticKey(t, stack, "dec-jwe-source", cryptoutilOpenapiModel.A256GCMDir)

	plaintext := []byte("test payload")

	jweBytes, err := stack.service.PostEncryptByElasticKeyID(stack.ctx, &ekIDJWE, &cryptoutilOpenapiModel.EncryptParams{}, plaintext)
	testify.NoError(t, err)

	// Decrypt with JWS elastic key — should fail because ES256 is not JWE.
	_, err = stack.service.PostDecryptByElasticKeyID(stack.ctx, &ekIDJWS, jweBytes)
	testify.Error(t, err)
}

func TestVerifyByElasticKeyID_NonJWSAlgorithm(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Create a JWE elastic key (A256GCM+Dir) — not a JWS algorithm.
	ekID := seedBarrierEncryptedElasticKey(t, stack, "verify-jwe-reject", cryptoutilOpenapiModel.A256GCMDir)

	// Create a JWS key to produce a valid JWS message.
	ekIDJWS := seedBarrierEncryptedElasticKey(t, stack, "verify-jws-source", cryptoutilOpenapiModel.ES256)

	payload := []byte("test payload")

	jwsBytes, err := stack.service.PostSignByElasticKeyID(stack.ctx, &ekIDJWS, payload)
	testify.NoError(t, err)

	// Verify with JWE elastic key — should fail because A256GCM+Dir is not JWS.
	_, err = stack.service.PostVerifyByElasticKeyID(stack.ctx, &ekID, jwsBytes)
	testify.Error(t, err)
}

func TestGetMaterialKeys_CachePath(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	// Seed two elastic keys with material keys — exercises the elastic key cache in GetMaterialKeys.
	ekID1 := seedElasticKey(t, stack, "cache-ek1", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID1)
	seedMaterialKey(t, stack, ekID1) // Second MK on same EK — cache hit.

	ekID2 := seedElasticKey(t, stack, "cache-ek2", cryptoutilOpenapiModel.ES256, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID2)

	// Get all material keys — this exercises the elasticKeyCache with multiple elastic keys.
	mks, err := stack.service.GetMaterialKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.GreaterOrEqual(t, len(mks), 3)
}

func TestGetMaterialKeysForElasticKey_WithPagination(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	ekID := seedElasticKey(t, stack, "mk-paginated", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID)
	seedMaterialKey(t, stack, ekID)
	seedMaterialKey(t, stack, ekID)

	page := 0
	size := 2
	params := &cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{Page: &page, Size: &size}

	mks, err := stack.service.GetMaterialKeysForElasticKey(stack.ctx, &ekID, params)
	testify.NoError(t, err)
	testify.Len(t, mks, 2) // First page of 3 items with size 2.
}

func TestGetElasticKeys_WithPagination(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	seedElasticKey(t, stack, "paginated-a", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedElasticKey(t, stack, "paginated-b", cryptoutilOpenapiModel.ES256, cryptoutilKmsServer.Active)
	seedElasticKey(t, stack, "paginated-c", cryptoutilOpenapiModel.HS256, cryptoutilKmsServer.Active)

	page := 0
	size := 2
	params := &cryptoutilOpenapiModel.ElasticKeysQueryParams{Page: &page, Size: &size}

	eks, err := stack.service.GetElasticKeys(stack.ctx, params)
	testify.NoError(t, err)
	testify.Len(t, eks, 2) // First page of 3 items with size 2.
}

func TestGetMaterialKeys_WithPagination(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	ekID := seedElasticKey(t, stack, "mk-paged", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID)
	seedMaterialKey(t, stack, ekID)
	seedMaterialKey(t, stack, ekID)

	page := 0
	size := 2
	params := &cryptoutilOpenapiModel.MaterialKeysQueryParams{Page: &page, Size: &size}

	mks, err := stack.service.GetMaterialKeys(stack.ctx, params)
	testify.NoError(t, err)
	testify.Len(t, mks, 2) // First page of 3 items with size 2.
}

func TestGetElasticKeys_InvalidPageParam(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	negativePage := -1
	params := &cryptoutilOpenapiModel.ElasticKeysQueryParams{Page: &negativePage}

	_, err := stack.service.GetElasticKeys(stack.ctx, params)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "invalid ElasticKeysQueryParams")
}

func TestGetMaterialKeysForElasticKey_NonexistentElasticKey(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	nonexistentID := googleUuid.New()

	_, err := stack.service.GetMaterialKeysForElasticKey(stack.ctx, &nonexistentID, nil)
	testify.Error(t, err)
}

func TestGetMaterialKeys_InvalidPageParam(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	negativePage := -1
	params := &cryptoutilOpenapiModel.MaterialKeysQueryParams{Page: &negativePage}

	_, err := stack.service.GetMaterialKeys(stack.ctx, params)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "invalid MaterialKeysQueryParams")
}
