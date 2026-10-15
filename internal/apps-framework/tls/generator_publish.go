// Copyright (c) 2025-2026 Justin Cranford.
//

package tls

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type publishPath struct {
	source      string
	destination string
	directory   bool
	published   bool
}

// validatePublishDestination rejects unsupported destination types before publication.
// In particular, symlinks are rejected so a generated output can never replace a target
// outside the configured output directory.
func validatePublishDestination(path string, directory bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("failed to inspect destination: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destination must not be a symbolic link")
	}

	if directory {
		if !info.IsDir() {
			return fmt.Errorf("destination exists but is not a directory")
		}

		return nil
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("destination exists but is not a regular file")
	}

	return nil
}

func ensureUninitializedOutputs(outputs []publishPath) error {
	for _, output := range outputs {
		_, err := os.Lstat(output.destination)
		if err == nil {
			return fmt.Errorf(
				"output %q already exists; pki-init initializes certificate material only and does not support certificate rotation",
				output.destination,
			)
		}

		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect output %q before initialization: %w", output.destination, err)
		}
	}

	return nil
}

// validateGeneratedMaterial checks the complete staged output before any consumer-visible
// path is changed. It validates the PEM chains, private-key/certificate pairs, CA bundle,
// and the CA material embedded in tls-config.yml.
func validateGeneratedMaterial(stagingRoot, tierID string) error {
	basePath := filepath.Join(stagingRoot, tierID)

	if err := validatePublishDestination(basePath, true); err != nil {
		return fmt.Errorf("staged tier directory: %w", err)
	}

	var certCount, keyCount, pkcs12Count int

	err := filepath.WalkDir(basePath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk staged path %q: %w", path, walkErr)
		}

		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("staged path %q must not be a symbolic link", path)
		}

		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect staged file %q: %w", path, err)
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("staged path %q is not a regular file", path)
		}

		if err := validateStagedFileMode(path, info.Mode()); err != nil {
			return err
		}

		relativePath, err := filepath.Rel(basePath, path)
		if err != nil {
			return fmt.Errorf("resolve staged file %q: %w", path, err)
		}

		data, err := readStagedFile(basePath, relativePath)
		if err != nil {
			return fmt.Errorf("read staged file %q: %w", path, err)
		}

		if len(data) == 0 {
			return fmt.Errorf("staged file %q is empty", path)
		}

		switch filepath.Ext(path) {
		case cryptoutilSharedMagic.PKIInitPublicCertFileExtension:
			certificates, err := parseCertificatePEM(path, data)
			if err != nil {
				return err
			}

			if err := validateCertificateChain(path, certificates); err != nil {
				return err
			}

			certCount++
		case cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension:
			certificateRelativePath := strings.TrimSuffix(relativePath, filepath.Ext(relativePath)) +
				cryptoutilSharedMagic.PKIInitPublicCertFileExtension

			certificatePEM, err := readStagedFile(basePath, certificateRelativePath)
			if err != nil {
				return fmt.Errorf("read certificate paired with staged key %q: %w", relativePath, err)
			}

			certificates, err := parseCertificatePEM(certificateRelativePath, certificatePEM)
			if err != nil {
				return err
			}

			if err := validatePrivateKeyMatchesCertificate(path, data, certificates[0]); err != nil {
				return err
			}

			keyCount++
		case cryptoutilSharedMagic.PKIInitPKCS12FileExtension:
			pkcs12Count++
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk staged certificate tree: %w", err)
	}

	if certCount == 0 || keyCount == 0 || pkcs12Count == 0 {
		return fmt.Errorf("staged certificate tree is incomplete: certificates=%d private-keys=%d pkcs12-files=%d",
			certCount, keyCount, pkcs12Count)
	}

	tlsConfigPath := filepath.Join(stagingRoot, "tls-config.yml")

	tlsConfigInfo, err := os.Stat(tlsConfigPath)
	if err != nil {
		return fmt.Errorf("inspect staged tls-config.yml: %w", err)
	}

	if err := validateStagedFileMode(tlsConfigPath, tlsConfigInfo.Mode()); err != nil {
		return fmt.Errorf("validate staged tls-config.yml permissions: %w", err)
	}

	if err := validateTLSConfig(tlsConfigPath); err != nil {
		return err
	}

	if err := validateAdminCABundle(filepath.Join(stagingRoot, cryptoutilSharedMagic.PKIInitAdminCABundleFilename)); err != nil {
		return err
	}

	return nil
}

func readStagedFile(rootPath, relativePath string) ([]byte, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open staged root %q: %w", rootPath, err)
	}

	file, err := root.Open(relativePath)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("open staged file %q: %w", relativePath, err),
			root.Close(),
		)
	}

	data, readErr := io.ReadAll(file)
	fileCloseErr := file.Close()

	rootCloseErr := root.Close()
	if err := errors.Join(readErr, fileCloseErr, rootCloseErr); err != nil {
		return nil, fmt.Errorf("read staged file %q: %w", relativePath, err)
	}

	return data, nil
}

func validateStagedFileMode(path string, mode os.FileMode) error {
	expectedMode := cryptoutilSharedMagic.PKIInitPublicCertFileMode
	privateMaterial := false

	switch filepath.Ext(path) {
	case cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension:
		expectedMode = cryptoutilSharedMagic.PKIInitPrivateKeyFileMode
		privateMaterial = true
	case cryptoutilSharedMagic.PKIInitPKCS12FileExtension:
		if filepath.Base(filepath.Dir(path)) != "truststore" {
			expectedMode = cryptoutilSharedMagic.PKIInitPrivateKeyFileMode
			privateMaterial = true
		}
	case ".yml":
		expectedMode = cryptoutilSharedMagic.PKIInitPrivateKeyFileMode
		privateMaterial = true
	}

	permissions := mode.Perm()
	if runtime.GOOS == "windows" {
		if permissions&0o222 != 0 {
			return fmt.Errorf("staged file %q has writable permissions on Windows", path)
		}

		if privateMaterial {
			if err := securePrivateFileACL(path); err != nil {
				return fmt.Errorf("secure staged private file %q ACL: %w", path, err)
			}
		}

		return nil
	}

	if permissions&0o400 == 0 || permissions&^expectedMode.Perm() != 0 {
		return fmt.Errorf(
			"staged file %q has permissions %04o, expected no broader than %04o and owner-readable",
			path,
			permissions,
			expectedMode.Perm(),
		)
	}

	return nil
}

func parseCertificatePEM(path string, data []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate

	remaining := data
	for len(bytes.TrimSpace(remaining)) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			return nil, fmt.Errorf("staged certificate file %q contains invalid PEM data", path)
		}

		remaining = rest

		if block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate {
			continue
		}

		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse staged certificate in %q: %w", path, err)
		}

		certificates = append(certificates, certificate)
	}

	if len(certificates) == 0 {
		return nil, fmt.Errorf("staged certificate file %q contains no certificates", path)
	}

	return certificates, nil
}

func validateCertificateChain(path string, certificates []*x509.Certificate) error {
	for i := 0; i < len(certificates)-1; i++ {
		if err := certificates[i].CheckSignatureFrom(certificates[i+1]); err != nil {
			return fmt.Errorf("staged certificate chain %q has invalid link %d: %w", path, i, err)
		}
	}

	return nil
}

func validatePrivateKeyMatchesCertificate(path string, keyPEM []byte, certificate *x509.Certificate) error {
	block, rest := pem.Decode(keyPEM)
	if block == nil || block.Type != cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey || len(bytes.TrimSpace(rest)) != 0 {
		return fmt.Errorf("staged private key %q is not a single PKCS#8 PEM block", path)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse staged private key %q: %w", path, err)
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return fmt.Errorf("staged private key %q does not implement crypto.Signer", path)
	}

	certificatePublicKey, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return fmt.Errorf("marshal certificate public key for %q: %w", path, err)
	}

	privatePublicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return fmt.Errorf("marshal private key public component for %q: %w", path, err)
	}

	if !bytes.Equal(certificatePublicKey, privatePublicKey) {
		return fmt.Errorf("staged private key %q does not match its certificate", path)
	}

	return nil
}

func validateTLSConfig(path string) error {
	data, err := readStagedFile(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return fmt.Errorf("read staged tls-config.yml: %w", err)
	}

	certPEM, err := readBase64ConfigValue(data, "tls-mixed-ca-cert-pem: ")
	if err != nil {
		return fmt.Errorf("validate staged tls-config.yml certificate: %w", err)
	}

	keyPEM, err := readBase64ConfigValue(data, "tls-mixed-ca-key-pem: ")
	if err != nil {
		return fmt.Errorf("validate staged tls-config.yml key: %w", err)
	}

	certificates, err := parseCertificatePEM(path, certPEM)
	if err != nil {
		return fmt.Errorf("validate staged tls-config.yml certificate PEM: %w", err)
	}

	if !certificates[0].IsCA {
		return fmt.Errorf("staged tls-config.yml does not contain a CA certificate")
	}

	if err := validateCertificateChain(path, certificates); err != nil {
		return fmt.Errorf("validate staged tls-config.yml certificate chain: %w", err)
	}

	if err := validatePrivateKeyMatchesCertificate(path, keyPEM, certificates[0]); err != nil {
		return fmt.Errorf("staged tls-config.yml CA key: %w", err)
	}

	return nil
}

func readBase64ConfigValue(data []byte, prefix string) ([]byte, error) {
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(line, prefix))
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", strings.TrimSuffix(prefix, " "), err)
		}

		return decoded, nil
	}

	return nil, fmt.Errorf("missing %s", strings.TrimSuffix(prefix, " "))
}

func validateAdminCABundle(path string) error {
	data, err := readStagedFile(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return fmt.Errorf("read staged admin CA bundle: %w", err)
	}

	certificates, err := parseCertificatePEM(path, data)
	if err != nil {
		return fmt.Errorf("validate staged admin CA bundle PEM: %w", err)
	}

	for i, certificate := range certificates {
		if !certificate.IsCA {
			return fmt.Errorf("staged admin CA bundle certificate %d is not a CA", i)
		}
	}

	return nil
}

func publishGeneratedMaterial(renameFn func(string, string) error, outputs []publishPath) error {
	if renameFn == nil {
		return fmt.Errorf("publication rename function is required")
	}

	for _, output := range outputs {
		if err := validatePublishDestination(output.destination, output.directory); err != nil {
			return fmt.Errorf("validate publication destination %q: %w", output.destination, err)
		}
	}

	if err := ensureUninitializedOutputs(outputs); err != nil {
		return fmt.Errorf("validate output publication state: %w", err)
	}

	for i := range outputs {
		if err := renameFn(outputs[i].source, outputs[i].destination); err != nil {
			rollbackErr := rollbackPublishedMaterial(outputs[:i+1])

			return errors.Join(fmt.Errorf("publish staged output %q: %w", outputs[i].destination, err), rollbackErr)
		}

		outputs[i].published = true
	}

	return nil
}

func rollbackPublishedMaterial(outputs []publishPath) error {
	var rollbackErrors []error

	for i := len(outputs) - 1; i >= 0; i-- {
		output := outputs[i]
		if output.published {
			if err := os.RemoveAll(output.destination); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("remove partially published output %q: %w", output.destination, err))
			}
		}
	}

	return errors.Join(rollbackErrors...)
}
