// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/sm-kms/server/pki/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/sm-kms/server/pki/profile/subject"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const estOtherProfileID = "other"

func TestSelectESTProfile(t *testing.T) {
	t.Parallel()

	serverProfile := &ProfileConfig{ID: cryptoutilSharedMagic.PKIProfileIDTLSServer}
	clientProfile := &ProfileConfig{ID: cryptoutilSharedMagic.PKIProfileIDTLSClient}

	tests := []struct {
		name       string
		profiles   map[string]*ProfileConfig
		wantID     string
		wantNoProf bool
	}{
		{name: "no configured profiles", profiles: map[string]*ProfileConfig{}, wantNoProf: true},
		{name: "single configured profile", profiles: map[string]*ProfileConfig{serverProfile.ID: serverProfile}, wantID: serverProfile.ID},
		{name: "TLS client profile takes precedence", profiles: map[string]*ProfileConfig{
			serverProfile.ID: serverProfile,
			clientProfile.ID: clientProfile,
		}, wantID: clientProfile.ID},
		{name: "ambiguous profiles without TLS client profile", profiles: map[string]*ProfileConfig{
			serverProfile.ID:  serverProfile,
			estOtherProfileID: {ID: estOtherProfileID},
		}, wantNoProf: true},
		{name: "mismatched profile ID is rejected", profiles: map[string]*ProfileConfig{
			cryptoutilSharedMagic.PKIProfileIDTLSClient: serverProfile,
		}, wantNoProf: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			profile, err := (&Handler{profiles: tc.profiles}).selectESTProfile()
			if tc.wantNoProf {
				require.ErrorIs(t, err, errESTNoProfile)
				require.Nil(t, profile)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantID, profile.ID)
		})
	}
}

func TestBuildESTIssueRequestAppliesConfiguredPolicy(t *testing.T) {
	t.Parallel()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrDER, err := x509.CreateCertificateRequest(crand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
		},
	}, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrDER)
	require.NoError(t, err)

	profile := &ProfileConfig{
		ID: cryptoutilSharedMagic.PKIProfileIDTLSClient,
		SubjectProfile: &cryptoutilCAProfileSubject.Profile{
			Constraints: cryptoutilCAProfileSubject.Constraints{RequireCommonName: true},
		},
		CertificateProfile: &cryptoutilCAProfileCertificate.Profile{
			Validity: cryptoutilCAProfileCertificate.ValidityConfig{
				Duration:    "24h",
				MaxDuration: "48h",
			},
		},
	}

	issueReq, err := (&Handler{}).buildESTIssueRequest(csr, profile)
	require.NoError(t, err)
	require.Equal(t, 24*time.Hour, issueReq.ValidityDuration)
	require.Equal(t, testDomainNameExample, issueReq.SubjectRequest.CommonName)
	require.Equal(t, []string{testOrganizationTest}, issueReq.SubjectRequest.Organization)
}

func TestBuildESTIssueRequestRejectsSubjectOutsideConfiguredPolicy(t *testing.T) {
	t.Parallel()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrDER, err := x509.CreateCertificateRequest(crand.Reader, &x509.CertificateRequest{}, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrDER)
	require.NoError(t, err)

	profile := &ProfileConfig{
		ID: cryptoutilSharedMagic.PKIProfileIDTLSClient,
		SubjectProfile: &cryptoutilCAProfileSubject.Profile{
			Constraints: cryptoutilCAProfileSubject.Constraints{RequireCommonName: true},
		},
	}

	_, err = (&Handler{}).buildESTIssueRequest(csr, profile)
	require.ErrorContains(t, err, "subject does not satisfy EST profile")
}
