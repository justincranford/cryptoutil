// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testVerificationMalformedManifest = "malformed manifest"
	testVerificationChangedPlan       = "protected plan changed"
	testVerificationMissingManifest   = "missing manifest"
	testVerificationGeneration        = "unsupported generation"
	testVerificationFilename          = "wrong filename"
	testVerificationManifest          = "verified manifest"
)

func TestProductionPKIVerifierLoadsExplicitManifestAndRejectsGeneration(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testVerificationManifest, testVerificationMissingManifest, testVerificationMalformedManifest, testVerificationGeneration, testVerificationFilename} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			input, request, manifest := pkiVerificationFixture(t)
			file := filepath.Join(t.TempDir(), "publication.json")
			raw, err := json.Marshal(manifest)
			require.NoError(t, err)

			if test == testVerificationMalformedManifest {
				raw = []byte("{")
			}

			if test == testVerificationGeneration {
				request.Input.Jobs[0].Args = []string{testPKIInitCommand, testPKIDomainArgument, testPKIOutputArgument}
			}

			if test == testVerificationFilename {
				request.Input.Jobs[0].Args[1] = "--input=/run/cryptoutil/secrets/owned/other.json"
			}

			if test != testVerificationMissingManifest {
				require.NoError(t, os.WriteFile(file, raw, 0o600))
			}

			process := lifecycleExecutionProcess(nil)

			prepared, err := process.verifyPKI(input, request, file)
			if test == testVerificationManifest {
				require.NoError(t, err)
				require.NotNil(t, prepared.VerifyPKI)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPostJobVerifierReevaluatesCertificatePolicy(t *testing.T) {
	t.Parallel()

	for _, test := range []string{testVerificationChangedPlan, "runtime binding changed"} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			input, request, manifest := pkiVerificationFixture(t)
			prepared, err := preparePKIVerification(input, request, manifest)
			require.NoError(t, err)

			if test == testVerificationChangedPlan {
				for index := range prepared.Secrets {
					if prepared.Secrets[index].Metadata.Name == testPKITrustSecret {
						prepared.Secrets[index].Data[testPKICAKey] = []byte(cryptoutilSharedMagic.DevSetupTestInvalidName)
					}
				}
			} else {
				input.RuntimeValues[0].Workloads[0].FrameworkConfigFile = ""
			}

			job := gen.LifecycleJobValues{LifecycleJobInput: request.Input.Jobs[0], Release: input.RuntimeValues[0].HelmRelease}
			job.Name += "-" + request.Model.Revision
			client := &verifyPublicationClient{}
			err = prepared.VerifyPKI(context.Background(), job, client)
			require.Error(t, err)
			require.Len(t, client.calls, 1+len(request.Secrets), "policy recheck follows read-only live identity/material checks")
		})
	}
}

func TestVerificationImportsNeverReadsUnownedNamespaceOrMount(t *testing.T) {
	t.Parallel()
	_, request, _ := pkiVerificationFixture(t)
	read := verificationImports("unowned", nil, request.Secrets)
	_, err := read("/run/cryptoutil/secrets/" + request.Secrets[0].Metadata.Name + "/" + pki.SecretDataPublicCertificate)
	require.Error(t, err)
	require.False(t, containsPKIName(nil, "unowned"))
}

func TestLifecycleCLIMissingOrRejectedPKIVerifierDoesNotExecute(t *testing.T) {
	t.Parallel()

	for _, missing := range []bool{true, false} {
		t.Run(map[bool]string{true: "missing adapter", false: "unsupported contract"}[missing], func(t *testing.T) {
			t.Parallel()
			args, _ := executionCLIArgs(t)
			calls := 0

			process := lifecycleExecutionIO{
				read: func(string) ([]byte, error) { return []byte(`{"version":1,"secrets":[],"jobs":[]}`), nil },
				load: func(string, string, string, func(string) ([]byte, error)) (gen.BootstrapPreflightInput, error) {
					return gen.BootstrapPreflightInput{}, nil
				},
				admit: func(gen.BootstrapPreflightInput) (gen.CapacityEstimate, error) { return gen.CapacityEstimate{}, nil },
				importSecrets: func(gen.Model, gen.Inventory, gen.LifecycleInput, string) ([]gen.LifecycleSecret, error) {
					return nil, nil
				},
				execute: func(context.Context, execution.Request, execution.KubectlConfig, execution.Policy) (execution.Result, error) {
					calls++

					return execution.Result{}, nil
				},
			}
			if !missing {
				process.verifyPKI = func(_ gen.BootstrapPreflightInput, r execution.Request, _ string) (execution.Request, error) {
					return r, execution.ErrUnsupportedPKI
				}
			}

			var output, stderr bytes.Buffer
			require.NotZero(t, lifecycleExecuteMainWithIO(args, &output, &stderr, process))
			require.Zero(t, calls)
		})
	}
}
