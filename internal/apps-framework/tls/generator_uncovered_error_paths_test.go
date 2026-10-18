// Copyright (c) 2025-2026 Justin Cranford.
//

package tls

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestFilesystemValidation_NonNotExistErrors(t *testing.T) {
	t.Parallel()

	invalidPath := "invalid\x00path"
	tests := []struct {
		name    string
		run     func() error
		wantErr string
	}{
		{
			name: "target directory",
			run: func() error {
				return (&Generator{}).validateTargetDir(invalidPath)
			},
			wantErr: "failed to stat target directory",
		},
		{
			name: "publish destination",
			run: func() error {
				return validatePublishDestination(invalidPath, false)
			},
			wantErr: "failed to inspect destination",
		},
		{
			name: "uninitialized output",
			run: func() error {
				return ensureUninitializedOutputs([]publishPath{{destination: invalidPath}})
			},
			wantErr: "inspect output",
		},
		{
			name: testTLSStagedTierDirectory,
			run: func() error {
				return validateGeneratedMaterial(invalidPath, testPublishTierID)
			},
			wantErr: testTLSStagedTierDirectory,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, tc.run(), tc.wantErr)
		})
	}
}

func TestValidatePublishDestination_LstatSymlink(t *testing.T) {
	t.Parallel()

	err := validatePublishDestinationWithLstat(
		"staged-link",
		false,
		func(string) (os.FileInfo, error) {
			return validationFileInfo{mode: os.ModeSymlink}, nil
		},
	)

	require.ErrorContains(t, err, "symbolic link")
}

func TestValidatePublishDestination_FileInsteadOfDirectory(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "destination")
	require.NoError(t, os.WriteFile(path, []byte("file"), cryptoutilSharedMagic.PKIInitPublicCertFileMode))

	require.ErrorContains(t, validatePublishDestination(path, true), "not a directory")
}

func TestReadStagedFile_ReadFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "file"), []byte("staged"), cryptoutilSharedMagic.PKIInitPublicCertFileMode))

	_, err := readStagedFileWithReadAll(root, "file", func(io.Reader) ([]byte, error) {
		return nil, errors.New("injected staged read failure")
	})

	require.ErrorContains(t, err, "read staged file")
	require.ErrorContains(t, err, "injected staged read failure")
}

func TestValidateTLSConfig_ErrorPaths(t *testing.T) {
	t.Parallel()

	caFixture := newPublishValidationFixture(t)
	nonCAFixture := newPublishValidationFixtureWithCA(t, false)
	otherFixture := newPublishValidationFixture(t)
	invalidChain := append(append([]byte{}, caFixture.certificatePEM...), otherFixture.certificatePEM...)

	tests := []struct {
		name    string
		data    []byte
		missing bool
		wantErr string
	}{
		{name: "missing file", missing: true, wantErr: "read staged tls-config.yml"},
		{name: "missing certificate setting", data: []byte(testTLSConfigKeyPrefix + "AA==\n"), wantErr: testTLSMissingCertPEMSetting},
		{name: testTLSCertificatePEMInvalid, data: makeTLSConfig([]byte("not pem"), caFixture.privateKeyPEM), wantErr: "certificate PEM"},
		{name: testTLSNonCACertificate, data: makeTLSConfig(nonCAFixture.certificatePEM, nonCAFixture.privateKeyPEM), wantErr: "does not contain a CA certificate"},
		{name: testTLSInvalidCertificateChain, data: makeTLSConfig(invalidChain, caFixture.privateKeyPEM), wantErr: "certificate chain"},
		{name: "CA private key mismatch", data: makeTLSConfig(caFixture.certificatePEM, otherFixture.privateKeyPEM), wantErr: testTLSPrivateKeyDoesNotMatch},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "tls-config.yml")
			if !tc.missing {
				require.NoError(t, os.WriteFile(path, tc.data, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode))
			}

			require.ErrorContains(t, validateTLSConfig(path), tc.wantErr)
		})
	}
}

func TestValidateGeneratedMaterial_PropagatesConfigAndBundleFailures(t *testing.T) {
	t.Parallel()

	caFixture := newPublishValidationFixture(t)
	nonCAFixture := newPublishValidationFixtureWithCA(t, false)
	tests := []struct {
		name      string
		configure func(*testing.T, string)
		wantErr   string
	}{
		{
			name: "TLS config validation failure",
			configure: func(t *testing.T, root string) {
				t.Helper()

				path := filepath.Join(root, "tls-config.yml")
				writePublishFile(t, path, makeTLSConfig([]byte("not pem"), caFixture.privateKeyPEM), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
			},
			wantErr: "validate staged tls-config.yml certificate PEM",
		},
		{
			name: "admin bundle validation failure",
			configure: func(t *testing.T, root string) {
				t.Helper()

				path := filepath.Join(root, cryptoutilSharedMagic.PKIInitAdminCABundleFilename)
				writePublishFile(t, path, nonCAFixture.certificatePEM, cryptoutilSharedMagic.PKIInitPublicCertFileMode)
			},
			wantErr: "admin CA bundle certificate",
		},
		{
			name: "TLS config permission failure",
			configure: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.Chmod(filepath.Join(root, "tls-config.yml"), cryptoutilSharedMagic.FilePermAllReadWrite))
			},
			wantErr: testTLSPermissions,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeValidPublishTree(t, root, caFixture)
			tc.configure(t, root)

			err := validateGeneratedMaterial(root, testPublishTierID)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestValidateGeneratedMaterialWithOps_WalkAndReadFailures(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	tests := []struct {
		name        string
		entryName   string
		entryType   fs.FileMode
		info        fs.FileInfo
		infoErr     error
		walkErr     error
		relativeErr error
		readErr     error
		pairedKey   bool
		wantErr     string
	}{
		{name: "walk error", walkErr: errors.New("injected walk failure"), wantErr: testTLSWalkStagedPath},
		{name: "symlink entry", entryName: "linked.crt", entryType: os.ModeSymlink, wantErr: testTLSPublishSymbolicLink},
		{name: "entry info error", entryName: "broken.crt", infoErr: errors.New("injected entry info failure"), wantErr: "inspect staged file"},
		{name: "non-regular entry", entryName: testNamedPipeCertificateName, info: validationFileInfo{mode: os.ModeNamedPipe | 0o600}, wantErr: testTLSNotRegularFile},
		{name: "relative path error", entryName: testCertificateName, relativeErr: errors.New("injected relative path failure"), wantErr: "resolve staged file"},
		{name: "staged read error", entryName: testCertificateName, readErr: errors.New("injected staged read failure"), wantErr: "read staged file"},
		{name: testTLSInvalidPairedCertificate, entryName: testPrivateKeyName, pairedKey: true, wantErr: testTLSInvalidPEM},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			basePath := filepath.Join(root, testPublishTierID)
			require.NoError(t, os.Mkdir(basePath, cryptoutilSharedMagic.PKIInitCertsDirMode))
			entryPath := filepath.Join(basePath, tc.entryName)

			info := tc.info
			if info == nil && tc.infoErr == nil && tc.walkErr == nil && tc.entryType&os.ModeSymlink == 0 {
				require.NoError(t, os.WriteFile(entryPath, []byte("staged"), cryptoutilSharedMagic.PKIInitPublicCertFileMode))
				info, _ = os.Stat(entryPath)
			}

			entry := validationDirEntry{name: tc.entryName, entryType: tc.entryType, info: info, infoErr: tc.infoErr}
			readCalls := 0

			ops := generatedMaterialValidationOps{
				walkDirFn: func(_ string, walkFn fs.WalkDirFunc) error {
					return walkFn(entryPath, entry, tc.walkErr)
				},
				relativePathFn: filepath.Rel,
				readStagedFileFn: func(_, _ string) ([]byte, error) {
					readCalls++
					switch {
					case tc.pairedKey && readCalls == 1:
						return fixture.privateKeyPEM, nil
					case tc.pairedKey:
						return []byte("not pem"), nil
					case tc.readErr != nil:
						return nil, tc.readErr
					default:
						return []byte("staged"), nil
					}
				},
				validateStagedFileMode: func(string, os.FileMode) error { return nil },
			}
			if tc.relativeErr != nil {
				ops.relativePathFn = func(string, string) (string, error) {
					return "", tc.relativeErr
				}
			}

			err := validateGeneratedMaterialWithOps(root, testPublishTierID, ops)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

type validationDirEntry struct {
	name      string
	entryType fs.FileMode
	info      fs.FileInfo
	infoErr   error
}

func (entry validationDirEntry) Name() string               { return entry.name }
func (entry validationDirEntry) IsDir() bool                { return false }
func (entry validationDirEntry) Type() fs.FileMode          { return entry.entryType }
func (entry validationDirEntry) Info() (fs.FileInfo, error) { return entry.info, entry.infoErr }

type validationFileInfo struct {
	mode fs.FileMode
}

func (info validationFileInfo) Name() string       { return testNamedPipeCertificateName }
func (info validationFileInfo) Size() int64        { return 1 }
func (info validationFileInfo) Mode() fs.FileMode  { return info.mode }
func (info validationFileInfo) ModTime() time.Time { return time.Time{} }
func (info validationFileInfo) IsDir() bool        { return false }
func (info validationFileInfo) Sys() any           { return nil }

func TestValidatePrivateKeyMatchesCertificate_UnsupportedPublicKey(t *testing.T) {
	t.Parallel()

	fixture := newPublishValidationFixture(t)
	certificate := *fixture.certificate
	certificate.PublicKey = struct{}{}

	require.ErrorContains(
		t,
		validatePrivateKeyMatchesCertificate(testPrivateKeyName, fixture.privateKeyPEM, &certificate),
		"marshal certificate public key",
	)
}

func TestWriteKeystore_MkdirFailure(t *testing.T) {
	t.Parallel()

	gen := &Generator{
		mkdirAllFn: func(string, os.FileMode) error {
			return errors.New("injected keystore directory failure")
		},
	}

	require.ErrorContains(t, gen.writeKeystore(t.TempDir(), "keystore", nil), "injected keystore directory failure")
}

func TestGenerateCAChain_RootKeystoreFailure(t *testing.T) {
	t.Parallel()

	gen := &Generator{
		getKeyPairFn: func() *cryptoutilSharedCryptoKeygen.KeyPair {
			return nil
		},
		mkdirAllFn: func(string, os.FileMode) error {
			return errors.New("injected root keystore directory failure")
		},
		createCAFn: func(
			_ *cryptoutilSharedCryptoCertificate.Subject,
			_ any,
			_ string,
			_ *cryptoutilSharedCryptoKeygen.KeyPair,
			_ time.Duration,
			_ int,
		) (*cryptoutilSharedCryptoCertificate.Subject, error) {
			return &cryptoutilSharedCryptoCertificate.Subject{}, nil
		},
	}

	_, err := gen.generateCAChain(t.TempDir(), "root", "test")
	require.ErrorContains(t, err, "root CA keystore")
	require.ErrorContains(t, err, "injected root keystore directory failure")
}
