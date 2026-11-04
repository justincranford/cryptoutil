// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	sha256 "crypto/sha256"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/hex"
	json "encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const verifyTestPrefix = "/run/cryptoutil/secrets/verified/"
const (
	verifyTestRead               = "verifier read failure"
	verifyTestMissingKey         = "verifier missing key"
	verifyTestHelp               = "verifier help"
	verifyTestZeroTime           = "verifier zero time"
	verifyTestDuplicate          = "verifier duplicate"
	verifyTestBadPath            = "verifier unsafe path"
	verifyTestDecode             = "verifier decode"
	verifyTestCancelled          = "verifier cancelled"
	verifyTestFlags              = "verifier flags"
	verifyTestChangedBytes       = "changed bytes"
	verifyTestOutput             = "verifier output"
	verifyTestDuplicateCert      = "duplicate certificate"
	verifyTestPath               = "verifier path"
	verifyTestEmptyInput         = "empty input"
	verifyTestInputRead          = "verifier input read"
	verifyTestForgedDigest       = "forged digest"
	verifyTestMissingCertificate = "missing certificate"
	verifyTestWrongRole          = "wrong role"
	verifyTestMissingDNS         = "missing server DNS"
	verifyTestWrongDNS           = "wrong DNS"
	verifyTestUncoveredKey       = "uncovered private key"
	verifyTestVersion            = cryptoutilSharedMagic.CLIVersionCommand
	verifyTestValid              = "valid"
	verifyTestUnknown            = cryptoutilSharedMagic.STRATEGY_UNKNOWN
	verifyTestTrailing           = "trailing"
	verifyTestVerification       = "verify"
)

func verifierFixture(t *testing.T) (VerificationInput, map[string][]byte, time.Time) {
	t.Helper()
	f := newTestPKIFixture(t)
	data := map[string][]byte{
		verifyTestPrefix + cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate: f.tlsBundle[SecretDataPublicCertificate],
		verifyTestPrefix + cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey:  f.tlsBundle[SecretDataPublicPrivateKey],
		verifyTestPrefix + "ca.crt": f.trust.PublicServerPEM,
	}
	input := VerificationInput{Version: 1, Files: map[string]string{}, Certificates: []VerificationCertificate{{
		Certificate: verifyTestPrefix + cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate, Key: verifyTestPrefix + cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey, Trust: verifyTestPrefix + "ca.crt",
		Usage: cryptoutilSharedMagic.CMD_SERVER, DNSNames: testExpectedPublicDNSNames(f.config, f.service, testNamespace),
	}}}

	for file, raw := range data {
		sum := sha256.Sum256(raw)
		input.Files[file] = hex.EncodeToString(sum[:])
	}

	return input, data, f.now
}

func TestVerifyPublishedFilesReadOnlyPolicy(t *testing.T) {
	t.Parallel()

	for _, test := range []string{verifyTestValid, verifyTestChangedBytes, verifyTestForgedDigest, verifyTestBadPath, verifyTestEmptyInput, verifyTestMissingCertificate, verifyTestWrongRole, verifyTestMissingDNS, verifyTestMissingKey, verifyTestDuplicateCert, verifyTestWrongDNS, verifyTestUncoveredKey, verifyTestCancelled, verifyTestRead, verifyTestZeroTime, verifyTestVersion} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			input, data, now := verifierFixture(t)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			switch test {
			case verifyTestChangedBytes:
				data[verifyTestPrefix+cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate] = []byte("DO-NOT-LOG")
			case verifyTestForgedDigest:
				input.Files[verifyTestPrefix+"ca.crt"] = "z" + input.Files[verifyTestPrefix+"ca.crt"][1:]
			case verifyTestBadPath:
				input.Files["/etc/private"] = "!"
			case verifyTestEmptyInput:
				input.Files = nil
			case verifyTestMissingCertificate:
				input.Certificates = nil
			case verifyTestWrongRole:
				input.Certificates[0].Usage = cryptoutilSharedMagic.DevSetupTestInvalidName
			case verifyTestMissingDNS:
				input.Certificates[0].DNSNames = nil
			case verifyTestMissingKey:
				input.Certificates[0].Key = verifyTestPrefix + "missing"
			case verifyTestDuplicateCert:
				input.Certificates = append(input.Certificates, input.Certificates[0])
			case verifyTestWrongDNS:
				input.Certificates[0].DNSNames = []string{testWrongDNSName}
			case verifyTestUncoveredKey:
				data[verifyTestPrefix+"extra.key"] = data[verifyTestPrefix+cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]
				input.Files[verifyTestPrefix+"extra.key"] = input.Files[verifyTestPrefix+cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey]
			case verifyTestCancelled:
				cancel()
			case verifyTestZeroTime:
				now = time.Time{}
			case verifyTestVersion:
				input.Version = 2
			}

			err := VerifyPublishedFiles(ctx, input, func(file string) ([]byte, error) {
				if test == verifyTestRead {
					return nil, errors.New("DO-NOT-LOG")
				}

				return data[file], nil
			}, now)
			if test == verifyTestValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "DO-NOT-LOG")
			}

			if test == verifyTestCancelled {
				require.ErrorIs(t, err, context.Canceled)
			}
		})
	}
}

func TestVerifyMainAndStrictInput(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testCommandSuccess, verifyTestHelp, verifyTestFlags, verifyTestPath, verifyTestInputRead, verifyTestDecode, verifyTestDuplicate, verifyTestUnknown, verifyTestTrailing, verifyTestVerification, verifyTestOutput} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			input, data, _ := verifierFixture(t)
			raw, err := json.Marshal(input)
			require.NoError(t, err)

			args := []string{"--input=" + verifyTestPrefix + VerificationInputName}

			switch test {
			case verifyTestHelp:
				args = []string{testCommandHelp}
			case verifyTestFlags:
				args = []string{"-unknown"}
			case verifyTestPath:
				args = []string{"--input=/tmp/unsafe"}
			case verifyTestDecode:
				raw = []byte(cryptoutilSharedMagic.DevSetupTestInvalidName)
			case verifyTestDuplicate:
				raw = []byte(`{"version":1,"version":2}`)
			case verifyTestUnknown:
				raw = []byte(`{"unexpected":true}`)
			case verifyTestTrailing:
				raw = append(raw, []byte("{}")...)
			case verifyTestVerification:
				data[verifyTestPrefix+"ca.crt"] = nil
			}

			var out, stderr bytes.Buffer

			var stdout interface{ Write([]byte) (int, error) } = &out
			if test == verifyTestOutput {
				stdout = verifyFailureWriter{}
			}

			code := VerifyMainWith(args, stdout, &stderr, func(file string) ([]byte, error) {
				if test == verifyTestInputRead {
					return nil, errors.New("DO-NOT-LOG")
				}

				if file == verifyTestPrefix+VerificationInputName {
					return raw, nil
				}

				return data[file], nil
			})
			if test == testCommandSuccess || test == verifyTestHelp {
				require.Zero(t, code)
			} else {
				require.NotZero(t, code)
			}

			require.NotContains(t, out.String()+stderr.String(), "DO-NOT-LOG")
		})
	}
}

type verifyFailureWriter struct{}

func (verifyFailureWriter) Write([]byte) (int, error) { return 0, errors.New("DO-NOT-LOG") }

func TestVerifyMainProductionBoundary(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		args       []string
		failOutput bool
		success    bool
	}{
		{"production help", []string{testCommandHelp}, false, true},
		{"help output rejected", []string{testCommandHelp}, true, false},
		{"default protected read unavailable", []string{"--input=" + verifyTestPrefix + VerificationInputName}, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var (
				stdout, stderr bytes.Buffer
				output         interface{ Write([]byte) (int, error) } = &stdout
			)
			if test.failOutput {
				output = verifyFailureWriter{}
			}

			code := VerifyMain(test.args, nil, output, &stderr)
			require.Equal(t, test.success, code == 0)
			require.NotContains(t, stdout.String()+stderr.String(), "DO-NOT-LOG")
		})
	}
}

func TestVerifyPublishedFilesHonorsPostReadCancellation(t *testing.T) {
	t.Parallel()

	input, data, now := verifierFixture(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	reads := 0
	err := VerifyPublishedFiles(ctx, input, func(file string) ([]byte, error) {
		reads++
		if reads == len(input.Files) {
			cancel()
		}

		return data[file], nil
	}, now)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, len(input.Files), reads, "every hash-bound read finished before final cancellation check")
}
