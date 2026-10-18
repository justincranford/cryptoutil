// Copyright (c) 2025-2026 Justin Cranford.
package userauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

const (
	testCaseStoresChallenge               = "stores challenge"
	testCaseCredentialLookupFails         = "credential lookup fails"
	testCaseChallengeStorageFails         = "challenge storage fails"
	testCaseNilResponse                   = "nil response"
	testCaseMalformedChallengeID          = "malformed challenge ID"
	testMalformedChallengeID              = "not-a-uuid"
	testInvalidChallengeID                = "invalid challenge ID"
	testCaseMissingChallenge              = "missing challenge"
	testChallengeNotFound                 = "challenge not found"
	testCaseExpiredChallenge              = "expired challenge"
	testChallengeExpired                  = "challenge expired"
	testInvalidSessionData                = "invalid session data"
	testInvalidUserHandle                 = "invalid user handle"
	testChallengeSessionData              = "challenge"
	testCredentialLookupError             = "failed to get user credentials"
	testChallengeStorageError             = "failed to store WebAuthn challenge"
	testWebAuthnLoginValidationFailed     = "failed to validate WebAuthn login"
	testCredentialID                      = "credential-id"
	testCaseInvalidAllowedCredentialIDs   = "invalid allowed credential IDs"
	testAllowedCredentialIDsMetadataField = "allowed_credential_ids"
)

type coverageCredentialStore struct {
	credentials     map[string]*Credential
	getCredentials  error
	getCredential   error
	storeCredential error
}

func (s *coverageCredentialStore) StoreCredential(_ context.Context, credential *Credential) error {
	if s.storeCredential != nil {
		return s.storeCredential
	}

	if s.credentials == nil {
		s.credentials = make(map[string]*Credential)
	}

	s.credentials[credential.ID] = credential

	return nil
}

func (s *coverageCredentialStore) GetCredential(_ context.Context, id string) (*Credential, error) {
	if s.getCredential != nil {
		return nil, s.getCredential
	}

	credential, ok := s.credentials[id]
	if !ok {
		return nil, errors.New("credential not found")
	}

	return credential, nil
}

func (s *coverageCredentialStore) GetUserCredentials(_ context.Context, userID string) ([]*Credential, error) {
	if s.getCredentials != nil {
		return nil, s.getCredentials
	}

	credentials := make([]*Credential, 0, len(s.credentials))
	for _, credential := range s.credentials {
		if credential.UserID == userID {
			credentials = append(credentials, credential)
		}
	}

	return credentials, nil
}

func (s *coverageCredentialStore) DeleteCredential(_ context.Context, id string) error {
	delete(s.credentials, id)

	return nil
}

type coverageChallengeStore struct {
	challenge   *AuthChallenge
	secret      string
	retrieveErr error
	storeErr    error
	deleteErr   error
}

func (s *coverageChallengeStore) Store(_ context.Context, challenge *AuthChallenge, secret string) error {
	if s.storeErr != nil {
		return s.storeErr
	}

	s.challenge = challenge
	s.secret = secret

	return nil
}

func (s *coverageChallengeStore) Retrieve(_ context.Context, _ googleUuid.UUID) (*AuthChallenge, string, error) {
	if s.retrieveErr != nil {
		return nil, "", s.retrieveErr
	}

	if s.challenge == nil {
		return nil, "", errors.New(testChallengeNotFound)
	}

	return s.challenge, s.secret, nil
}

func (s *coverageChallengeStore) Update(_ context.Context, challenge *AuthChallenge) error {
	s.challenge = challenge

	return nil
}

func (s *coverageChallengeStore) Delete(_ context.Context, _ googleUuid.UUID) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}

	s.challenge = nil

	return nil
}

func newCoverageWebAuthn(t *testing.T, credentials CredentialStore, challenges ChallengeStore) *WebAuthnAuthenticator {
	t.Helper()

	authenticator, err := NewWebAuthnAuthenticator(&WebAuthnConfig{
		RPID:          testExampleCom,
		RPDisplayName: testExampleCorp,
		RPOrigins:     []string{testHTTPSExampleCom},
		Timeout:       time.Hour,
	}, credentials, challenges)
	require.NoError(t, err)

	return authenticator
}

func coverageWebAuthnUser() *cryptoutilIdentityDomain.User {
	return &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		PreferredUsername: testTestuser,
		Name:              testTestUser,
	}
}

func TestWebAuthnAuthenticator_BeginRegistrationCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		getCredentialsErr error
		storeErr          error
		wantErr           string
	}{
		{name: testCaseStoresChallenge, wantErr: ""},
		{name: testCaseCredentialLookupFails, getCredentialsErr: errors.New("credential lookup failed"), wantErr: testCredentialLookupError},
		{name: testCaseChallengeStorageFails, storeErr: errors.New("challenge storage failed"), wantErr: testChallengeStorageError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			credentialStore := &coverageCredentialStore{getCredentials: tc.getCredentialsErr}
			challengeStore := &coverageChallengeStore{storeErr: tc.storeErr}
			authenticator := newCoverageWebAuthn(t, credentialStore, challengeStore)

			creation, err := authenticator.BeginRegistration(context.Background(), coverageWebAuthnUser())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, creation)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, creation)
			require.NotNil(t, challengeStore.challenge)
			require.Equal(t, "registration", challengeStore.challenge.Metadata[valueOperation])
			require.NotEmpty(t, challengeStore.challenge.Metadata[valueSessionData])
		})
	}
}

func TestWebAuthnAuthenticator_FinishRegistrationValidationCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		userIsNil         bool
		responseIsNil     bool
		challengeID       string
		challenge         *AuthChallenge
		retrieveErr       error
		getCredentialsErr error
		wantErr           string
	}{
		{
			name:          "nil user",
			userIsNil:     true,
			responseIsNil: true,
			wantErr:       "user cannot be nil",
		},
		{
			name:          testCaseNilResponse,
			responseIsNil: true,
			wantErr:       "credential creation response cannot be nil",
		},
		{
			name:        testCaseMalformedChallengeID,
			challengeID: testMalformedChallengeID,
			wantErr:     testInvalidChallengeID,
		},
		{
			name:        testCaseMissingChallenge,
			retrieveErr: errors.New(testChallengeNotFound),
			wantErr:     testChallengeNotFound,
		},
		{
			name: testCaseExpiredChallenge,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			},
			wantErr: testChallengeExpired,
		},
		{
			name: testInvalidSessionData,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{valueSessionData: 7},
			},
			wantErr: testInvalidSessionData,
		},
		{
			name: testInvalidUserHandle,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{valueSessionData: testChallengeSessionData},
			},
			wantErr: testInvalidUserHandle,
		},
		{
			name: testCaseCredentialLookupFails,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata: map[string]any{
					valueSessionData: testChallengeSessionData,
					valueUserHandle:  []byte("user-handle"),
				},
			},
			getCredentialsErr: errors.New("credential lookup failed"),
			wantErr:           testCredentialLookupError,
		},
		{
			name: "library rejects malformed attestation",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata: map[string]any{
					valueSessionData: testChallengeSessionData,
					valueUserHandle:  []byte("user-handle"),
				},
			},
			wantErr: "failed to create WebAuthn credential",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user := coverageWebAuthnUser()
			if tc.userIsNil {
				user = nil
			}

			challengeID := tc.challengeID
			if challengeID == "" {
				challengeID = googleUuid.Must(googleUuid.NewV7()).String()
			}

			challengeStore := &coverageChallengeStore{
				challenge:   tc.challenge,
				retrieveErr: tc.retrieveErr,
			}
			if challengeStore.challenge != nil {
				challengeStore.challenge.ID, _ = googleUuid.Parse(challengeID)
			}

			credentialStore := &coverageCredentialStore{getCredentials: tc.getCredentialsErr}
			authenticator := newCoverageWebAuthn(t, credentialStore, challengeStore)

			var response *protocol.ParsedCredentialCreationData
			if !tc.responseIsNil {
				response = &protocol.ParsedCredentialCreationData{}
			}

			credential, err := authenticator.FinishRegistration(context.Background(), user, challengeID, response)
			require.ErrorContains(t, err, tc.wantErr)
			require.Nil(t, credential)
		})
	}
}

func TestWebAuthnAuthenticator_InitiateAuthCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		getCredentialsErr error
		storeErr          error
		withCredential    bool
		wantErr           string
	}{
		{name: testCaseCredentialLookupFails, getCredentialsErr: errors.New("credential lookup failed"), wantErr: testCredentialLookupError},
		{name: "no credentials", wantErr: "no WebAuthn credentials registered"},
		{name: testCaseChallengeStorageFails, withCredential: true, storeErr: errors.New("challenge storage failed"), wantErr: testChallengeStorageError},
		{name: testCaseStoresChallenge, withCredential: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user := coverageWebAuthnUser()

			credentialStore := &coverageCredentialStore{getCredentials: tc.getCredentialsErr}
			if tc.withCredential {
				credentialStore.credentials = map[string]*Credential{
					testCredentialID: {
						ID:        testCredentialID,
						UserID:    user.ID.String(),
						PublicKey: []byte("public-key"),
					},
				}
			}

			challengeStore := &coverageChallengeStore{storeErr: tc.storeErr}
			authenticator := newCoverageWebAuthn(t, credentialStore, challengeStore)

			challenge, err := authenticator.InitiateAuth(context.Background(), user.ID.String())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, challenge)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, challenge)
			require.Equal(t, "authentication", challenge.Metadata[valueOperation])
			require.NotEmpty(t, challenge.Metadata[valueSessionData])
			require.NotNil(t, challenge.Metadata["assertion"])
		})
	}
}

func TestWebAuthnAuthenticator_VerifyAuthValidationCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		responseIsNil     bool
		challengeID       string
		challenge         *AuthChallenge
		retrieveErr       error
		getCredentialsErr error
		deleteErr         error
		includeCredential bool
		wantErr           string
	}{
		{name: testCaseNilResponse, responseIsNil: true, wantErr: "credential assertion response cannot be nil"},
		{name: testCaseMalformedChallengeID, challengeID: testMalformedChallengeID, wantErr: testInvalidChallengeID},
		{name: testCaseMissingChallenge, retrieveErr: errors.New(testChallengeNotFound), wantErr: testChallengeNotFound},
		{
			name: testCaseExpiredChallenge,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			},
			wantErr: testChallengeExpired,
		},
		{
			name: "expired challenge cleanup failure",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(-time.Hour),
			},
			deleteErr: errors.New("injected expired challenge cleanup failure"),
			wantErr:   testChallengeExpired,
		},
		{
			name: testInvalidSessionData,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{valueSessionData: 7},
			},
			wantErr: testInvalidSessionData,
		},
		{
			name: testInvalidUserHandle,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata:  map[string]any{valueSessionData: testChallengeSessionData},
			},
			wantErr: testInvalidUserHandle,
		},
		{
			name: testCaseInvalidAllowedCredentialIDs,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata: map[string]any{
					valueSessionData: testChallengeSessionData,
					valueUserHandle:  []byte("user-handle"),
				},
			},
			wantErr: testCaseInvalidAllowedCredentialIDs,
		},
		{
			name: testCaseCredentialLookupFails,
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				Metadata: map[string]any{
					valueSessionData:                      testChallengeSessionData,
					valueUserHandle:                       []byte("user-handle"),
					testAllowedCredentialIDsMetadataField: [][]byte{},
				},
			},
			getCredentialsErr: errors.New("credential lookup failed"),
			wantErr:           testCredentialLookupError,
		},
		{
			name: "library rejects malformed assertion",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				UserID:    googleUuid.Must(googleUuid.NewV7()).String(),
				Metadata: map[string]any{
					valueSessionData:                      testChallengeSessionData,
					valueUserHandle:                       []byte("user-handle"),
					testAllowedCredentialIDsMetadataField: [][]byte{},
				},
			},
			wantErr: testWebAuthnLoginValidationFailed,
		},
		{
			name: "credential conversion precedes assertion validation",
			challenge: &AuthChallenge{
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				UserID:    googleUuid.Must(googleUuid.NewV7()).String(),
				Metadata: map[string]any{
					valueSessionData:                      testChallengeSessionData,
					valueUserHandle:                       []byte("user-handle"),
					testAllowedCredentialIDsMetadataField: [][]byte{},
				},
			},
			includeCredential: true,
			wantErr:           testWebAuthnLoginValidationFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			challengeID := tc.challengeID
			if challengeID == "" {
				challengeID = googleUuid.Must(googleUuid.NewV7()).String()
			}

			challengeStore := &coverageChallengeStore{
				challenge:   tc.challenge,
				retrieveErr: tc.retrieveErr,
				deleteErr:   tc.deleteErr,
			}
			if challengeStore.challenge != nil {
				challengeStore.challenge.ID, _ = googleUuid.Parse(challengeID)
			}

			credentials := make(map[string]*Credential)
			if tc.includeCredential {
				credentials[testCredentialID] = &Credential{
					ID:     testCredentialID,
					UserID: tc.challenge.UserID,
				}
			}

			authenticator := newCoverageWebAuthn(
				t,
				&coverageCredentialStore{
					credentials:    credentials,
					getCredentials: tc.getCredentialsErr,
				},
				challengeStore,
			)

			var response *protocol.ParsedCredentialAssertionData
			if !tc.responseIsNil {
				response = &protocol.ParsedCredentialAssertionData{}
			}

			user, err := authenticator.VerifyAuth(context.Background(), challengeID, response)
			require.ErrorContains(t, err, tc.wantErr)
			require.Nil(t, user)
		})
	}
}
