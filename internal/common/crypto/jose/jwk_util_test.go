package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"sync"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwkTestKeys struct {
	rsaPrivJwk   joseJwk.Key
	rsaPubJwk    joseJwk.Key
	ecdsaPrivJwk joseJwk.Key
	ecdsaPubJwk  joseJwk.Key
	ecdhPrivJwk  joseJwk.Key
	ecdhPubJwk   joseJwk.Key
	okpPrivJwk   joseJwk.Key
	okpPubJwk    joseJwk.Key
	symJwk       joseJwk.Key
}

var (
	testKeys     *jwkTestKeys
	testKeysOnce sync.Once
)

func getTestKeys(t *testing.T) *jwkTestKeys {
	t.Helper()

	testKeysOnce.Do(func() {
		testKeys = &jwkTestKeys{}

		rsaGenFunc := cryptoutilKeyGen.GenerateRSAKeyPairFunction(2048)
		ecdsaGenFunc := cryptoutilKeyGen.GenerateECDSAKeyPairFunction(elliptic.P256())
		ecdhGenFunc := cryptoutilKeyGen.GenerateECDHKeyPairFunction(ecdh.P256())
		eddsaGenFunc := cryptoutilKeyGen.GenerateEDDSAKeyPairFunction("Ed25519")
		aesGenFunc := cryptoutilKeyGen.GenerateAESKeyFunction(256)

		var wg sync.WaitGroup
		type keyGenResult struct {
			keyType string
			key     interface{}
			err     error
		}
		results := make(chan keyGenResult, 5) // Buffer for all results

		wg.Add(1)
		go func() {
			defer wg.Done()
			keyPair, err := rsaGenFunc()
			results <- keyGenResult{"rsa", keyPair, err}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyPair, err := ecdsaGenFunc()
			results <- keyGenResult{"ecdsa", keyPair, err}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyPair, err := ecdhGenFunc()
			results <- keyGenResult{"ecdh", keyPair, err}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyPair, err := eddsaGenFunc()
			results <- keyGenResult{"eddsa", keyPair, err}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			key, err := aesGenFunc()
			results <- keyGenResult{"aes", key, err}
		}()
		go func() {
			wg.Wait()
			close(results)
		}()

		// Process results and generate JWKs
		var rsaKeyPair, ecdsaKeyPair, ecdhKeyPair, ed25519KeyPair *cryptoutilKeyGen.KeyPair
		var aesKey cryptoutilKeyGen.SecretKey

		for result := range results {
			require.NoError(t, result.err, "failed to generate %s key", result.keyType)

			switch result.keyType {
			case "rsa":
				rsaKeyPair = result.key.(*cryptoutilKeyGen.KeyPair)
			case "ecdsa":
				ecdsaKeyPair = result.key.(*cryptoutilKeyGen.KeyPair)
			case "ecdh":
				ecdhKeyPair = result.key.(*cryptoutilKeyGen.KeyPair)
			case "eddsa":
				ed25519KeyPair = result.key.(*cryptoutilKeyGen.KeyPair)
			case "aes":
				aesKey = result.key.(cryptoutilKeyGen.SecretKey)
			}
		}

		// Create JWKs from the generated keys concurrently
		type jwkImportResult struct {
			keyType string
			jwk     joseJwk.Key
			err     error
		}

		jwkResults := make(chan jwkImportResult, 9) // Buffer for all JWK results
		var jwkWg sync.WaitGroup

		// Helper function for concurrent JWK import
		importJWK := func(keyType string, rawKey interface{}) {
			defer jwkWg.Done()
			jwk, err := joseJwk.Import(rawKey)
			jwkResults <- jwkImportResult{keyType, jwk, err}
		}

		// Start all JWK import operations concurrently
		jwkWg.Add(9)
		go importJWK("rsaPriv", rsaKeyPair.Private.(*rsa.PrivateKey))
		go importJWK("rsaPub", rsaKeyPair.Public.(*rsa.PublicKey))
		go importJWK("ecdsaPriv", ecdsaKeyPair.Private.(*ecdsa.PrivateKey))
		go importJWK("ecdsaPub", ecdsaKeyPair.Public.(*ecdsa.PublicKey))
		go importJWK("ecdhPriv", ecdhKeyPair.Private.(*ecdh.PrivateKey))
		go importJWK("ecdhPub", ecdhKeyPair.Public.(*ecdh.PublicKey))
		go importJWK("okpPriv", ed25519KeyPair.Private.(ed25519.PrivateKey))
		go importJWK("okpPub", ed25519KeyPair.Public.(ed25519.PublicKey))
		go importJWK("sym", []byte(aesKey))

		// Close jwkResults channel when all JWK imports are done
		go func() {
			jwkWg.Wait()
			close(jwkResults)
		}()

		// Process JWK import results
		for result := range jwkResults {
			require.NoError(t, result.err, "failed to import %s key to JWK", result.keyType)

			switch result.keyType {
			case "rsaPriv":
				testKeys.rsaPrivJwk = result.jwk
			case "rsaPub":
				testKeys.rsaPubJwk = result.jwk
			case "ecdsaPriv":
				testKeys.ecdsaPrivJwk = result.jwk
			case "ecdsaPub":
				testKeys.ecdsaPubJwk = result.jwk
			case "ecdhPriv":
				testKeys.ecdhPrivJwk = result.jwk
			case "ecdhPub":
				testKeys.ecdhPubJwk = result.jwk
			case "okpPriv":
				testKeys.okpPrivJwk = result.jwk
			case "okpPub":
				testKeys.okpPubJwk = result.jwk
			case "sym":
				testKeys.symJwk = result.jwk
			}
		}
	})

	return testKeys
}

func TestIsPrivateJwk(t *testing.T) {
	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

	keys := getTestKeys(t)
	tests := []testCase{
		{
			name:     "nil JWK",
			jwk:      nil,
			expected: false,
			wantErr:  cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:     "RSA private key",
			jwk:      keys.rsaPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      keys.rsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      keys.ecdsaPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecdsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH private key",
			jwk:      keys.ecdhPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDH public key",
			jwk:      keys.ecdhPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      keys.okpPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      keys.okpPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      keys.symJwk,
			expected: false,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPrivate, err := IsPrivateJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isPrivate)
			}
		})
	}
}

func TestIsPublicJwk(t *testing.T) {
	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

	keys := getTestKeys(t)
	tests := []testCase{
		{
			name:     "nil JWK",
			jwk:      nil,
			expected: false,
			wantErr:  cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:     "RSA private key",
			jwk:      keys.rsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      keys.rsaPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      keys.ecdsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecdsaPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDH private key",
			jwk:      keys.ecdhPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH public key",
			jwk:      keys.ecdhPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      keys.okpPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      keys.okpPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      keys.symJwk,
			expected: false,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPublic, err := IsPublicJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isPublic)
			}
		})
	}
}

func TestIsSymmetricJwk(t *testing.T) {
	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

	keys := getTestKeys(t)
	tests := []testCase{
		{
			name:     "nil JWK",
			jwk:      nil,
			expected: false,
			wantErr:  cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:     "RSA private key",
			jwk:      keys.rsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      keys.rsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      keys.ecdsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecdsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH private key",
			jwk:      keys.ecdhPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH public key",
			jwk:      keys.ecdhPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      keys.okpPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      keys.okpPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      keys.symJwk,
			expected: true,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSymmetric, err := IsSymmetricJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isSymmetric)
			}
		})
	}
}
