// Copyright (c) 2025-2026 Justin Cranford.
package userauth

import (
	"context"
	"errors"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRiskBasedAuthenticator_VerifyAuthExpiredChallenge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		deleteErr error
	}{
		{name: "expired challenge is rejected"},
		{name: "expired challenge remains rejected when cleanup fails", deleteErr: errors.New("cleanup failed")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			challengeID := googleUuid.Must(googleUuid.NewV7())
			challengeStore := &expiredRiskChallengeStore{
				challenge: &AuthChallenge{
					ID:        challengeID,
					ExpiresAt: time.Now().UTC().Add(-time.Minute),
				},
				deleteErr: tc.deleteErr,
			}
			authenticator := NewRiskBasedAuthenticator(nil, nil, challengeStore, nil, nil)

			user, err := authenticator.VerifyAuth(context.Background(), challengeID.String(), "unused")

			require.ErrorContains(t, err, "risk-based challenge expired")
			require.Nil(t, user)
			require.Equal(t, 1, challengeStore.deleteCalls)
		})
	}
}

type expiredRiskChallengeStore struct {
	challenge   *AuthChallenge
	deleteErr   error
	deleteCalls int
}

func (s *expiredRiskChallengeStore) Store(context.Context, *AuthChallenge, string) error {
	return nil
}

func (s *expiredRiskChallengeStore) Retrieve(context.Context, googleUuid.UUID) (*AuthChallenge, string, error) {
	return s.challenge, "", nil
}

func (s *expiredRiskChallengeStore) Update(context.Context, *AuthChallenge) error {
	return nil
}

func (s *expiredRiskChallengeStore) Delete(context.Context, googleUuid.UUID) error {
	s.deleteCalls++

	return s.deleteErr
}
