// Copyright (c) 2025 Justin Cranford
//
//

package storage

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// newTestCert creates a StoredCertificate for testing.
func newTestCert(serialNumber, profileID, requesterID, issuerDN string, notAfter time.Time) *StoredCertificate {
	return &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: serialNumber,
		SubjectDN:    "CN=test",
		IssuerDN:     issuerDN,
		NotBefore:    time.Now().UTC(),
		NotAfter:     notAfter,
		Status:       StatusActive,
		ProfileID:    profileID,
		RequesterID:  requesterID,
	}
}

// TestMemoryStore_Store_DuplicateSerialNumber tests that storing a certificate with
// a duplicate serial number returns an error.
func TestMemoryStore_Store_DuplicateSerialNumber(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert1 := newTestCert("serial-dup-001", "profile-1", "req-1", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

	err := store.Store(ctx, cert1)
	require.NoError(t, err)

	// Store another cert with the same serial number.
	cert2 := newTestCert("serial-dup-001", "profile-2", "req-2", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

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

	cert := newTestCert("serial-offset-001", "profile-1", "req-1", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	filter := &ListFilter{Offset: 100}

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

	for i := range 3 {
		cert := newTestCert(
			"serial-within-"+string(rune('A'+i)),
			"profile-1", "req-1", "CN=CA",
			time.Now().UTC().Add(365*24*time.Hour),
		)

		err := store.Store(ctx, cert)
		require.NoError(t, err)
	}

	filter := &ListFilter{Offset: 1}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, results, 2)
}

// TestMemoryStore_List_LimitResults tests that listing with Limit truncates the results.
func TestMemoryStore_List_LimitResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	for i := range 3 {
		cert := newTestCert(
			"serial-limit-"+string(rune('A'+i)),
			"profile-1", "req-1", "CN=CA",
			time.Now().UTC().Add(365*24*time.Hour),
		)

		err := store.Store(ctx, cert)
		require.NoError(t, err)
	}

	filter := &ListFilter{Limit: 1}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, results, 1)
}

// TestMemoryStore_matchesFilter_StatusMismatch tests that a certificate is excluded
// when its status does not match the filter status.
func TestMemoryStore_matchesFilter_StatusMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := newTestCert("serial-status-001", "profile-1", "req-1", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

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

	cert := newTestCert("serial-profile-001", "profile-A", "req-1", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

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

	cert1 := newTestCert("serial-dup-id-001", "profile-1", "req-1", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

	err := store.Store(ctx, cert1)
	require.NoError(t, err)

	// Create cert2 with the SAME ID but different serial.
	cert2 := &StoredCertificate{
		ID:           cert1.ID,
		SerialNumber: "serial-dup-id-002",
		SubjectDN:    "CN=test2",
		IssuerDN:     "CN=CA",
		NotBefore:    time.Now().UTC(),
		NotAfter:     time.Now().UTC().Add(365 * 24 * time.Hour),
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

	cert := newTestCert("serial-req-001", "profile-1", "requester-A", "CN=CA", time.Now().UTC().Add(365*24*time.Hour))

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	differentRequester := "requester-B"
	filter := &ListFilter{RequesterID: &differentRequester}

	results, total, err := store.List(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, results)
}
