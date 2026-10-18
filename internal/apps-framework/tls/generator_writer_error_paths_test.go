package tls

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestWriterFunctions_UnsupportedPrivateKey(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	subject := &cryptoutilSharedCryptoCertificate.Subject{
		KeyMaterial: cryptoutilSharedCryptoCertificate.KeyMaterial{
			CertificateChain: []*x509.Certificate{fixture.certificate},
			PrivateKey:       struct{}{},
		},
	}
	gen := &Generator{
		mkdirAllFn: func(string, os.FileMode) error {
			return nil
		},
		writeFileFn: func(string, []byte, os.FileMode) error {
			return nil
		},
		encodePKCS12Fn: func(any, *x509.Certificate, []*x509.Certificate) ([]byte, error) {
			return []byte("pkcs12"), nil
		},
		encodeTrustPKCS12Fn: func([]*x509.Certificate) ([]byte, error) {
			return []byte("truststore"), nil
		},
	}

	tests := []struct {
		name      string
		run       func(*testing.T) error
		wantError string
	}{
		{
			name:      "keystore",
			run:       func(t *testing.T) error { return gen.writeKeystore(t.TempDir(), "keystore", subject) },
			wantError: "pem encode key",
		},
		{
			name:      "TLS config",
			run:       func(t *testing.T) error { return gen.writeTLSConfigYAML(t.TempDir(), subject) },
			wantError: "pem encode issuing CA key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, tc.run(t), tc.wantError)
		})
	}
}

func TestWriteTruststore_MkdirFailure(t *testing.T) {
	t.Parallel()

	gen := &Generator{
		mkdirAllFn: func(string, os.FileMode) error {
			return errors.New("injected directory failure")
		},
	}

	err := gen.writeTruststore(t.TempDir(), "truststore", nil)

	require.ErrorContains(t, err, "injected directory failure")
}

func TestReadStagedFile_OpenFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rootPath  string
		relative  string
		wantError string
	}{
		{
			name:      "root directory missing",
			rootPath:  filepath.Join(t.TempDir(), "missing"),
			relative:  "file",
			wantError: "open staged root",
		},
		{
			name:      "file missing within staged root",
			rootPath:  t.TempDir(),
			relative:  "missing",
			wantError: "open staged file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := readStagedFile(tc.rootPath, tc.relative)

			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestPublishGeneratedMaterial_RejectsInvalidState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, destination string)
		wantError string
	}{
		{
			name: "destination is directory instead of file",
			setup: func(t *testing.T, destination string) {
				t.Helper()
				require.NoError(t, os.Mkdir(destination, cryptoutilSharedMagic.PKIInitCertsDirMode))
			},
			wantError: testTLSNotRegularFile,
		},
		{
			name: "destination already initialized",
			setup: func(t *testing.T, destination string) {
				t.Helper()
				require.NoError(t, os.WriteFile(destination, []byte("existing"), cryptoutilSharedMagic.PKIInitPublicCertFileMode))
			},
			wantError: "already exists",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			destination := filepath.Join(t.TempDir(), "destination")
			tc.setup(t, destination)

			err := publishGeneratedMaterial(func(string, string) error {
				return nil
			}, []publishPath{{source: "source", destination: destination}})

			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestGenerate_TargetDirectoryCreationFailure(t *testing.T) {
	t.Parallel()

	outputRoot := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(outputRoot, []byte("not a directory"), cryptoutilSharedMagic.PKIInitPublicCertFileMode))

	gen := newPublishInfrastructureFailureGenerator()
	err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, outputRoot)

	require.ErrorContains(t, err, "failed to create output directory")
}

func TestGenerate_StagingInfrastructureFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configure  func(*Generator)
		wantErrors []string
	}{
		{
			name: "create output directory",
			configure: func(gen *Generator) {
				gen.mkdirAllFn = func(string, os.FileMode) error {
					return errors.New("injected output directory failure")
				}
			},
			wantErrors: []string{"failed to create output directory", "injected output directory failure"},
		},
		{
			name: "create staging directory",
			configure: func(gen *Generator) {
				gen.mkdirTempFn = func(string, string) (string, error) {
					return "", errors.New("injected staging directory failure")
				}
			},
			wantErrors: []string{"failed to create private staging directory", "injected staging directory failure"},
		},
		{
			name: "secure staging directory",
			configure: func(gen *Generator) {
				gen.secureStagingDirFn = func(string) error {
					return errors.New(testTLSInjectedStagingACLFailure)
				}
			},
			wantErrors: []string{"failed to secure private staging directory", testTLSInjectedStagingACLFailure},
		},
		{
			name: "cleanup staging directory",
			configure: func(gen *Generator) {
				gen.secureStagingDirFn = func(string) error {
					return errors.New(testTLSInjectedStagingACLFailure)
				}
				gen.removeAllFn = func(string) error {
					return errors.New("injected staging cleanup failure")
				}
			},
			wantErrors: []string{testTLSInjectedStagingACLFailure, "injected staging cleanup failure"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gen := newPublishInfrastructureFailureGenerator()
			tc.configure(gen)

			err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, t.TempDir())
			for _, wantErr := range tc.wantErrors {
				require.ErrorContains(t, err, wantErr)
			}
		})
	}
}

func TestGenerate_OutputFileIsNotRegular(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(outputDir, "tls-config.yml"), cryptoutilSharedMagic.PKIInitCertsDirMode))

	err := newPublishInfrastructureFailureGenerator().Generate(
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
		outputDir,
	)

	require.ErrorContains(t, err, "not a regular file")
}

func newPublishInfrastructureFailureGenerator() *Generator {
	return &Generator{
		mkdirAllFn:         os.MkdirAll,
		mkdirTempFn:        os.MkdirTemp,
		writeFileFn:        os.WriteFile,
		renameFn:           os.Rename,
		removeAllFn:        os.RemoveAll,
		secureStagingDirFn: securePrivateStagingDirectory,
	}
}
