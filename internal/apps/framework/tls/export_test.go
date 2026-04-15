// Copyright (c) 2025 Justin Cranford
//

package tls

import (
	"context"
	"crypto/x509"
	"io"
	"net"
	"os"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TelemetryFnType is the seam type for telemetry service creation.
type TelemetryFnType = func(context.Context) (*cryptoutilSharedTelemetry.TelemetryService, error)

// GeneratorFnType is the seam type for generator creation.
type GeneratorFnType = func(context.Context, *cryptoutilSharedTelemetry.TelemetryService) (*Generator, error)

// ExportedInitRun calls unexported initRun with the given seam functions, enabling parallel tests.
func ExportedInitRun(args []string, _ io.Reader, stdout, stderr io.Writer, telemetryFn TelemetryFnType, generatorFn GeneratorFnType) int {
	return initRun(args, stdout, stderr, telemetryFn, generatorFn)
}

// ExportedNewTestGenerator creates a Generator with injectable seams for testing.
func ExportedNewTestGenerator(
	mkdirAllFn func(string, os.FileMode) error,
	writeFileFn func(string, []byte, os.FileMode) error,
	createCAFn func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error),
	createLeafFn func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error),
	getKeyPairFn func() *cryptoutilSharedCryptoKeygen.KeyPair,
	encodePKCS12Fn func(priv any, cert *x509.Certificate, chain []*x509.Certificate) ([]byte, error),
	encodeTrustPKCS12Fn func(certs []*x509.Certificate) ([]byte, error),
	getRealmsForPSIDFn func(psID string) ([]string, error),
) *Generator {
	return &Generator{
		getKeyPairFn:        getKeyPairFn,
		mkdirAllFn:          mkdirAllFn,
		writeFileFn:         writeFileFn,
		createCAFn:          createCAFn,
		createLeafFn:        createLeafFn,
		encodePKCS12Fn:      encodePKCS12Fn,
		encodeTrustPKCS12Fn: encodeTrustPKCS12Fn,
		getRealmsForPSIDFn:  getRealmsForPSIDFn,
	}
}

// ExportedDefaultIPs exposes defaultIPs for testing.
func ExportedDefaultIPs() []net.IP {
	return defaultIPs()
}

// ExportedGenerateCAChain wraps generateCAChain for direct error-path testing.
func (g *Generator) ExportedGenerateCAChain(basePath, prefix, suffix string) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	return g.generateCAChain(basePath, prefix, suffix)
}

// ExportedWriteKeystore wraps writeKeystore for direct error-path testing.
func (g *Generator) ExportedWriteKeystore(basePath, dirName string, subject *cryptoutilSharedCryptoCertificate.Subject) error {
	return g.writeKeystore(basePath, dirName, subject)
}

// ExportedWriteTruststore wraps writeTruststore for direct error-path testing.
func (g *Generator) ExportedWriteTruststore(basePath, dirName string, subject *cryptoutilSharedCryptoCertificate.Subject) error {
	return g.writeTruststore(basePath, dirName, subject)
}

// ExportedGenerateServerLeafDir wraps generateServerLeafDir for direct testing.
func (g *Generator) ExportedGenerateServerLeafDir(basePath, dirName string, issuer *cryptoutilSharedCryptoCertificate.Subject, dns []string, ips []net.IP) error {
	return g.generateServerLeafDir(basePath, dirName, issuer, dns, ips)
}

// ExportedGenerateClientLeafDir wraps generateClientLeafDir for direct testing.
func (g *Generator) ExportedGenerateClientLeafDir(basePath, dirName string, issuer *cryptoutilSharedCryptoCertificate.Subject) error {
	return g.generateClientLeafDir(basePath, dirName, issuer)
}

// ExportedGenerateMutualLeafDir wraps generateMutualLeafDir for direct testing.
func (g *Generator) ExportedGenerateMutualLeafDir(basePath, dirName string, issuer *cryptoutilSharedCryptoCertificate.Subject, dns []string, ips []net.IP) error {
	return g.generateMutualLeafDir(basePath, dirName, issuer, dns, ips)
}

// ExportedReadRealmsForPSID exposes readRealmsForPSID for testing.
func ExportedReadRealmsForPSID(registryPath, psID string) ([]string, error) {
	return readRealmsForPSID(registryPath, psID)
}

// ExportedDefaultRealms exposes defaultRealms for testing.
func ExportedDefaultRealms() []string {
	return defaultRealms()
}
