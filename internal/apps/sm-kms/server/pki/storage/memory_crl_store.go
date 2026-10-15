// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"math/big"
	"strings"
)

// GetRevoked returns all revoked certificates for CRL generation.
func (s *MemoryStore) GetRevoked(_ context.Context, issuerDN string) ([]*StoredCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*StoredCertificate

	for _, cert := range s.certificates {
		if cert.Status == StatusRevoked && (issuerDN == "" || cert.IssuerDN == issuerDN) {
			results = append(results, cert)
		}
	}

	return results, nil
}

// GetRevokedByIssuerKeyID returns revoked certificates across tenants for one issuer.
func (s *MemoryStore) GetRevokedByIssuerKeyID(_ context.Context, issuerKeyID string) ([]*StoredCertificate, error) {
	issuerKeyID = strings.ToLower(strings.TrimSpace(issuerKeyID))
	if issuerKeyID == "" {
		return nil, ErrInvalidIssuerKeyID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*StoredCertificate

	for _, cert := range s.certificates {
		if cert.Status == StatusRevoked && strings.EqualFold(cert.AuthorityKeyID, issuerKeyID) {
			results = append(results, cert)
		}
	}

	return results, nil
}

// AllocateCRLNumber allocates a monotonic CRL number for a single issuer in memory.
func (s *MemoryStore) AllocateCRLNumber(_ context.Context, issuerKeyID string) (*big.Int, error) {
	issuerKeyID = strings.ToLower(strings.TrimSpace(issuerKeyID))
	if issuerKeyID == "" {
		return nil, ErrInvalidIssuerKeyID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.crlSequences[issuerKeyID] == nil {
		s.crlSequences[issuerKeyID] = new(big.Int)
	}

	s.crlSequences[issuerKeyID].Add(s.crlSequences[issuerKeyID], big.NewInt(1))

	return new(big.Int).Set(s.crlSequences[issuerKeyID]), nil
}
