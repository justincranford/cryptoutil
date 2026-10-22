// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	json "encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
)

func TestSubmitEnrollment_UnknownProfileAndStorageError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		profile        string
		profiles       map[string]*ProfileConfig
		storeErr       error
		wantStatusCode int
	}{
		{
			name:           "unknown profile",
			profile:        cryptoutilSharedMagic.PKIProfileIDTLSServer,
			profiles:       map[string]*ProfileConfig{},
			wantStatusCode: fiber.StatusBadRequest,
		},
		{
			name:    testStorageFailure,
			profile: cryptoutilSharedMagic.PKIProfileIDTLSServer,
			profiles: map[string]*ProfileConfig{
				cryptoutilSharedMagic.PKIProfileIDTLSServer: {
					ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
					Name:        testProfileNameTLSServer,
					Description: testProfileDescriptionTLSServer,
					Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
				},
			},
			storeErr:       errors.New("injected enrollment persistence failure"),
			wantStatusCode: fiber.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testSetup := createTestIssuer(t)
			store := &errorStore{
				inner:    cryptoutilCAStorage.NewMemoryStore(),
				storeErr: tc.storeErr,
			}
			handler := &Handler{
				storage:           store,
				issuer:            testSetup.Issuer,
				profiles:          tc.profiles,
				enrollmentTracker: newEnrollmentTracker(cryptoutilSharedMagic.JoseJAMaxMaterials),
			}

			requestBody, err := json.Marshal(cryptoutilApiCaServer.EnrollmentRequest{
				CSR:     createValidCSRPEM(t),
				Profile: tc.profile,
			})
			require.NoError(t, err)

			app := fiber.New()
			app.Post("/enrollments", handler.SubmitEnrollment)

			req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", testContentTypeJSON)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatusCode, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestGetEnrollmentStatus_StorageErrorAndFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		getErr           error
		storeCertificate bool
		wantStatusCode   int
	}{
		{
			name:           testStorageFailure,
			getErr:         errors.New("injected enrollment lookup failure"),
			wantStatusCode: fiber.StatusInternalServerError,
		},
		{
			name:             "certificate found",
			storeCertificate: true,
			wantStatusCode:   fiber.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requestID := googleUuid.Must(googleUuid.NewV7())

			inner := cryptoutilCAStorage.NewMemoryStore()
			if tc.storeCertificate {
				err := inner.Store(context.Background(), &cryptoutilCAStorage.StoredCertificate{
					ID:             googleUuid.Must(googleUuid.NewV7()),
					RequestID:      requestID,
					SerialNumber:   testSerialNumberHex,
					CertificatePEM: "certificate-pem",
					NotBefore:      time.Now().UTC(),
					NotAfter:       time.Now().UTC().Add(time.Hour),
				})
				require.NoError(t, err)
			}

			handler := &Handler{
				storage: &errorStore{
					inner:  inner,
					getErr: tc.getErr,
				},
			}
			app := fiber.New()
			app.Get("/enrollments", func(c *fiber.Ctx) error {
				return handler.GetEnrollmentStatus(c, requestID)
			})

			req := httptest.NewRequest(http.MethodGet, "/enrollments", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatusCode, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestGetCRL_RevocationStorageAndMetadataErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		revoked         *cryptoutilCAStorage.StoredCertificate
		storageError    error
		generationError bool
		format          string
	}{
		{
			name:         testStorageFailure,
			storageError: errors.New("injected CRL lookup failure"),
		},
		{
			name:            "DER generation failure",
			generationError: true,
		},
		{
			name:            "PEM generation failure",
			generationError: true,
			format:          cryptoutilSharedMagic.PKICRLFormatPEM,
		},
		{
			name: "missing revocation metadata",
			revoked: &cryptoutilCAStorage.StoredCertificate{
				SerialNumber: "10",
				Status:       cryptoutilCAStorage.StatusRevoked,
			},
		},
		{
			name: "invalid serial number",
			revoked: &cryptoutilCAStorage.StoredCertificate{
				SerialNumber:   "not-hex",
				Status:         cryptoutilCAStorage.StatusRevoked,
				RevocationTime: timePointer(time.Now().UTC()),
				RevocationReason: reasonPointer(
					cryptoutilCAStorage.ReasonKeyCompromise,
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testSetup := createTestIssuer(t)
			signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
			require.True(t, ok)

			crlConfig := &cryptoutilCAServiceRevocation.CRLConfig{
				Issuer:     testSetup.Issuer.GetCAConfig().Certificate,
				PrivateKey: signer,
				Provider:   testSetup.Provider,
				Validity:   time.Hour,
			}
			if tc.generationError {
				crlConfig.AllocateCRLNumber = func(context.Context, string) (*big.Int, error) {
					return nil, errors.New("injected CRL number allocation failure")
				}
			}

			crlService, err := cryptoutilCAServiceRevocation.NewCRLService(crlConfig)
			require.NoError(t, err)

			inner := cryptoutilCAStorage.NewMemoryStore()

			if tc.revoked != nil {
				tc.revoked.ID = googleUuid.Must(googleUuid.NewV7())
				tc.revoked.RequestID = googleUuid.Must(googleUuid.NewV7())
				tc.revoked.AuthorityKeyID = fmtIssuerKeyID(testSetup.Issuer.GetCAConfig().Certificate)
				err = inner.Store(context.Background(), tc.revoked)
				require.NoError(t, err)
			}

			handler := &Handler{
				storage: &errorStore{
					inner:      inner,
					revokedErr: tc.storageError,
				},
				issuer:     testSetup.Issuer,
				crlService: crlService,
			}
			app := fiber.New()
			app.Get("/ca/crl", func(c *fiber.Ctx) error {
				params := cryptoutilApiCaServer.GetCRLParams{}

				if tc.format != "" {
					format := cryptoutilApiCaServer.GetCRLParamsFormat(tc.format)
					params.Format = &format
				}

				return handler.GetCRL(c, testSetup.Issuer.GetCAConfig().Name, params)
			})

			req := httptest.NewRequest(http.MethodGet, "/ca/crl", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestHandlers_JSONEncoderErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		add  func(*testing.T, *Handler, *fiber.App)
		body func(*testing.T) []byte
	}{
		{
			name: "list profiles",
			path: testProfilesPath,
			add: func(_ *testing.T, handler *Handler, app *fiber.App) {
				app.Get("/profiles", handler.ListProfiles)
			},
		},
		{
			name: "get profile",
			path: testProfilesPath,
			add: func(_ *testing.T, handler *Handler, app *fiber.App) {
				app.Get("/profiles", func(c *fiber.Ctx) error {
					return handler.GetProfile(c, cryptoutilSharedMagic.PKIProfileIDTLSServer)
				})
			},
		},
		{
			name: "get enrollment status",
			path: testEnrollmentsPath,
			add: func(t *testing.T, handler *Handler, app *fiber.App) {
				t.Helper()

				requestID := googleUuid.Must(googleUuid.NewV7())
				err := handler.storage.Store(context.Background(), &cryptoutilCAStorage.StoredCertificate{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					RequestID: requestID,
				})
				require.NoError(t, err)
				app.Get("/enrollments", func(c *fiber.Ctx) error {
					return handler.GetEnrollmentStatus(c, requestID)
				})
			},
		},
		{
			name: "list CAs",
			path: testCAsPath,
			add: func(t *testing.T, handler *Handler, app *fiber.App) {
				t.Helper()
				handler.issuer = createTestIssuer(t).Issuer
				app.Get("/cas", handler.ListCAs)
			},
		},
		{
			name: "get CA",
			path: testCAsPath,
			add: func(t *testing.T, handler *Handler, app *fiber.App) {
				t.Helper()
				handler.issuer = createTestIssuer(t).Issuer

				app.Get("/cas", func(c *fiber.Ctx) error {
					return handler.GetCA(c, handler.issuer.GetCAConfig().Name)
				})
			},
		},
		{
			name: "submit enrollment",
			path: "/enrollments",
			add: func(t *testing.T, handler *Handler, app *fiber.App) {
				t.Helper()
				setup := createTestIssuer(t)
				handler.issuer = setup.Issuer
				handler.profiles[cryptoutilSharedMagic.PKIProfileIDTLSServer] = &ProfileConfig{
					ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
					Name:        testProfileNameTLSServer,
					Description: testProfileDescriptionTLSServer,
					Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
				}
				handler.enrollmentTracker = newEnrollmentTracker(cryptoutilSharedMagic.JoseJAMaxMaterials)
				app.Post("/enrollments", handler.SubmitEnrollment)
			},
			body: func(t *testing.T) []byte {
				t.Helper()
				body, err := json.Marshal(cryptoutilApiCaServer.EnrollmentRequest{
					CSR:     createValidCSRPEM(t),
					Profile: cryptoutilSharedMagic.PKIProfileIDTLSServer,
				})
				require.NoError(t, err)

				return body
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				storage: cryptoutilCAStorage.NewMemoryStore(),
				profiles: map[string]*ProfileConfig{cryptoutilSharedMagic.PKIProfileIDTLSServer: {
					ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
					Name:        testProfileNameTLSServer,
					Description: testProfileDescriptionTLSServer,
					Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
				}},
			}
			app := fiber.New(fiber.Config{
				JSONEncoder: func(any) ([]byte, error) {
					return nil, errors.New("injected JSON encoder failure")
				},
			})
			tc.add(t, handler, app)

			var body []byte
			if tc.body != nil {
				body = tc.body(t)
			}

			req := httptest.NewRequest(http.MethodGet, tc.path, bytes.NewReader(body))
			if tc.body != nil {
				req.Method = http.MethodPost
				req.Header.Set("Content-Type", testContentTypeJSON)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestHandleOCSP_IssuerConfigurationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		noIssuer    bool
		noCert      bool
		noSubjectID bool
	}{
		{name: "issuer not configured", noIssuer: true},
		{name: "issuer certificate not configured", noCert: true},
		{name: "issuer subject key ID missing", noSubjectID: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setup := createTestIssuer(t)
			caCertificate := setup.Issuer.GetCAConfig().Certificate
			signer, ok := setup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
			require.True(t, ok)

			crlService, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
				Issuer:     caCertificate,
				PrivateKey: signer,
				Provider:   setup.Provider,
				Validity:   time.Hour,
			})
			require.NoError(t, err)
			ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(&cryptoutilCAServiceRevocation.OCSPConfig{
				Issuer:       caCertificate,
				Responder:    caCertificate,
				ResponderKey: signer,
				Provider:     setup.Provider,
				Validity:     time.Hour,
			}, crlService)
			require.NoError(t, err)

			handler := &Handler{
				storage:     cryptoutilCAStorage.NewMemoryStore(),
				issuer:      setup.Issuer,
				ocspService: ocspService,
			}

			switch {
			case tc.noIssuer:
				handler.issuer = nil
			case tc.noCert:
				setup.Issuer.GetCAConfig().Certificate = nil
			case tc.noSubjectID:
				caCertificate.SubjectKeyId = nil
			}

			app := fiber.New()
			app.Post("/ocsp", handler.HandleOCSP)

			req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte(cryptoutilSharedMagic.TLSClientPolicyRequest)))
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestHandleOCSP_BodyStreamReadError(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	caCertificate := setup.Issuer.GetCAConfig().Certificate
	signer, ok := setup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     caCertificate,
		PrivateKey: signer,
		Provider:   setup.Provider,
		Validity:   time.Hour,
	})
	require.NoError(t, err)
	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(&cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       caCertificate,
		Responder:    caCertificate,
		ResponderKey: signer,
		Provider:     setup.Provider,
		Validity:     time.Hour,
	}, crlService)
	require.NoError(t, err)

	handler := &Handler{issuer: setup.Issuer, ocspService: ocspService}

	app := fiber.New()
	app.Post("/ocsp", func(c *fiber.Ctx) error {
		c.Request().SetBodyStream(failingBodyReader{}, -1)

		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ocsp", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestESTSimpleEnroll_ErrorPaths(t *testing.T) {
	t.Parallel()

	validCSR := createValidCSRDER(t)
	invalidSignatureCSR := append([]byte(nil), validCSR...)
	invalidSignatureCSR[len(invalidSignatureCSR)-1] ^= 0xff

	tests := []struct {
		name           string
		body           []byte
		storeErr       error
		wantStatusCode int
	}{
		{
			name:           testEmptyCSR,
			wantStatusCode: fiber.StatusBadRequest,
		},
		{
			name:           "invalid CSR signature",
			body:           invalidSignatureCSR,
			wantStatusCode: fiber.StatusBadRequest,
		},
		{
			name:           "certificate storage failure",
			body:           validCSR,
			storeErr:       errors.New("injected EST persistence failure"),
			wantStatusCode: fiber.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setup := createTestIssuer(t)
			handler := &Handler{
				storage: &errorStore{
					inner:    cryptoutilCAStorage.NewMemoryStore(),
					storeErr: tc.storeErr,
				},
				issuer: setup.Issuer,
				profiles: map[string]*ProfileConfig{
					cryptoutilSharedMagic.PKIProfileIDTLSServer: {
						ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
						Name:        testProfileNameTLSServer,
						Description: testProfileDescriptionTLSServer,
						Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
					},
				},
			}
			app := fiber.New()
			app.Post("/est/simpleenroll", handler.ESTSimpleEnroll)

			req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(tc.body))
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatusCode, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestESTSimpleReenrollAuthorizationAndIdentityChecks(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	caCertificate := setup.Issuer.GetCAConfig().Certificate
	clientCertificate := newESTReenrollmentTestCertificate(t, caCertificate, testIssuerSigner(t, setup))
	validCSR := createValidCSRDER(t)
	mismatchedKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	mismatchedCSR, err := x509.CreateCertificateRequest(crand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: "different.example.test"},
	}, mismatchedKey)
	require.NoError(t, err)

	tests := []struct {
		name           string
		body           []byte
		clientCert     bool
		storageState   string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           testESTMissingClientCertificate,
			body:           validCSR,
			wantStatusCode: fiber.StatusUnauthorized,
			wantBody:       cryptoutilSharedMagic.PKIErrorCodeClientCertificateRequired,
		},
		{
			name:           "client certificate not stored",
			body:           validCSR,
			clientCert:     true,
			wantStatusCode: fiber.StatusForbidden,
			wantBody:       cryptoutilSharedMagic.PKIErrorCodeReenrollmentNotAuthorized,
		},
		{
			name:           "certificate lookup failure",
			body:           validCSR,
			clientCert:     true,
			storageState:   cryptoutilSharedMagic.StringError,
			wantStatusCode: fiber.StatusInternalServerError,
			wantBody:       "injected reenrollment lookup failure",
		},
		{
			name:           testESTCSRIdentityChange,
			body:           mismatchedCSR,
			clientCert:     true,
			storageState:   testESTStorageStateStored,
			wantStatusCode: fiber.StatusForbidden,
			wantBody:       cryptoutilSharedMagic.PKIMessageReenrollmentIdentityMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			memoryStore := cryptoutilCAStorage.NewMemoryStore()

			var store cryptoutilCAStorage.Store = memoryStore

			switch tc.storageState {
			case cryptoutilSharedMagic.StringError:
				store = &errorStore{
					inner:  memoryStore,
					getErr: errors.New("injected reenrollment lookup failure"),
				}
			case testESTStorageStateStored:
				storeESTReenrollmentCertificate(
					t,
					memoryStore,
					caCertificate,
					clientCertificate,
					cryptoutilCAStorage.StatusActive,
					clientCertificate.Raw,
				)
			}

			handler := &Handler{
				storage: store,
				issuer:  setup.Issuer,
				profiles: map[string]*ProfileConfig{
					cryptoutilSharedMagic.PKIProfileIDTLSServer: {
						ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
						Name:        testProfileNameTLSServer,
						Description: testProfileDescriptionTLSServer,
						Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
					},
				},
			}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/est/simplereenroll", func(c *fiber.Ctx) error {
				ctx := c.UserContext()
				if tc.clientCert {
					ctx = withESTClientCertificate(ctx, clientCertificate)
				}

				return handler.estSimpleEnroll(c, ctx, true)
			})

			req := httptest.NewRequest(http.MethodPost, "/est/simplereenroll", bytes.NewReader(tc.body))
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatusCode, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tc.wantBody)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestGetCRL_MissingSubjectKeyIdentifier(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	setup.Issuer.GetCAConfig().Certificate.SubjectKeyId = nil

	handler := &Handler{
		issuer:     setup.Issuer,
		storage:    cryptoutilCAStorage.NewMemoryStore(),
		crlService: &cryptoutilCAServiceRevocation.CRLService{},
	}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/ca/crl", func(c *fiber.Ctx) error {
		return handler.GetCRL(c, setup.Issuer.GetCAConfig().Name, cryptoutilApiCaServer.GetCRLParams{})
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ca/crl", nil), -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "ca_config_error")
	require.NoError(t, resp.Body.Close())
}

func TestLookupCertificateStatusBySerial_ErrorAndRevocationPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		serialNumber     *big.Int
		getErr           error
		certificatePEM   string
		status           cryptoutilCAStorage.CertificateStatus
		revocationTime   *time.Time
		revocationReason *cryptoutilCAStorage.RevocationReason
		wantError        string
		wantCertificate  bool
		wantRevoked      bool
	}{
		{
			name:      "nil serial number",
			wantError: "serial number is required",
		},
		{
			name:         strictHandlerTestStorageError,
			serialNumber: big.NewInt(42),
			getErr:       errors.New("injected OCSP storage failure"),
			wantError:    "failed to look up certificate",
		},
		{
			name:           testHandlerInvalidPEM,
			serialNumber:   big.NewInt(42),
			certificatePEM: "not pem",
			wantError:      testStoredCertificatePEMInvalid,
		},
		{
			name:           testInvalidDERCertificate,
			serialNumber:   big.NewInt(42),
			certificatePEM: string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: []byte("invalid DER")})),
			wantError:      testFailedToParseStoredCertificate,
		},
		{
			name:            testActiveCertificate,
			serialNumber:    big.NewInt(42),
			status:          cryptoutilCAStorage.StatusActive,
			wantCertificate: true,
		},
		{
			name:         "revoked certificate missing metadata",
			serialNumber: big.NewInt(42),
			status:       cryptoutilCAStorage.StatusRevoked,
			wantError:    "revoked certificate is missing revocation metadata",
		},
		{
			name:             "revoked certificate with metadata",
			serialNumber:     big.NewInt(42),
			status:           cryptoutilCAStorage.StatusRevoked,
			revocationTime:   timePointer(time.Now().UTC()),
			revocationReason: reasonPointer(cryptoutilCAStorage.ReasonKeyCompromise),
			wantCertificate:  true,
			wantRevoked:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testSetup := createTestIssuer(t)
			certificatePEM := tc.certificatePEM

			issuerKeyID := fmtIssuerKeyID(testSetup.Issuer.GetCAConfig().Certificate)
			if certificatePEM == "" && tc.serialNumber != nil {
				certificatePEM = createOCSPTestCertificatePEM(
					t,
					testSetup.Issuer.GetCAConfig().Certificate,
					testIssuerSigner(t, testSetup),
					tc.serialNumber,
				)
			}

			inner := cryptoutilCAStorage.NewMemoryStore()
			if tc.serialNumber != nil && tc.getErr == nil {
				err := inner.Store(context.Background(), &cryptoutilCAStorage.StoredCertificate{
					ID:               googleUuid.Must(googleUuid.NewV7()),
					RequestID:        googleUuid.Must(googleUuid.NewV7()),
					SerialNumber:     tc.serialNumber.Text(cryptoutilSharedMagic.HexBase),
					AuthorityKeyID:   issuerKeyID,
					CertificatePEM:   certificatePEM,
					Status:           tc.status,
					RevocationTime:   tc.revocationTime,
					RevocationReason: tc.revocationReason,
				})
				require.NoError(t, err)
			}

			handler := &Handler{storage: &errorStore{inner: inner, getErr: tc.getErr}}

			gotCert, gotRevoked, err := handler.lookupCertificateStatusBySerial(
				context.Background(),
				tc.serialNumber,
				issuerKeyID,
			)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantCertificate, gotCert != nil)
			require.Equal(t, tc.wantRevoked, gotRevoked != nil)
		})
	}
}

func createValidCSRPEM(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: testDomainNameExample},
	}, key)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCSR, Bytes: csrDER}))
}

func createValidCSRDER(t *testing.T) []byte {
	t.Helper()

	block, _ := pem.Decode([]byte(createValidCSRPEM(t)))
	require.NotNil(t, block)

	return block.Bytes
}

func createOCSPTestCertificatePEM(
	t *testing.T,
	issuerCertificate *x509.Certificate,
	issuerKey crypto.Signer,
	serialNumber *big.Int,
) string {
	t.Helper()

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	now := time.Now().UTC()
	der, err := x509.CreateCertificate(crand.Reader, &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: testDomainNameExample},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(24 * time.Hour),
	}, issuerCertificate, &leafKey.PublicKey, issuerKey)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: der}))
}

type failingBodyReader struct{}

func (failingBodyReader) Read([]byte) (int, error) {
	return 0, errors.New("injected request stream failure")
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func reasonPointer(value cryptoutilCAStorage.RevocationReason) *cryptoutilCAStorage.RevocationReason {
	return &value
}

func fmtIssuerKeyID(certificate *x509.Certificate) string {
	return fmt.Sprintf("%x", certificate.SubjectKeyId)
}
