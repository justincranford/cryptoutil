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
	"log/slog"
	"math/big"
	"os"
	"testing"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/stretchr/testify/assert"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
)

func TestMain(m *testing.M) {
	telemetryService, err := cryptoutilTelemetry.NewService(testCtx, "asn1_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	testTelemetryService = telemetryService
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

func TestPemEncodeDecodeRSA(t *testing.T) {
	keyPairOriginal, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	privateKeyOriginal := keyPairOriginal
	publicKeyOriginal := &keyPairOriginal.PublicKey
	assert.IsType(t, &rsa.PrivateKey{}, privateKeyOriginal)
	assert.IsType(t, &rsa.PublicKey{}, publicKeyOriginal)

	privateKeyPemBytes, err := PemEncode(privateKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("RSA Private", "pem", string(privateKeyPemBytes))

	privateKeyDecoded, err := PemDecode(privateKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &rsa.PrivateKey{}, privateKeyDecoded)
	assert.Equal(t, privateKeyOriginal, privateKeyDecoded.(*rsa.PrivateKey))

	publicKeyPemBytes, err := PemEncode(publicKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("RSA Public", "pem", string(privateKeyPemBytes))

	publicKeyDecoded, err := PemDecode(publicKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &rsa.PublicKey{}, publicKeyDecoded)
	assert.Equal(t, publicKeyOriginal, publicKeyDecoded.(*rsa.PublicKey))
}

func TestPemEncodeDecodeECDSA(t *testing.T) {
	keyPairOriginal, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	privateKeyOriginal := keyPairOriginal
	publicKeyOriginal := &keyPairOriginal.PublicKey
	assert.IsType(t, &ecdsa.PrivateKey{}, privateKeyOriginal)
	assert.IsType(t, &ecdsa.PublicKey{}, publicKeyOriginal)

	privateKeyPemBytes, err := PemEncode(privateKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDSA Private", "pem", string(privateKeyPemBytes))

	privateKeyDecoded, err := PemDecode(privateKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &ecdsa.PrivateKey{}, privateKeyDecoded)
	assert.Equal(t, privateKeyOriginal, privateKeyDecoded.(*ecdsa.PrivateKey))

	publicKeyPemBytes, err := PemEncode(publicKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDSA Public", "pem", string(privateKeyPemBytes))

	publicKeyDecoded, err := PemDecode(publicKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &ecdsa.PublicKey{}, publicKeyDecoded)
	assert.Equal(t, publicKeyOriginal, publicKeyDecoded.(*ecdsa.PublicKey))
}

func TestPemEncodeDecodeECDH(t *testing.T) {
	t.Skip("Blocked by bug: https://github.com/golang/go/issues/71919")
	keyPairOriginal, err := ecdh.P256().GenerateKey(rand.Reader)
	assert.NoError(t, err)
	privateKeyOriginal := keyPairOriginal
	publicKeyOriginal := keyPairOriginal.PublicKey()
	assert.IsType(t, &ecdh.PrivateKey{}, privateKeyOriginal)
	assert.IsType(t, &ecdh.PublicKey{}, publicKeyOriginal)

	privateKeyPemBytes, err := PemEncode(privateKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDH Private", "pem", string(privateKeyPemBytes))

	privateKeyDecoded, err := PemDecode(privateKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &ecdh.PrivateKey{}, privateKeyDecoded)
	assert.Equal(t, privateKeyOriginal, privateKeyDecoded.(*ecdh.PrivateKey))

	publicKeyPemBytes, err := PemEncode(publicKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("ECDH Public", "pem", string(privateKeyPemBytes))

	publicKeyDecoded, err := PemDecode(publicKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &ecdh.PublicKey{}, publicKeyDecoded)
	assert.Equal(t, publicKeyOriginal, publicKeyDecoded.(*ecdh.PublicKey))
}

func TestPemEncodeDecodeEdDSA(t *testing.T) {
	publicKeyOriginal, privateKeyOriginal, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	assert.IsType(t, ed25519.PrivateKey{}, privateKeyOriginal)
	assert.IsType(t, ed25519.PublicKey{}, publicKeyOriginal)

	privateKeyPemBytes, err := PemEncode(privateKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("ED Private", "pem", string(privateKeyPemBytes))

	privateKeyDecoded, err := PemDecode(privateKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, ed25519.PrivateKey{}, privateKeyDecoded)
	assert.Equal(t, privateKeyOriginal, privateKeyDecoded.(ed25519.PrivateKey))

	publicKeyPemBytes, err := PemEncode(publicKeyOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("ED Public", "pem", string(privateKeyPemBytes))

	publicKeyDecoded, err := PemDecode(publicKeyPemBytes)
	assert.NoError(t, err)

	assert.IsType(t, ed25519.PublicKey{}, publicKeyDecoded)
	assert.Equal(t, publicKeyOriginal, publicKeyDecoded.(ed25519.PublicKey))
}

func TestPemEncodeDecodeCertificate(t *testing.T) {
	privateKeyOriginal, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	certificateTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	certificateDerBytes, err := x509.CreateCertificate(rand.Reader, certificateTemplate, certificateTemplate, &privateKeyOriginal.PublicKey, privateKeyOriginal)
	assert.NoError(t, err)

	certificateOriginal, err := x509.ParseCertificate(certificateDerBytes)
	assert.NoError(t, err)

	certificatePemBytes, err := PemEncode(certificateOriginal)
	assert.NoError(t, err)
	testTelemetryService.Slogger.Info("Cert", "pem", string(certificatePemBytes))

	certificateDecoded, err := PemDecode(certificatePemBytes)
	assert.NoError(t, err)

	assert.IsType(t, &x509.Certificate{}, certificateDecoded)
	assert.Equal(t, certificateOriginal, certificateDecoded.(*x509.Certificate))
}
