// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net"
	"net/url"
	"slices"
	"testing"

	cryptoutilCAProfileSubject "cryptoutil/internal/apps/sm-kms/server/pki/profile/subject"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testESTIdentityPrimaryDNS     = "est-primary.example.test"
	testESTIdentitySecondaryDNS   = "est-secondary.example.test"
	testESTIdentityPrimaryEmail   = "primary@example.test"
	testESTIdentitySecondaryEmail = "secondary@example.test"
	testESTIdentityPrimaryIP      = "192.0.2.10"
	testESTIdentitySecondaryIP    = "2001:db8::10"
	testESTIdentityPrimaryURI     = "spiffe://example.test/workload/primary"
	testESTIdentitySecondaryURI   = "spiffe://example.test/workload/secondary"
	testESTIdentityDifferentName  = "different.example.test"
	testESTIdentityDifferentEmail = "different@example.test"
	testESTIdentityDifferentURI   = "spiffe://example.test/workload/different"
)

func TestESTClientCertificateContext(t *testing.T) {
	t.Parallel()

	certificate := &x509.Certificate{}
	tests := []struct {
		name string
		ctx  context.Context
		cert *x509.Certificate
		want *x509.Certificate
	}{
		{name: "nil context and certificate"},
		{name: "nil context with certificate", cert: certificate, want: certificate},
		{
			name: "nil certificate preserves existing context value",
			ctx:  withESTClientCertificate(context.Background(), certificate),
			want: certificate,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := withESTClientCertificate(tc.ctx, tc.cert)
			require.NotNil(t, ctx)
			require.Same(t, tc.want, estClientCertificateFromContext(ctx))
		})
	}
}

func TestESTClientCertificateFromNilContext(t *testing.T) {
	t.Parallel()

	var nilContext context.Context
	require.Nil(t, estClientCertificateFromContext(nilContext))
}

func TestStoredESTCertificateMatches(t *testing.T) {
	t.Parallel()

	cert := &x509.Certificate{Raw: []byte("certificate DER")}
	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: cert.Raw,
	})
	wrongPEMType := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
		Bytes: cert.Raw,
	})

	tests := []struct {
		name    string
		stored  *cryptoutilCAStorage.StoredCertificate
		cert    *x509.Certificate
		matches bool
	}{
		{name: testHandlerNilStoredCertificate, cert: cert},
		{name: testHandlerNilClientCertificate, stored: &cryptoutilCAStorage.StoredCertificate{}},
		{
			name:    "DER matches",
			stored:  &cryptoutilCAStorage.StoredCertificate{CertificateDER: cert.Raw},
			cert:    cert,
			matches: true,
		},
		{
			name:   "DER does not match",
			stored: &cryptoutilCAStorage.StoredCertificate{CertificateDER: []byte("different DER")},
			cert:   cert,
		},
		{
			name:    "certificate PEM matches",
			stored:  &cryptoutilCAStorage.StoredCertificate{CertificatePEM: string(certificatePEM)},
			cert:    cert,
			matches: true,
		},
		{
			name:   testHandlerWrongPEMBlockType,
			stored: &cryptoutilCAStorage.StoredCertificate{CertificatePEM: string(wrongPEMType)},
			cert:   cert,
		},
		{
			name:   testHandlerInvalidPEM,
			stored: &cryptoutilCAStorage.StoredCertificate{CertificatePEM: testHandlerNotPEM},
			cert:   cert,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.matches, storedESTCertificateMatches(tc.stored, tc.cert))
		})
	}
}

func TestSameESTCertificateIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		nilFirst     bool
		nilSecond    bool
		mutateSecond func(*testing.T, *x509.Certificate)
		wantMatch    bool
	}{
		{
			name:      "same identity with reordered values",
			wantMatch: true,
			mutateSecond: func(_ *testing.T, cert *x509.Certificate) {
				slices.Reverse(cert.DNSNames)
				slices.Reverse(cert.EmailAddresses)
				slices.Reverse(cert.IPAddresses)
				slices.Reverse(cert.URIs)
			},
		},
		{name: "nil first certificate", nilFirst: true},
		{name: "nil second certificate", nilSecond: true},
		{
			name: testESTIdentitySubjectMismatch,
			mutateSecond: func(_ *testing.T, cert *x509.Certificate) {
				cert.Subject.CommonName = testESTIdentityDifferentName
			},
		},
		{
			name: testESTIdentityDNSMismatch,
			mutateSecond: func(_ *testing.T, cert *x509.Certificate) {
				cert.DNSNames[0] = testESTIdentityDifferentName
			},
		},
		{
			name: testESTIdentityEmailMismatch,
			mutateSecond: func(_ *testing.T, cert *x509.Certificate) {
				cert.EmailAddresses[0] = testESTIdentityDifferentEmail
			},
		},
		{
			name: testESTIdentityIPMismatch,
			mutateSecond: func(_ *testing.T, cert *x509.Certificate) {
				cert.IPAddresses[0] = net.ParseIP("192.0.2.11")
			},
		},
		{
			name: testESTIdentityURIMismatch,
			mutateSecond: func(t *testing.T, cert *x509.Certificate) {
				t.Helper()

				uri, err := url.Parse(testESTIdentityDifferentURI)
				require.NoError(t, err)

				cert.URIs[0] = uri
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var first, second *x509.Certificate
			if !tc.nilFirst {
				first = newESTIdentityTestCertificate(t)
			}

			if !tc.nilSecond {
				second = newESTIdentityTestCertificate(t)
			}

			if tc.mutateSecond != nil {
				tc.mutateSecond(t, second)
			}

			require.Equal(t, tc.wantMatch, sameESTCertificateIdentity(first, second))
		})
	}
}

func TestSameESTCertificateIdentityRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		nilCert    bool
		nilRequest bool
		mutate     func(*testing.T, *cryptoutilCAProfileSubject.Request)
		wantMatch  bool
	}{
		{name: "matching request", wantMatch: true},
		{name: "nil certificate", nilCert: true},
		{name: "nil request", nilRequest: true},
		{
			name: testESTIdentitySubjectMismatch,
			mutate: func(_ *testing.T, request *cryptoutilCAProfileSubject.Request) {
				request.CommonName = testESTIdentityDifferentName
			},
		},
		{
			name: testESTIdentityDNSMismatch,
			mutate: func(_ *testing.T, request *cryptoutilCAProfileSubject.Request) {
				request.DNSNames[0] = testESTIdentityDifferentName
			},
		},
		{
			name: testESTIdentityEmailMismatch,
			mutate: func(_ *testing.T, request *cryptoutilCAProfileSubject.Request) {
				request.EmailAddresses[0] = testESTIdentityDifferentEmail
			},
		},
		{
			name: testESTIdentityIPMismatch,
			mutate: func(_ *testing.T, request *cryptoutilCAProfileSubject.Request) {
				request.IPAddresses[0] = "192.0.2.11"
			},
		},
		{
			name: testESTIdentityURIMismatch,
			mutate: func(_ *testing.T, request *cryptoutilCAProfileSubject.Request) {
				request.URIs[0] = testESTIdentityDifferentURI
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert := newESTIdentityTestCertificate(t)
			request := newESTIdentityTestRequest(t)

			if tc.nilCert {
				cert = nil
			}

			if tc.nilRequest {
				request = nil
			}

			if tc.mutate != nil {
				tc.mutate(t, request)
			}

			require.Equal(t, tc.wantMatch, sameESTCertificateIdentityRequest(cert, request))
		})
	}
}

func newESTIdentityTestCertificate(t *testing.T) *x509.Certificate {
	t.Helper()

	primaryURI, err := url.Parse(testESTIdentityPrimaryURI)
	require.NoError(t, err)
	secondaryURI, err := url.Parse(testESTIdentitySecondaryURI)
	require.NoError(t, err)

	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
		},
		DNSNames:       []string{testESTIdentityPrimaryDNS, testESTIdentitySecondaryDNS},
		EmailAddresses: []string{testESTIdentityPrimaryEmail, testESTIdentitySecondaryEmail},
		IPAddresses: []net.IP{
			net.ParseIP(testESTIdentityPrimaryIP),
			net.ParseIP(testESTIdentitySecondaryIP),
		},
		URIs: []*url.URL{primaryURI, secondaryURI},
	}
}

func newESTIdentityTestRequest(t *testing.T) *cryptoutilCAProfileSubject.Request {
	t.Helper()

	return &cryptoutilCAProfileSubject.Request{
		CommonName:     testDomainNameExample,
		Organization:   []string{testOrganizationTest},
		DNSNames:       []string{testESTIdentityPrimaryDNS, testESTIdentitySecondaryDNS},
		EmailAddresses: []string{testESTIdentityPrimaryEmail, testESTIdentitySecondaryEmail},
		IPAddresses:    []string{testESTIdentityPrimaryIP, testESTIdentitySecondaryIP},
		URIs:           []string{testESTIdentityPrimaryURI, testESTIdentitySecondaryURI},
	}
}
