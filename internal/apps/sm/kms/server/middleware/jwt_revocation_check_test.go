package middleware

import (
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestCheckRevocation_MockServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		response   string
		statusCode int
		wantActive bool
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "active token",
			response:   `{"active":true}`,
			statusCode: http.StatusOK,
			wantActive: true,
			wantErr:    false,
		},
		{
			name:       "revoked token",
			response:   `{"active":false}`,
			statusCode: http.StatusOK,
			wantActive: false,
			wantErr:    false,
		},
		{
			name:       "server error",
			response:   `Internal Server Error`,
			statusCode: http.StatusInternalServerError,
			wantActive: false,
			wantErr:    true,
			errMsg:     "introspection returned status 500",
		},
		{
			name:       "invalid json response",
			response:   `not-json`,
			statusCode: http.StatusOK,
			wantActive: false,
			wantErr:    true,
			errMsg:     "failed to parse introspection response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			introspectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, writeErr := w.Write([]byte(tc.response))
				require.NoError(t, writeErr)
			}))
			t.Cleanup(introspectionServer.Close)

			validator, err := NewJWTValidator(JWTValidatorConfig{
				JWKSURL:          "https://localhost/.well-known/jwks.json",
				IntrospectionURL: introspectionServer.URL,
			})
			require.NoError(t, err)

			active, err := validator.checkRevocation(t.Context(), "test-token")

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantActive, active)
			}
		})
	}
}

func TestCheckRevocation_WithClientAuth(t *testing.T) {
	t.Parallel()

	var receivedAuth string

	introspectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusOK)

		_, writeErr := w.Write([]byte(`{"active":true}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(introspectionServer.Close)

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:                   "https://localhost/.well-known/jwks.json",
		IntrospectionURL:          introspectionServer.URL,
		IntrospectionClientID:     "client-id",
		IntrospectionClientSecret: "client-secret",
	})
	require.NoError(t, err)

	active, err := validator.checkRevocation(t.Context(), "test-token")
	require.NoError(t, err)
	require.True(t, active)
	require.NotEmpty(t, receivedAuth)
	require.Contains(t, receivedAuth, "Basic")
}

func TestGetJWKS_CacheHit(t *testing.T) {
	t.Parallel()

	fetchCount := 0

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetchCount++

		w.Header().Set("Content-Type", "application/json")

		keySet := joseJwk.NewSet()
		buf, marshalErr := json.Marshal(keySet)
		require.NoError(t, marshalErr)

		_, writeErr := w.Write(buf)
		require.NoError(t, writeErr)
	}))
	t.Cleanup(jwksServer.Close)

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:  jwksServer.URL,
		CacheTTL: 1 * time.Hour,
	})
	require.NoError(t, err)

	// First call should fetch.
	_, err = validator.getJWKS(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, fetchCount)

	// Second call should use cache.
	_, err = validator.getJWKS(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, fetchCount)
}

func TestPerformRevocationCheck_Modes(t *testing.T) {
	t.Parallel()

	introspectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, writeErr := w.Write([]byte(`{"active":true}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(introspectionServer.Close)

	tests := []struct {
		name    string
		mode    RevocationCheckMode
		claims  *JWTClaims
		wantErr bool
	}{
		{
			name:    "disabled mode skips check",
			mode:    RevocationCheckDisabled,
			claims:  &JWTClaims{Scopes: []string{"read"}},
			wantErr: false,
		},
		{
			name: "every-request mode performs check",
			mode: RevocationCheckEveryRequest,
			claims: &JWTClaims{
				Subject: "user-1",
				Scopes:  []string{"read"},
			},
			wantErr: false,
		},
		{
			name: "sensitive-only mode skips non-sensitive",
			mode: RevocationCheckSensitiveOnly,
			claims: &JWTClaims{
				Subject: "user-2",
				Scopes:  []string{"read"},
			},
			wantErr: false,
		},
		{
			name: "sensitive-only mode checks sensitive scope",
			mode: RevocationCheckSensitiveOnly,
			claims: &JWTClaims{
				Subject: "user-3",
				Scopes:  []string{"admin"},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator, err := NewJWTValidator(JWTValidatorConfig{
				JWKSURL:             "https://localhost/.well-known/jwks.json",
				RevocationCheckMode: tc.mode,
				IntrospectionURL:    introspectionServer.URL,
				SensitiveScopes:     []string{"admin", "write"},
			})
			require.NoError(t, err)

			err = validator.performRevocationCheck(t.Context(), "test-token", tc.claims)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPerformRevocationCheck_RevokedToken(t *testing.T) {
	t.Parallel()

	introspectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, writeErr := w.Write([]byte(`{"active":false}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(introspectionServer.Close)

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:             "https://localhost/.well-known/jwks.json",
		RevocationCheckMode: RevocationCheckEveryRequest,
		IntrospectionURL:    introspectionServer.URL,
	})
	require.NoError(t, err)

	err = validator.performRevocationCheck(t.Context(), "revoked-token", &JWTClaims{Subject: "user-1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "revoked")
}
