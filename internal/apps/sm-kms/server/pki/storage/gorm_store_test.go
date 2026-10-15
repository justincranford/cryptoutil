// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"database/sql"
	"math/big"
	"path/filepath"
	"sync"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	gormTestProfileID        = "server-profile"
	gormTestRequesterA       = "requester-a"
	gormTestRequesterB       = "requester-b"
	gormTestUpdatedProfileID = "updated-profile"
	gormTestIssuerKeyA       = "issuer-key-a"
	gormTestSharedIssuerDN   = "CN=shared"
	testLiteralSource        = "source"
	testLiteralEnrollment    = "enrollment"
)

func TestGORMStore_TenantScopedLifecycle(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}, &StoredAuthority{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	tenantA := googleUuid.Must(googleUuid.NewV7())
	tenantB := googleUuid.Must(googleUuid.NewV7())
	requestID := googleUuid.Must(googleUuid.NewV7())
	serialNumber := googleUuid.Must(googleUuid.NewV7()).String()
	ctxA := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: tenantA})
	ctxB := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: tenantB})

	certA := newTestCert(serialNumber, gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	certA.RequestID = requestID
	certA.Metadata = map[string]string{testLiteralSource: testLiteralEnrollment}
	require.NoError(t, store.Store(ctxA, certA))

	certB := newTestCert(serialNumber, gormTestProfileID, gormTestRequesterB, testLiteralCNCA, newTestExpiration())
	require.NoError(t, store.Store(ctxB, certB))

	gotA, err := store.Get(ctxA, certA.ID)
	require.NoError(t, err)
	require.Equal(t, tenantA, gotA.TenantID)
	require.Equal(t, requestID, gotA.RequestID)
	require.Equal(t, certA.Metadata, gotA.Metadata)
	require.Equal(t, certA.SerialNumber, gotA.SerialNumber)

	gotBySerial, err := store.GetBySerialNumber(ctxA, serialNumber)
	require.NoError(t, err)
	require.Equal(t, certA.ID, gotBySerial.ID)

	gotByRequest, err := store.GetByRequestID(ctxA, requestID)
	require.NoError(t, err)
	require.Equal(t, certA.ID, gotByRequest.ID)

	_, err = store.Get(ctxB, certA.ID)
	require.ErrorIs(t, err, ErrCertificateNotFound)
	_, err = store.GetByRequestID(ctxB, requestID)
	require.ErrorIs(t, err, ErrCertificateNotFound)

	duplicate := newTestCert(serialNumber, gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	require.ErrorIs(t, store.Store(ctxA, duplicate), ErrCertificateExists)

	activeStatus := StatusActive
	filter := &ListFilter{Status: &activeStatus}
	certificates, total, err := store.List(ctxA, filter)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, certificates, 1)

	require.NoError(t, store.Revoke(ctxA, certA.ID, ReasonKeyCompromise))
	require.ErrorIs(t, store.Revoke(ctxA, certA.ID, ReasonKeyCompromise), ErrCertificateRevoked)

	revoked, err := store.GetRevoked(ctxA, testLiteralCNCA)
	require.NoError(t, err)
	require.Len(t, revoked, 1)
	require.Equal(t, certA.ID, revoked[0].ID)

	counts, err := store.CountByStatus(ctxA)
	require.NoError(t, err)
	require.Equal(t, int64(1), counts[StatusRevoked])

	gotA.ProfileID = gormTestUpdatedProfileID
	require.NoError(t, store.Update(ctxA, gotA))

	updated, err := store.Get(ctxA, certA.ID)
	require.NoError(t, err)
	require.Equal(t, gormTestUpdatedProfileID, updated.ProfileID)

	require.NoError(t, store.Delete(ctxA, certA.ID))
	_, err = store.Get(ctxA, certA.ID)
	require.ErrorIs(t, err, ErrCertificateNotFound)

	require.NoError(t, store.Close())
}

func TestGORMStore_RequiresTenantContext(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	_, _, err = store.List(context.Background(), nil)
	require.ErrorContains(t, err, "tenant context is required")
	require.ErrorContains(t, store.Store(context.Background(), nil), "invalid certificate")
	require.ErrorContains(t, store.Delete(context.Background(), googleUuid.Must(googleUuid.NewV7())), "tenant context is required")
}

func TestGORMAuthorityStore_RoundTrip(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredAuthority{}))

	store, err := NewGORMAuthorityStore(db)
	require.NoError(t, err)

	_, err = store.Get(context.Background(), "missing-authority")
	require.ErrorIs(t, err, ErrAuthorityNotFound)

	authority := &StoredAuthority{
		ID:                        googleUuid.Must(googleUuid.NewV7()).String(),
		CertificateDER:            []byte("certificate"),
		PrivateKeyEncrypted:       []byte("encrypted-key"),
		PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
	}
	require.NoError(t, store.Create(context.Background(), authority))

	got, err := store.Get(context.Background(), authority.ID)
	require.NoError(t, err)
	require.Equal(t, authority.CertificateDER, got.CertificateDER)
	require.Equal(t, authority.PrivateKeyEncrypted, got.PrivateKeyEncrypted)
	require.Equal(t, authority.PrivateKeyEnvelopeVersion, got.PrivateKeyEnvelopeVersion)

	gotAny, err := store.GetAny(context.Background())
	require.NoError(t, err)
	require.Equal(t, authority.ID, gotAny.ID)

	listed, err := store.List(context.Background())
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, authority.ID, listed[0].ID)

	secondAuthority := &StoredAuthority{
		ID:                        googleUuid.Must(googleUuid.NewV7()).String(),
		CertificateDER:            []byte("second-certificate"),
		PrivateKeyEncrypted:       []byte("second-encrypted-key"),
		PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
	}
	require.NoError(t, store.Create(context.Background(), secondAuthority))
	_, err = store.GetAny(context.Background())
	require.ErrorIs(t, err, ErrMultipleAuthorities)
	listed, err = store.List(context.Background())
	require.NoError(t, err)
	require.Len(t, listed, 2)
	require.Equal(t, []string{authority.ID, secondAuthority.ID}, []string{listed[0].ID, listed[1].ID})

	var nilContext context.Context

	_, err = store.List(nilContext)
	require.ErrorContains(t, err, "context is required")
}

func TestGORMStore_RequestIDUniquenessIsTenantScopedAndRequiresRequestID(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	tenantA := googleUuid.Must(googleUuid.NewV7())
	tenantB := googleUuid.Must(googleUuid.NewV7())
	requestID := googleUuid.Must(googleUuid.NewV7())
	ctxA := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: tenantA})
	ctxB := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: tenantB})

	first := newTestCert("request-id-serial-a", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	first.RequestID = requestID
	require.NoError(t, store.Store(ctxA, first))

	first.RequestID = googleUuid.UUID{}
	require.ErrorIs(t, store.Update(ctxA, first), ErrInvalidRequestID)

	duplicate := newTestCert("request-id-serial-b", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	duplicate.RequestID = requestID
	require.ErrorIs(t, store.Store(ctxA, duplicate), ErrCertificateExists)

	otherTenant := newTestCert("request-id-serial-c", gormTestProfileID, gormTestRequesterB, testLiteralCNCA, newTestExpiration())
	otherTenant.RequestID = requestID
	require.NoError(t, store.Store(ctxB, otherTenant))

	nilRequestOne := newTestCert("nil-request-serial-a", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	nilRequestOne.RequestID = googleUuid.UUID{}
	nilRequestTwo := newTestCert("nil-request-serial-b", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	nilRequestTwo.RequestID = googleUuid.UUID{}

	require.ErrorIs(t, store.Store(ctxA, nilRequestOne), ErrInvalidRequestID)
	require.ErrorIs(t, store.Store(ctxA, nilRequestTwo), ErrInvalidRequestID)
}

func TestGORMStore_UpdateCannotUndoRevocation(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	tenantID := googleUuid.Must(googleUuid.NewV7())
	ctx := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: tenantID})
	certificate := newTestCert("stale-revocation-serial", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	certificate.AuthorityKeyID = gormTestIssuerKeyA
	require.NoError(t, store.Store(ctx, certificate))

	staleSnapshot, err := store.Get(ctx, certificate.ID)
	require.NoError(t, err)
	require.NoError(t, store.Revoke(ctx, certificate.ID, ReasonKeyCompromise))

	staleSnapshot.ProfileID = gormTestUpdatedProfileID
	require.NoError(t, store.Update(ctx, staleSnapshot))

	updated, err := store.Get(ctx, certificate.ID)
	require.NoError(t, err)
	require.Equal(t, StatusRevoked, updated.Status)
	require.NotNil(t, updated.RevocationTime)
	require.NotNil(t, updated.RevocationReason)
	require.Equal(t, ReasonKeyCompromise, *updated.RevocationReason)
	require.Equal(t, gormTestUpdatedProfileID, updated.ProfileID)
}

func TestGORMStore_GetRevokedByIssuerKeyIDAggregatesTenants(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	require.NoError(t, db.AutoMigrate(&StoredCertificate{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	ctxA := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: googleUuid.Must(googleUuid.NewV7())})
	ctxB := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: googleUuid.Must(googleUuid.NewV7())})
	ctxC := cryptoutilSharedDatabase.WithTenantContext(context.Background(), &cryptoutilSharedDatabase.TenantContext{TenantID: googleUuid.Must(googleUuid.NewV7())})

	for _, testCase := range []struct {
		ctx         context.Context
		serial      string
		issuerKeyID string
		issuerDN    string
	}{
		{ctxA, "issuer-scope-a", gormTestIssuerKeyA, gormTestSharedIssuerDN},
		{ctxB, "issuer-scope-b", gormTestIssuerKeyA, "CN=different-dn"},
		{ctxC, "issuer-scope-c", "issuer-key-b", gormTestSharedIssuerDN},
	} {
		certificate := newTestCert(testCase.serial, gormTestProfileID, gormTestRequesterA, testCase.issuerDN, newTestExpiration())
		certificate.AuthorityKeyID = testCase.issuerKeyID
		require.NoError(t, store.Store(testCase.ctx, certificate))
		require.NoError(t, store.Revoke(testCase.ctx, certificate.ID, ReasonKeyCompromise))
	}

	revoked, err := store.GetRevokedByIssuerKeyID(context.Background(), "ISSUER-KEY-A")
	require.NoError(t, err)
	require.Len(t, revoked, 2)
	require.Equal(t, gormTestIssuerKeyA, revoked[0].AuthorityKeyID)
	require.Equal(t, gormTestIssuerKeyA, revoked[1].AuthorityKeyID)

	issuerScopedCertificate, err := store.GetBySerialNumberForIssuer(
		context.Background(),
		"issuer-scope-b",
		gormTestIssuerKeyA,
	)
	require.NoError(t, err)
	require.Equal(t, "issuer-scope-b", issuerScopedCertificate.SerialNumber)
	require.Equal(t, gormTestIssuerKeyA, issuerScopedCertificate.AuthorityKeyID)

	_, err = store.GetBySerialNumberForIssuer(context.Background(), "issuer-scope-b", "unknown-issuer")
	require.ErrorIs(t, err, ErrCertificateNotFound)
}

func TestGORMStore_AllocateCRLNumbersConcurrentAndDurable(t *testing.T) {
	t.Parallel()

	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "pki-sequences.db")) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(24)
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&StoredCRLSequence{}))

	store, err := NewGORMStore(db)
	require.NoError(t, err)

	const allocationCount = 24

	numbers := make(chan *big.Int, allocationCount)
	errs := make(chan error, allocationCount)

	var wg sync.WaitGroup
	for range allocationCount {
		wg.Add(1)

		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			number, allocateErr := store.AllocateCRLNumber(ctx, gormTestIssuerKeyA)
			if allocateErr != nil {
				errs <- allocateErr

				return
			}

			numbers <- number
		}()
	}

	wg.Wait()
	close(numbers)
	close(errs)

	for allocateErr := range errs {
		require.NoError(t, allocateErr)
	}

	seen := make(map[string]bool, allocationCount)

	for number := range numbers {
		require.NotNil(t, number)
		require.True(t, number.IsInt64())
		seen[number.String()] = true
	}

	require.Len(t, seen, allocationCount)

	for i := int64(1); i <= allocationCount; i++ {
		require.True(t, seen[big.NewInt(i).String()], "missing CRL number %d", i)
	}

	require.NoError(t, store.Close())

	restartedSQLDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = restartedSQLDB.Close()
	})
	restartedSQLDB.SetMaxOpenConns(1)
	restartedDB, err := gorm.Open(sqlite.Dialector{Conn: restartedSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)
	restartedStore, err := NewGORMStore(restartedDB)
	require.NoError(t, err)
	restartedNumber, err := restartedStore.AllocateCRLNumber(context.Background(), gormTestIssuerKeyA)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(allocationCount+1), restartedNumber)

	otherIssuerNumber, err := restartedStore.AllocateCRLNumber(context.Background(), "issuer-key-b")
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1), otherIssuerNumber)

	require.NoError(t, restartedStore.Close())
}

func newTestExpiration() time.Time {
	return time.Now().UTC().Add(cryptoutilSharedMagic.Days7)
}
