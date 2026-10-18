// Copyright (c) 2025-2026 Justin Cranford.

package apis_test

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity-idp/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestHandleLogout_DeleteFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := cryptoutilIdentityConfig.DefaultConfig()
	config.Database.DSN = cryptoutilSharedMagic.SQLiteMemoryPlaceholder

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, repoFactory.Close())
	})
	require.NoError(t, repoFactory.DB().AutoMigrate(&cryptoutilIdentityDomain.User{}, &cryptoutilIdentityDomain.Session{}))

	user := &cryptoutilIdentityDomain.User{Sub: googleUuid.Must(googleUuid.NewV7()).String()}
	require.NoError(t, repoFactory.UserRepository().Create(ctx, user))

	session := &cryptoutilIdentityDomain.Session{
		UserID:     user.ID,
		SessionID:  googleUuid.Must(googleUuid.NewV7()).String(),
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(time.Hour),
		LastSeenAt: time.Now().UTC(),
		Active:     boolPtr(true),
	}
	require.NoError(t, repoFactory.SessionRepository().Create(ctx, session))

	require.NoError(t, repoFactory.DB().Exec("CREATE TRIGGER fail_session_delete BEFORE DELETE ON sessions BEGIN SELECT RAISE(ABORT, 'injected delete failure'); END").Error)

	service := cryptoutilIdentityIdp.NewService(config, repoFactory, nil)
	app := fiber.New()
	service.RegisterRoutes(app)

	request := httptest.NewRequest(http.MethodPost, "/oidc/v1/logout", nil)
	request.AddCookie(&http.Cookie{
		Name:     config.Sessions.CookieName,
		Value:    session.SessionID,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	response, err := app.Test(request, -1)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }() //nolint:errcheck // Test cleanup.

	require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
}
