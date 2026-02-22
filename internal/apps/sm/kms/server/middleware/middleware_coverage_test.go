package middleware

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestIsAllowedValue(t *testing.T) {
	t.Parallel()

	mw := &ServiceAuthMiddleware{}

	tests := []struct {
		name     string
		value    string
		allowed  []string
		expected bool
	}{
		{name: "match found", value: "foo", allowed: []string{"foo", "bar"}, expected: true},
		{name: "no match", value: "baz", allowed: []string{"foo", "bar"}, expected: false},
		{name: "empty allowed list", value: "foo", allowed: []string{}, expected: false},
		{name: "nil allowed list", value: "foo", allowed: nil, expected: false},
		{name: "empty value matches empty entry", value: "", allowed: []string{""}, expected: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, mw.isAllowedValue(tc.value, tc.allowed))
		})
	}
}

func TestAuthenticateJWT_NilValidator(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodJWT},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticateJWT_EmptyToken(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodJWT},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServiceAuth_UnauthorizedWithDetails(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods:   []AuthMethod{AuthMethodAPIKey},
		ErrorDetailLevel: errorDetailLevelStd,
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticateMTLS_NoTLSConnection(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodMTLS},
		MTLSConfig: &MTLSConfig{
			RequireClientCert: true,
		},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_EmptyBearerToken(t *testing.T) {
	t.Parallel()

	jwksServer := newTestJWKSServer(t)

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL: jwksServer.server.URL,
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(validator.JWTMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSessionMiddleware_EmptyServiceToken(t *testing.T) {
	t.Parallel()

	mockValidator := &mockSessionValidator{}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(SessionMiddleware(mockValidator, "session_token", false))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPerformRevocationCheck_CheckRevocationError(t *testing.T) {
	t.Parallel()

	closedServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedServer.Close()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:             "https://localhost/.well-known/jwks.json",
		RevocationCheckMode: RevocationCheckEveryRequest,
		IntrospectionURL:    closedServer.URL,
	})
	require.NoError(t, err)

	err = validator.performRevocationCheck(t.Context(), "test-token", &JWTClaims{Subject: "user"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "revocation check failed")
}

func TestCheckRevocation_HttpExecuteError(t *testing.T) {
	t.Parallel()

	closedServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedServer.Close()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:          "https://localhost/.well-known/jwks.json",
		IntrospectionURL: closedServer.URL,
	})
	require.NoError(t, err)

	active, checkErr := validator.checkRevocation(t.Context(), "test-token")
	require.Error(t, checkErr)
	require.False(t, active)
	require.Contains(t, checkErr.Error(), "introspection request failed")
}

func TestCheckRevocation_BadURL(t *testing.T) {
	t.Parallel()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:          "https://localhost/.well-known/jwks.json",
		IntrospectionURL: "://\x00invalid",
	})
	require.NoError(t, err)

	active, checkErr := validator.checkRevocation(t.Context(), "test-token")
	require.Error(t, checkErr)
	require.False(t, active)
}

func TestRefreshJWKS_FetchError(t *testing.T) {
	t.Parallel()

	closedServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedServer.Close()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL: closedServer.URL,
	})
	require.NoError(t, err)

	_, fetchErr := validator.refreshJWKS(t.Context())
	require.Error(t, fetchErr)
	require.Contains(t, fetchErr.Error(), "failed to fetch JWKS")
}

func TestPerformIntrospection_BadURL(t *testing.T) {
	t.Parallel()

	introspector, err := NewBatchIntrospector(IntrospectionConfig{
		IntrospectionURL: "://\x00invalid",
	})
	require.NoError(t, err)

	result, introErr := introspector.performIntrospection(context.Background(), "test-token")
	require.Error(t, introErr)
	require.Nil(t, result)
}

func TestExtractFromMap_MarshalError(t *testing.T) {
	t.Parallel()

	extractor := NewClaimsExtractor()

	rawClaims := map[string]any{
		"sub": make(chan int),
	}

	claims, err := extractor.ExtractFromMap(rawClaims)
	require.Error(t, err)
	require.Nil(t, claims)
	require.Contains(t, err.Error(), "failed to marshal claims")
}

func TestClientCredentials_NoBearerToken(t *testing.T) {
	t.Parallel()

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodClientCredentials},
		ClientCredentialsConfig: &ClientCredentialsConfig{
			TokenEndpoint: "https://auth.example.com/token",
		},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func generateTestCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().UTC().Add(-time.Hour),
		NotAfter:              time.Now().UTC().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caDER, err := x509.CreateCertificate(crand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caCert, err := x509.ParseCertificate(caDER)
	require.NoError(t, err)

	return caCert, caKey
}

func generateTestCert(
	t *testing.T,
	caCert *x509.Certificate,
	caKey *ecdsa.PrivateKey,
	cn string,
	dnsNames []string,
) (tls.Certificate, *x509.Certificate) {
	t.Helper()

	certKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:         cn,
			OrganizationalUnit: []string{"Test OU"},
		},
		DNSNames:    dnsNames,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:   time.Now().UTC().Add(-time.Hour),
		NotAfter:    time.Now().UTC().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
	}

	certDER, err := x509.CreateCertificate(crand.Reader, certTemplate, caCert, &certKey.PublicKey, caKey)
	require.NoError(t, err)

	parsedCert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(certKey)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	return tlsCert, parsedCert
}

func TestAuthenticateMTLS_SuccessfulAuth(t *testing.T) {
	t.Parallel()

	caCert, caKey := generateTestCA(t)
	serverCert, _ := generateTestCert(t, caCert, caKey, "test-server", []string{"localhost"})
	clientCert, _ := generateTestCert(t, caCert, caKey, "test-client", []string{"client.local"})

	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	mw, err := NewServiceAuthMiddleware(ServiceAuthConfig{
		AllowedMethods: []AuthMethod{AuthMethodMTLS},
		MTLSConfig:     &MTLSConfig{RequireClientCert: true},
	})
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	require.NoError(t, err)

	defer func() { _ = ln.Close() }()

	go func() {
		_ = app.Listener(ln) //nolint:errcheck // Server runs in background goroutine.
	}()

	clientTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS13,
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: clientTLSConfig},
	}

	url := fmt.Sprintf("https://%s/test", ln.Addr().String())

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
