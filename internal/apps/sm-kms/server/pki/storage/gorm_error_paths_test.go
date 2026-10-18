// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testGORMErrorProfileID   = "gorm-error-profile"
	testGORMErrorRequester   = "gorm-error-requester"
	testGORMErrorIssuerDN    = "CN=gorm-error-issuer"
	testGORMErrorIssuerKeyID = "issuer-key"
)

func TestGORMStores_DatabaseFailurePaths(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	certificateStore, err := NewGORMStore(db)
	require.NoError(t, err)
	authorityStore, err := NewGORMAuthorityStore(db)
	require.NoError(t, err)

	tenantID := googleUuid.Must(googleUuid.NewV7())
	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{
		TenantID: tenantID,
	})

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "store",
			run: func() error {
				certificate := newTestCert("store-error-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())

				return certificateStore.Store(ctx, certificate)
			},
		},
		{
			name: "get",
			run: func() error {
				_, getErr := certificateStore.Get(ctx, googleUuid.Must(googleUuid.NewV7()))

				return getErr
			},
		},
		{
			name: "get by serial number",
			run: func() error {
				_, getErr := certificateStore.GetBySerialNumber(ctx, "missing-table-serial")

				return getErr
			},
		},
		{
			name: "get by serial number and issuer",
			run: func() error {
				_, getErr := certificateStore.GetBySerialNumberForIssuer(ctx, "missing-table-serial", "issuer-key")

				return getErr
			},
		},
		{
			name: "get by request ID",
			run: func() error {
				_, getErr := certificateStore.GetByRequestID(ctx, googleUuid.Must(googleUuid.NewV7()))

				return getErr
			},
		},
		{
			name: "list count",
			run: func() error {
				_, _, listErr := certificateStore.List(ctx, nil)

				return listErr
			},
		},
		{
			name: "update",
			run: func() error {
				certificate := newTestCert("update-error-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())

				return certificateStore.Update(ctx, certificate)
			},
		},
		{
			name: "delete",
			run: func() error {
				return certificateStore.Delete(ctx, googleUuid.Must(googleUuid.NewV7()))
			},
		},
		{
			name: "revoke",
			run: func() error {
				return certificateStore.Revoke(ctx, googleUuid.Must(googleUuid.NewV7()), ReasonKeyCompromise)
			},
		},
		{
			name: "get revoked by tenant",
			run: func() error {
				_, getErr := certificateStore.GetRevoked(ctx, testGORMErrorIssuerDN)

				return getErr
			},
		},
		{
			name: "get revoked by issuer key ID",
			run: func() error {
				_, getErr := certificateStore.GetRevokedByIssuerKeyID(ctx, "issuer-key")

				return getErr
			},
		},
		{
			name: "count by status",
			run: func() error {
				_, countErr := certificateStore.CountByStatus(ctx)

				return countErr
			},
		},
		{
			name: "allocate CRL number",
			run: func() error {
				_, allocateErr := certificateStore.AllocateCRLNumber(ctx, "issuer-key")

				return allocateErr
			},
		},
		{
			name: "get authority",
			run: func() error {
				_, getErr := authorityStore.Get(ctx, "missing-authority")

				return getErr
			},
		},
		{
			name: "get any authority",
			run: func() error {
				_, getErr := authorityStore.GetAny(ctx)

				return getErr
			},
		},
		{
			name: "list authorities",
			run: func() error {
				_, listErr := authorityStore.List(ctx)

				return listErr
			},
		},
		{
			name: "create authority",
			run: func() error {
				return authorityStore.Create(ctx, &StoredAuthority{ID: "missing-table-authority"})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Error(t, tc.run())
		})
	}
}

func TestGORMStore_ListAppliesEveryFilter(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{
		TenantID: googleUuid.Must(googleUuid.NewV7()),
	})
	certificate := newTestCert("filtered-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())
	require.NoError(t, store.Store(ctx, certificate))

	status := StatusActive
	profileID := testGORMErrorProfileID
	requesterID := testGORMErrorRequester
	issuerDN := testGORMErrorIssuerDN
	notAfterFrom := certificate.NotAfter.Add(-1)
	notAfterTo := certificate.NotAfter.Add(1)

	certificates, total, err := store.List(ctx, &ListFilter{
		Status:       &status,
		ProfileID:    &profileID,
		RequesterID:  &requesterID,
		IssuerDN:     &issuerDN,
		NotAfterFrom: &notAfterFrom,
		NotAfterTo:   &notAfterTo,
		Limit:        1,
		Offset:       0,
	})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, certificates, 1)
	require.Equal(t, certificate.ID, certificates[0].ID)
}

func TestGORMStore_RequiresContextAndIssuerKeyID(t *testing.T) {
	t.Parallel()

	_, err := NewGORMStore(nil)
	require.ErrorContains(t, err, "database is required")

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	var nilContext context.Context

	_, err = store.GetBySerialNumberForIssuer(nilContext, "serial", "issuer-key")
	require.ErrorContains(t, err, "context is required")
	_, err = store.GetBySerialNumberForIssuer(context.Background(), "serial", " ")
	require.ErrorIs(t, err, ErrInvalidIssuerKeyID)
	_, err = store.GetByRequestID(context.Background(), googleUuid.Nil)
	require.ErrorIs(t, err, ErrCertificateNotFound)
	_, err = store.AllocateCRLNumber(nilContext, "issuer-key")
	require.ErrorContains(t, err, "context is required")
	_, err = store.AllocateCRLNumber(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidIssuerKeyID)
	_, err = store.GetRevokedByIssuerKeyID(nilContext, "issuer-key")
	require.ErrorContains(t, err, "context is required")
	_, err = store.GetRevokedByIssuerKeyID(context.Background(), " ")
	require.ErrorIs(t, err, ErrInvalidIssuerKeyID)

	_, err = requireTenantID(nilContext)
	require.ErrorContains(t, err, "context is required")
}

func TestGORMAuthorityStore_ConstructorAndValidation(t *testing.T) {
	t.Parallel()

	_, err := NewGORMAuthorityStore(nil)
	require.ErrorContains(t, err, "database is required")

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	store, err := NewGORMAuthorityStore(db)
	require.NoError(t, err)

	var nilContext context.Context

	_, err = store.Get(nilContext, "authority")
	require.ErrorContains(t, err, "context is required")
	_, err = store.GetAny(nilContext)
	require.ErrorContains(t, err, "context is required")
	_, err = store.List(nilContext)
	require.ErrorContains(t, err, "context is required")
	require.ErrorContains(t, store.Create(nilContext, &StoredAuthority{ID: "authority"}), "context is required")
	require.ErrorContains(t, store.Create(context.Background(), nil), "PKI authority is required")
}

func TestGORMStore_UpdateAndDeleteMissingCertificate(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{
		TenantID: googleUuid.Must(googleUuid.NewV7()),
	})
	certificate := newTestCert("missing-row-update", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())
	require.ErrorIs(t, store.Update(ctx, certificate), ErrCertificateNotFound)
	require.ErrorIs(t, store.Delete(ctx, googleUuid.Must(googleUuid.NewV7())), ErrCertificateNotFound)
	require.ErrorIs(t, store.Revoke(ctx, googleUuid.Must(googleUuid.NewV7()), ReasonKeyCompromise), ErrCertificateNotFound)
}

func TestGORMStore_ListOffsetAndUpdateNilMetadata(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{
		TenantID: googleUuid.Must(googleUuid.NewV7()),
	})
	first := newTestCert("offset-serial-first", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())
	second := newTestCert("offset-serial-second", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())

	require.NoError(t, store.Store(ctx, first))
	require.NoError(t, store.Store(ctx, second))

	certificates, total, err := store.List(ctx, &ListFilter{Limit: 1, Offset: 1})
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, certificates, 1)

	first.Metadata = nil
	require.NoError(t, store.Update(ctx, first))
	updated, err := store.Get(ctx, first.ID)
	require.NoError(t, err)
	require.Empty(t, updated.Metadata)
}

func TestGORMStore_StoreGeneratesMissingCertificateID(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{
		TenantID: googleUuid.Must(googleUuid.NewV7()),
	})
	certificate := newTestCert("generated-id-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())
	certificate.ID = googleUuid.Nil
	require.NoError(t, store.Store(ctx, certificate))
	require.NotEqual(t, googleUuid.Nil, certificate.ID)
}

func TestGORMStore_StoreInsertFailure(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	triggerName := "reject_pki_certificate_insert"
	trigger := "CREATE TRIGGER " + triggerName + " BEFORE INSERT ON " + (StoredCertificate{}).TableName() +
		" BEGIN SELECT RAISE(FAIL, 'injected insert failure'); END"
	require.NoError(t, db.Exec(trigger).Error)

	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{
		TenantID: googleUuid.Must(googleUuid.NewV7()),
	})
	certificate := newTestCert("trigger-error-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())
	require.ErrorContains(t, store.Store(ctx, certificate), "injected insert failure")
}

func TestGORMAuthorityStore_GetAnyEmpty(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredAuthority{}))
	store, err := NewGORMAuthorityStore(db)
	require.NoError(t, err)

	_, err = store.GetAny(context.Background())
	require.ErrorIs(t, err, ErrAuthorityNotFound)
}

func TestGORMStore_AllocateCRLNumberExhaustedSequence(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCRLSequence{}))
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	require.NoError(t, db.Create(&StoredCRLSequence{
		IssuerKeyID: "exhausted-issuer",
		Number:      math.MaxInt64,
	}).Error)

	_, err = store.AllocateCRLNumber(context.Background(), "exhausted-issuer")
	require.ErrorContains(t, err, "exhausted or missing")
}

type cancelOnDoneContext struct {
	context.Context
	cancel context.CancelFunc
}

func (ctx cancelOnDoneContext) Done() <-chan struct{} {
	ctx.cancel()

	return ctx.Context.Done()
}

func TestWaitForCRLSequenceRetry(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name: "retry delay elapsed",
			ctx:  context.Background(),
		},
		{
			name:    "context canceled while waiting",
			ctx:     cancelOnDoneContext{Context: ctx, cancel: cancel},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := waitForCRLSequenceRetry(tc.ctx, 0)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestMemoryStore_RejectsOrphanedLookupIndexes(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	serialID := googleUuid.Must(googleUuid.NewV7())
	requestID := googleUuid.Must(googleUuid.NewV7())
	store.bySerial["orphan-serial"] = serialID
	store.byRequestID[requestID] = serialID

	_, err := store.GetBySerialNumber(context.Background(), "orphan-serial")
	require.ErrorIs(t, err, ErrCertificateNotFound)
	_, err = store.GetByRequestID(context.Background(), requestID)
	require.ErrorIs(t, err, ErrCertificateNotFound)
}

func TestMemoryStore_GetBySerialNumberForIssuerAmbiguous(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	first := newTestCert("ambiguous-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, time.Now().UTC().Add(time.Hour))
	first.AuthorityKeyID = testGORMErrorIssuerKeyID
	require.NoError(t, store.Store(context.Background(), first))

	second := newTestCert("ambiguous-serial", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, time.Now().UTC().Add(time.Hour))
	second.AuthorityKeyID = testGORMErrorIssuerKeyID
	store.certificates[second.ID] = second

	_, err := store.GetBySerialNumberForIssuer(context.Background(), "ambiguous-serial", testGORMErrorIssuerKeyID)
	require.ErrorIs(t, err, ErrCertificateAmbiguous)
}

func TestMemoryStore_UpdatePreservesRevocation(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	certificate := newTestCert("memory-revoke-update", testGORMErrorProfileID, testGORMErrorRequester, testGORMErrorIssuerDN, newTestExpiration())
	require.NoError(t, store.Store(context.Background(), certificate))

	staleCertificate, err := store.Get(context.Background(), certificate.ID)
	require.NoError(t, err)

	staleSnapshot := *staleCertificate

	require.NoError(t, store.Revoke(context.Background(), certificate.ID, ReasonKeyCompromise))

	staleSnapshot.Status = StatusActive
	require.NoError(t, store.Update(context.Background(), &staleSnapshot))

	updated, err := store.Get(context.Background(), certificate.ID)
	require.NoError(t, err)
	require.Equal(t, StatusRevoked, updated.Status)
	require.NotNil(t, updated.RevocationTime)
	require.NotNil(t, updated.RevocationReason)
	require.Equal(t, ReasonKeyCompromise, *updated.RevocationReason)
}

func TestAllocateCRLNumberWithRetry_WaitCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := allocateCRLNumberWithRetry(
		ctx,
		cryptoutilSharedMagic.TestDatabaseSQLite,
		func() (int64, error) { return 0, errors.New("database is locked") },
		waitForCRLSequenceRetry,
	)
	require.ErrorIs(t, err, context.Canceled)

	number, err := allocateCRLNumberWithRetry(
		context.Background(),
		cryptoutilSharedMagic.TestDatabaseSQLite,
		func() (int64, error) { return 42, nil },
		waitForCRLSequenceRetry,
	)
	require.NoError(t, err)
	require.Equal(t, int64(42), number)
}

func TestNewGORMStore_ClosedDatabase(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	_, err = NewGORMStore(db)
	require.ErrorContains(t, err, "failed to determine SQLite version")
}

func TestGORMStore_AllocateCRLNumberRejectsNonpositiveResult(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCRLSequence{}))
	store, err := NewGORMStore(db)
	require.NoError(t, err)

	require.NoError(t, db.Create(&StoredCRLSequence{IssuerKeyID: "negative-sequence", Number: -1}).Error)

	_, err = store.AllocateCRLNumber(context.Background(), "negative-sequence")
	require.ErrorContains(t, err, "returned invalid number")
}
