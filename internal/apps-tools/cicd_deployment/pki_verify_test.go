// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	sha256 "crypto/sha256"
	"encoding/hex"
	json "encoding/json"
	"errors"
	"strings"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testVerifyPlanSecret = "verification-plan"
const (
	testVerifyAPIVersion          = "v1"
	testVerifyServerRole          = cryptoutilSharedMagic.CMD_SERVER
	testVerifyJobFailure          = "job"
	testVerifyForgedValue         = "forged"
	testVerifyInvalidManifest     = "verification invalid manifest"
	testVerifyBadCertificate      = "verification bad certificate"
	testVerifyDatabaseJob         = "db-init"
	testVerifyChangedMaterial     = "changed bytes"
	testVerifyExtraFile           = "extra manifest file"
	testVerifyGeneratingJob       = "generation"
	testVerifySupported           = "supported"
	testVerifyMissingInput        = "missing input"
	testVerifyUncoveredMountedKey = "uncovered mounted key"
	testVerifyWritePermission     = "write permission"
	testVerifyRuntimeIdentity     = "invalid runtime identity"
	testVerifyMissingAPIs         = "missing verifier APIs"
	testVerifyReplacedJob         = "job replaced"
	testVerifyReplacedSecret      = "secret replaced"
	testVerifyNotPKI              = "not PKI"
)

func pkiVerificationFixture(t *testing.T) (gen.BootstrapPreflightInput, execution.Request, PKIPublicationInput) {
	t.Helper()
	values, secrets, manifest := publicationFixture(t)
	namespace := values[0].Namespace
	plan := pki.VerificationInput{Version: 1, Files: map[string]string{}}
	file := func(secret, key string) string { return "/run/cryptoutil/secrets/" + secret + "/" + key }
	trust := file(testPKITrustSecret, testPKICAKey)

	for _, secret := range secrets {
		for key, raw := range secret.Data {
			sum := sha256.Sum256(raw)
			plan.Files[file(secret.Metadata.Name, key)] = hex.EncodeToString(sum[:])
		}
	}

	publicNames := manifest.Pods[0].ExpectedPublicDNSNames
	plan.Certificates = []pki.VerificationCertificate{
		{Certificate: file(secrets[0].Metadata.Name, pki.SecretDataPublicCertificate), Key: file(secrets[0].Metadata.Name, pki.SecretDataPublicPrivateKey), Trust: trust, Usage: testVerifyServerRole, DNSNames: publicNames},
		{Certificate: file(secrets[0].Metadata.Name, pki.SecretDataAdminCertificate), Key: file(secrets[0].Metadata.Name, pki.SecretDataAdminPrivateKey), Trust: trust, Usage: testVerifyServerRole, DNSNames: []string{manifest.Pods[0].ExpectedAdminDNS}},
		{Certificate: file(secrets[1].Metadata.Name, pki.SecretDataCertificate), Key: file(secrets[1].Metadata.Name, pki.SecretDataPrivateKey), Trust: trust, Usage: testPKIClientFailure},
		{Certificate: file(secrets[2].Metadata.Name, pki.SecretDataCertificate), Key: file(secrets[2].Metadata.Name, pki.SecretDataPrivateKey), Trust: trust, Usage: testPKIClientFailure},
	}
	raw, err := json.Marshal(plan)
	require.NoError(t, err)

	secrets = append(secrets, gen.LifecycleSecret{
		APIVersion: testVerifyAPIVersion, Kind: testPKISecretKind, Type: testPKIOpaque,
		Metadata: gen.LifecycleMetadata{Name: testVerifyPlanSecret, Namespace: namespace, Labels: map[string]string{testPKIRunLabel: testPKIRun, testPKIOwnerLabel: testPKIImportOwner}},
		Data:     map[string][]byte{pki.VerificationInputName: raw},
	})

	names := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		names = append(names, secret.Metadata.Name)
	}

	job := gen.LifecycleJobInput{Name: testPKIInitCommand, Namespace: namespace, Args: []string{pki.VerifyCommand, "--input=" + file(testVerifyPlanSecret, pki.VerificationInputName)}, SecretNames: names}
	request := execution.Request{Model: gen.Model{Run: testPKIRun, Revision: "r-certified"}, Input: gen.LifecycleInput{Jobs: []gen.LifecycleJobInput{job}}, Secrets: secrets}

	return gen.BootstrapPreflightInput{RuntimeValues: values}, request, manifest
}

type verifyPublicationClient struct {
	execution.Client
	calls []string
	fail  string
}

func (c *verifyPublicationClient) VerifyJob(context.Context, execution.Identity) error {
	c.calls = append(c.calls, "job")
	if c.fail == testVerifyJobFailure {
		return execution.ErrReplaced
	}

	return nil
}

func (c *verifyPublicationClient) VerifySecret(_ context.Context, secret gen.LifecycleSecret) error {
	c.calls = append(c.calls, secret.Metadata.Name)
	if c.fail == testPKISecretFailure {
		return execution.ErrCollision
	}

	return nil
}

func TestPreparePKIVerificationAndPostJobReadOnlyBoundary(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testVerifySupported, testVerifyGeneratingJob, testVerifyMissingInput, testVerifyInvalidManifest, testVerifyChangedMaterial, testVerifyUncoveredMountedKey, testVerifyExtraFile, testVerifyBadCertificate, testVerifyWritePermission, testVerifyRuntimeIdentity, testVerifyMissingAPIs, testVerifyReplacedJob, testVerifyReplacedSecret, testVerifyNotPKI} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			input, request, manifest := pkiVerificationFixture(t)

			switch test {
			case testVerifyGeneratingJob:
				request.Input.Jobs[0].Args = []string{testPKIInitCommand, testPKIDomainArgument, testPKIOutputArgument}
			case testVerifyMissingInput:
				request.Input.Jobs[0].Args[1] = "--input=/run/cryptoutil/secrets/absent/" + pki.VerificationInputName
			case testVerifyInvalidManifest:
				request.Secrets[len(request.Secrets)-1].Data[pki.VerificationInputName] = []byte("{")
			case testVerifyChangedMaterial:
				request.Secrets[3].Data[testPKICAKey] = []byte(testVerifyForgedValue)
			case testVerifyUncoveredMountedKey:
				request.Secrets[3].Data["uncovered"] = []byte("bounded")
			case testVerifyExtraFile:
				var plan pki.VerificationInput
				require.NoError(t, json.Unmarshal(request.Secrets[len(request.Secrets)-1].Data[pki.VerificationInputName], &plan))
				plan.Files["/run/cryptoutil/secrets/unowned/x"] = strings.Repeat("a", 64)
				raw, err := json.Marshal(plan)
				require.NoError(t, err)

				request.Secrets[len(request.Secrets)-1].Data[pki.VerificationInputName] = raw
			case testVerifyBadCertificate:
				request.Secrets[0].Data[pki.SecretDataPublicCertificate] = []byte(testVerifyForgedValue)
			case testVerifyWritePermission:
				request.Input.Jobs[0].WriteSecretNames = []string{request.Secrets[0].Metadata.Name}
			case testVerifyRuntimeIdentity:
				manifest.Pods[0].ExpectedAdminDNS = testVerifyForgedValue
			case testVerifyNotPKI:
				request.Input.Jobs[0].Name = testVerifyDatabaseJob
			}

			prepared, err := preparePKIVerification(input, request, manifest)

			switch test {
			case testVerifySupported, testVerifyMissingAPIs, testVerifyReplacedJob, testVerifyReplacedSecret, testVerifyNotPKI:
				require.NoError(t, err)

				if test == testVerifyNotPKI {
					return
				}

				client := &verifyPublicationClient{}
				if test == testVerifyReplacedJob {
					client.fail = testVerifyJobFailure
				}

				if test == testVerifyReplacedSecret {
					client.fail = testPKISecretFailure
				}

				var boundary execution.Client = client
				if test == testVerifyMissingAPIs {
					boundary = nil
				}

				job := gen.LifecycleJobValues{LifecycleJobInput: request.Input.Jobs[0], Release: input.RuntimeValues[0].HelmRelease}
				job.Name += "-" + request.Model.Revision

				err = prepared.VerifyPKI(context.Background(), job, boundary)
				if test == testVerifySupported {
					require.NoError(t, err)
					require.Len(t, client.calls, 1+len(request.Secrets))
				} else {
					require.Error(t, err)
					require.True(t, errors.Is(err, execution.ErrPrerequisite) || errors.Is(err, execution.ErrReplaced) || errors.Is(err, execution.ErrCollision))
				}
			default:
				require.Error(t, err)
			}
		})
	}
}
