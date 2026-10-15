// Copyright (c) 2025-2026 Justin Cranford.
//
//

package storage

import (
	"context"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// newTestCert creates a StoredCertificate for testing.
func newTestCert(serialNumber, profileID, requesterID, issuerDN string, notAfter time.Time) *StoredCertificate {
	return &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		RequestID:    googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: serialNumber,
		SubjectDN:    testLiteralCNTest,
		IssuerDN:     issuerDN,
		NotBefore:    time.Now().UTC(),
		NotAfter:     notAfter,
		Status:       StatusActive,
		ProfileID:    profileID,
		RequesterID:  requesterID,
	}
}

func TestMemoryStore_StoreRequiresRequestID(t *testing.T) {
	t.Parallel()

	certificate := newTestCert("missing-request-id", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	certificate.RequestID = googleUuid.UUID{}

	err := NewMemoryStore().Store(context.Background(), certificate)
	require.ErrorIs(t, err, ErrInvalidRequestID)
}

func TestMemoryStore_UpdateRequiresRequestID(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	certificate := newTestCert("update-missing-request-id", gormTestProfileID, gormTestRequesterA, testLiteralCNCA, newTestExpiration())
	require.NoError(t, store.Store(context.Background(), certificate))

	certificate.RequestID = googleUuid.UUID{}
	require.ErrorIs(t, store.Update(context.Background(), certificate), ErrInvalidRequestID)
}

// TestMemoryStore_Store_DuplicateSerialNumber tests that storing a certificate with
// a duplicate serial number returns an error.
func TestMemoryStore_Store_DuplicateSerialNumber(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert1 := newTestCert("serial-dup-001", "profile-1", "req-1", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err := store.Store(ctx, cert1)
	require.NoError(t, err)

	// Store another cert with the same serial number.
	cert2 := newTestCert("serial-dup-001", "profile-2", "req-2", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err = store.Store(ctx, cert2)

	require.Error(t, err)
	require.Contains(t, err.Error(), "serial number already exists")
}

// TestMemoryStore_List_OffsetBeyondResults tests that listing with Offset >= total results
// returns an empty slice (not an error).
func TestMemoryStore_List_OffsetBeyondResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := newTestCert("serial-offset-001", "profile-1", "req-1", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	filter := &ListFilter{Offset: cryptoutilSharedMagic.JoseJAMaxMaterials}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Empty(t, results)
}

// TestMemoryStore_List_OffsetWithinResults tests that listing with Offset within results
// returns the remaining slice starting at offset.
func TestMemoryStore_List_OffsetWithinResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	for i := range testNumericThree {
		cert := newTestCert(
			"serial-within-"+string(rune('A'+i)),
			"profile-1", "req-1", testLiteralCNCA,
			time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour),
		)

		err := store.Store(ctx, cert)
		require.NoError(t, err)
	}

	filter := &ListFilter{Offset: 1}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, testNumericThree, total)
	require.Len(t, results, 2)
}

// TestMemoryStore_List_LimitResults tests that listing with Limit truncates the results.
func TestMemoryStore_List_LimitResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	for i := range testNumericThree {
		cert := newTestCert(
			"serial-limit-"+string(rune('A'+i)),
			"profile-1", "req-1", testLiteralCNCA,
			time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour),
		)

		err := store.Store(ctx, cert)
		require.NoError(t, err)
	}

	filter := &ListFilter{Limit: 1}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, testNumericThree, total)
	require.Len(t, results, 1)
}

// TestMemoryStore_matchesFilter_StatusMismatch tests that a certificate is excluded
// when its status does not match the filter status.
func TestMemoryStore_matchesFilter_StatusMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := newTestCert("serial-status-001", "profile-1", "req-1", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	revokedStatus := StatusRevoked
	filter := &ListFilter{Status: &revokedStatus}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, results)
}

// TestMemoryStore_matchesFilter_ProfileIDMismatch tests that a certificate is excluded
// when its profile ID does not match the filter profile ID.
func TestMemoryStore_matchesFilter_ProfileIDMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := newTestCert("serial-profile-001", "profile-A", "req-1", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	differentProfile := "profile-B"
	filter := &ListFilter{ProfileID: &differentProfile}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, results)
}

// TestMemoryStore_Store_DuplicateID tests that storing a certificate with
// a duplicate ID returns ErrCertificateExists.
func TestMemoryStore_Store_DuplicateID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert1 := newTestCert("serial-dup-id-001", "profile-1", "req-1", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err := store.Store(ctx, cert1)
	require.NoError(t, err)

	// Create cert2 with the SAME ID but different serial.
	cert2 := &StoredCertificate{
		ID:           cert1.ID,
		RequestID:    googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "serial-dup-id-002",
		SubjectDN:    testLiteralCNTest2,
		IssuerDN:     testLiteralCNCA,
		NotBefore:    time.Now().UTC(),
		NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		Status:       StatusActive,
	}

	err = store.Store(ctx, cert2)

	require.ErrorIs(t, err, ErrCertificateExists)
}

// TestMemoryStore_matchesFilter_RequesterIDMismatch tests that a certificate is excluded
// when its requester ID does not match the filter requester ID.
func TestMemoryStore_matchesFilter_RequesterIDMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := newTestCert("serial-req-001", "profile-1", "requester-A", testLiteralCNCA, time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year*cryptoutilSharedMagic.HoursPerDay*time.Hour))

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	differentRequester := "requester-B"
	filter := &ListFilter{RequesterID: &differentRequester}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, results)
}

func TestMemoryStore_matchesFilter_IssuerAndNotAfter(t *testing.T) {
	t.Parallel()

	profileID := "filter-profile"
	requesterID := "filter-requester"
	notAfter := time.Now().UTC().Add(time.Hour)
	activeStatus := StatusActive
	matchingIssuer := testLiteralCNCA
	differentIssuer := testLiteralCNTest2
	notAfterFromEqual := notAfter
	notAfterFromLater := notAfter.Add(time.Second)
	notAfterToEqual := notAfter
	notAfterToEarlier := notAfter.Add(-time.Second)

	tests := []struct {
		name      string
		filter    *ListFilter
		wantTotal int
	}{
		{
			name: "all filter fields match including date boundaries",
			filter: &ListFilter{
				Status:       &activeStatus,
				ProfileID:    &profileID,
				RequesterID:  &requesterID,
				IssuerDN:     &matchingIssuer,
				NotAfterFrom: &notAfterFromEqual,
				NotAfterTo:   &notAfterToEqual,
			},
			wantTotal: 1,
		},
		{
			name:      "issuer does not match",
			filter:    &ListFilter{IssuerDN: &differentIssuer},
			wantTotal: 0,
		},
		{
			name:      "not after from includes equal boundary",
			filter:    &ListFilter{NotAfterFrom: &notAfterFromEqual},
			wantTotal: 1,
		},
		{
			name:      "not after from excludes later boundary",
			filter:    &ListFilter{NotAfterFrom: &notAfterFromLater},
			wantTotal: 0,
		},
		{
			name:      "not after to includes equal boundary",
			filter:    &ListFilter{NotAfterTo: &notAfterToEqual},
			wantTotal: 1,
		},
		{
			name:      "not after to excludes earlier boundary",
			filter:    &ListFilter{NotAfterTo: &notAfterToEarlier},
			wantTotal: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := NewMemoryStore()
			cert := newTestCert("serial-filter", profileID, requesterID, matchingIssuer, notAfter)
			err := store.Store(context.Background(), cert)
			require.NoError(t, err)

			results, total, err := store.List(context.Background(), tc.filter)
			require.NoError(t, err)
			require.Equal(t, tc.wantTotal, total)
			require.Len(t, results, tc.wantTotal)
		})
	}
}
