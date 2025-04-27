package jose

import (
	cryptoutilAppErr "cryptoutil/internal/apperr"
	"encoding/json"
	"fmt"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	KtyOct              = joseJwa.OctetSeq()                             // KeyType
	AlgKekDIRECT        = joseJwa.DIRECT()                               // KeyEncryptionAlgorithm
	AlgKekA256GCMKW     = joseJwa.A256GCMKW()                            // KeyEncryptionAlgorithm
	AlgKekA192GCMKW     = joseJwa.A192GCMKW()                            // KeyEncryptionAlgorithm
	AlgKekA128GCMKW     = joseJwa.A128GCMKW()                            // KeyEncryptionAlgorithm
	AlgCekA256GCM       = joseJwa.A256GCM()                              // ContentEncryptionAlgorithm
	AlgCekA192GCM       = joseJwa.A192GCM()                              // ContentEncryptionAlgorithm
	AlgCekA128GCM       = joseJwa.A128GCM()                              // ContentEncryptionAlgorithm
	AlgCekA256CBC_HS512 = joseJwa.A256CBC_HS512()                        // ContentEncryptionAlgorithm
	AlgCekA192CBC_HS384 = joseJwa.A192CBC_HS384()                        // ContentEncryptionAlgorithm
	AlgCekA128CBC_HS256 = joseJwa.A128CBC_HS256()                        // ContentEncryptionAlgorithm
	OpsEncDec           = joseJwk.KeyOperationList{"encrypt", "decrypt"} // []KeyOperation
)

func EncryptBytes(jwks []joseJwk.Key, clearBytes []byte) (*joseJwe.Message, []byte, error) {
	if jwks == nil {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	cekAlgs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	jweEncryptOptions := make([]joseJwe.EncryptOption, 0, len(jwks))
	if len(jwks) > 1 {
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithJSON()) // if more than one JWK, must use JSON encoding instead of default Compact encoding
	}
	for i, jwk := range jwks {
		kekAlg, cekAlg, err := getJwkAlgAndEnc(jwk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		if len(cekAlgs) == 0 {
			jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithContentEncryption(*cekAlg)) // only add CEK alg once
		}
		cekAlgs[*cekAlg] = struct{}{} // ensure CEK alg is the same for all JWKs
		if len(cekAlgs) != 1 {
			return nil, nil, fmt.Errorf("can't use JWK %d 'enc' attributes; only one unique 'enc' attribute is allowed", i)
		}
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithKey(*kekAlg, jwk)) // add ALG+JWK tuple for each JWK
	}

	encodedJweMessage, err := joseJwe.Encrypt(clearBytes, jweEncryptOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt clearBytes: %w", err)
	}

	jweMessage, err := joseJwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse encrypted JWE message bytes: %w", err)
	}

	return jweMessage, encodedJweMessage, nil
}

func DecryptBytes(jwks []joseJwk.Key, jweMessageBytes []byte) ([]byte, error) {
	if jwks == nil {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeNil)
	} else if len(jwks) == 0 {
		return nil, fmt.Errorf("invalid JWKs: %w", cryptoutilAppErr.ErrCantBeEmpty)
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted JWE message bytes: %w", err)
	}

	cekAlgs := make(map[joseJwa.ContentEncryptionAlgorithm]struct{})
	jweDecryptOptions := make([]joseJwe.DecryptOption, 0, len(jwks))
	for i, jwk := range jwks {
		kekAlg, cekAlg, err := getJwkAlgAndEnc(jwk, i)
		if err != nil {
			return nil, fmt.Errorf("JWK %d invalid: %w", i, err)
		}
		cekAlgs[*cekAlg] = struct{}{}
		if len(cekAlgs) != 1 {
			return nil, fmt.Errorf("can't use JWK %d 'enc' attributes; only one unique 'enc' attribute is allowed", i)
		}
		jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithKey(*kekAlg, jwk))
	}
	jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithMessage(jweMessage))

	decryptedBytes, err := joseJwe.Decrypt(jweMessageBytes, jweDecryptOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWE message bytes: %w", err)
	}

	return decryptedBytes, nil
}

func EncryptKey(keks []joseJwk.Key, clearCek joseJwk.Key) (*joseJwe.Message, []byte, error) {
	clearCekBytes, err := json.Marshal(clearCek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode CEK: %w", err)
	}
	return EncryptBytes(keks, clearCekBytes)
}

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

func JSONHeadersString(jweMessage *joseJwe.Message) (string, error) {
	jweHeadersString, err := json.Marshal(jweMessage.ProtectedHeaders())
	if err != nil {
		return "", fmt.Errorf("failed to marshall JWE headers: %w", err)
	}
	return string(jweHeadersString), err
}

func getJwkAlgAndEnc(jwk joseJwk.Key, i int) (*joseJwa.KeyAlgorithm, *joseJwa.ContentEncryptionAlgorithm, error) {
	if jwk == nil {
		return nil, nil, fmt.Errorf("JWK %d invalid: %w", i, cryptoutilAppErr.ErrCantBeNil)
	}

	var kekAlg joseJwa.KeyAlgorithm
	err := jwk.Get(joseJwk.AlgorithmKey, &kekAlg) // Example: A256GCMKW, A192GCMKW, A128GCMKW, AlgDIRECT
	if err != nil {
		return nil, nil, fmt.Errorf("can't get JWK %d 'alg' attribute: %w", i, err)
	}

	var cekAlg joseJwa.ContentEncryptionAlgorithm
	err = jwk.Get("enc", &cekAlg) // Example: A256GCM, A192GCM, A128GCM, A256CBC-HS512, A192CBC-HS384, A128CBC-HS256
	if err != nil {
		// Workaround: If JWK was serialized (for encryption) and parsed (after decryption), 'enc' header incorrect gets parsed as string, so try getting as string converting it to joseJwa.ContentEncryptionAlgorithm
		var cekAlgString string
		err = jwk.Get("enc", &cekAlgString)
		if err != nil {
			return nil, nil, fmt.Errorf("can't get JWK %d 'enc' attribute: %w", i, err)
		}
		cekAlg = joseJwa.NewContentEncryptionAlgorithm(cekAlgString)
	}
	return &kekAlg, &cekAlg, nil
}
