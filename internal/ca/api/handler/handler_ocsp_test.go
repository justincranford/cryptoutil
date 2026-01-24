// Copyright (c) 2025 Justin Cranford

package handler

import (
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

// TestHandleOCSP_NoService tests OCSP when service not configured.
func TestHandleOCSP_NoService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Get("/ocsp/:serial", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ocsp/12345", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// OCSP service not configured, should return 503.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestHandleOCSP_POST_NoService tests OCSP POST when service not configured.
func TestHandleOCSP_POST_NoService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	// Create minimal OCSP request.
	ocspReq := createMinimalOCSPRequest(t)

	app := fiber.New()

	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader(ocspReq))
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// OCSP service not configured, should return 503.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestHandleOCSP_GET_EmptySerialNumber tests OCSP GET with empty serial.
func TestHandleOCSP_GET_EmptySerialNumber(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Get("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ocsp", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should return 503 (service not configured) or error.
	require.True(t, resp.StatusCode >= 400)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// createMinimalOCSPRequest creates a minimal OCSP request for testing.
func createMinimalOCSPRequest(t *testing.T) []byte {
	t.Helper()

	// Create minimal ASN.1 OCSP request structure.
	// This is a simplified structure that won't parse correctly but tests error handling.
	type CertID struct {
		HashAlgorithm  asn1.ObjectIdentifier
		IssuerNameHash []byte
		IssuerKeyHash  []byte
		SerialNumber   *big.Int
	}

	type Request struct {
		ReqCert CertID
	}

	type TBSRequest struct {
		Version     int `asn1:"optional,explicit,tag:0"`
		RequestList []Request
	}

	type OCSPRequest struct {
		TBSRequest TBSRequest
	}

	// SHA-1 OID.
	sha1OID := asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}

	issuerNameHash := sha256.Sum256([]byte("test-issuer"))
	issuerKeyHash := sha256.Sum256([]byte("test-key"))

	certID := CertID{
		HashAlgorithm:  sha1OID,
		IssuerNameHash: issuerNameHash[:20], // Use first 20 bytes for SHA-1 size.
		IssuerKeyHash:  issuerKeyHash[:20],
		SerialNumber:   big.NewInt(12345),
	}

	req := OCSPRequest{
		TBSRequest: TBSRequest{
			Version:     0,
			RequestList: []Request{{ReqCert: certID}},
		},
	}

	der, err := asn1.Marshal(req)
	require.NoError(t, err)

	return der
}

// TestHandleOCSP_POST_MalformedRequest tests OCSP POST with malformed request.
func TestHandleOCSP_POST_MalformedRequest(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte("malformed-ocsp")))
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should return 503 (service not configured) before parsing.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestHandleOCSP_POST_EmptyBody tests OCSP POST with empty body.
func TestHandleOCSP_POST_EmptyBody(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should return 503 (service not configured) before checking empty body.
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}
