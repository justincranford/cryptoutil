package businesslogic

import (
	crand "crypto/rand"
	"strings"
	"testing"
	"time"

	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	testify "github.com/stretchr/testify/require"
)

// testTamperedB64 is a base64url-encoded value used to replace JWE/JWS compact
// serialization parts in tamper tests.
const testTamperedB64 = "dGFtcGVyZWQ"

// TestAddElasticKey_UnsupportedAlgorithm covers the generateJWK else branch
// and the AddElasticKey "failed to generate first MaterialKey" error path.
func TestAddElasticKey_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	unsupportedAlg := "UNSUPPORTED_ALGO_FOR_COVERAGE"

	_, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
		Name:      "unsupported-alg-ek",
		Algorithm: unsupportedAlg,
		Provider:  providerInternal,
	})

	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to generate first MaterialKey")
}

// TestGenerateMaterialKey_UnsupportedAlgorithm covers the GenerateMaterialKeyInElasticKey
// "failed to generate new MaterialKey" error path (step 2 error via generateJWK else branch).
// The elastic key has an active status but an algorithm that generateJWK cannot handle.
func TestGenerateMaterialKey_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	// Insert elastic key directly with an unsupported algorithm (no DB constraint on algorithm).
	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := googleUuid.New()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              "gen-mk-unsupported-alg",
		ElasticKeyDescription:       "test-desc",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.ElasticKeyAlgorithm("NOT_JWE_OR_JWS"),
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            cryptoutilKmsServer.Active,
	}

	testify.NoError(t, stack.core.DB.Create(ek).Error)

	_, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, &ekID, nil)

	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to generate new MaterialKey")
}

// TestGetElasticKeys_ValidPagination ensures the pagination + mapping path works end-to-end.
func TestGetElasticKeys_ValidPagination(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	seedElasticKey(t, stack, "pag-ek1", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedElasticKey(t, stack, "pag-ek2", cryptoutilOpenapiModel.ES256, cryptoutilKmsServer.Active)

	page := 0
	size := 2
	params := &cryptoutilOpenapiModel.ElasticKeysQueryParams{Page: &page, Size: &size}

	eks, err := stack.service.GetElasticKeys(stack.ctx, params)
	testify.NoError(t, err)
	testify.Len(t, eks, 2)
}

// setupCryptoTestStackHighConns creates a testStack similar to setupCryptoTestStack
// for tests that need barrier-encrypted material keys but only for coverage tests.
var _ = setupCryptoTestStack // ensure function is reachable (used in roundtrip tests).

// seedBarrierElasticKeyForCoverage creates a barrier-encrypted elastic key for coverage testing.
func seedBarrierElasticKeyForCoverage(t *testing.T, stack *testStack, name string, alg cryptoutilOpenapiModel.ElasticKeyAlgorithm) googleUuid.UUID {
	t.Helper()

	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := stack.service.jwkGenService.GenerateUUIDv7()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                *ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       "coverage-desc",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         alg,
		ElasticKeyVersioningAllowed: false,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            cryptoutilKmsServer.Active,
	}

	testify.NoError(t, stack.core.DB.Create(ek).Error)

	materialKeyID, _, _, clearNonPublicBytes, clearPublicBytes, err := stack.service.generateJWK(&alg)
	testify.NoError(t, err)

	encryptedNonPublicBytes, err := stack.service.barrierService.EncryptContentWithContext(stack.ctx, clearNonPublicBytes)
	testify.NoError(t, err)

	now := time.Now().UTC().UnixMilli()
	mk := &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  *ekID,
		MaterialKeyID:                 *materialKeyID,
		MaterialKeyClearPublic:        clearPublicBytes,
		MaterialKeyEncryptedNonPublic: encryptedNonPublicBytes,
		MaterialKeyGenerateDate:       &now,
	}

	testify.NoError(t, stack.core.DB.Create(mk).Error)

	return *ekID
}

// TestGetMaterialKeyByElasticKeyAndMaterialKeyID_NotFoundMaterialKey covers the
// "failed to get MaterialKeys by ElasticKeyID and MaterialKeyID" error path.
func TestGetMaterialKeyByElasticKeyAndMaterialKeyID_NotFoundMaterialKey(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	ekID := seedElasticKey(t, stack, "matkey-notfound", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	nonExistentMKID := googleUuid.New()

	_, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(stack.ctx, &ekID, &nonExistentMKID)

	testify.Error(t, err)
}

// TestGetMaterialKeys_CacheHitExercise seeds multiple MKs across EKs to
// exercise the elastic key cache hit path in GetMaterialKeys.
func TestGetMaterialKeys_CacheHitExercise(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	// Seed two elastic keys; both get two material keys so the cache hit path fires.
	ekID1 := seedElasticKey(t, stack, "cache-cov-ek1", cryptoutilOpenapiModel.A256GCMDir, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID1)
	seedMaterialKey(t, stack, ekID1) // Second MK -> cache hit for ekID1.

	ekID2 := seedElasticKey(t, stack, "cache-cov-ek2", cryptoutilOpenapiModel.ES256, cryptoutilKmsServer.Active)
	seedMaterialKey(t, stack, ekID2)
	seedMaterialKey(t, stack, ekID2) // Second MK -> cache hit for ekID2.

	mks, err := stack.service.GetMaterialKeys(stack.ctx, nil)
	testify.NoError(t, err)
	testify.GreaterOrEqual(t, len(mks), 4)
}

// TestPostSign_MapMaterialKeyError exercises GetMaterialKeyByElasticKeyAndMaterialKeyID
// when the elastic key exists but the material key does not.
func TestGetMaterialKey_ElasticKeyNotFound(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	nonExistentEKID := googleUuid.New()
	nonExistentMKID := googleUuid.New()

	_, err := stack.service.GetMaterialKeyByElasticKeyAndMaterialKeyID(stack.ctx, &nonExistentEKID, &nonExistentMKID)

	testify.Error(t, err)
}

// TestGenerateMaterialKeyInElasticKey_VersioningMaterialKey generates a second material
// key version to exercise the full GenerateMaterialKeyInElasticKey success path via the service.
// Uses setupCryptoTestStack to allow nested barrier transactions without deadlocking.
func TestGenerateMaterialKeyInElasticKey_VersioningMaterialKey(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	sqlDB, err := stack.core.DB.DB()
	testify.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnectionsForGORM)

	// Seed a versioning-allowed elastic key directly.
	tenantID := cryptoutilKmsMiddleware.GetRealmContext(stack.ctx).TenantID
	ekID := stack.service.jwkGenService.GenerateUUIDv7()
	ek := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                *ekID,
		TenantID:                    tenantID,
		ElasticKeyName:              "versioning-ek",
		ElasticKeyDescription:       "test",
		ElasticKeyProvider:          "Internal",
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMDir,
		ElasticKeyVersioningAllowed: true,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            cryptoutilKmsServer.Active,
	}

	testify.NoError(t, stack.core.DB.Create(ek).Error)

	// Seed first material key.
	alg := cryptoutilOpenapiModel.A256GCMDir
	mk1ID, _, _, clearNonPublic1, clearPublic1, err := stack.service.generateJWK(&alg)
	testify.NoError(t, err)

	encNonPublic1, err := stack.service.barrierService.EncryptContentWithContext(stack.ctx, clearNonPublic1)
	testify.NoError(t, err)

	now1 := time.Now().UTC().UnixMilli()
	mk1 := &cryptoutilOrmRepository.MaterialKey{
		ElasticKeyID:                  *ekID,
		MaterialKeyID:                 *mk1ID,
		MaterialKeyClearPublic:        clearPublic1,
		MaterialKeyEncryptedNonPublic: encNonPublic1,
		MaterialKeyGenerateDate:       &now1,
	}
	testify.NoError(t, stack.core.DB.Create(mk1).Error)

	// Now generate a second material key through the service.
	mk2, err := stack.service.GenerateMaterialKeyInElasticKey(stack.ctx, ekID, nil)
	testify.NoError(t, err)
	testify.NotNil(t, mk2)
}

// TestPostDecrypt_JWEWithoutKid covers the "failed to get JWE message header kid" path.
// A valid JWE without a kid header in the protected headers triggers this error.
func TestPostDecrypt_JWEWithoutKid(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	// Create a valid JWE using direct encryption without setting a kid header.
	rawKey := make([]byte, cryptoutilSharedMagic.AES256KeySize)

	_, err := crand.Read(rawKey)
	testify.NoError(t, err)

	jweBytes, err := joseJwe.Encrypt(
		[]byte("test payload"),
		joseJwe.WithKey(joseJwa.DIRECT(), rawKey),
		joseJwe.WithContentEncryption(joseJwa.A256GCM()),
	)
	testify.NoError(t, err)

	anyEKID := googleUuid.New()

	_, err = stack.service.PostDecryptByElasticKeyID(stack.ctx, &anyEKID, jweBytes)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to get JWE message header kid")
}

// TestPostDecrypt_WrongAlgorithmType covers the "decrypt not supported by KeyMaterial" path.
// A JWE whose kid references a JWS (non-JWE) elastic key triggers this error.
func TestPostDecrypt_WrongAlgorithmType(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Seed a JWS elastic key (ES256) with a barrier-encrypted material key.
	jwsEKID := seedBarrierElasticKeyForCoverage(t, stack, "postdec-wrong-alg", cryptoutilOpenapiModel.ES256)

	// Query the DB to get the material key ID.
	var ormMK cryptoutilOrmRepository.MaterialKey

	testify.NoError(t, stack.core.DB.Where("elastic_key_id = ?", jwsEKID).First(&ormMK).Error)

	materialKeyID := ormMK.MaterialKeyID

	// Generate a fresh JWE JWK (A256GCMDir) and override its kid to materialKeyID.
	jweAlg := cryptoutilOpenapiModel.A256GCMDir

	_, jweJWK, _, _, _, err := stack.service.generateJWK(&jweAlg)
	testify.NoError(t, err)
	testify.NoError(t, jweJWK.Set(joseJwk.KeyIDKey, materialKeyID.String()))

	// Encrypt to produce JWE bytes where kid = materialKeyID of the JWS EK.
	_, jweBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{jweJWK}, []byte("test payload"))
	testify.NoError(t, err)

	// PostDecryptByElasticKeyID with JWS EK id and JWE with that EK's kid → !IsJWE error.
	_, err = stack.service.PostDecryptByElasticKeyID(stack.ctx, &jwsEKID, jweBytes)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "decrypt not supported by KeyMaterial")
}

// TestPostVerify_JWSWithoutKid covers the "failed to get JWS message headers kid and alg" path.
// A valid JWS without a kid header in the signature protected headers triggers this error.
func TestPostVerify_JWSWithoutKid(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)

	// Create a valid compact JWS using HMAC-SHA256 without setting a kid header.
	rawKey := make([]byte, cryptoutilSharedMagic.HMACSHA256KeySize)

	_, err := crand.Read(rawKey)
	testify.NoError(t, err)

	jwsBytes, err := joseJws.Sign(
		[]byte("test payload"),
		joseJws.WithKey(joseJwa.HS256(), rawKey),
	)
	testify.NoError(t, err)

	anyEKID := googleUuid.New()

	_, err = stack.service.PostVerifyByElasticKeyID(stack.ctx, &anyEKID, jwsBytes)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to get JWS message headers kid and alg")
}

// TestPostVerify_WrongAlgorithmType covers the "verify not supported by KeyMaterial" path.
// A JWS whose kid references a JWE (non-JWS) elastic key triggers this error.
func TestPostVerify_WrongAlgorithmType(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Seed a JWE elastic key (A256GCMDir) with a barrier-encrypted material key.
	jweEKID := seedBarrierElasticKeyForCoverage(t, stack, "postver-wrong-alg", cryptoutilOpenapiModel.A256GCMDir)

	// Query the DB to get the material key ID.
	var ormMK cryptoutilOrmRepository.MaterialKey

	testify.NoError(t, stack.core.DB.Where("elastic_key_id = ?", jweEKID).First(&ormMK).Error)

	materialKeyID := ormMK.MaterialKeyID

	// Generate a fresh JWS JWK (ES256) and override its kid to materialKeyID.
	jwsAlg := cryptoutilOpenapiModel.ES256

	_, jwsJWK, _, _, _, err := stack.service.generateJWK(&jwsAlg)
	testify.NoError(t, err)
	testify.NoError(t, jwsJWK.Set(joseJwk.KeyIDKey, materialKeyID.String()))

	// Sign to produce JWS bytes where kid = materialKeyID of the JWE EK.
	_, jwsBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{jwsJWK}, []byte("test payload"))
	testify.NoError(t, err)

	// PostVerifyByElasticKeyID with JWE EK id and JWS with that EK's kid → !IsJWS error.
	_, err = stack.service.PostVerifyByElasticKeyID(stack.ctx, &jweEKID, jwsBytes)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "verify not supported by KeyMaterial")
}

// TestGetMaterialKeysForElasticKey_InvalidPage covers businesslogic.go:275 – the
// toOrmGetMaterialKeysForElasticKeyQueryParams validation-error return path.
func TestGetMaterialKeysForElasticKey_InvalidPage(t *testing.T) {
	t.Parallel()

	stack := setupTestStack(t)
	anyEKID := googleUuid.New() // validation fails before any DB call

	negativePage := -1
	invalidParams := &cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
		Page: &negativePage,
	}

	_, err := stack.service.GetMaterialKeysForElasticKey(stack.ctx, &anyEKID, invalidParams)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to map MaterialKeys for ElasticKey query parameters")
}

// TestPostDecrypt_TamperedCiphertext covers businesslogic_crypto.go:82 – the
// DecryptBytes failure path when ciphertext is tampered after the material key
// is successfully resolved.
func TestPostDecrypt_TamperedCiphertext(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Create JWE-type EK; AddElasticKey generates the first material key automatically.
	ek, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
		Name:        "tamper-jwe-dec",
		Description: ptr("tampered-jwe-test"),
		Algorithm:   string(cryptoutilOpenapiModel.A256GCMDir),
		Provider:    providerInternal,
	})
	testify.NoError(t, err)

	ekID := ek.ElasticKeyID

	// Encrypt a valid payload to obtain a well-formed JWE compact serialization.
	jweBytes, err := stack.service.PostEncryptByElasticKeyID(
		stack.ctx, ekID, &cryptoutilOpenapiModel.EncryptParams{}, []byte("test-payload"),
	)
	testify.NoError(t, err)

	// JWE compact has 5 parts: header.encryptedKey.iv.ciphertext.tag
	// Replace ciphertext (part[3]) with different bytes so the GCM tag no longer matches.
	parts := strings.Split(string(jweBytes), ".")
	testify.Len(t, parts, cryptoutilSharedMagic.JWECompactParts)

	parts[3] = testTamperedB64 // base64url for "tampered"

	tamperedJWE := []byte(strings.Join(parts, "."))

	// Decryption must fail with the "failed to decrypt bytes" error (line 82).
	_, err = stack.service.PostDecryptByElasticKeyID(stack.ctx, ekID, tamperedJWE)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to decrypt bytes with MaterialKey")
}

// TestPostVerify_TamperedPayload covers businesslogic_crypto.go:138 – the
// VerifyBytes failure path when the JWS payload is tampered after the material
// key is successfully resolved.
func TestPostVerify_TamperedPayload(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Create JWS-type EK; AddElasticKey generates the first material key automatically.
	ek, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
		Name: "tamper-jws-ver", Description: ptr("tampered-jws-test"), Algorithm: string(cryptoutilOpenapiModel.ES256),
		Provider: providerInternal,
	})
	testify.NoError(t, err)

	ekID := ek.ElasticKeyID

	// Sign a valid payload to obtain a well-formed JWS compact serialization.
	jwsBytes, err := stack.service.PostSignByElasticKeyID(stack.ctx, ekID, []byte("test-payload"))
	testify.NoError(t, err)

	// JWS compact has 3 parts: header.payload.signature
	// Replace payload (part[1]) with different bytes so the signature no longer verifies.
	parts := strings.Split(string(jwsBytes), ".")
	testify.Len(t, parts, 3)

	parts[1] = testTamperedB64 // base64url for "tampered"

	tamperedJWS := []byte(strings.Join(parts, "."))

	// Verification must fail with the "failed to verify bytes" error (line 138).
	_, err = stack.service.PostVerifyByElasticKeyID(stack.ctx, ekID, tamperedJWS)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to verify bytes with MaterialKey")
}

// TestPostSign_JWEKeyType covers businesslogic_crypto.go:101 – the SignBytes
// failure path when the material key is an AES-GCM (JWE) key that cannot be
// used for JWS signing.
func TestPostSign_JWEKeyType(t *testing.T) {
	t.Parallel()

	stack := setupCryptoTestStack(t)

	// Create JWE-type EK; AddElasticKey generates the first AES material key automatically.
	ek, err := stack.service.AddElasticKey(stack.ctx, &cryptoutilKmsServer.ElasticKeyCreate{
		Name: "sign-jwe-type", Description: ptr("sign-jwe-type-test"), Algorithm: string(cryptoutilOpenapiModel.A256GCMDir),
		Provider: providerInternal,
	})
	testify.NoError(t, err)

	ekID := ek.ElasticKeyID

	// PostSign with an AES (JWE-type) key must fail at SignBytes because A256GCM
	// is not a valid JWS signing algorithm.
	_, err = stack.service.PostSignByElasticKeyID(stack.ctx, ekID, []byte("test-payload"))
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to sign bytes with latest MaterialKey")
}
