// Copyright (c) 2025 Justin Cranford

package handler

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	sha256 "crypto/sha256"
	"encoding/asn1"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestTsaTimestamp_NoService tests TSA when service not configured.
func TestTsaTimestamp_NoService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	// Create valid timestamp request.
	tsReq := createValidTimestampRequest(t)

	app := fiber.New()

	app.Post("/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/timestamp", bytes.NewReader(tsReq))
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// TSA not configured, should return 503 Service Unavailable.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestTsaTimestamp_MalformedRequest tests TSA with malformed request (service not configured).
func TestTsaTimestamp_MalformedRequest(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Post("/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/timestamp", bytes.NewReader([]byte("invalid-asn1-data")))
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// Should return 503 (no service configured) before parsing request.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestTsaTimestamp_EmptyBody tests TSA with empty request body.
func TestTsaTimestamp_EmptyBody(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Post("/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/timestamp", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	// Should return 503 (no service configured) before checking empty body.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// createValidTimestampRequest creates a minimal valid RFC 3161 timestamp request.
func createValidTimestampRequest(t *testing.T) []byte {
	t.Helper()

	// RFC 3161 TimeStampReq structure (simplified for testing).
	type MessageImprint struct {
		HashAlgorithm asn1.ObjectIdentifier
		HashedMessage []byte
	}

	type TimeStampReq struct {
		Version        int
		MessageImprint MessageImprint
		ReqPolicy      asn1.ObjectIdentifier `asn1:"optional"`
		Nonce          *big.Int              `asn1:"optional"`
		CertReq        bool                  `asn1:"optional"`
	}

	// SHA-256 OID.
	sha256OID := asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, 3, 4, 2, 1}

	// Hash some data.
	data := []byte("test data for timestamp")
	hash := sha256.Sum256(data)

	tsReq := TimeStampReq{
		Version: 1,
		MessageImprint: MessageImprint{
			HashAlgorithm: sha256OID,
			HashedMessage: hash[:],
		},
		Nonce:   big.NewInt(12345),
		CertReq: true,
	}

	der, err := asn1.Marshal(tsReq)
	require.NoError(t, err)

	return der
}
