// Copyright (c) 2025 Justin Cranford

package jwks

import (
	json "encoding/json"
	"errors"
	"log/slog"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		logger  bool
		keyRepo bool
		wantErr bool
	}{
		{
			name:    "valid handler creation",
			logger:  true,
			keyRepo: true,
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  false,
			keyRepo: true,
			wantErr: true,
		},
		{
			name:    "nil keyRepo",
			logger:  true,
			keyRepo: false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.Default()
			if !tc.logger {
				logger = nil
			}

			var keyRepo cryptoutilIdentityRepository.KeyRepository = &MockKeyRepository{}
			if !tc.keyRepo {
				keyRepo = nil
			}

			handler, err := NewHandler(logger, keyRepo)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		setupMock      func(*MockKeyRepository)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:   "successful GET request",
			method: http.MethodGet,
			setupMock: func(repo *MockKeyRepository) {
				// Create test RSA signing key.
				kid := googleUuid.Must(googleUuid.NewV7())
				privateKey, publicKey := generateTestRSAKeyPair(t)

				key := &cryptoutilIdentityDomain.Key{
					ID:         kid,
					Usage:      cryptoutilIdentityMagic.KeyUsageSigning,
					Algorithm:  joseJwa.RS256().String(),
					PrivateKey: privateKey,
					PublicKey:  publicKey,
					Active:     true,
				}

				repo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
					Return([]*cryptoutilIdentityDomain.Key{key}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()

				var jwksResponse struct {
					Keys []json.RawMessage `json:"keys"`
				}

				err := json.Unmarshal(body, &jwksResponse)
				require.NoError(t, err)
				require.Greater(t, len(jwksResponse.Keys), 0, "JWKS should contain at least one key")
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			setupMock: func(_ *MockKeyRepository) {
				// No mock setup needed - request rejected before repository access.
			},
			expectedStatus: http.StatusMethodNotAllowed,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()
				require.Contains(t, string(body), "Method not allowed")
			},
		},
		{
			name:   "repository error",
			method: http.MethodGet,
			setupMock: func(repo *MockKeyRepository) {
				repo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
					Return(nil, errors.New("repository error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()
				require.Contains(t, string(body), "Internal server error")
			},
		},
		{
			name:   "no active signing keys",
			method: http.MethodGet,
			setupMock: func(repo *MockKeyRepository) {
				repo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
					Return([]*cryptoutilIdentityDomain.Key{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()

				var jwksResponse struct {
					Keys []json.RawMessage `json:"keys"`
				}

				err := json.Unmarshal(body, &jwksResponse)
				require.NoError(t, err)
				require.Equal(t, 0, len(jwksResponse.Keys), "JWKS should be empty when no keys exist")
			},
		},
		{
			name:   "invalid public key in repository",
			method: http.MethodGet,
			setupMock: func(repo *MockKeyRepository) {
				// Create key with invalid PublicKey (malformed JSON).
				kid := googleUuid.Must(googleUuid.NewV7())
				key := &cryptoutilIdentityDomain.Key{
					ID:         kid,
					Usage:      cryptoutilIdentityMagic.KeyUsageSigning,
					Algorithm:  joseJwa.RS256().String(),
					PrivateKey: `{"kty":"RSA"}`, // Valid private key.
					PublicKey:  `invalid-json-not-a-jwk`,
					Active:     true,
				}

				repo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
					Return([]*cryptoutilIdentityDomain.Key{key}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()

				var jwksResponse struct {
					Keys []json.RawMessage `json:"keys"`
				}

				err := json.Unmarshal(body, &jwksResponse)
				require.NoError(t, err)
				// Invalid key should be skipped, resulting in empty JWKS.
				require.Equal(t, 0, len(jwksResponse.Keys), "JWKS should skip invalid public keys")
			},
		},
		{
			name:   "symmetric key without public key",
			method: http.MethodGet,
			setupMock: func(repo *MockKeyRepository) {
				// Create symmetric key (no PublicKey field).
				kid := googleUuid.Must(googleUuid.NewV7())
				key := &cryptoutilIdentityDomain.Key{
					ID:         kid,
					Usage:      cryptoutilIdentityMagic.KeyUsageSigning,
					Algorithm:  joseJwa.HS256().String(),
					PrivateKey: `{"kty":"oct","k":"secret"}`,
					PublicKey:  "", // No public key for symmetric algorithm.
					Active:     true,
				}

				repo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
					Return([]*cryptoutilIdentityDomain.Key{key}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				t.Helper()

				var jwksResponse struct {
					Keys []json.RawMessage `json:"keys"`
				}

				err := json.Unmarshal(body, &jwksResponse)
				require.NoError(t, err)
				// Symmetric key should be skipped (no public key), resulting in empty JWKS.
				require.Equal(t, 0, len(jwksResponse.Keys), "JWKS should skip symmetric keys without public keys")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.Default()
			keyRepo := &MockKeyRepository{}
			tc.setupMock(keyRepo)

			handler, err := NewHandler(logger, keyRepo)
			require.NoError(t, err)

			req := httptest.NewRequest(tc.method, cryptoutilIdentityMagic.PathJWKS, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			closeErr := resp.Body.Close()
			require.NoError(t, closeErr)

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.validateBody != nil {
				body := w.Body.Bytes()
				tc.validateBody(t, body)
			}

			if tc.expectedStatus == http.StatusOK {
				require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
				require.Contains(t, resp.Header.Get("Cache-Control"), "public")
			}
		})
	}
}

func generateTestRSAKeyPair(t *testing.T) (string, string) {
	t.Helper()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.RS256()

	// Generate RSA key pair using existing crypto utilities.
	_, privateJWK, publicJWK, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	// Set kid on both keys.
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, kid.String()))

	if publicJWK != nil {
		require.NoError(t, publicJWK.Set(joseJwk.KeyIDKey, kid.String()))
	}

	// Marshal to JSON strings.
	privateBytes, err := json.Marshal(privateJWK)
	require.NoError(t, err)

	publicBytes, err := json.Marshal(publicJWK)
	require.NoError(t, err)

	return string(privateBytes), string(publicBytes)
}
