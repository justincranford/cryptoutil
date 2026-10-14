// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	gormTestProfileID        = "server-profile"
	gormTestRequesterA       = "requester-a"
	gormTestRequesterB       = "requester-b"
	gormTestUpdatedProfileID = "updated-profile"
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
		ID:                  googleUuid.Must(googleUuid.NewV7()).String(),
		CertificateDER:      []byte("certificate"),
		PrivateKeyEncrypted: []byte("encrypted-key"),
	}
	require.NoError(t, store.Create(context.Background(), authority))

	got, err := store.Get(context.Background(), authority.ID)
	require.NoError(t, err)
	require.Equal(t, authority.CertificateDER, got.CertificateDER)
	require.Equal(t, authority.PrivateKeyEncrypted, got.PrivateKeyEncrypted)
}

func newTestExpiration() time.Time {
	return time.Now().UTC().Add(cryptoutilSharedMagic.Days7)
}
