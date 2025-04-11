package jose

import (
	"fmt"

	"encoding/json"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	KtyOct              = joseJwa.OctetSeq()                             // KeyType
	AlgDIRECT           = joseJwa.DIRECT()                               // KeyEncryptionAlgorithm
	AlgA256GCMKW        = joseJwa.A256GCMKW()                            // KeyEncryptionAlgorithm
	AlgA256GCM          = joseJwa.A256GCM()                              // ContentEncryptionAlgorithm
	OpsEncDec           = joseJwk.KeyOperationList{"encrypt", "decrypt"} // []KeyOperation
	ErrCantBeNil        = fmt.Errorf("jwk can't be nil")
	ErrCantBeEmpty      = fmt.Errorf("jwks can't be empty")
	ErrKidCantBeNilUuid = fmt.Errorf("jwk kid can't be nil uuid")
	ErrKidCantBeMaxUuid = fmt.Errorf("jwk kid can't be max uuid")
)

func EncryptBytes(ceks []joseJwk.Key, clearBytes []byte) (*joseJwe.Message, []byte, error) {
	if ceks == nil {
		return nil, nil, ErrCantBeNil
	} else if len(ceks) == 0 {
		return nil, nil, ErrCantBeEmpty
	}

	jweEncryptOptions := make([]joseJwe.EncryptOption, 0, len(ceks))
	if len(ceks) > 1 {
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithJSON()) // use JSON encoding instead of default Compact encoding
	}
	for _, cek := range ceks {
		var alg joseJwa.KeyAlgorithm
		err := cek.Get(joseJwk.AlgorithmKey, &alg)
		if err != nil {
			return nil, nil, fmt.Errorf("algorithm not found in CEK: %w", err)
		}
		jweEncryptOptions = append(jweEncryptOptions, joseJwe.WithKey(alg, cek))
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

func DecryptBytes(cdks []joseJwk.Key, jweMessageBytes []byte) ([]byte, error) {
	if cdks == nil {
		return nil, ErrCantBeNil
	} else if len(cdks) == 0 {
		return nil, ErrCantBeEmpty
	}

	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted JWE message bytes: %w", err)
	}

	jweDecryptOptions := make([]joseJwe.DecryptOption, 0, len(cdks))
	for i, cdk := range cdks {
		var alg joseJwa.KeyAlgorithm
		err := cdk.Get(joseJwk.AlgorithmKey, &alg) // Example: A256GCMKW
		if err != nil {
			return nil, fmt.Errorf("algorithm not found in CDK %d: %w", i, err)
		}
		jweDecryptOptions = append(jweDecryptOptions, joseJwe.WithKey(alg, cdk))
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
