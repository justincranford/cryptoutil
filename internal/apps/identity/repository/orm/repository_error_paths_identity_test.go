// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestIdentityRepositories_DatabaseErrors(t *testing.T) {
	t.Parallel()

	testDB := setupTestDB(t)
	sqlDB, err := testDB.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	ctx := context.Background()
	id := googleUuid.Must(googleUuid.NewV7())
	before := time.Now().UTC()
	session := NewSessionRepository(testDB.db)
	token := NewTokenRepository(testDB.db)
	user := NewUserRepository(testDB.db)
	client := NewClientRepository(testDB.db)
	webauthn, err := NewWebAuthnCredentialRepository(testDB.db)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation func() error
	}{
		{"session_create", func() error { return session.Create(ctx, &cryptoutilIdentityDomain.Session{}) }},
		{"session_get_by_id", func() error {
			_, err := session.GetByID(ctx, id)

			return err
		}},
		{"session_get_by_session_id", func() error {
			_, err := session.GetBySessionID(ctx, "session")

			return err
		}},
		{"session_update", func() error { return session.Update(ctx, &cryptoutilIdentityDomain.Session{}) }},
		{"session_delete", func() error { return session.Delete(ctx, id) }},
		{"session_terminate_by_id", func() error { return session.TerminateByID(ctx, id) }},
		{"session_terminate_by_session_id", func() error { return session.TerminateBySessionID(ctx, "session") }},
		{"session_delete_expired", func() error { return session.DeleteExpired(ctx) }},
		{"session_delete_expired_before", func() error {
			_, err := session.DeleteExpiredBefore(ctx, before)

			return err
		}},
		{"session_list", func() error {
			_, err := session.List(ctx, 0, 1)

			return err
		}},
		{"session_count", func() error {
			_, err := session.Count(ctx)

			return err
		}},
		{"token_create", func() error { return token.Create(ctx, &cryptoutilIdentityDomain.Token{}) }},
		{"token_get_by_id", func() error {
			_, err := token.GetByID(ctx, id)

			return err
		}},
		{"token_get_by_value", func() error {
			_, err := token.GetByTokenValue(ctx, cryptoutilSharedMagic.ParamToken)

			return err
		}},
		{"token_update", func() error { return token.Update(ctx, &cryptoutilIdentityDomain.Token{}) }},
		{"token_delete", func() error { return token.Delete(ctx, id) }},
		{"token_revoke_by_id", func() error { return token.RevokeByID(ctx, id) }},
		{"token_revoke_by_value", func() error { return token.RevokeByTokenValue(ctx, cryptoutilSharedMagic.ParamToken) }},
		{"token_delete_expired", func() error { return token.DeleteExpired(ctx) }},
		{"token_delete_expired_before", func() error {
			_, err := token.DeleteExpiredBefore(ctx, before)

			return err
		}},
		{"token_list", func() error {
			_, err := token.List(ctx, 0, 1)

			return err
		}},
		{"token_count", func() error {
			_, err := token.Count(ctx)

			return err
		}},
		{"user_create", func() error { return user.Create(ctx, &cryptoutilIdentityDomain.User{}) }},
		{"user_get_by_id", func() error {
			_, err := user.GetByID(ctx, id)

			return err
		}},
		{"user_get_by_sub", func() error {
			_, err := user.GetBySub(ctx, "subject")

			return err
		}},
		{"user_get_by_username", func() error {
			_, err := user.GetByUsername(ctx, "username")

			return err
		}},
		{"user_get_by_email", func() error {
			_, err := user.GetByEmail(ctx, "user@example.com")

			return err
		}},
		{"user_update", func() error { return user.Update(ctx, &cryptoutilIdentityDomain.User{}) }},
		{"user_delete", func() error { return user.Delete(ctx, id) }},
		{"user_list", func() error {
			_, err := user.List(ctx, 0, 1)

			return err
		}},
		{"user_count", func() error {
			_, err := user.Count(ctx)

			return err
		}},
		{"client_create", func() error { return client.Create(ctx, &cryptoutilIdentityDomain.Client{}) }},
		{"client_get_by_id", func() error {
			_, err := client.GetByID(ctx, id)

			return err
		}},
		{"client_get_by_client_id", func() error {
			_, err := client.GetByClientID(ctx, "client")

			return err
		}},
		{"client_get_all", func() error {
			_, err := client.GetAll(ctx)

			return err
		}},
		{"client_update", func() error { return client.Update(ctx, &cryptoutilIdentityDomain.Client{}) }},
		{"client_delete", func() error { return client.Delete(ctx, id) }},
		{"client_list", func() error {
			_, err := client.List(ctx, 0, 1)

			return err
		}},
		{"client_count", func() error {
			_, err := client.Count(ctx)

			return err
		}},
		{"client_rotate_secret", func() error { return client.RotateSecret(ctx, id, "hash", "test", "rotation") }},
		{"client_secret_history", func() error {
			_, err := client.GetSecretHistory(ctx, id)

			return err
		}},
		{"webauthn_store_credential", func() error {
			return webauthn.StoreCredential(ctx, &Credential{ID: "credential", UserID: id.String()})
		}},
		{"webauthn_get_credential", func() error {
			_, err := webauthn.GetCredential(ctx, "credential")

			return err
		}},
		{"webauthn_get_user_credentials", func() error {
			_, err := webauthn.GetUserCredentials(ctx, id.String())

			return err
		}},
		{"webauthn_delete_credential", func() error { return webauthn.DeleteCredential(ctx, "credential") }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, tc.operation())
		})
	}
}
