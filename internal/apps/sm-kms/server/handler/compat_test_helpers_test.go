// Copyright (c) 2025-2026 Justin Cranford.
//

package handler

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testCompatInvalidUUID             = "not-a-uuid"
	testCompatMessageIDParam          = ":messageID"
	testCompatInvalidKeyIDCase        = "invalid key id"
	testCompatInvalidMessageIDCase    = "invalid message id"
	testCompatJWKSPath                = "/jwks"
	testCompatMessagePath             = "/messages"
	testCompatMessageSendPath         = "/messages/send"
	testCompatMessageReceivePath      = "/messages/receive"
	testCompatActiveMaterialPathStart = "/elastic-keys/"
	testCompatElasticKeyIDParam       = ":elasticKeyID"
	testCompatActiveMaterialPathEnd   = "/material-keys/active"
	testCompatRotatePathEnd           = "/rotate"
	testCompatTenantQuery             = "tenant_id"
	testCompatRecipientQuery          = "recipient_id"
	testCompatMaterialKID             = "test-material-kid"
	testCompatMessageJWE              = "test-jwe"
	testCompatUnexpectedRepoCall      = "unexpected compatibility repository call"
)

func newCompatTestUUID(t *testing.T) googleUuid.UUID {
	t.Helper()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)

	return id
}

func runCompatFiberRequest(t *testing.T, register func(*fiber.App), method, target string) (int, []byte) {
	t.Helper()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	register(app)

	request := httptest.NewRequest(method, target, nil)
	response, err := app.Test(request, -1)
	require.NoError(t, err)

	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	require.NoError(t, errors.Join(readErr, closeErr))

	return response.StatusCode, body
}
