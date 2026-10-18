// Copyright (c) 2025-2026 Justin Cranford.
package clientauth

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	clientauthTestClientLookupError          = "failed to get client"
	clientauthTestInvalidHashClientID        = "malformed-hash-client"
	clientauthTestProvidedSecret             = "provided-secret"
	clientauthTestInvalidBasicMethodClientID = "wrong-basic-method"
	clientauthTestInvalidPostMethodClientID  = "wrong-post-method"
	clientauthTestJWTMissingIssuer           = "missing issuer"
	clientauthTestJWTInvalidIssuer           = "invalid issuer"
	clientauthTestJWTInvalidSubject          = "invalid subject"
	clientauthTestOtherClient                = "other-client"
	clientauthTestJWTMissingAudience         = "missing audience"
	clientauthTestJWTMissingAudienceClaim    = "missing audience claim"
	clientauthTestJWTInvalidAudience         = "invalid audience"
	clientauthTestJWTIssuedInFuture          = "JWT issued in the future"
	clientauthTestJWTAssertionLifetime       = "assertion lifetime"
	clientauthTestJWTMissingJTI              = "missing jti"
	clientauthTestInvalidSaltEncoding        = "invalid salt encoding"
	clientauthTestInvalidHashEncoding        = "invalid hash encoding"
	clientauthTestPlaintextSecret            = "plaintext"
	clientauthTestHashedSecret               = "$pbkdf2-sha256$100000$salt$hash"
	clientauthTestClientDisabledError        = "client is disabled"
	clientauthTestDisabledClientCase         = "disabled client"
	clientauthTestInvalidCredentialsError    = "invalid client credentials"
)

type lookupErrorClientRepository struct {
	mockClientRepo
	err error
}

func (r *lookupErrorClientRepository) GetByClientID(ctx context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.mockClientRepo.GetByClientID(ctx, clientID)
}

func TestBasicAndPostAuthenticators_RepositoryAndCredentialErrors(t *testing.T) {
	t.Parallel()

	repositoryErr := errors.New("repository unavailable")
	validStoredSecret, err := cryptoutilSharedCryptoHash.HashSecretPBKDF2("provided-secret")
	require.NoError(t, err)

	tests := []struct {
		name             string
		client           *cryptoutilIdentityDomain.Client
		clientID         string
		repository       *lookupErrorClientRepository
		secret           string
		wantErr          error
		wantErrContains  string
		wantBasicErr     bool
		wantPostErr      bool
		wantBasicSuccess bool
		wantPostSuccess  bool
	}{
		{
			name:            "repository failure",
			clientID:        "repository-failure",
			repository:      &lookupErrorClientRepository{err: repositoryErr},
			wantErrContains: clientauthTestClientLookupError,
		},
		{
			name: "malformed stored hash",
			client: &cryptoutilIdentityDomain.Client{
				ClientID:     clientauthTestInvalidHashClientID,
				ClientSecret: "not-a-pbkdf2-hash",
			},
			clientID:        clientauthTestInvalidHashClientID,
			secret:          clientauthTestProvidedSecret,
			wantErrContains: "failed to compare client secret",
		},
		{
			name: "invalid basic method",
			client: &cryptoutilIdentityDomain.Client{
				ClientID:                clientauthTestInvalidBasicMethodClientID,
				ClientSecret:            validStoredSecret,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
			},
			clientID:        clientauthTestInvalidBasicMethodClientID,
			secret:          clientauthTestProvidedSecret,
			wantErr:         cryptoutilIdentityAppErr.ErrInvalidClientAuth,
			wantBasicErr:    true,
			wantPostSuccess: true,
		},
		{
			name: "invalid post method",
			client: &cryptoutilIdentityDomain.Client{
				ClientID:                clientauthTestInvalidPostMethodClientID,
				ClientSecret:            validStoredSecret,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
			},
			clientID:         clientauthTestInvalidPostMethodClientID,
			secret:           clientauthTestProvidedSecret,
			wantErr:          cryptoutilIdentityAppErr.ErrInvalidClientAuth,
			wantPostErr:      true,
			wantBasicSuccess: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repository := tc.repository
			if repository == nil {
				repository = &lookupErrorClientRepository{
					mockClientRepo: mockClientRepo{
						clients: map[string]*cryptoutilIdentityDomain.Client{
							tc.client.ClientID: tc.client,
						},
					},
				}
			}

			basic := NewBasicAuthenticator(repository)
			post := NewPostAuthenticator(repository)

			secret := tc.secret
			if secret == "" {
				secret = clientauthTestProvidedSecret
			}

			basicClient, basicErr := basic.Authenticate(context.Background(), tc.clientID, secret)

			postClient, postErr := post.Authenticate(context.Background(), tc.clientID, secret)
			if tc.wantErr != nil && (tc.wantBasicErr || (!tc.wantBasicErr && !tc.wantPostErr)) {
				require.ErrorIs(t, basicErr, tc.wantErr)
			}

			if tc.wantErr != nil && (tc.wantPostErr || (!tc.wantBasicErr && !tc.wantPostErr)) {
				require.ErrorIs(t, postErr, tc.wantErr)
			}

			if tc.wantErrContains != "" {
				require.ErrorContains(t, basicErr, tc.wantErrContains)
				require.ErrorContains(t, postErr, tc.wantErrContains)
			}

			require.Equal(t, tc.wantBasicSuccess, basicErr == nil)
			require.Equal(t, tc.wantPostSuccess, postErr == nil)

			if basicErr == nil {
				require.NotNil(t, basicClient)
			} else {
				require.Nil(t, basicClient)
			}

			if postErr == nil {
				require.NotNil(t, postClient)
			} else {
				require.Nil(t, postClient)
			}
		})
	}
}

func TestClientSecretJWTValidator_ClaimValidationErrors(t *testing.T) {
	t.Parallel()

	client := &cryptoutilIdentityDomain.Client{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     "claim-validation-client",
		ClientSecret: testClientSecretStr,
	}
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)
	claimIssuer := joseJwt.IssuerKey
	claimSubject := joseJwt.SubjectKey
	claimAudience := joseJwt.AudienceKey
	claimExpiration := joseJwt.ExpirationKey
	claimIssuedAt := joseJwt.IssuedAtKey
	claimID := joseJwt.JwtIDKey

	tests := []struct {
		name          string
		omit          string
		overrides     map[string]any
		wantErrSubstr string
	}{
		{name: clientauthTestJWTMissingIssuer, omit: claimIssuer, wantErrSubstr: clientauthTestJWTInvalidIssuer},
		{name: clientauthTestJWTInvalidIssuer, overrides: map[string]any{claimIssuer: clientauthTestOtherClient}, wantErrSubstr: clientauthTestJWTInvalidIssuer},
		{name: "missing subject", omit: claimSubject, wantErrSubstr: clientauthTestJWTInvalidSubject},
		{name: clientauthTestJWTInvalidSubject, overrides: map[string]any{claimSubject: clientauthTestOtherClient}, wantErrSubstr: clientauthTestJWTInvalidSubject},
		{name: clientauthTestJWTMissingAudience, omit: claimAudience, wantErrSubstr: clientauthTestJWTMissingAudienceClaim},
		{name: clientauthTestJWTInvalidAudience, overrides: map[string]any{claimAudience: []string{"https://elsewhere.example.test"}}, wantErrSubstr: clientauthTestJWTInvalidAudience},
		{name: "missing expiration", omit: claimExpiration, wantErrSubstr: "missing expiration claim"},
		{name: "expired", overrides: map[string]any{claimExpiration: time.Now().UTC().Add(-time.Second)}, wantErrSubstr: "JWT expired"},
		{name: "missing issued at", omit: claimIssuedAt, wantErrSubstr: "missing issued at claim"},
		{name: "future issued at", overrides: map[string]any{claimIssuedAt: time.Now().UTC().Add(time.Hour)}, wantErrSubstr: clientauthTestJWTIssuedInFuture},
		{name: "missing jti", omit: claimID, wantErrSubstr: clientauthTestJWTMissingJTI},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := time.Now().UTC()
			claims := map[string]any{
				claimIssuer:     client.ClientID,
				claimSubject:    client.ClientID,
				claimAudience:   []string{testTokenEndpointURL},
				claimExpiration: now.Add(cryptoutilSharedMagic.JWTAssertionMaxLifetime / 2),
				claimIssuedAt:   now,
				claimID:         googleUuid.NewString(),
			}
			delete(claims, tc.omit)

			for claim, value := range tc.overrides {
				claims[claim] = value
			}

			token := joseJwt.New()
			for claim, value := range claims {
				require.NoError(t, token.Set(claim, value))
			}

			err := validator.validateClaims(context.Background(), token, client)
			require.ErrorContains(t, err, tc.wantErrSubstr)
		})
	}
}

func TestClientSecretJWTValidator_ValidateJWT_MissingSecretAndMalformedToken(t *testing.T) {
	t.Parallel()

	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)
	tests := []struct {
		name          string
		client        *cryptoutilIdentityDomain.Client
		token         string
		wantErrSubstr string
	}{
		{
			name:          "no client secret",
			client:        &cryptoutilIdentityDomain.Client{ClientID: "missing-secret"},
			token:         cryptoutilSharedMagic.ParamToken,
			wantErrSubstr: "client has no secret configured",
		},
		{
			name: "malformed signed token",
			client: &cryptoutilIdentityDomain.Client{
				ClientID:     "malformed-token",
				ClientSecret: testClientSecretStr,
			},
			token:         "not-a-jwt",
			wantErrSubstr: "failed to parse and verify JWT",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := validator.ValidateJWT(context.Background(), tc.token, tc.client)
			require.ErrorContains(t, err, tc.wantErrSubstr)
		})
	}
}

func TestPBKDF2Hasher_CompareSecret_MalformedEncodedFields(t *testing.T) {
	t.Parallel()

	hasher := NewPBKDF2Hasher()
	prefix := "$" + cryptoutilSharedMagic.PBKDF2DefaultHashName + "$"
	tests := []struct {
		name       string
		encoded    string
		wantErrMsg string
	}{
		{name: "invalid iteration count", encoded: prefix + "not-a-number$c2FsdA$aGFzaA", wantErrMsg: "invalid iterations"},
		{name: clientauthTestInvalidSaltEncoding, encoded: prefix + "100000$%%%$aGFzaA", wantErrMsg: clientauthTestInvalidSaltEncoding},
		{name: clientauthTestInvalidHashEncoding, encoded: prefix + "100000$c2FsdA$%%%", wantErrMsg: clientauthTestInvalidHashEncoding},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, hasher.CompareSecret(tc.encoded, "secret"), tc.wantErrMsg)
		})
	}
}
