// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	"crypto/x509"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGenerateSerialNumber_RandIntError tests the error path when crand.Int fails.
func TestGenerateSerialNumber_RandIntError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	serial, err := GenerateSerialNumber()
	require.Error(t, err)
	require.Nil(t, serial)
	require.Contains(t, err.Error(), "failed to generate random serial number offset")
}

// TestGenerateNotBeforeNotAfter_RandIntError tests the error path when crand.Int fails
// inside generateNotBeforeNotAfter.
func TestGenerateNotBeforeNotAfter_RandIntError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	now := time.Now().UTC()

	_, _, err := generateNotBeforeNotAfter(now, cryptoutilSharedMagic.HoursPerDay*time.Hour, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate random range offset")
}

// TestCertificateTemplateCA_SerialNumberError tests the error path when
// GenerateSerialNumber fails inside CertificateTemplateCA.
func TestCertificateTemplateCA_SerialNumberError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	_, err := CertificateTemplateCA("Issuer", "Subject", cryptoutilSharedMagic.HoursPerDay*time.Hour, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate serial number")
}

// TestCertificateTemplateEndEntity_SerialNumberError tests the error path when
// GenerateSerialNumber fails inside CertificateTemplateEndEntity.
func TestCertificateTemplateEndEntity_SerialNumberError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	_, err := CertificateTemplateEndEntity("Issuer", "Subject", cryptoutilSharedMagic.HoursPerDay*time.Hour, nil, nil, nil, nil, 0, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate serial number")
}

// TestRandomizedNotBeforeNotAfterCA_RandIntError tests the error path when
// generateNotBeforeNotAfter fails via randIntFn inside randomizedNotBeforeNotAfterCA.
func TestRandomizedNotBeforeNotAfterCA_RandIntError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	_, _, err := randomizedNotBeforeNotAfterCA(time.Now().UTC(), cryptoutilSharedMagic.HoursPerDay*time.Hour, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate notBefore/notAfter")
}

// TestRandomizedNotBeforeNotAfterEndEntity_RandIntError tests the error path when
// generateNotBeforeNotAfter fails via randIntFn inside randomizedNotBeforeNotAfterEndEntity.
func TestRandomizedNotBeforeNotAfterEndEntity_RandIntError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	_, _, err := randomizedNotBeforeNotAfterEndEntity(time.Now().UTC(), cryptoutilSharedMagic.HoursPerDay*time.Hour, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate notBefore/notAfter")
}

// TestCertificateTemplateCA_ValidityPeriodError tests the error path when
// randomizedNotBeforeNotAfterCA fails inside CertificateTemplateCA (serial succeeds, notBefore fails).
func TestCertificateTemplateCA_ValidityPeriodError(t *testing.T) {
	callCount := 0
	original := randIntFn

	randIntFn = func(r io.Reader, max *big.Int) (*big.Int, error) {
		callCount++
		if callCount == 1 {
			return original(r, max) // First call (serial number) succeeds.
		}

		return nil, fmt.Errorf("injected randInt error on notBefore") // Second call (notBefore) fails.
	}

	defer func() { randIntFn = original }()

	_, err := CertificateTemplateCA("Issuer", "Subject", cryptoutilSharedMagic.HoursPerDay*time.Hour, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate certificate validity period")
}

// TestCertificateTemplateEndEntity_ValidityPeriodError tests the error path when
// randomizedNotBeforeNotAfterEndEntity fails inside CertificateTemplateEndEntity.
func TestCertificateTemplateEndEntity_ValidityPeriodError(t *testing.T) {
	callCount := 0
	original := randIntFn

	randIntFn = func(r io.Reader, max *big.Int) (*big.Int, error) {
		callCount++
		if callCount == 1 {
			return original(r, max) // First call (serial number) succeeds.
		}

		return nil, fmt.Errorf("injected randInt error on notBefore") // Second call (notBefore) fails.
	}

	defer func() { randIntFn = original }()

	_, err := CertificateTemplateEndEntity("Issuer", "Subject", cryptoutilSharedMagic.HoursPerDay*time.Hour, nil, nil, nil, nil, 0, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate certificate validity period")
}

// TestCreateCASubject_CertificateTemplateError tests the error path when
// CertificateTemplateCA fails inside CreateCASubject via randIntFn failure.
func TestCreateCASubject_CertificateTemplateError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	keyPair := testKeyGenPool.GetMany(1)[0]

	_, err := CreateCASubject(nil, nil, "Test CA", keyPair, cryptoutilSharedMagic.HoursPerDay*time.Hour, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create CA certificate template")
}

// TestCreateCASubjects_CertificateTemplateError tests the error path when
// CertificateTemplateCA fails inside CreateCASubjects via randIntFn failure.
func TestCreateCASubjects_CertificateTemplateError(t *testing.T) {
	original := randIntFn
	randIntFn = func(_ io.Reader, _ *big.Int) (*big.Int, error) {
		return nil, fmt.Errorf("injected randInt error")
	}

	defer func() { randIntFn = original }()

	keyPairs := testKeyGenPool.GetMany(2)

	_, err := CreateCASubjects(keyPairs, "Test CA", cryptoutilSharedMagic.HoursPerDay*time.Hour)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create CA subject")
}

// TestSignCertificate_ParseCertificateError tests the error path when
// x509.ParseCertificate fails inside SignCertificate.
func TestSignCertificate_ParseCertificateError(t *testing.T) {
	original := parseCertificateFn
	parseCertificateFn = func(_ []byte) (*x509.Certificate, error) {
		return nil, fmt.Errorf("injected parse certificate error")
	}

	defer func() { parseCertificateFn = original }()

	keyPair := testKeyGenPool.GetMany(1)[0]

	certTemplate, err := CertificateTemplateCA("Issuer", "Test CA", testCACertValidity10Years, 0)
	require.NoError(t, err)

	_, _, _, err = SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse certificate")
}

// TestSerializeSubjects_JSONMarshalError tests the error path when
// json.Marshal fails inside SerializeSubjects.
func TestSerializeSubjects_JSONMarshalError(t *testing.T) {
	original := jsonMarshalFn
	jsonMarshalFn = func(_ any) ([]byte, error) {
		return nil, fmt.Errorf("injected json marshal error")
	}

	defer func() { jsonMarshalFn = original }()

	keyPair := testKeyGenPool.GetMany(1)[0]

	certTemplate, err := CertificateTemplateCA("Issuer", "Test CA", testCACertValidity10Years, 0)
	require.NoError(t, err)

	cert, _, _, err := SignCertificate(nil, keyPair.Private, certTemplate, keyPair.Public, x509.ECDSAWithSHA256)
	require.NoError(t, err)

	subjects := []*Subject{{
		SubjectName: "Test CA",
		IssuerName:  "Issuer",
		Duration:    testCACertValidity10Years,
		IsCA:        true,
		MaxPathLen:  0,
		KeyMaterial: KeyMaterial{
			CertificateChain: []*x509.Certificate{cert},
			PublicKey:        keyPair.Public,
		},
	}}

	_, err = SerializeSubjects(subjects, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to serialize KeyMaterialEncoded")
}
