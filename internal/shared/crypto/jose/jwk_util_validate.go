// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	rsa "crypto/rsa"
	"fmt"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// JWK algorithm and key type constants and error values.
func validateOrGenerateRSAJWK(key cryptoutilSharedCryptoKeygen.Key, keyBitsLength int) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
	if key == nil {
		generatedKey, err := generateRSAKeyPair(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA %d key: %w", keyBitsLength, err)
		}

		return generatedKey, nil
	}

	keyPair, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
	}

	rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *rsa.PrivateKey", keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("invalid nil RSA private key")
	}

	rsaPublicKey, ok := keyPair.Public.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *rsa.PublicKey", keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("invalid nil RSA public key")
	}

	return keyPair, nil
}

func validateOrGenerateEcdsaJWK(key cryptoutilSharedCryptoKeygen.Key, curve elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
	if key == nil {
		generatedKey, err := generateECDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA %s key pair: %w", curve, err)
		}

		return generatedKey, nil
	}

	keyPair, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
	}

	rsaPrivateKey, ok := keyPair.Private.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *ecdsa.PrivateKey", keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("invalid nil ECDSA private key")
	}

	rsaPublicKey, ok := keyPair.Public.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use *ecdsa.PublicKey", keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("invalid nil ECDSA public key")
	}

	return keyPair, nil
}

func validateOrGenerateEddsaJWK(key cryptoutilSharedCryptoKeygen.Key, curve string) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
	if key == nil {
		generatedKey, err := generateEDDSAKeyPair(curve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ed29919 key pair: %w", err)
		}

		return generatedKey, nil
	}

	keyPair, ok := key.(*cryptoutilSharedCryptoKeygen.KeyPair)
	if !ok {
		return nil, fmt.Errorf("unsupported key type %T; use *cryptoutilKeyGen.KeyPair", key)
	}

	rsaPrivateKey, ok := keyPair.Private.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use ed25519.PrivateKey", keyPair.Private)
	}

	if rsaPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("invalid nil Ed29919 private key")
	}

	rsaPublicKey, ok := keyPair.Public.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use ed25519.PublicKey", keyPair.Public)
	}

	if rsaPublicKey == nil {
		return nil, fmt.Errorf("invalid nil Ed29919 public key")
	}

	return keyPair, nil
}

func validateOrGenerateHMACJWK(key cryptoutilSharedCryptoKeygen.Key, keyBitsLength int) (cryptoutilSharedCryptoKeygen.SecretKey, error) { // pragma: allowlist secret
	if key == nil {
		generatedKey, err := generateHMACKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate HMAC %d key: %w", keyBitsLength, err)
		}

		return generatedKey, nil
	}

	hmacKey, ok := key.(cryptoutilSharedCryptoKeygen.SecretKey) // pragma: allowlist secret
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use cryptoKeygen.SecretKey", key) // pragma: allowlist secret
	}

	if hmacKey == nil {
		return nil, fmt.Errorf("invalid nil key bytes")
	}

	if len(hmacKey) != keyBitsLength/cryptoutilSharedMagic.BitsToBytes {
		return nil, fmt.Errorf("invalid key length %d; use HMAC %d", len(hmacKey), keyBitsLength)
	}

	return hmacKey, nil
}

func validateOrGenerateAESJWK(key cryptoutilSharedCryptoKeygen.Key, keyBitsLength int) (cryptoutilSharedCryptoKeygen.SecretKey, error) { // pragma: allowlist secret
	if key == nil {
		generatedKey, err := generateAESKey(keyBitsLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AES %d key: %w", keyBitsLength, err)
		}

		return generatedKey, nil
	}

	aesKey, ok := key.(cryptoutilSharedCryptoKeygen.SecretKey) // pragma: allowlist secret
	if !ok {
		return nil, fmt.Errorf("invalid key type %T; use cryptoKeygen.SecretKey", key) // pragma: allowlist secret
	}

	if aesKey == nil {
		return nil, fmt.Errorf("invalid nil key bytes")
	}

	if len(aesKey) != keyBitsLength/cryptoutilSharedMagic.BitsToBytes {
		return nil, fmt.Errorf("invalid key length %d; use AES %d", len(aesKey), keyBitsLength)
	}

	return aesKey, nil
}

// IsPublicJWK returns true if the JWK is a public key.
func IsPublicJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.ECDSAPrivateKey, joseJwk.OKPPrivateKey, joseJwk.SymmetricKey:
		return false, nil
	case joseJwk.RSAPublicKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPublicKey:
		return true, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsPrivateJWK returns true if the JWK is a private key.
func IsPrivateJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.ECDSAPrivateKey, joseJwk.OKPPrivateKey:
		return true, nil
	case joseJwk.RSAPublicKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPublicKey, joseJwk.SymmetricKey:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsAsymmetricJWK returns true if the JWK is an asymmetric key.
func IsAsymmetricJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.RSAPublicKey, joseJwk.ECDSAPrivateKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return true, nil
	case joseJwk.SymmetricKey:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsSymmetricJWK returns true if the JWK is a symmetric key.
func IsSymmetricJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("invalid jwk: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	switch jwk.(type) {
	case joseJwk.RSAPrivateKey, joseJwk.RSAPublicKey, joseJwk.ECDSAPrivateKey, joseJwk.ECDSAPublicKey, joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return false, nil
	case joseJwk.SymmetricKey:
		return true, nil
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsEncryptJWK returns true if the JWK can be used for encryption.
func IsEncryptJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, _, err := ExtractAlgEncFromJWEJWK(jwk, 0)
	if err != nil {
		// Missing enc/alg headers means not an encryption key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	// At this point, JWK is confirmed to have an enc header
	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return false, nil // private key can be used for encrypt, but shouldn't be used in practice
	case joseJwk.RSAPrivateKey:
		return false, nil // private key can be used for encrypt, but shouldn't be used in practice
	case joseJwk.ECDSAPublicKey:
		return true, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but encrypt alg header narrows it down to ECDH
	case joseJwk.RSAPublicKey:
		return true, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to AES only
	case joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return false, nil // Ed25519/Ed448 are signing only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsDecryptJWK returns true if the JWK can be used for decryption.
func IsDecryptJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, _, err := ExtractAlgEncFromJWEJWK(jwk, 0)
	if err != nil {
		// Missing enc/alg headers means not a decryption key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	// At this point, JWK is confirmed to have an enc header
	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return true, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but encrypt alg header narrows it down to ECDH
	case joseJwk.RSAPrivateKey:
		return true, nil
	case joseJwk.ECDSAPublicKey:
		return false, nil
	case joseJwk.RSAPublicKey:
		return false, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to AES only
	case joseJwk.OKPPrivateKey, joseJwk.OKPPublicKey:
		return false, nil // Ed25519/Ed448 are signing only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsSignJWK returns true if the JWK can be used for signing.
func IsSignJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, err := ExtractAlgFromJWSJWK(jwk, 0)
	if err != nil {
		// Missing alg header means not a signing key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return true, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but signature alg header narrows it down to ECDSA
	case joseJwk.RSAPrivateKey:
		return true, nil
	case joseJwk.OKPPrivateKey:
		return true, nil
	case joseJwk.ECDSAPublicKey:
		return false, nil
	case joseJwk.RSAPublicKey:
		return false, nil
	case joseJwk.OKPPublicKey:
		return false, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to HMAC only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// IsVerifyJWK returns true if the JWK can be used for signature verification.
func IsVerifyJWK(jwk joseJwk.Key) (bool, error) {
	if jwk == nil {
		return false, fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	_, err := ExtractAlgFromJWSJWK(jwk, 0)
	if err != nil {
		// Missing alg header means not a verification key
		return false, nil //nolint:nilerr // Missing headers = incompatible key type, not an error
	}

	switch jwk.(type) {
	case joseJwk.ECDSAPrivateKey:
		return false, nil // jwx uses ECDSAPrivateKey for both ECDSA or ECDH, but signature alg header narrows it down to ECDSA
	case joseJwk.RSAPrivateKey:
		return false, nil
	case joseJwk.OKPPrivateKey:
		return false, nil
	case joseJwk.RSAPublicKey:
		return true, nil
	case joseJwk.ECDSAPublicKey:
		return true, nil
	case joseJwk.OKPPublicKey:
		return true, nil
	case joseJwk.SymmetricKey:
		return true, nil // jwx SymmetricKey can be AES and HMAC, but enc header narrows it down to HMAC only
	default:
		return false, fmt.Errorf("unsupported JWK type %T", jwk)
	}
}

// EnsureSignatureAlgorithmType validates that the JWK has a recognized signature algorithm.
// In JWX v3, algorithm fields are stored as typed structs (not strings), so this function
// gets the algorithm as a typed joseJwa.SignatureAlgorithm and validates it is supported.
func EnsureSignatureAlgorithmType(jwk joseJwk.Key) error {
	if jwk == nil {
		return fmt.Errorf("JWK invalid: %w", cryptoutilSharedApperr.ErrCantBeNil)
	}

	// Get the algorithm as a typed SignatureAlgorithm (JWX v3 stores typed algorithm structs).
	var alg joseJwa.SignatureAlgorithm

	err := jwk.Get(joseJwk.AlgorithmKey, &alg)
	if err != nil {
		return fmt.Errorf("failed to get algorithm from JWK: %w", err)
	}

	// Validate that the algorithm is a known supported signature algorithm.
	switch alg.String() {
	case algStrHS256, algStrHS384, algStrHS512,
		algStrRS256, algStrRS384, algStrRS512,
		algStrPS256, algStrPS384, algStrPS512,
		algStrES256, algStrES384, algStrES512,
		algStrEdDSA:
		// Valid signature algorithm, already properly typed.
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", alg)
	}

	// Set the properly typed algorithm back on the JWK to ensure type consistency.
	err = jwk.Set(joseJwk.AlgorithmKey, alg)
	if err != nil {
		return fmt.Errorf("failed to set typed algorithm on JWK: %w", err)
	}

	return nil
}
