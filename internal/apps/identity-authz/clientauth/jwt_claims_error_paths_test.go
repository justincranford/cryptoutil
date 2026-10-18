// Copyright (c) 2025-2026 Justin Cranford.
//
//

package clientauth

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type jtiReplayCacheStub struct {
	err error
}

func (r jtiReplayCacheStub) Store(context.Context, string, googleUuid.UUID, time.Time) error {
	return r.err
}

func (jtiReplayCacheStub) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func (jtiReplayCacheStub) DeleteExpired(context.Context) (int64, error) {
	return 0, nil
}

func TestClientSecretJWTValidator_ValidateClaims_ErrorPaths(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	replayErr := errors.New("JTI already used")
	tests := []struct {
		name          string
		updateClaims  func(*testing.T, joseJwt.Token)
		jtiRepository cryptoutilIdentityRepository.JTIReplayCacheRepository
		wantErr       string
	}{
		{
			name: clientauthTestJWTMissingIssuer,
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				require.NoError(t, token.Remove(joseJwt.IssuerKey))
			},
			wantErr: clientauthTestJWTInvalidIssuer,
		},
		{
			name: "invalid subject",
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				require.NoError(t, token.Set(joseJwt.SubjectKey, "different-client"))
			},
			wantErr: clientauthTestJWTInvalidSubject,
		},
		{
			name: clientauthTestJWTMissingAudience,
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				require.NoError(t, token.Remove(joseJwt.AudienceKey))
			},
			wantErr: clientauthTestJWTMissingAudienceClaim,
		},
		{
			name: clientauthTestJWTInvalidAudience,
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				require.NoError(t, token.Set(joseJwt.AudienceKey, []string{"https://other.example.com/token"}))
			},
			wantErr: clientauthTestJWTInvalidAudience,
		},
		{
			name: "issued in the future",
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				future := now.Add(time.Hour)
				require.NoError(t, token.Set(joseJwt.IssuedAtKey, future))
				require.NoError(t, token.Set(joseJwt.ExpirationKey, future.Add(time.Minute)))
			},
			wantErr: clientauthTestJWTIssuedInFuture,
		},
		{
			name: "assertion lifetime exceeds limit",
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				require.NoError(t, token.Set(
					joseJwt.ExpirationKey,
					now.Add(cryptoutilSharedMagic.JWTAssertionMaxLifetime+time.Minute),
				))
			},
			wantErr: clientauthTestJWTAssertionLifetime,
		},
		{
			name: "missing JTI",
			updateClaims: func(t *testing.T, token joseJwt.Token) {
				require.NoError(t, token.Remove(joseJwt.JwtIDKey))
			},
			wantErr: clientauthTestJWTMissingJTI,
		},
		{
			name: "replayed JTI",
			jtiRepository: jtiReplayCacheStub{
				err: replayErr,
			},
			wantErr: "JTI replay detected",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token := joseJwt.New()
			require.NoError(t, token.Set(joseJwt.IssuerKey, cryptoutilSharedMagic.TestClientID))
			require.NoError(t, token.Set(joseJwt.SubjectKey, cryptoutilSharedMagic.TestClientID))
			require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
			require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(time.Minute)))
			require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
			require.NoError(t, token.Set(joseJwt.JwtIDKey, "unique-jti"))

			if tc.updateClaims != nil {
				tc.updateClaims(t, token)
			}

			validator := NewClientSecretJWTValidator(testTokenEndpointURL, tc.jtiRepository)
			client := &cryptoutilIdentityDomain.Client{
				ID:       googleUuid.Must(googleUuid.NewV7()),
				ClientID: cryptoutilSharedMagic.TestClientID,
			}
			err := validator.validateClaims(context.Background(), token, client)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
