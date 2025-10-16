package asn1

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("der_pem_test")
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, testSettings)
		defer testTelemetryService.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestPEMEncodeDecodeRSA(t *testing.T) {
	keyPairOriginal, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyOriginal := keyPairOriginal
	publicKeyOriginal := &keyPairOriginal.PublicKey

	require.IsType(t, &rsa.PrivateKey{}, privateKeyOriginal)
	require.IsType(t, &rsa.PublicKey{}, publicKeyOriginal)

	privateKeyPEMBytes, err := PEMEncode(privateKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("RSA Private", "pem", string(privateKeyPEMBytes))

	privateKeyDecoded, err := PEMDecode(privateKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, &rsa.PrivateKey{}, privateKeyDecoded)
	privateKeyDecodedTyped, ok := privateKeyDecoded.(*rsa.PrivateKey)
	require.True(t, ok, "privateKeyDecoded should be *rsa.PrivateKey")
	require.Equal(t, privateKeyOriginal, privateKeyDecodedTyped)

	publicKeyPEMBytes, err := PEMEncode(publicKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("RSA Public", "pem", string(privateKeyPEMBytes))

	publicKeyDecoded, err := PEMDecode(publicKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, &rsa.PublicKey{}, publicKeyDecoded)
	publicKeyDecodedTyped, ok := publicKeyDecoded.(*rsa.PublicKey)
	require.True(t, ok, "publicKeyDecoded should be *rsa.PublicKey")
	require.Equal(t, publicKeyOriginal, publicKeyDecodedTyped)
}

func TestPEMEncodeDecodeECDSA(t *testing.T) {
	keyPairOriginal, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	privateKeyOriginal := keyPairOriginal
	publicKeyOriginal := &keyPairOriginal.PublicKey

	require.IsType(t, &ecdsa.PrivateKey{}, privateKeyOriginal)
	require.IsType(t, &ecdsa.PublicKey{}, publicKeyOriginal)

	privateKeyPEMBytes, err := PEMEncode(privateKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDSA Private", "pem", string(privateKeyPEMBytes))

	privateKeyDecoded, err := PEMDecode(privateKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, &ecdsa.PrivateKey{}, privateKeyDecoded)
	privateKeyDecodedTyped, ok := privateKeyDecoded.(*ecdsa.PrivateKey)
	require.True(t, ok, "privateKeyDecoded should be *ecdsa.PrivateKey")
	require.Equal(t, privateKeyOriginal, privateKeyDecodedTyped)

	publicKeyPEMBytes, err := PEMEncode(publicKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDSA Public", "pem", string(privateKeyPEMBytes))

	publicKeyDecoded, err := PEMDecode(publicKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, &ecdsa.PublicKey{}, publicKeyDecoded)
	publicKeyDecodedTyped, ok := publicKeyDecoded.(*ecdsa.PublicKey)
	require.True(t, ok, "publicKeyDecoded should be *ecdsa.PublicKey")
	require.Equal(t, publicKeyOriginal, publicKeyDecodedTyped)
}

func TestPEMEncodeDecodeECDH(t *testing.T) {
	t.Skip("Blocked by bug: https://github.com/golang/go/issues/71919")

	keyPairOriginal, err := ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)

	privateKeyOriginal := keyPairOriginal
	publicKeyOriginal := keyPairOriginal.PublicKey()

	require.IsType(t, &ecdh.PrivateKey{}, privateKeyOriginal)
	require.IsType(t, &ecdh.PublicKey{}, publicKeyOriginal)

	privateKeyPEMBytes, err := PEMEncode(privateKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDH Private", "pem", string(privateKeyPEMBytes))

	privateKeyDecoded, err := PEMDecode(privateKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, &ecdh.PrivateKey{}, privateKeyDecoded)
	privateKeyDecodedTyped, ok := privateKeyDecoded.(*ecdh.PrivateKey)
	require.True(t, ok, "privateKeyDecoded should be *ecdh.PrivateKey")
	require.Equal(t, privateKeyOriginal, privateKeyDecodedTyped)

	publicKeyPEMBytes, err := PEMEncode(publicKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDH Public", "pem", string(privateKeyPEMBytes))

	publicKeyDecoded, err := PEMDecode(publicKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, &ecdh.PublicKey{}, publicKeyDecoded)
	publicKeyDecodedTyped, ok := publicKeyDecoded.(*ecdh.PublicKey)
	require.True(t, ok, "publicKeyDecoded should be *ecdh.PublicKey")
	require.Equal(t, publicKeyOriginal, publicKeyDecodedTyped)
}

func TestPEMEncodeDecodeEdDSA(t *testing.T) {
	publicKeyOriginal, privateKeyOriginal, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	require.IsType(t, ed25519.PrivateKey{}, privateKeyOriginal)
	require.IsType(t, ed25519.PublicKey{}, publicKeyOriginal)

	privateKeyPEMBytes, err := PEMEncode(privateKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("ED Private", "pem", string(privateKeyPEMBytes))

	privateKeyDecoded, err := PEMDecode(privateKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, ed25519.PrivateKey{}, privateKeyDecoded)
	privateKeyDecodedTyped, ok := privateKeyDecoded.(ed25519.PrivateKey)
	require.True(t, ok, "privateKeyDecoded should be ed25519.PrivateKey")
	require.Equal(t, privateKeyOriginal, privateKeyDecodedTyped)

	publicKeyPEMBytes, err := PEMEncode(publicKeyOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("ED Public", "pem", string(privateKeyPEMBytes))

	publicKeyDecoded, err := PEMDecode(publicKeyPEMBytes)
	require.NoError(t, err)

	require.IsType(t, ed25519.PublicKey{}, publicKeyDecoded)
	publicKeyDecodedTyped, ok := publicKeyDecoded.(ed25519.PublicKey)
	require.True(t, ok, "publicKeyDecoded should be ed25519.PublicKey")
	require.Equal(t, publicKeyOriginal, publicKeyDecodedTyped)
}

func TestPEMEncodeDecodeCertificate(t *testing.T) {
	privateKeyOriginal, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	certificateTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	certificateDERBytes, err := x509.CreateCertificate(rand.Reader, certificateTemplate, certificateTemplate, &privateKeyOriginal.PublicKey, privateKeyOriginal)
	require.NoError(t, err)

	certificateOriginal, err := x509.ParseCertificate(certificateDERBytes)
	require.NoError(t, err)

	certificatePEMBytes, err := PEMEncode(certificateOriginal)
	require.NoError(t, err)
	testTelemetryService.Slogger.Info("Cert", "pem", string(certificatePEMBytes))

	certificateDecoded, err := PEMDecode(certificatePEMBytes)
	require.NoError(t, err)

	require.IsType(t, &x509.Certificate{}, certificateDecoded)
	certificateDecodedTyped, ok := certificateDecoded.(*x509.Certificate)
	require.True(t, ok, "certificateDecoded should be *x509.Certificate")
	require.Equal(t, certificateOriginal, certificateDecodedTyped)
}
