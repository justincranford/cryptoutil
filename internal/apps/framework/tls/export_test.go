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
) *Generator {
	return &Generator{
		getKeyPairFn: getKeyPairFn,
		mkdirAllFn:   mkdirAllFn,
		writeFileFn:  writeFileFn,
		createCAFn:   createCAFn,
		createLeafFn: createLeafFn,
	}
}

// ExportedDefaultIPs exposes defaultIPs for testing.
func ExportedDefaultIPs() []net.IP {
	return defaultIPs()
}
