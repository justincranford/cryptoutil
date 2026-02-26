// Copyright (c) 2025 Justin Cranford
//
//

package certificate

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"crypto"
"crypto/x509"
"fmt"
"io"
"testing"
"time"

cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

"github.com/stretchr/testify/require"
)

// caTestFailSigner is a crypto.Signer whose Sign method always returns an error.
// Used to trigger x509.CreateCertificate failures in coverage tests.
type caTestFailSigner struct {
pub crypto.PublicKey
}

func (s caTestFailSigner) Public() crypto.PublicKey { return s.pub }

func (s caTestFailSigner) Sign(_ io.Reader, _ []byte, _ crypto.SignerOpts) ([]byte, error) {
return nil, fmt.Errorf("intentional sign failure for testing")
}

// TestCreateCASubject_ValidationErrors tests all validation error paths in CreateCASubject.
func TestCreateCASubject_ValidationErrors(t *testing.T) {
t.Parallel()

keyPair := testKeyGenPool.GetMany(1)[0]

tests := []struct {
name           string
issuerSubject  *Subject
issuerPrivKey  any
subjectName    string
subjectKeyPair *cryptoutilSharedCryptoKeygen.KeyPair
maxPathLen     int
wantErrContain string
}{
{
name:           "nil issuerSubject with non-nil issuerPrivKey",
issuerSubject:  nil,
issuerPrivKey:  keyPair.Private,
subjectName:    "Test CA",
subjectKeyPair: keyPair,
maxPathLen:     0,
wantErrContain: "issuerSubject is nil but issuerPrivateKey is not nil",
},
{
name: "non-nil issuerSubject with nil issuerPrivKey",
issuerSubject: &Subject{
SubjectName: "Root CA",
},
issuerPrivKey:  nil,
subjectName:    "Test CA",
subjectKeyPair: keyPair,
maxPathLen:     0,
wantErrContain: "issuerSubject is not nil but issuerPrivateKey is nil",
},
{
name:           "empty subjectName",
issuerSubject:  nil,
issuerPrivKey:  nil,
subjectName:    "",
subjectKeyPair: keyPair,
maxPathLen:     0,
wantErrContain: "subjectName should not be empty",
},
{
name:           "nil subjectKeyPair",
issuerSubject:  nil,
issuerPrivKey:  nil,
subjectName:    "Test CA",
subjectKeyPair: nil,
maxPathLen:     0,
wantErrContain: "subjectKeyPair should not be nil",
},
{
name:          "nil subjectKeyPair.Public",
issuerSubject: nil,
issuerPrivKey: nil,
subjectName:   "Test CA",
subjectKeyPair: &cryptoutilSharedCryptoKeygen.KeyPair{
Public:  nil,
Private: keyPair.Private,
},
maxPathLen:     0,
wantErrContain: "subjectKeyPair.Public should not be nil",
},
{
name:          "nil subjectKeyPair.Private",
issuerSubject: nil,
issuerPrivKey: nil,
subjectName:   "Test CA",
subjectKeyPair: &cryptoutilSharedCryptoKeygen.KeyPair{
Public:  keyPair.Public,
Private: nil,
},
maxPathLen:     0,
wantErrContain: "subjectKeyPair.Private should not be nil",
},
{
name:           "negative maxPathLen",
issuerSubject:  nil,
issuerPrivKey:  nil,
subjectName:    "Test CA",
subjectKeyPair: keyPair,
maxPathLen:     -1,
wantErrContain: "maxPathLen should not be negative",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
t.Parallel()

_, err := CreateCASubject(
tt.issuerSubject,
tt.issuerPrivKey,
tt.subjectName,
tt.subjectKeyPair,
testCACertValidity10Years,
tt.maxPathLen,
)
require.Error(t, err)
require.Contains(t, err.Error(), tt.wantErrContain)
})
}
}

// TestCreateCASubjects_ErrorFromNilPublicKey triggers the error path in CreateCASubjects
// when a key pair with a nil Public key is passed.
func TestCreateCASubjects_ErrorFromNilPublicKey(t *testing.T) {
t.Parallel()

keyPair := testKeyGenPool.GetMany(1)[0]

nilPublicKeyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
Public:  nil,
Private: keyPair.Private,
}

_, err := CreateCASubjects(
[]*cryptoutilSharedCryptoKeygen.KeyPair{nilPublicKeyPair},
"Error Test CA",
testCACertValidity10Years,
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to create CA subject")
}

// TestBuildTLSCertificate_ValidationErrors tests validation error paths in BuildTLSCertificate.
func TestBuildTLSCertificate_ValidationErrors(t *testing.T) {
t.Parallel()

keyPair := testKeyGenPool.GetMany(1)[0]

tests := []struct {
name           string
subject        *Subject
wantErrContain string
}{
{
name: "empty certificate chain",
subject: &Subject{
KeyMaterial: KeyMaterial{
CertificateChain: []*x509.Certificate{},
PublicKey:        keyPair.Public,
PrivateKey:       keyPair.Private,
},
},
wantErrContain: "certificate chain is empty",
},
{
name: "nil private key",
subject: &Subject{
KeyMaterial: KeyMaterial{
CertificateChain: []*x509.Certificate{{}},
PublicKey:        keyPair.Public,
PrivateKey:       nil,
},
},
wantErrContain: "private key is nil",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
t.Parallel()

_, _, _, err := BuildTLSCertificate(tt.subject)
require.Error(t, err)
require.Contains(t, err.Error(), tt.wantErrContain)
})
}
}

// TestCertificateTemplateEndEntity_ExceedsDuration tests that CertificateTemplateEndEntity
// returns an error when the duration exceeds the subscriber cert maximum.
func TestCertificateTemplateEndEntity_ExceedsDuration(t *testing.T) {
t.Parallel()

// 398 days exceeds TLSDefaultSubscriberCertDuration (397 days).
excessiveDuration := cryptoutilSharedMagic.TLSMaxValidityEndEntityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour

_, err := CertificateTemplateEndEntity(
"Issuer CA",
"Test End Entity",
excessiveDuration,
[]string{"example.com"},
nil, nil, nil,
x509.KeyUsageDigitalSignature,
[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to generate certificate validity period")
}

// TestCreateEndEntitySubject_TemplateError tests that CreateEndEntitySubject
// returns an error when CertificateTemplateEndEntity fails (excessive duration).
func TestCreateEndEntitySubject_TemplateError(t *testing.T) {
t.Parallel()

keyPairs := testKeyGenPool.GetMany(2)

caSubject, err := CreateCASubject(nil, nil, "Test CA", keyPairs[1], testCACertValidity10Years, 0)
require.NoError(t, err)

// 398 days exceeds TLSDefaultSubscriberCertDuration.
excessiveDuration := cryptoutilSharedMagic.TLSMaxValidityEndEntityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour

_, err = CreateEndEntitySubject(
caSubject,
keyPairs[0],
"Test End Entity",
excessiveDuration,
[]string{"example.com"},
nil, nil, nil,
x509.KeyUsageDigitalSignature,
[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to create end entity certificate template")
}

// TestCreateEndEntitySubject_SignError tests that CreateEndEntitySubject
// returns an error when SignCertificate fails (non-Signer private key).
func TestCreateEndEntitySubject_SignError(t *testing.T) {
t.Parallel()

keyPairs := testKeyGenPool.GetMany(2)

// Create a valid CA first (real cert chain).
realCASubject, err := CreateCASubject(nil, nil, "Real Test CA", keyPairs[1], testCACertValidity10Years, 0)
require.NoError(t, err)

// Clone the CA subject but replace private key with a non-signer.
fakeCASubject := &Subject{
SubjectName: realCASubject.SubjectName,
IssuerName:  realCASubject.IssuerName,
Duration:    realCASubject.Duration,
IsCA:        true,
MaxPathLen:  0,
KeyMaterial: KeyMaterial{
CertificateChain: realCASubject.KeyMaterial.CertificateChain,
PublicKey:        realCASubject.KeyMaterial.PublicKey,
PrivateKey:       "not-a-crypto-signer", // Not a crypto.Signer.
},
}

_, err = CreateEndEntitySubject(
fakeCASubject,
keyPairs[0],
"Test End Entity",
testEndEntityCertValidity396Days,
[]string{"example.com"},
nil, nil, nil,
x509.KeyUsageDigitalSignature,
[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to sign end entity certificate")
}

// TestSignCertificate_NotSigner tests that SignCertificate errors when the
// issuer private key does not implement crypto.Signer.
func TestSignCertificate_NotSigner(t *testing.T) {
t.Parallel()

keyPair := testKeyGenPool.GetMany(1)[0]

template, err := CertificateTemplateCA("Issuer CA", "Test CA", testCACertValidity10Years, 0)
require.NoError(t, err)

_, _, _, err = SignCertificate(nil, "not-a-signer", template, keyPair.Public, x509.ECDSAWithSHA256)
require.Error(t, err)
require.Contains(t, err.Error(), "issuer private key is not a crypto.Signer")
}

// TestCreateCASubject_FailingSignerError tests that CreateCASubject returns an error
// when SignCertificate fails because the issuer private key Sign() method returns an error.
// This covers the SignCertificate error path (line 136) and the x509.CreateCertificate
// failure branch (line 275) in certificates.go.
func TestCreateCASubject_FailingSignerError(t *testing.T) {
t.Parallel()

keyPairs := testKeyGenPool.GetMany(2)

// Create a valid root CA as issuer.
issuerSubject, err := CreateCASubject(nil, nil, "Root Test CA", keyPairs[0], testCACertValidity10Years, 1)
require.NoError(t, err)

// caTestFailSigner implements crypto.Signer but Sign() always fails.
// x509.CreateCertificate will call Sign() and propagate the error.
failingKey := caTestFailSigner{pub: keyPairs[1].Public}

_, err = CreateCASubject(
issuerSubject,
failingKey,
"Intermediate CA",
keyPairs[1],
testCACertValidity10Years,
0,
)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to sign CA certificate")
}
