// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"encoding/json"
	"fmt"
	"time"

	cryptoutilAppErr "cryptoutil/internal/shared/apperr"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// EncryptBytes encrypts bytes using the provided JWKs and returns a JWE message.
func EncryptBytes(jwks []joseJwk.Key, clearBytes []byte) (*joseJwe.Message, []byte, error) {
	return EncryptBytesWithContext(jwks, clearBytes, nil)
}

// EncryptBytesWithContext encrypts bytes with additional authenticated data context.
func EncryptBytesWithContext(jwks []joseJwk.Key, clearBytes []byte, context []byte) (*joseJwe.Message, []byte, error) {
	if jwks == nil {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	} else if clearBytes == nil {
		return nil, nil, fmt.Errorf("invalid clearBytes: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(clearBytes) == 0 {
		return nil, nil, fmt.Errorf("invalid clearBytes: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	for _, jwk := range jwks {
		isEncryptJWK, err := IsEncryptJWK(jwk)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid JWK: %w", err)
		} else if !isEncryptJWK {
			return nil, nil, fmt.Errorf("invalid JWK: %w", cryptoutilAppErr.ErrJWKMustBeEncryptJWK)
		}
	}

	encs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	algs := make(map[joseJwa.KeyEncryptionAlgorithm]struct{})

	jweEncryptOptions := make([]joseJwe.EncryptOption, 0, len(jwks))
	if len(jwks) > 1 { // more than one JWK requires using JSON encoding, instead of default Compact encoding
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithJSON())
	}

	jweProtectedHeaders := joseJwe.NewHeaders()
	if err := jweProtectedHeaders.Set("iat", time.Now().UTC().Unix()); err != nil {
		return nil, nil, fmt.Errorf("failed to set iat header: %w", err)
	}

	if len(context) > 0 {
		if err := jweProtectedHeaders.Set(joseJwe.AuthenticatedDataKey, context); err != nil {
			return nil, nil, fmt.Errorf("failed to set aad header: %w", err)
		}
	}

	jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithProtectedHeaders(jweProtectedHeaders))

	for i, jwk := range jwks {
		kid, err := ExtractKidUUID(jwk)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}

		enc, alg, err := ExtractAlgEncFromJWEJWK(jwk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}

		if len(encs) == 0 {
			jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithContentEncryption(*enc)) // only add CEK alg once
		}

		encs[*enc] = struct{}{} // track ContentEncryptionAlgorithm counts
		if len(encs) != 1 {     // validate that one-and-only-one ContentEncryptionAlgorithm is used across all JWKs
			return nil, nil, fmt.Errorf("can't use JWK %d 'enc' attribute; only one unique 'enc' attribute is allowed", i)
		}

		algs[*alg] = struct{}{} // track KeyEncryptionAlgorithm counts
		if len(algs) != 1 {     // validate that one-and-only-one KeyEncryptionAlgorithm is used across all JWKs
			return nil, nil, fmt.Errorf("can't use JWK %d 'alg' attribute; only one unique 'alg' attribute is allowed", i)
		}

		jweProtectedHeaders := joseJwe.NewHeaders()
		if err := jweProtectedHeaders.Set(joseJwk.KeyIDKey, kid.String()); err != nil {
			return nil, nil, fmt.Errorf("failed to set kid header: %w", err)
		}

		if err := jweProtectedHeaders.Set(`enc`, *enc); err != nil {
			return nil, nil, fmt.Errorf("failed to set enc header: %w", err)
		}

		if err := jweProtectedHeaders.Set(joseJwk.AlgorithmKey, *alg); err != nil {
			return nil, nil, fmt.Errorf("failed to set alg header: %w", err)
		}

		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithKey(*alg, jwk, joseJwe.WithPerRecipientHeaders(jweProtectedHeaders)))
	}

	jweMessageBytes, err := joseJwe.Encrypt(clearBytes, jweEncryptOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}

	return jweMessage, jweMessageBytes, nil
}

// DecryptBytes decrypts a JWE message using the provided JWKs.
func DecryptBytes(jwks []joseJwk.Key, jweMessageBytes []byte) ([]byte, error) {
	return DecryptBytesWithContext(jwks, jweMessageBytes, nil)
}

// DecryptBytesWithContext decrypts a JWE message with additional authenticated data context.
func DecryptBytesWithContext(jwks []joseJwk.Key, jweMessageBytes []byte, _ []byte) ([]byte, error) {
	if jwks == nil {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	} else if jweMessageBytes == nil {
		return nil, fmt.Errorf("invalid jweMessageBytes: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jweMessageBytes) == 0 {
		return nil, fmt.Errorf("invalid jweMessageBytes: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	for _, jwk := range jwks {
		isDecryptJWK, err := IsDecryptJWK(jwk)
		if err != nil {
			return nil, fmt.Errorf("invalid JWK: %w", err)
		} else if !isDecryptJWK {
			return nil, fmt.Errorf("invalid JWK: %w", cryptoutilAppErr.ErrJWKMustBeDecryptJWK)
		}
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}

	encs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	algs := make(map[joseJwa.KeyEncryptionAlgorithm]struct{})
	jweDecryptOptions := make([]joseJwe.DecryptOption, 0, len(jwks))

	for i, jwk := range jwks {
		enc, alg, err := ExtractAlgEncFromJWEJWK(jwk, i)
		if err != nil {
			return nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}

		encs[*enc] = struct{}{}
		if len(encs) != 1 {
			return nil, fmt.Errorf("can't use JWK %d 'enc' attribute; only one unique 'enc' attribute is allowed", i)
		}

		algs[*alg] = struct{}{} // track KeyEncryptionAlgorithm counts
		// jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithKey(*alg, jwk))
		if len(algs) != 1 { // validate that one-and-only-one KeyEncryptionAlgorithm is used across all JWKs
			return nil, fmt.Errorf("can't use JWK %d 'alg' attribute; only one unique 'alg' attribute is allowed", i)
		}
	}

	jwkSet := joseJwk.NewSet()
	if err := jwkSet.Set("keys", jwks); err != nil {
		return nil, fmt.Errorf("failed to set keys in JWK set: %w", err)
	}

	jwkSetOptions := []joseJwe.WithKeySetSuboption{joseJwe.WithRequireKid(true)}
	jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithKeySet(jwkSet, jwkSetOptions...), joseJwe.WithMessage(jweMessage))

	decryptedBytes, err := joseJwe.Decrypt(jweMessageBytes, jweDecryptOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE message bytes: %w", err)
	}

	return decryptedBytes, nil
}

// EncryptKey encrypts a content encryption key using key encryption keys.
func EncryptKey(keks []joseJwk.Key, clearCek joseJwk.Key) (*joseJwe.Message, []byte, error) {
	clearCekBytes, err := json.Marshal(clearCek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode CEK: %w", err)
	}

	return EncryptBytes(keks, clearCekBytes)
}

// DecryptKey decrypts a content encryption key using key decryption keys.
func DecryptKey(kdks []joseJwk.Key, encryptedCdkBytes []byte) (joseJwk.Key, error) {
	decryptedCdkBytes, err := DecryptBytes(kdks, encryptedCdkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CDK bytes: %w", err)
	}

	decryptedCdk, err := joseJwk.ParseKey(decryptedCdkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to derypt CDK: %w", err)
	}

	return decryptedCdk, nil
}

// JWEHeadersString returns a string representation of JWE message headers.
func JWEHeadersString(jweMessage *joseJwe.Message) (string, error) {
	if jweMessage == nil {
		return "", fmt.Errorf("invalid jweMessage: %w", cryptoutilAppErr.ErrCantBeNil)
	}

	jweHeadersString, err := json.Marshal(jweMessage.ProtectedHeaders())
	if err != nil {
		return "", fmt.Errorf("failed to marshall JWE headers: %w", err)
	}

	return string(jweHeadersString), err
}

// ExtractKidFromJWEMessage extracts the key ID from a JWE message.
func ExtractKidFromJWEMessage(jweMessage *joseJwe.Message) (*googleUuid.UUID, error) {
	if jweMessage == nil {
		return nil, fmt.Errorf("invalid jweMessage: %w", cryptoutilAppErr.ErrCantBeNil)
	}

	var kidUUIDString string

	err := jweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &kidUUIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to get kid UUID: %w", err)
	}

	kidUUID, err := googleUuid.Parse(kidUUIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid UUID: %w", err)
	}

	return &kidUUID, nil
}

// ExtractKidEncAlgFromJWEMessage extracts the key ID, encryption, and algorithm from a JWE message.
func ExtractKidEncAlgFromJWEMessage(jweMessage *joseJwe.Message) (*googleUuid.UUID, *joseJwa.ContentEncryptionAlgorithm, *joseJwa.KeyEncryptionAlgorithm, error) {
	kidUUID, err := ExtractKidFromJWEMessage(jweMessage)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get kid UUID: %w", err)
	}

	var enc joseJwa.ContentEncryptionAlgorithm

	err = jweMessage.ProtectedHeaders().Get("enc", &enc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get enc: %w", err)
	}

	var alg joseJwa.KeyEncryptionAlgorithm

	err = jweMessage.ProtectedHeaders().Get(joseJwk.AlgorithmKey, &alg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get alg: %w", err)
	}

	return kidUUID, &enc, &alg, nil
}
