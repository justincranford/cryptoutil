// Copyright (c) 2025-2026 Justin Cranford.
// Package handler provides EST profile selection and re-enrollment authorization.
package handler

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilCAProfileSubject "cryptoutil/internal/apps/sm-kms/server/pki/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	errESTClientCertificateRequired = errors.New("EST client certificate is required")
	errESTReenrollmentDenied        = errors.New("EST client certificate is not authorized for re-enrollment")
	errESTNoProfile                 = errors.New("no unambiguous EST profile is configured")
)

type estClientCertificateContextKey struct{}

func withESTClientCertificate(ctx context.Context, cert *x509.Certificate) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if cert == nil {
		return ctx
	}

	return context.WithValue(ctx, estClientCertificateContextKey{}, cert)
}

func estClientCertificateFromContext(ctx context.Context) *x509.Certificate {
	if ctx == nil {
		return nil
	}

	cert, _ := ctx.Value(estClientCertificateContextKey{}).(*x509.Certificate)

	return cert
}

func estClientCertificateFromFiberContext(c *fiber.Ctx) *x509.Certificate {
	tlsState := c.Context().TLSConnectionState()
	if tlsState == nil || len(tlsState.PeerCertificates) == 0 {
		return nil
	}

	return tlsState.PeerCertificates[0]
}

func (h *Handler) selectESTProfile() (*ProfileConfig, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clientProfile, exists := h.profiles[cryptoutilSharedMagic.PKIProfileIDTLSClient]
	if exists {
		if clientProfile == nil || clientProfile.ID != cryptoutilSharedMagic.PKIProfileIDTLSClient {
			return nil, errESTNoProfile
		}

		return clientProfile, nil
	}

	if len(h.profiles) == 1 {
		for id, profile := range h.profiles {
			if profile != nil && profile.ID == id {
				return profile, nil
			}
		}
	}

	return nil, errESTNoProfile
}

func (h *Handler) buildESTIssueRequest(
	csr *x509.CertificateRequest,
	profile *ProfileConfig,
) (*cryptoutilCAServiceIssuer.CertificateRequest, error) {
	if csr == nil {
		return nil, fmt.Errorf("CSR is required")
	}

	if profile == nil {
		return nil, fmt.Errorf("EST certificate profile is required")
	}

	// Build subject request from CSR.
	subjectReq := &cryptoutilCAProfileSubject.Request{
		CommonName:         csr.Subject.CommonName,
		Organization:       csr.Subject.Organization,
		OrganizationalUnit: csr.Subject.OrganizationalUnit,
		Country:            csr.Subject.Country,
		State:              csr.Subject.Province,
		Locality:           csr.Subject.Locality,
		DNSNames:           csr.DNSNames,
		IPAddresses:        h.ipsToStrings(csr.IPAddresses),
		EmailAddresses:     csr.EmailAddresses,
		URIs:               h.urisToStrings(csr.URIs),
	}

	if profile.SubjectProfile != nil {
		resolved, err := profile.SubjectProfile.Resolve(subjectReq)
		if err != nil {
			return nil, fmt.Errorf("subject does not satisfy EST profile: %w", err)
		}

		subjectReq = &cryptoutilCAProfileSubject.Request{
			CommonName:         resolved.DN.CommonName,
			Organization:       resolved.DN.Organization,
			OrganizationalUnit: resolved.DN.OrganizationalUnit,
			Country:            resolved.DN.Country,
			State:              resolved.DN.Province,
			Locality:           resolved.DN.Locality,
			DNSNames:           resolved.DNSNames,
			IPAddresses:        h.ipsToStrings(resolved.IPAddresses),
			EmailAddresses:     resolved.EmailAddresses,
			URIs:               h.urisToStrings(resolved.URIs),
		}
	}

	validityDuration := time.Duration(cryptoutilSharedMagic.DaysPerYear) * cryptoutilSharedMagic.HoursPerDay * time.Hour

	if profile.CertificateProfile != nil {
		configuredValidity, err := profile.CertificateProfile.Validity.GetDuration()
		if err != nil {
			return nil, fmt.Errorf("invalid EST certificate profile validity: %w", err)
		}

		if err := profile.CertificateProfile.Validity.ValidateDuration(configuredValidity); err != nil {
			return nil, fmt.Errorf("invalid EST certificate profile validity: %w", err)
		}

		validityDuration = configuredValidity
	}

	return &cryptoutilCAServiceIssuer.CertificateRequest{
		PublicKey:        csr.PublicKey,
		SubjectRequest:   subjectReq,
		ValidityDuration: validityDuration,
	}, nil
}

func (h *Handler) authorizeESTReenrollment(ctx context.Context, clientCert *x509.Certificate) error {
	if clientCert == nil {
		return errESTClientCertificateRequired
	}

	if h.issuer == nil || h.storage == nil {
		return fmt.Errorf("EST re-enrollment dependencies are not configured")
	}

	caConfig := h.issuer.GetCAConfig()
	if caConfig == nil || caConfig.Certificate == nil || len(caConfig.Certificate.SubjectKeyId) == 0 {
		return fmt.Errorf("EST issuing authority is not configured")
	}

	if !bytes.Equal(clientCert.RawIssuer, caConfig.Certificate.RawSubject) ||
		!bytes.Equal(clientCert.AuthorityKeyId, caConfig.Certificate.SubjectKeyId) ||
		clientCert.CheckSignatureFrom(caConfig.Certificate) != nil {
		return errESTReenrollmentDenied
	}

	if ctx == nil {
		ctx = context.Background()
	}

	stored, err := h.storage.GetBySerialNumberForIssuer(
		ctx,
		clientCert.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId),
	)
	if errors.Is(err, cryptoutilCAStorage.ErrCertificateNotFound) {
		return errESTReenrollmentDenied
	}

	if err != nil {
		return fmt.Errorf("failed to look up EST re-enrollment certificate: %w", err)
	}

	if stored == nil || stored.Status != cryptoutilCAStorage.StatusActive || !storedESTCertificateMatches(stored, clientCert) {
		return errESTReenrollmentDenied
	}

	return nil
}

func storedESTCertificateMatches(stored *cryptoutilCAStorage.StoredCertificate, cert *x509.Certificate) bool {
	if stored == nil || cert == nil {
		return false
	}

	if len(stored.CertificateDER) > 0 {
		return bytes.Equal(stored.CertificateDER, cert.Raw)
	}

	block, _ := pem.Decode([]byte(stored.CertificatePEM))

	return block != nil &&
		block.Type == cryptoutilSharedMagic.StringPEMTypeCertificate &&
		bytes.Equal(block.Bytes, cert.Raw)
}

func sameESTCertificateIdentity(first, second *x509.Certificate) bool {
	if first == nil || second == nil ||
		!reflect.DeepEqual(first.Subject.ToRDNSequence(), second.Subject.ToRDNSequence()) {
		return false
	}

	return sameESTStringValues(first.DNSNames, second.DNSNames) &&
		sameESTStringValues(first.EmailAddresses, second.EmailAddresses) &&
		sameESTStringValues(estIPAddresses(first), estIPAddresses(second)) &&
		sameESTStringValues(estURIs(first), estURIs(second))
}

func sameESTCertificateIdentityRequest(
	cert *x509.Certificate,
	request *cryptoutilCAProfileSubject.Request,
) bool {
	if cert == nil || request == nil {
		return false
	}

	subject := pkix.Name{
		CommonName:         request.CommonName,
		Organization:       request.Organization,
		OrganizationalUnit: request.OrganizationalUnit,
		Country:            request.Country,
		Province:           request.State,
		Locality:           request.Locality,
	}
	if !reflect.DeepEqual(cert.Subject.ToRDNSequence(), subject.ToRDNSequence()) {
		return false
	}

	return sameESTStringValues(cert.DNSNames, request.DNSNames) &&
		sameESTStringValues(cert.EmailAddresses, request.EmailAddresses) &&
		sameESTStringValues(estIPAddresses(cert), request.IPAddresses) &&
		sameESTStringValues(estURIs(cert), request.URIs)
}

func sameESTStringValues(first, second []string) bool {
	firstCopy := slices.Clone(first)
	secondCopy := slices.Clone(second)

	slices.Sort(firstCopy)
	slices.Sort(secondCopy)

	return slices.Equal(firstCopy, secondCopy)
}

func estIPAddresses(cert *x509.Certificate) []string {
	addresses := make([]string, 0, len(cert.IPAddresses))
	for _, address := range cert.IPAddresses {
		addresses = append(addresses, address.String())
	}

	return addresses
}

func estURIs(cert *x509.Certificate) []string {
	uris := make([]string, 0, len(cert.URIs))
	for _, uri := range cert.URIs {
		uris = append(uris, uri.String())
	}

	return uris
}
