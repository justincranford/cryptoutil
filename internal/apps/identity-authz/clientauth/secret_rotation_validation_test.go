// Copyright (c) 2025-2026 Justin Cranford.
package clientauth

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type migrationRepositoryStub struct {
	mockClientRepo
	clients   []*cryptoutilIdentityDomain.Client
	getAllErr error
	updateErr error
}

func (r *migrationRepositoryStub) GetAll(context.Context) ([]*cryptoutilIdentityDomain.Client, error) {
	return r.clients, r.getAllErr
}

func (r *migrationRepositoryStub) Update(context.Context, *cryptoutilIdentityDomain.Client) error {
	return r.updateErr
}

type migrationHasherStub struct {
	hashed string
	err    error
}

func (h migrationHasherStub) HashLowEntropyNonDeterministic(string) (string, error) {
	return h.hashed, h.err
}

func (migrationHasherStub) CompareSecret(string, string) error {
	return nil
}

func TestMigrateClientSecrets_ErrorAndSkipPaths(t *testing.T) {
	t.Parallel()

	getAllErr := errors.New("list clients failed")
	hashErr := errors.New("secret hash failed")
	updateErr := errors.New("client update failed")

	tests := []struct {
		name        string
		repository  *migrationRepositoryStub
		hasher      SecretHasher
		wantCount   int
		wantErrText string
	}{
		{
			name:        "client list error",
			repository:  &migrationRepositoryStub{getAllErr: getAllErr},
			hasher:      migrationHasherStub{},
			wantErrText: "failed to fetch clients for migration",
		},
		{
			name: "hash error",
			repository: &migrationRepositoryStub{
				clients: []*cryptoutilIdentityDomain.Client{{ClientID: "hash-error", ClientSecret: clientauthTestPlaintextSecret}},
			},
			hasher:      migrationHasherStub{err: hashErr},
			wantErrText: "failed to hash secret for client hash-error",
		},
		{
			name: "update error",
			repository: &migrationRepositoryStub{
				clients:   []*cryptoutilIdentityDomain.Client{{ClientID: "update-error", ClientSecret: clientauthTestPlaintextSecret}},
				updateErr: updateErr,
			},
			hasher:      migrationHasherStub{hashed: clientauthTestHashedSecret},
			wantErrText: "failed to update client update-error with hashed secret",
		},
		{
			name: "public and previously hashed clients are skipped",
			repository: &migrationRepositoryStub{
				clients: []*cryptoutilIdentityDomain.Client{
					{ClientID: cryptoutilSharedMagic.PublicLabel, ClientSecret: ""},
					{ClientID: "hashed", ClientSecret: clientauthTestHashedSecret},
				},
			},
			hasher: migrationHasherStub{hashed: clientauthTestHashedSecret},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			migrated, err := MigrateClientSecrets(context.Background(), tc.repository, tc.hasher)
			require.Equal(t, tc.wantCount, migrated)

			if tc.wantErrText != "" {
				require.ErrorContains(t, err, tc.wantErrText)

				return
			}

			require.NoError(t, err)
		})
	}
}

type secretAuthenticatorRepositoryStub struct {
	mockClientRepo
	client *cryptoutilIdentityDomain.Client
	err    error
}

func (r *secretAuthenticatorRepositoryStub) GetByClientID(context.Context, string) (*cryptoutilIdentityDomain.Client, error) {
	return r.client, r.err
}

type secretAuthenticatorHasherStub struct {
	compareErr error
}

func (secretAuthenticatorHasherStub) HashLowEntropyNonDeterministic(string) (string, error) {
	return "", nil
}

func (h secretAuthenticatorHasherStub) CompareSecret(string, string) error {
	return h.compareErr
}

type rotationServiceStub struct {
	valid   bool
	version int
	err     error
}

func (r rotationServiceStub) ValidateSecretDuringGracePeriod(context.Context, googleUuid.UUID, string) (bool, int, error) {
	return r.valid, r.version, r.err
}

func TestSecretBasedAuthenticator_AuthenticateBasic_ValidationPaths(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("client lookup failed")
	rotationErr := errors.New("rotation check failed")
	compareErr := errors.New("secret comparison failed")
	enabled := true
	disabled := false

	client := &cryptoutilIdentityDomain.Client{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     "enabled-client",
		ClientSecret: "stored-secret",
		Enabled:      &enabled,
	}

	tests := []struct {
		name          string
		repository    *secretAuthenticatorRepositoryStub
		hasher        SecretHasher
		rotation      RotationService
		wantClient    bool
		wantErrSubstr string
	}{
		{
			name:          "repository error",
			repository:    &secretAuthenticatorRepositoryStub{err: repoErr},
			wantErrSubstr: "client authentication failed",
		},
		{
			name:          "missing enabled state",
			repository:    &secretAuthenticatorRepositoryStub{client: &cryptoutilIdentityDomain.Client{ClientID: "no-enabled"}},
			wantErrSubstr: clientauthTestClientDisabledError,
		},
		{
			name:          clientauthTestDisabledClientCase,
			repository:    &secretAuthenticatorRepositoryStub{client: &cryptoutilIdentityDomain.Client{Enabled: &disabled}},
			wantErrSubstr: clientauthTestClientDisabledError,
		},
		{
			name:          "rotation validation error",
			repository:    &secretAuthenticatorRepositoryStub{client: client},
			rotation:      rotationServiceStub{err: rotationErr},
			wantErrSubstr: "failed to validate secret",
		},
		{
			name:          "invalid rotated credential",
			repository:    &secretAuthenticatorRepositoryStub{client: client},
			rotation:      rotationServiceStub{},
			wantErrSubstr: clientauthTestInvalidCredentialsError,
		},
		{
			name:       "valid rotated credential",
			repository: &secretAuthenticatorRepositoryStub{client: client},
			rotation:   rotationServiceStub{valid: true, version: 2},
			wantClient: true,
		},
		{
			name:          "fallback comparison error",
			repository:    &secretAuthenticatorRepositoryStub{client: client},
			hasher:        secretAuthenticatorHasherStub{compareErr: compareErr},
			wantErrSubstr: clientauthTestInvalidCredentialsError,
		},
		{
			name:       "fallback comparison success",
			repository: &secretAuthenticatorRepositoryStub{client: client},
			hasher:     secretAuthenticatorHasherStub{},
			wantClient: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authenticator := &SecretBasedAuthenticator{
				clientRepo:      tc.repository,
				hasher:          tc.hasher,
				rotationService: tc.rotation,
			}

			result, err := authenticator.AuthenticateBasic(context.Background(), client.ClientID, clientauthTestProvidedSecret)
			if tc.wantErrSubstr != "" {
				require.ErrorContains(t, err, tc.wantErrSubstr)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)

			if tc.wantClient {
				require.Same(t, client, result)
			}
		})
	}
}
