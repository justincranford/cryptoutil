// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"encoding/hex"
	"errors"
	http "net/http"
	"strings"
	"testing"
	"time"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStrictServer_CAResponses(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	strict := NewStrictServer(&Handler{issuer: setup.Issuer})
	ctx := strictTestContext()

	listResponse := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.ListCAs(ctx, cryptoutilApiCaServer.ListCAsRequestObject{})
	})
	listBody := strictResponseBody[cryptoutilApiCaServer.CAListResponse](t, listResponse)
	require.Len(t, listBody.Authorities, 1)
	require.Equal(t, cryptoutilApiCaServer.CASummaryTypeIntermediate, listBody.Authorities[0].Type)

	caID := setup.Issuer.GetCAConfig().Name
	getResponse := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.GetCA(ctx, cryptoutilApiCaServer.GetCARequestObject{CAID: caID})
	})
	getBody := strictResponseBody[cryptoutilApiCaServer.CAResponse](t, getResponse)
	require.Equal(t, caID, getBody.ID)
	require.Equal(t, cryptoutilApiCaServer.CAResponseTypeIntermediate, getBody.Type)
	require.NotEmpty(t, getBody.CertificatePEM)
	require.NotNil(t, getBody.PathLength)

	assertStrictCall(t, http.StatusNotFound, func() (any, error) {
		return strict.GetCA(ctx, cryptoutilApiCaServer.GetCARequestObject{CAID: "unknown-ca"})
	})

	rootStrict := NewStrictServer(&Handler{issuer: setup.RootIssuer})
	rootResponse := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return rootStrict.ListCAs(ctx, cryptoutilApiCaServer.ListCAsRequestObject{})
	})
	rootBody := strictResponseBody[cryptoutilApiCaServer.CAListResponse](t, rootResponse)
	require.Equal(t, cryptoutilApiCaServer.CASummaryTypeRoot, rootBody.Authorities[0].Type)

	tests := []struct {
		name   string
		issuer *cryptoutilCAServiceIssuer.Issuer
		want   int
	}{
		{name: strictHandlerTestIssuerAbsent, issuer: nil, want: http.StatusInternalServerError},
		{name: "CA configuration absent", issuer: &cryptoutilCAServiceIssuer.Issuer{}, want: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := NewStrictServer(&Handler{issuer: tc.issuer})
			assertStrictCall(t, tc.want, func() (any, error) {
				return server.ListCAs(ctx, cryptoutilApiCaServer.ListCAsRequestObject{})
			})
		})
	}
}

func TestStrictServer_GetCRLFormatsAndRevocations(t *testing.T) {
	t.Parallel()

	setup, handler, strict := newStrictServerFixture(t)
	handler.crlService = newStrictCRLService(t, setup)

	caCert := setup.Issuer.GetCAConfig().Certificate
	issuerKeyID := strings.ToLower(hex.EncodeToString(caCert.SubjectKeyId))
	stored := strictTestStoredCertificate(t, "2A")
	now := time.Now().UTC()
	reason := cryptoutilCAStorage.ReasonKeyCompromise
	stored.Status = cryptoutilCAStorage.StatusRevoked
	stored.RevocationTime = &now
	stored.RevocationReason = &reason
	stored.AuthorityKeyID = issuerKeyID
	require.NoError(t, handler.storage.Store(context.Background(), stored))

	tests := []struct {
		name        string
		format      *cryptoutilApiCaServer.GetCRLParamsFormat
		contentType string
		filename    string
	}{
		{
			name:        "default DER",
			contentType: cryptoutilSharedMagic.PKIContentTypeCRLDER,
			filename:    setup.Issuer.GetCAConfig().Name + ".crl",
		},
		{
			name:        "PEM",
			format:      strictCRLFormat(cryptoutilSharedMagic.PKICRLFormatPEM),
			contentType: cryptoutilSharedMagic.PKIContentTypeCRLPEM,
			filename:    setup.Issuer.GetCAConfig().Name + ".crl.pem",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := assertStrictCall(t, http.StatusOK, func() (any, error) {
				return strict.GetCRL(strictTestContext(), cryptoutilApiCaServer.GetCRLRequestObject{
					CAID: setup.Issuer.GetCAConfig().Name,
					Params: cryptoutilApiCaServer.GetCRLParams{
						Format: tc.format,
					},
				})
			})
			require.NotEmpty(t, result.body)
			require.Equal(t, tc.contentType, result.contentType)
			require.Equal(t, "attachment; filename=\""+tc.filename+"\"", result.headers[0][1])
		})
	}
}

func TestStrictServer_GetCRLRejectsInvalidStoredRevocations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata bool
		serial   string
	}{
		{name: strictHandlerTestMissingRevocationMetadata, metadata: false, serial: "2A"},
		{name: "invalid hexadecimal serial", metadata: true, serial: strictHandlerTestSerialNotHex},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setup, handler, strict := newStrictServerFixture(t)
			handler.crlService = newStrictCRLService(t, setup)
			stored := strictTestStoredCertificate(t, tc.serial)
			stored.Status = cryptoutilCAStorage.StatusRevoked
			stored.AuthorityKeyID = hex.EncodeToString(setup.Issuer.GetCAConfig().Certificate.SubjectKeyId)

			if tc.metadata {
				now := time.Now().UTC()
				reason := cryptoutilCAStorage.ReasonKeyCompromise
				stored.RevocationTime = &now
				stored.RevocationReason = &reason
			}

			require.NoError(t, handler.storage.Store(context.Background(), stored))
			result := assertStrictCall(t, http.StatusInternalServerError, func() (any, error) {
				return strict.GetCRL(strictTestContext(), cryptoutilApiCaServer.GetCRLRequestObject{
					CAID: setup.Issuer.GetCAConfig().Name,
				})
			})
			require.Contains(t, strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, result).Error, "storage_error")
		})
	}
}

func TestStrictServer_ListAndGetCertificates(t *testing.T) {
	t.Parallel()

	_, handler, strict := newStrictServerFixture(t)
	ctx := strictTestContext()

	certificates := []*cryptoutilCAStorage.StoredCertificate{
		strictTestStoredCertificate(t, "101"),
		strictTestStoredCertificate(t, "102"),
		strictTestStoredCertificate(t, "103"),
	}

	certificates[2].Status = cryptoutilCAStorage.StatusRevoked
	for _, cert := range certificates {
		require.NoError(t, handler.storage.Store(ctx, cert))
	}

	page, pageSize := 2, 1
	profileID := cryptoutilSharedMagic.PKIProfileIDTLSServer
	status := cryptoutilApiCaServer.CertificateStatusActive
	listResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.ListCertificates(ctx, cryptoutilApiCaServer.ListCertificatesRequestObject{
			Params: cryptoutilApiCaServer.ListCertificatesParams{
				Page:     &page,
				PageSize: &pageSize,
				Profile:  &profileID,
				Status:   &status,
			},
		})
	})
	listBody := strictResponseBody[cryptoutilApiCaServer.CertificateListResponse](t, listResult)
	require.Equal(t, page, listBody.Page)
	require.Equal(t, pageSize, listBody.PageSize)
	require.Equal(t, 2, listBody.Total)
	require.Len(t, listBody.Certificates, 1)

	getResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.GetCertificate(ctx, cryptoutilApiCaServer.GetCertificateRequestObject{
			SerialNumber: certificates[0].SerialNumber,
		})
	})
	getBody := strictResponseBody[cryptoutilApiCaServer.CertificateResponse](t, getResult)
	require.Equal(t, certificates[0].SerialNumber, getBody.SerialNumber)
	require.Equal(t, certificates[0].CertificatePEM, getBody.CertificatePEM)

	chainResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.GetCertificateChain(ctx, cryptoutilApiCaServer.GetCertificateChainRequestObject{
			SerialNumber: certificates[0].SerialNumber,
		})
	})
	chainBody := strictResponseBody[cryptoutilApiCaServer.CertificateChainResponse](t, chainResult)
	require.Len(t, chainBody.Certificates, 1)
	require.Equal(t, certificates[0].CertificatePEM, chainBody.Certificates[0].CertificatePEM)

	assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.ListCertificates(ctx, cryptoutilApiCaServer.ListCertificatesRequestObject{})
	})
}

func TestStrictServer_ListAndGetCertificateErrors(t *testing.T) {
	t.Parallel()

	ctx := strictTestContext()
	storeErr := errors.New("injected strict storage failure")
	strict := NewStrictServer(&Handler{storage: &errorStore{
		inner:   cryptoutilCAStorage.NewMemoryStore(),
		listErr: storeErr,
		getErr:  storeErr,
	}})

	listResult := assertStrictCall(t, http.StatusInternalServerError, func() (any, error) {
		return strict.ListCertificates(ctx, cryptoutilApiCaServer.ListCertificatesRequestObject{})
	})
	listError := strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, listResult)
	require.NotNil(t, listError.Message)
	require.Contains(t, *listError.Message, "injected")

	assertStrictCall(t, http.StatusBadRequest, func() (any, error) {
		return strict.GetCertificate(ctx, cryptoutilApiCaServer.GetCertificateRequestObject{})
	})
	assertStrictCall(t, http.StatusBadRequest, func() (any, error) {
		return strict.GetCertificateChain(ctx, cryptoutilApiCaServer.GetCertificateChainRequestObject{})
	})
	assertStrictCall(t, http.StatusInternalServerError, func() (any, error) {
		return strict.GetCertificate(ctx, cryptoutilApiCaServer.GetCertificateRequestObject{SerialNumber: strictHandlerTestSerialX})
	})
	assertStrictCall(t, http.StatusInternalServerError, func() (any, error) {
		return strict.GetCertificateChain(ctx, cryptoutilApiCaServer.GetCertificateChainRequestObject{SerialNumber: strictHandlerTestSerialX})
	})

	emptyStoreStrict := NewStrictServer(&Handler{storage: cryptoutilCAStorage.NewMemoryStore()})

	assertStrictCall(t, http.StatusNotFound, func() (any, error) {
		return emptyStoreStrict.GetCertificate(ctx, cryptoutilApiCaServer.GetCertificateRequestObject{SerialNumber: strictServerTestMissing})
	})
	assertStrictCall(t, http.StatusNotFound, func() (any, error) {
		return emptyStoreStrict.GetCertificateChain(ctx, cryptoutilApiCaServer.GetCertificateChainRequestObject{SerialNumber: strictServerTestMissing})
	})
}

func TestStrictServer_ProfileResponses(t *testing.T) {
	t.Parallel()

	_, _, strict := newStrictServerFixture(t)

	listResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.ListProfiles(strictTestContext(), cryptoutilApiCaServer.ListProfilesRequestObject{})
	})
	listBody := strictResponseBody[cryptoutilApiCaServer.ProfileListResponse](t, listResult)
	require.Len(t, listBody.Profiles, 1)
	require.Equal(t, testProfileNameTLSServer, listBody.Profiles[0].Name)

	profileID := cryptoutilSharedMagic.PKIProfileIDTLSServer
	getResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.GetProfile(strictTestContext(), cryptoutilApiCaServer.GetProfileRequestObject{ProfileID: profileID})
	})
	profileBody := strictResponseBody[*cryptoutilApiCaServer.ProfileResponse](t, getResult)
	require.NotNil(t, profileBody)
	require.Equal(t, profileID, profileBody.ID)
	require.Equal(t, testProfileNameTLSServer, profileBody.Name)
	require.NotNil(t, profileBody.Category)

	assertStrictCall(t, http.StatusNotFound, func() (any, error) {
		return strict.GetProfile(strictTestContext(), cryptoutilApiCaServer.GetProfileRequestObject{ProfileID: strictServerTestMissing})
	})

	emptyStrict := NewStrictServer(&Handler{profiles: map[string]*ProfileConfig{}})
	emptyResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return emptyStrict.ListProfiles(strictTestContext(), cryptoutilApiCaServer.ListProfilesRequestObject{})
	})
	require.Empty(t, strictResponseBody[cryptoutilApiCaServer.ProfileListResponse](t, emptyResult).Profiles)
}

func TestStrictServer_GetEnrollmentStatusResponses(t *testing.T) {
	t.Parallel()

	ctx := strictTestContext()
	_, handler, strict := newStrictServerFixture(t)
	stored := strictTestStoredCertificate(t, "201")
	require.NoError(t, handler.storage.Store(ctx, stored))

	result := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.GetEnrollmentStatus(ctx, cryptoutilApiCaServer.GetEnrollmentStatusRequestObject{
			RequestID: stored.RequestID,
		})
	})
	body := strictResponseBody[cryptoutilApiCaServer.EnrollmentStatusResponse](t, result)
	require.Equal(t, stored.RequestID, body.RequestID)
	require.Equal(t, cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, body.Status)
	require.NotNil(t, body.Certificate)
	require.Equal(t, stored.SerialNumber, body.Certificate.SerialNumber)

	unknownID, err := googleUuid.NewV7()
	require.NoError(t, err)
	assertStrictCall(t, http.StatusNotFound, func() (any, error) {
		return strict.GetEnrollmentStatus(ctx, cryptoutilApiCaServer.GetEnrollmentStatusRequestObject{
			RequestID: unknownID,
		})
	})

	errorStrict := NewStrictServer(&Handler{storage: &errorStore{
		inner:  cryptoutilCAStorage.NewMemoryStore(),
		getErr: errors.New("injected request lookup failure"),
	}})

	assertStrictCall(t, http.StatusInternalServerError, func() (any, error) {
		return errorStrict.GetEnrollmentStatus(ctx, cryptoutilApiCaServer.GetEnrollmentStatusRequestObject{
			RequestID: unknownID,
		})
	})
}

func strictCRLFormat(value string) *cryptoutilApiCaServer.GetCRLParamsFormat {
	format := cryptoutilApiCaServer.GetCRLParamsFormat(value)

	return &format
}
