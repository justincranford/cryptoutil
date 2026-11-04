// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	sha256 "crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	json "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	VerifyCommand         = "pki-verify"
	VerificationInputName = "pki-verification.json"
)

// VerificationInput binds a verify-only Job to exact host-prepublished bytes.
// It contains hashes and explicit public certificate expectations, not material.
type VerificationInput struct {
	Version      int                       `json:"version"`
	Files        map[string]string         `json:"files"`
	Certificates []VerificationCertificate `json:"certificates"`
}

type VerificationCertificate struct {
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Trust       string   `json:"trust"`
	Usage       string   `json:"usage"`
	DNSNames    []string `json:"dnsNames"`
	CommonName  string   `json:"commonName"`
}

func verificationPath(value string) bool {
	if !strings.HasPrefix(value, "/run/cryptoutil/secrets/") || path.Clean(value) != value || strings.ContainsAny(value, "\\\x00\r\n") {
		return false
	}

	parts := strings.Split(strings.TrimPrefix(value, "/run/cryptoutil/secrets/"), "/")

	return len(parts) == 2 && validDNSLabel(parts[0]) && safeConfigPath(parts[1])
}

func DecodeVerificationInput(raw []byte) (VerificationInput, error) {
	var input VerificationInput
	if len(raw) == 0 || len(raw) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes {
		return input, errors.New("verification input size rejected")
	}
	// Shared strict JSON decoding rejects duplicates/case aliases at all levels.
	var unique map[string]any
	if err := yamlVerificationFields(raw, &unique); err != nil {
		return input, err
	}

	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return input, errors.New("verification input rejected")
	}

	var tail any
	if err := decoder.Decode(&tail); !errors.Is(err, io.EOF) {
		return input, errors.New("verification trailing input rejected")
	}

	if !verificationFieldsMatch(raw, input) {
		return input, errors.New("verification property case or type rejected")
	}

	return input, nil
}

// VerifyPublishedFiles is read-only. Hashes bind every requested byte; the
// existing materializer policy verifies trust/key/chain/SAN/EKU/validity.
func VerifyPublishedFiles(ctx context.Context, input VerificationInput, read func(string) ([]byte, error), now time.Time) error {
	if ctx == nil || read == nil || now.IsZero() || input.Version != 1 || len(input.Files) == 0 || len(input.Files) > 1024 || len(input.Certificates) == 0 {
		return errors.New("verification contract required")
	}

	data := map[string][]byte{}
	total := 0

	for file, digest := range input.Files {
		if !verificationPath(file) || path.Base(file) == VerificationInputName || len(digest) != sha256.Size*2 {
			return errors.New("verification file binding rejected")
		}

		expected, err := hex.DecodeString(digest)
		if err != nil {
			return errors.New("verification digest rejected")
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("verification context: %w", err)
		}

		raw, err := read(file)
		if err != nil || len(raw) == 0 || len(raw) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes-total {
			return errors.New("verification protected material rejected")
		}

		total += len(raw)

		actual := sha256.Sum256(raw)
		if !bytesVerificationEqual(actual[:], expected) {
			return errors.New("verification material collision")
		}

		data[file] = raw
	}

	covered := map[string]bool{}

	for _, certificate := range input.Certificates {
		usage := x509.ExtKeyUsageClientAuth

		switch certificate.Usage {
		case cryptoutilSharedMagic.CMD_SERVER:
			usage = x509.ExtKeyUsageServerAuth

			if len(certificate.DNSNames) == 0 {
				return errors.New("verification server identity required")
			}
		case "client":
		default:
			return errors.New("verification certificate role rejected")
		}

		if covered[certificate.Certificate] || len(data[certificate.Certificate]) == 0 || len(data[certificate.Key]) == 0 || len(data[certificate.Trust]) == 0 {
			return errors.New("verification certificate binding rejected")
		}

		if err := validateCertificateMaterial(certificateMaterial{
			certPEM: data[certificate.Certificate], keyPEM: data[certificate.Key], trustPEM: data[certificate.Trust],
			expectedDNSNames: certificate.DNSNames, expectedCommonName: certificate.CommonName, usage: usage,
		}, now); err != nil {
			return errors.New("verification certificate policy rejected")
		}

		covered[certificate.Certificate], covered[certificate.Key], covered[certificate.Trust] = true, true, true
	}

	for file, raw := range data {
		if (strings.Contains(string(raw), "-----BEGIN CERTIFICATE-----") || strings.Contains(string(raw), "PRIVATE KEY-----")) && !covered[file] {
			return errors.New("verification has unchecked certificate material")
		}
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("verification context: %w", err)
	}

	return nil
}

func VerifyMain(args []string, _ io.Reader, stdout, stderr io.Writer) int {
	return VerifyMainWith(args, stdout, stderr, boundedCommandRead)
}

func VerifyMainWith(args []string, stdout, stderr io.Writer, read func(string) ([]byte, error)) int {
	flags := flag.NewFlagSet(VerifyCommand, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	file := flags.String("input", "", "explicit protected verification manifest")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = fmt.Fprintln(stdout, "cryptoutil pki-verify --input=/run/cryptoutil/secrets/OWNED-SECRET/"+VerificationInputName)
			if err == nil {
				return 0
			}
		}

		return commandFailure(stderr, "PKI verification flags rejected")
	}

	if flags.NArg() != 0 || !verificationPath(*file) || path.Base(*file) != VerificationInputName {
		return commandFailure(stderr, "PKI verification protected input required")
	}

	raw, err := read(*file)
	if err != nil {
		return commandFailure(stderr, "PKI verification input unavailable")
	}

	input, err := DecodeVerificationInput(raw)
	if err != nil {
		return commandFailure(stderr, "PKI verification input rejected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DeploymentMaterializerOverallTimeout)
	defer cancel()

	if err := VerifyPublishedFiles(ctx, input, read, time.Now().UTC()); err != nil {
		return commandFailure(stderr, "PKI verification rejected")
	}

	if _, err := fmt.Fprintln(stdout, "PASS: existing published PKI material verified; no generation or publication"); err != nil {
		return 1
	}

	return 0
}
