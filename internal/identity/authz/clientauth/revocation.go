// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	http "net/http"
	"sync"
	"time"

	"golang.org/x/crypto/ocsp"
)

// RevocationChecker defines the interface for certificate revocation checking.
type RevocationChecker interface {
	CheckRevocation(ctx context.Context, cert *x509.Certificate, issuer *x509.Certificate) error
}

// CRLCache caches downloaded Certificate Revocation Lists.
type CRLCache struct {
	mu     sync.RWMutex
	crls   map[string]*cachedCRL
	maxAge time.Duration
}

type cachedCRL struct {
	crl       *pkix.CertificateList
	fetchedAt time.Time
}

// NewCRLCache creates a new CRL cache with the specified max age.
func NewCRLCache(maxAge time.Duration) *CRLCache {
	return &CRLCache{
		crls:   make(map[string]*cachedCRL),
		maxAge: maxAge,
	}
}

// GetCRL retrieves a CRL from cache or fetches it if not cached/expired.
func (c *CRLCache) GetCRL(ctx context.Context, url string) (*pkix.CertificateList, error) {
	// Check cache first.
	c.mu.RLock()
	cached, exists := c.crls[url]
	c.mu.RUnlock()

	if exists && time.Since(cached.fetchedAt) < c.maxAge {
		return cached.crl, nil
	}

	// Fetch CRL from URL.
	crl, err := c.fetchCRL(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CRL from %s: %w", url, err)
	}

	// Cache the CRL.
	c.mu.Lock()
	c.crls[url] = &cachedCRL{
		crl:       crl,
		fetchedAt: time.Now(),
	}
	c.mu.Unlock()

	return crl, nil
}

// fetchCRL downloads a CRL from the given URL.
func (c *CRLCache) fetchCRL(ctx context.Context, url string) (*pkix.CertificateList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRL request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download CRL: %w", err)
	}

	defer func() {
		//nolint:errcheck // Best effort close; error logged elsewhere if critical.
		_ = resp.Body.Close()
	}() // Best effort close; ignore error.

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CRL download returned status %d", resp.StatusCode)
	}

	crlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CRL response: %w", err)
	}

	revocationList, err := x509.ParseRevocationList(crlBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CRL: %w", err)
	}

	// Convert x509.RevocationList to deprecated pkix.CertificateList for backward compatibility.
	//nolint:staticcheck // pkix.CertificateList deprecated but still used in existing code.
	crl := &pkix.CertificateList{
		TBSCertList: pkix.TBSCertificateList{
			RevokedCertificates: make([]pkix.RevokedCertificate, len(revocationList.RevokedCertificateEntries)),
		},
	}

	for i, entry := range revocationList.RevokedCertificateEntries {
		crl.TBSCertList.RevokedCertificates[i] = pkix.RevokedCertificate{
			SerialNumber:   entry.SerialNumber,
			RevocationTime: entry.RevocationTime,
		}
	}

	return crl, nil
}

// IsRevoked checks if a certificate is revoked according to the CRL.
func (c *CRLCache) IsRevoked(crl *pkix.CertificateList, serialNumber *big.Int) bool {
	for _, revokedCert := range crl.TBSCertList.RevokedCertificates {
		if revokedCert.SerialNumber.Cmp(serialNumber) == 0 {
			return true
		}
	}

	return false
}

// CRLRevocationChecker checks certificate revocation using CRLs.
type CRLRevocationChecker struct {
	cache   *CRLCache
	timeout time.Duration
}

// NewCRLRevocationChecker creates a new CRL-based revocation checker.
func NewCRLRevocationChecker(cacheMaxAge, timeout time.Duration) *CRLRevocationChecker {
	return &CRLRevocationChecker{
		cache:   NewCRLCache(cacheMaxAge),
		timeout: timeout,
	}
}

// CheckRevocation checks if a certificate has been revoked using CRLs.
func (r *CRLRevocationChecker) CheckRevocation(ctx context.Context, cert *x509.Certificate, issuer *x509.Certificate) error {
	if len(cert.CRLDistributionPoints) == 0 {
		// No CRL distribution points - cannot verify revocation.
		// This may be acceptable depending on security policy.
		return nil
	}

	// Create context with timeout.
	checkCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Check each CRL distribution point.
	for _, crlURL := range cert.CRLDistributionPoints {
		crl, err := r.cache.GetCRL(checkCtx, crlURL)
		if err != nil {
			// Log error but continue to next CRL URL.
			continue
		}

		// Verify CRL signature using deprecated issuer.CheckCRLSignature.
		//nolint:staticcheck // CheckCRLSignature deprecated but pkix.CertificateList also deprecated.
		if err := issuer.CheckCRLSignature(crl); err != nil {
			// CRL signature invalid - skip this CRL.
			continue
		}

		// Check if certificate is in the revocation list.
		if r.cache.IsRevoked(crl, cert.SerialNumber) {
			return fmt.Errorf("certificate has been revoked (CRL)")
		}
	}

	return nil
}

// OCSPRevocationChecker checks certificate revocation using OCSP.
type OCSPRevocationChecker struct {
	timeout time.Duration
}

// NewOCSPRevocationChecker creates a new OCSP-based revocation checker.
func NewOCSPRevocationChecker(timeout time.Duration) *OCSPRevocationChecker {
	return &OCSPRevocationChecker{
		timeout: timeout,
	}
}

// CheckRevocation checks if a certificate has been revoked using OCSP.
func (r *OCSPRevocationChecker) CheckRevocation(ctx context.Context, cert *x509.Certificate, issuer *x509.Certificate) error {
	if len(cert.OCSPServer) == 0 {
		return fmt.Errorf("no OCSP server URLs in certificate")
	}

	// Create OCSP request.
	ocspRequest, err := ocsp.CreateRequest(cert, issuer, nil)
	if err != nil {
		return fmt.Errorf("failed to create OCSP request: %w", err)
	}

	// Create context with timeout.
	checkCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Try each OCSP server.
	for _, ocspURL := range cert.OCSPServer {
		req, err := http.NewRequestWithContext(checkCtx, http.MethodPost, ocspURL, bytes.NewReader(ocspRequest))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/ocsp-request")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		ocspResponseBytes, err := io.ReadAll(resp.Body)
		//nolint:errcheck // Best effort close; error logged elsewhere if critical.
		_ = resp.Body.Close()

		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			continue
		}

		ocspResponse, err := ocsp.ParseResponse(ocspResponseBytes, issuer)
		if err != nil {
			continue
		}

		// Check revocation status.
		switch ocspResponse.Status {
		case ocsp.Good:
			return nil // Certificate is valid.
		case ocsp.Revoked:
			return fmt.Errorf("certificate has been revoked (OCSP)")
		case ocsp.Unknown:
			// OCSP server doesn't know about this certificate.
			// Continue to next server or fail.
			continue
		}
	}

	return fmt.Errorf("OCSP check failed for all servers")
}

// CombinedRevocationChecker tries OCSP first, falls back to CRL.
type CombinedRevocationChecker struct {
	ocspChecker *OCSPRevocationChecker
	crlChecker  *CRLRevocationChecker
}

// NewCombinedRevocationChecker creates a checker that tries OCSP then CRL.
func NewCombinedRevocationChecker(ocspTimeout, crlTimeout, crlCacheMaxAge time.Duration) *CombinedRevocationChecker {
	return &CombinedRevocationChecker{
		ocspChecker: NewOCSPRevocationChecker(ocspTimeout),
		crlChecker:  NewCRLRevocationChecker(crlCacheMaxAge, crlTimeout),
	}
}

// CheckRevocation checks revocation using OCSP, falls back to CRL.
func (r *CombinedRevocationChecker) CheckRevocation(ctx context.Context, cert *x509.Certificate, issuer *x509.Certificate) error {
	// Try OCSP first (faster, real-time).
	if err := r.ocspChecker.CheckRevocation(ctx, cert, issuer); err == nil {
		return nil // OCSP check passed.
	}

	// Fallback to CRL checking.
	return r.crlChecker.CheckRevocation(ctx, cert, issuer)
}
