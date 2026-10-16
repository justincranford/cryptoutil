// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"strings"
	"testing"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_GetBySerialNumberForIssuer(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	serialNumber := googleUuid.Must(googleUuid.NewV7()).String()
	certificate := newTestCert(serialNumber, gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	certificate.AuthorityKeyID = gormTestIssuerKeyA
	require.NoError(t, store.Store(context.Background(), certificate))

	tests := []struct {
		name         string
		serialNumber string
		issuerKeyID  string
		wantErr      error
	}{
		{
			name:         testMissingIssuerKeyID,
			serialNumber: serialNumber,
			issuerKeyID:  testWhitespaceOnly,
			wantErr:      ErrInvalidIssuerKeyID,
		},
		{
			name:         testCertificateNotFound,
			serialNumber: serialNumber,
			issuerKeyID:  "missing-issuer",
			wantErr:      ErrCertificateNotFound,
		},
		{
			name:         "case insensitive serial and issuer lookup",
			serialNumber: strings.ToUpper(serialNumber),
			issuerKeyID:  strings.ToUpper(gormTestIssuerKeyA),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := store.GetBySerialNumberForIssuer(context.Background(), tc.serialNumber, tc.issuerKeyID)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, certificate.ID, got.ID)
		})
	}
}

func TestMemoryStore_GetByRequestID(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	certificate := newTestCert(
		googleUuid.Must(googleUuid.NewV7()).String(),
		gormTestProfileID,
		gormTestRequesterA,
		testLiteralCNCA,
		newTestExpiration(),
	)
	require.NoError(t, store.Store(context.Background(), certificate))

	tests := []struct {
		name      string
		requestID googleUuid.UUID
		wantErr   error
	}{
		{name: "missing request ID", requestID: googleUuid.UUID{}, wantErr: ErrCertificateNotFound},
		{name: testUnknownRequestID, requestID: googleUuid.Must(googleUuid.NewV7()), wantErr: ErrCertificateNotFound},
		{name: testStoredRequestID, requestID: certificate.RequestID},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := store.GetByRequestID(context.Background(), tc.requestID)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, certificate.ID, got.ID)
		})
	}
}

func TestGORMStore_GetBySerialNumberForIssuer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		seedCount     int
		nilContext    bool
		issuerKeyID   string
		wantErr       error
		wantErrorText string
	}{
		{
			name:          "nil context",
			nilContext:    true,
			issuerKeyID:   gormTestIssuerKeyA,
			wantErrorText: "context is required",
		},
		{name: testMissingIssuerKeyID, issuerKeyID: testWhitespaceOnly, wantErr: ErrInvalidIssuerKeyID},
		{name: testCertificateNotFound, issuerKeyID: gormTestIssuerKeyA, wantErr: ErrCertificateNotFound},
		{name: "single matching certificate", seedCount: 1, issuerKeyID: strings.ToUpper(gormTestIssuerKeyA)},
		{name: "ambiguous across tenants", seedCount: 2, issuerKeyID: gormTestIssuerKeyA, wantErr: ErrCertificateAmbiguous},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
			require.NoError(t, db.AutoMigrate(&StoredCertificate{}))
			store, err := NewGORMStore(db)
			require.NoError(t, err)

			serialNumber := googleUuid.Must(googleUuid.NewV7()).String()

			tenantIDs := make([]googleUuid.UUID, tc.seedCount)
			for index := range tenantIDs {
				tenantIDs[index] = googleUuid.Must(googleUuid.NewV7())
				ctx := cryptoutilSharedDatabase.WithTenantContext(
					context.Background(),
					&cryptoutilSharedDatabase.TenantContext{TenantID: tenantIDs[index]},
				)
				certificate := newTestCert(
					serialNumber,
					gormTestProfileID,
					gormTestRequesterA,
					testLiteralCNCA,
					newTestExpiration(),
				)
				certificate.AuthorityKeyID = gormTestIssuerKeyA
				require.NoError(t, store.Store(ctx, certificate))
			}

			ctx := context.Background()
			if tc.nilContext {
				ctx = nil
			}

			got, err := store.GetBySerialNumberForIssuer(ctx, strings.ToUpper(serialNumber), tc.issuerKeyID)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			if tc.wantErrorText != "" {
				require.ErrorContains(t, err, tc.wantErrorText)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, serialNumber, got.SerialNumber)
			require.Equal(t, gormTestIssuerKeyA, got.AuthorityKeyID)
		})
	}
}

func TestGORMStore_GetByRequestID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		nilRequestID     bool
		tenantContext    bool
		storeCertificate bool
		wantErr          error
		wantErrorText    string
	}{
		{name: "nil request ID", nilRequestID: true, wantErr: ErrCertificateNotFound},
		{name: "missing tenant context", wantErrorText: "tenant context is required"},
		{name: testUnknownRequestID, tenantContext: true, wantErr: ErrCertificateNotFound},
		{name: testStoredRequestID, tenantContext: true, storeCertificate: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
			require.NoError(t, db.AutoMigrate(&StoredCertificate{}))
			store, err := NewGORMStore(db)
			require.NoError(t, err)

			tenantID := googleUuid.Must(googleUuid.NewV7())
			requestID := googleUuid.Must(googleUuid.NewV7())

			ctx := context.Background()
			if tc.tenantContext {
				ctx = cryptoutilSharedDatabase.WithTenantContext(
					ctx,
					&cryptoutilSharedDatabase.TenantContext{TenantID: tenantID},
				)
			}

			certificate := newTestCert(
				googleUuid.Must(googleUuid.NewV7()).String(),
				gormTestProfileID,
				gormTestRequesterA,
				testLiteralCNCA,
				newTestExpiration(),
			)

			certificate.RequestID = requestID
			if tc.storeCertificate {
				require.NoError(t, store.Store(ctx, certificate))
			}

			if tc.nilRequestID {
				requestID = googleUuid.UUID{}
			}

			got, err := store.GetByRequestID(ctx, requestID)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			if tc.wantErrorText != "" {
				require.ErrorContains(t, err, tc.wantErrorText)

				return
			}

			require.NoError(t, err)
			require.Equal(t, certificate.ID, got.ID)
		})
	}
}
