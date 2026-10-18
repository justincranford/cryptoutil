// Copyright (c) 2025-2026 Justin Cranford.

package apis

import (
	"context"
	"errors"
	"io"
	"log/slog"
	http "net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	backchannelTestIssuer           = "https://issuer.example"
	backchannelSuccessPath          = "/success"
	backchannelFailurePath          = "/failure"
	backchannelTransportErrPath     = "/transport-error"
	backchannelInvalidURI           = "://invalid"
	backchannelExpectedFailureCount = 3
)

type backchannelRoundTripper func(*http.Request) (*http.Response, error)

func (f backchannelRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func newBackchannelTestTokenService(t *testing.T, withSigningKey bool) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	ctx := context.Background()
	keyRotationManager, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	if withSigningKey {
		require.NoError(t, keyRotationManager.RotateSigningKey(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm))
	}

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		backchannelTestIssuer,
		keyRotationManager,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		time.Hour,
		time.Hour,
	)
	require.NoError(t, err)

	return cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, nil, nil, &cryptoutilIdentityConfig.TokenConfig{
		Issuer:           backchannelTestIssuer,
		SigningAlgorithm: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
	})
}

func backchannelTestSession() *cryptoutilIdentityDomain.Session {
	return &cryptoutilIdentityDomain.Session{
		SessionID: googleUuid.Must(googleUuid.NewV7()).String(),
		UserID:    googleUuid.Must(googleUuid.NewV7()),
	}
}

func backchannelTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestBackChannelLogoutService_SendPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		withSigningKey bool
		clients        []*cryptoutilIdentityDomain.Client
		statusCodes    map[string]int
		transportErrs  map[string]error
		wantSuccess    int
		wantErrors     int
		wantToken      bool
	}{
		{
			name:           "skips clients without logout URI",
			withSigningKey: true,
			clients:        []*cryptoutilIdentityDomain.Client{{ClientID: "no-uri"}},
			wantSuccess:    0,
		},
		{
			name:       "reports token generation failures",
			clients:    []*cryptoutilIdentityDomain.Client{{ClientID: "no-signing-key", BackChannelLogoutURI: "https://client.example/logout"}},
			wantErrors: 1,
		},
		{
			name:           "delivers tokens and aggregates delivery errors",
			withSigningKey: true,
			clients: []*cryptoutilIdentityDomain.Client{
				{ClientID: "bad-uri", BackChannelLogoutURI: backchannelInvalidURI},
				{ClientID: "server-error", BackChannelLogoutURI: "https://client.example" + backchannelFailurePath},
				{ClientID: "network-error", BackChannelLogoutURI: "https://client.example" + backchannelTransportErrPath},
				{ClientID: "successful-client", BackChannelLogoutURI: "https://client.example" + backchannelSuccessPath, BackChannelLogoutSessionRequired: boolPtr(true)},
			},
			statusCodes:   map[string]int{backchannelFailurePath: http.StatusServiceUnavailable, backchannelSuccessPath: http.StatusNoContent},
			transportErrs: map[string]error{backchannelTransportErrPath: errors.New("transport failed")},
			wantSuccess:   1,
			wantErrors:    backchannelExpectedFailureCount,
			wantToken:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := NewBackChannelLogoutService(newBackchannelTestTokenService(t, tc.withSigningKey), backchannelTestIssuer, backchannelTestLogger())

			var deliveredToken string

			svc.httpClient = &http.Client{
				Transport: backchannelRoundTripper(func(request *http.Request) (*http.Response, error) {
					if err := tc.transportErrs[request.URL.Path]; err != nil {
						return nil, err
					}

					body, err := io.ReadAll(request.Body)
					require.NoError(t, err)
					form, err := url.ParseQuery(string(body))
					require.NoError(t, err)
					require.Equal(t, http.MethodPost, request.Method)
					require.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))

					if request.URL.Path == backchannelSuccessPath {
						deliveredToken = form.Get("logout_token")
					}

					statusCode := tc.statusCodes[request.URL.Path]
					if statusCode == 0 {
						statusCode = http.StatusOK
					}

					return &http.Response{
						StatusCode: statusCode,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("")),
						Request:    request,
					}, nil
				}),
			}

			session := backchannelTestSession()
			successCount, errs := svc.SendBackChannelLogout(context.Background(), session, tc.clients)
			require.Equal(t, tc.wantSuccess, successCount)
			require.Len(t, errs, tc.wantErrors)

			if !tc.wantToken {
				return
			}

			require.NotEmpty(t, deliveredToken)
			claims, err := svc.tokenService.ValidateIDToken(context.Background(), deliveredToken)
			require.NoError(t, err)
			require.Equal(t, backchannelTestIssuer, claims[cryptoutilSharedMagic.ClaimIss])
			require.Equal(t, "successful-client", claims[cryptoutilSharedMagic.ClaimAud])
			require.Equal(t, session.UserID.String(), claims[cryptoutilSharedMagic.ClaimSub])
		})
	}
}
