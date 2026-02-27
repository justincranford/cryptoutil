// Copyright (c) 2025 Justin Cranford

package storage

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStore(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	require.NotNil(t, store)
	require.NotNil(t, store.certificates)
	require.NotNil(t, store.bySerial)
}

func TestMemoryStore_Store(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	tests := []struct {
		name    string
		cert    *StoredCertificate
		wantErr bool
		errType error
	}{
		{
			name:    "nil certificate",
			cert:    nil,
			wantErr: true,
			errType: ErrInvalidCertificate,
		},
		{
			name: "valid certificate",
			cert: &StoredCertificate{
				ID:           googleUuid.Must(googleUuid.NewV7()),
				SerialNumber: "123456",
				SubjectDN:    "CN=test",
				IssuerDN:     "CN=CA",
				NotBefore:    time.Now().UTC(),
				NotAfter:     time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				Status:       StatusActive,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Not parallel because we're testing store state.
			err := store.Store(ctx, tc.cert)
			if tc.wantErr {
				require.Error(t, err)

				if tc.errType != nil {
					require.ErrorIs(t, err, tc.errType)
				}

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestMemoryStore_DuplicateSerial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert1 := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "duplicate-serial",
		SubjectDN:    "CN=test1",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	cert2 := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "duplicate-serial",
		SubjectDN:    "CN=test2",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	err := store.Store(ctx, cert1)
	require.NoError(t, err)

	err = store.Store(ctx, cert2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "serial number already exists")
}

func TestMemoryStore_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "get-test-123",
		SubjectDN:    "CN=test",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	// Get existing.
	result, err := store.Get(ctx, cert.ID)
	require.NoError(t, err)
	require.Equal(t, cert.ID, result.ID)

	// Get non-existent.
	_, err = store.Get(ctx, googleUuid.Must(googleUuid.NewV7()))
	require.ErrorIs(t, err, ErrCertificateNotFound)
}

func TestMemoryStore_GetBySerialNumber(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "serial-test-456",
		SubjectDN:    "CN=test",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	// Get existing.
	result, err := store.GetBySerialNumber(ctx, cert.SerialNumber)
	require.NoError(t, err)
	require.Equal(t, cert.ID, result.ID)

	// Get non-existent.
	_, err = store.GetBySerialNumber(ctx, "non-existent")
	require.ErrorIs(t, err, ErrCertificateNotFound)
}

func TestMemoryStore_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	// Create test certificates.
	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		status := StatusActive
		if i%2 == 0 {
			status = StatusRevoked
		}

		cert := &StoredCertificate{
			ID:           googleUuid.Must(googleUuid.NewV7()),
			SerialNumber: googleUuid.Must(googleUuid.NewV7()).String(),
			SubjectDN:    "CN=test",
			IssuerDN:     "CN=CA",
			Status:       status,
			ProfileID:    "tls-server",
		}

		err := store.Store(ctx, cert)
		require.NoError(t, err)
	}

	// List all.
	results, total, err := store.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, total)
	require.Len(t, results, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

	// List with status filter.
	activeStatus := StatusActive
	_, total, err = store.List(ctx, &ListFilter{Status: &activeStatus})
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, total)

	// List with pagination.
	results, total, err = store.List(ctx, &ListFilter{Limit: 3, Offset: 0})
	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.JoseJADefaultMaxMaterials, total)
	require.Len(t, results, 3)
}

func TestMemoryStore_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "update-test-789",
		SubjectDN:    "CN=test",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	// Update existing.
	cert.Status = StatusSuspended
	err = store.Update(ctx, cert)
	require.NoError(t, err)

	result, err := store.Get(ctx, cert.ID)
	require.NoError(t, err)
	require.Equal(t, StatusSuspended, result.Status)

	// Update non-existent.
	nonExistent := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "non-existent",
	}
	err = store.Update(ctx, nonExistent)
	require.ErrorIs(t, err, ErrCertificateNotFound)

	// Update nil.
	err = store.Update(ctx, nil)
	require.ErrorIs(t, err, ErrInvalidCertificate)
}

func TestMemoryStore_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "delete-test-abc",
		SubjectDN:    "CN=test",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	// Delete existing.
	err = store.Delete(ctx, cert.ID)
	require.NoError(t, err)

	// Verify deleted.
	_, err = store.Get(ctx, cert.ID)
	require.ErrorIs(t, err, ErrCertificateNotFound)

	// Delete non-existent.
	err = store.Delete(ctx, googleUuid.Must(googleUuid.NewV7()))
	require.ErrorIs(t, err, ErrCertificateNotFound)
}

func TestMemoryStore_Revoke(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	cert := &StoredCertificate{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		SerialNumber: "revoke-test-def",
		SubjectDN:    "CN=test",
		IssuerDN:     "CN=CA",
		Status:       StatusActive,
	}

	err := store.Store(ctx, cert)
	require.NoError(t, err)

	// Revoke.
	err = store.Revoke(ctx, cert.ID, ReasonKeyCompromise)
	require.NoError(t, err)

	result, err := store.Get(ctx, cert.ID)
	require.NoError(t, err)
	require.Equal(t, StatusRevoked, result.Status)
	require.NotNil(t, result.RevocationTime)
	require.NotNil(t, result.RevocationReason)
	require.Equal(t, ReasonKeyCompromise, *result.RevocationReason)

	// Revoke again should fail.
	err = store.Revoke(ctx, cert.ID, ReasonSuperseded)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already revoked")

	// Revoke non-existent.
	err = store.Revoke(ctx, googleUuid.Must(googleUuid.NewV7()), ReasonUnspecified)
	require.ErrorIs(t, err, ErrCertificateNotFound)
}

func TestMemoryStore_GetRevoked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	// Create certificates with different issuers and statuses.
	for i := 0; i < cryptoutilSharedMagic.DefaultEmailOTPLength; i++ {
		issuer := "CN=CA1"
		if i >= 3 {
			issuer = "CN=CA2"
		}

		cert := &StoredCertificate{
			ID:           googleUuid.Must(googleUuid.NewV7()),
			SerialNumber: googleUuid.Must(googleUuid.NewV7()).String(),
			SubjectDN:    "CN=test",
			IssuerDN:     issuer,
			Status:       StatusActive,
		}

		err := store.Store(ctx, cert)
		require.NoError(t, err)

		// Revoke every other certificate.
		if i%2 == 0 {
			err = store.Revoke(ctx, cert.ID, ReasonUnspecified)
			require.NoError(t, err)
		}
	}

	// Get all revoked.
	revoked, err := store.GetRevoked(ctx, "")
	require.NoError(t, err)
	require.Len(t, revoked, 3)

	// Get revoked by issuer.
	revoked, err = store.GetRevoked(ctx, "CN=CA1")
	require.NoError(t, err)
	require.Len(t, revoked, 2)
}

func TestMemoryStore_CountByStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()

	// Create certificates with different statuses.
	statuses := []CertificateStatus{StatusActive, StatusActive, StatusActive, StatusRevoked, StatusExpired}

	for i, status := range statuses {
		cert := &StoredCertificate{
			ID:           googleUuid.Must(googleUuid.NewV7()),
			SerialNumber: googleUuid.Must(googleUuid.NewV7()).String() + string(rune(i)),
			SubjectDN:    "CN=test",
			IssuerDN:     "CN=CA",
			Status:       status,
		}

		err := store.Store(ctx, cert)
		require.NoError(t, err)
	}

	counts, err := store.CountByStatus(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(3), counts[StatusActive])
	require.Equal(t, int64(1), counts[StatusRevoked])
	require.Equal(t, int64(1), counts[StatusExpired])
}

func TestMemoryStore_Close(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	err := store.Close()
	require.NoError(t, err)
}

func TestNewStoredCertificateFromX509(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cert        *x509.Certificate
		profileID   string
		requesterID string
		wantErr     bool
	}{
		{
			name:        "nil certificate",
			cert:        nil,
			profileID:   "tls-server",
			requesterID: "user-123",
			wantErr:     true,
		},
		{
			name: "valid certificate",
			cert: &x509.Certificate{
				Subject: pkix.Name{
					CommonName: "test.example.com",
				},
				Issuer: pkix.Name{
					CommonName: "Test CA",
				},
				NotBefore:      time.Now().UTC(),
				NotAfter:       time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
				SubjectKeyId:   []byte{1, 2, 3, 4},
				AuthorityKeyId: []byte{cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries, cryptoutilSharedMagic.DefaultEmailOTPLength, cryptoutilSharedMagic.GitRecentActivityDays, cryptoutilSharedMagic.IMMinPasswordLength},
				Raw:            []byte{0x30, 0x00}, // Minimal DER.
			},
			profileID:   "tls-server",
			requesterID: "user-123",
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := NewStoredCertificateFromX509(tc.cert, tc.profileID, tc.requesterID)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.profileID, result.ProfileID)
			require.Equal(t, tc.requesterID, result.RequesterID)
			require.Equal(t, StatusActive, result.Status)
		})
	}
}

func TestCertificateStatus_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   CertificateStatus
		expected string
	}{
		{StatusActive, "active"},
		{StatusRevoked, "revoked"},
		{StatusExpired, "expired"},
		{StatusSuspended, "suspended"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestRevocationReason_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		reason   RevocationReason
		expected int
	}{
		{ReasonUnspecified, 0},
		{ReasonKeyCompromise, 1},
		{ReasonCACompromise, 2},
		{ReasonAffiliationChanged, 3},
		{ReasonSuperseded, 4},
		{ReasonCessationOfOperation, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		{ReasonCertificateHold, cryptoutilSharedMagic.DefaultEmailOTPLength},
		{ReasonRemoveFromCRL, cryptoutilSharedMagic.IMMinPasswordLength},
		{ReasonPrivilegeWithdrawn, 9},
		{ReasonAACompromise, cryptoutilSharedMagic.JoseJADefaultMaxMaterials},
	}

	for _, tc := range tests {
		t.Run(string(rune(tc.expected)), func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, int(tc.reason))
		})
	}
}
