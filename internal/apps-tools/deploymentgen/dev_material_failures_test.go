// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	"crypto/elliptic"
	json "encoding/json"
	"errors"
	"testing"
	"time"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	"github.com/stretchr/testify/require"
)

const (
	testMaterialExtraKey           = "incomplete generated key set"
	testMaterialGraph              = "invalid graph"
	testMaterialLifecycle          = "foreign lifecycle"
	testMaterialNilContext         = "nil context"
	testMaterialNilGenerator       = "nil generator"
	testMaterialLeafEntropy        = "leaf entropy"
	testMaterialIssuerShape        = "issuer shape"
	testMaterialLeafShape          = "leaf shape"
	testMaterialCanceledLeaf       = "leaf cancellation"
	testMaterialEncoding           = "materializer encoding"
	testMaterialAbsentLeaf         = "missing leaf source"
	testMaterialDuplicateLeaf      = "duplicate leaf source"
	testMaterialAbsentAttestation  = "missing runtime attestation"
	testMaterialE2ELeaf            = "e2e leaf rejection"
	testMaterialInfraLeaf          = "infrastructure leaf rejection"
	testMaterialPrivateEncoding    = "private encoding rejection"
	testMaterialTrustMount         = "verification trust mount rejection"
	testMaterialRolesPerPod        = 4
	testMaterialInfraOffset        = 3
	testMaterialUnsupportedPrivate = "unsupported-private-key"
	testMaterialKeysetDrift        = "late protected keyset drift"
	testMaterialMismatchedPrivate  = "generated certificate-key mismatch"
	testMaterialUnissuedFile       = "protected/unissued.key"
)

func TestDevelopmentMaterialRejectsIncompleteGenerationAndTopology(t *testing.T) {
	t.Parallel()

	for _, test := range []string{
		testMaterialGraph, testMaterialLifecycle, testMaterialNilContext, testMaterialNilGenerator, testMaterialLeafEntropy,
		testMaterialIssuerShape, testMaterialLeafShape, testMaterialCanceledLeaf, testMaterialEncoding, testMaterialAbsentLeaf, testMaterialDuplicateLeaf, testMaterialAbsentAttestation,
		testMaterialExtraKey,
		testMaterialE2ELeaf, testMaterialInfraLeaf, testMaterialPrivateEncoding, testMaterialTrustMount,
		testMaterialKeysetDrift, testMaterialMismatchedPrivate,
	} {
		t.Run(test, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			plan, err := BuildDevelopmentRuntimePlan(model, lock, testDevelopmentRun, testServiceSMKMSMemSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			generate := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair

			switch test {
			case testMaterialGraph:
				plan.Model.ClusterDomain = ""
			case testMaterialLifecycle:
				plan.Lifecycle.Version = 0
			case testMaterialNilContext:
				ctx = nil
			case testMaterialNilGenerator:
				generate = nil
			case testMaterialLeafEntropy, testMaterialIssuerShape, testMaterialLeafShape, testMaterialCanceledLeaf:
				calls := 0
				generate = func(curve elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					calls++
					if test == testMaterialIssuerShape || (test == testMaterialLeafShape && calls > 1) {
						return &cryptoutilSharedCryptoKeygen.KeyPair{Private: testMaterialUnsupportedPrivate, Public: "unsupported-public-key"}, nil
					}

					if test == testMaterialLeafEntropy && calls > 1 {
						return nil, errors.New("PRIVATE")
					}

					pair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(curve)

					if test == testMaterialCanceledLeaf {
						cancel()
					}

					return pair, err
				}
			case testMaterialEncoding:
				plan.runtimeValues[0].Workloads[0].MaterializerConfig = testOpenJSONObject
			case testMaterialAbsentLeaf, testMaterialDuplicateLeaf:
				var input pki.CommandInput
				require.NoError(t, json.Unmarshal([]byte(plan.runtimeValues[0].Workloads[0].MaterializerConfig), &input))

				if test == testMaterialAbsentLeaf {
					input.Pods[0].PodName = "no-imported-leaf-source"
				} else {
					input.Pods = append(input.Pods, input.Pods[0])
				}

				raw, err := json.Marshal(input)
				require.NoError(t, err)

				plan.runtimeValues[0].Workloads[0].MaterializerConfig = string(raw)
			case testMaterialAbsentAttestation:
				plan.runtimeValues = nil
			case testMaterialExtraKey:
				plan.Lifecycle.Secrets[0].Files["unissued.key"] = testMaterialUnissuedFile
			case testMaterialTrustMount:
				plan.runtimeValues[0].Workloads[0].TrustSecretName = "unprovided-verification-trust"
			case testMaterialE2ELeaf, testMaterialInfraLeaf, testMaterialPrivateEncoding, testMaterialKeysetDrift, testMaterialMismatchedPrivate:
				calls := 0
				appLeaves := len(plan.runtimeValues[0].Workloads[0].InstanceServices) * testMaterialRolesPerPod
				generate = func(curve elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
					calls++
					if test == testMaterialKeysetDrift && calls == 1 {
						plan.Lifecycle.Secrets[0].Files["unissued.key"] = testMaterialUnissuedFile
					}

					if (test == testMaterialE2ELeaf && calls == appLeaves+2) || (test == testMaterialInfraLeaf && calls == appLeaves+testMaterialInfraOffset) {
						return nil, errors.New("PRIVATE")
					}

					pair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(curve)
					if err == nil && test == testMaterialPrivateEncoding && calls > 1 {
						pair.Private = testMaterialUnsupportedPrivate
					}

					if err == nil && test == testMaterialMismatchedPrivate && calls > 1 {
						other, otherErr := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(curve)
						require.NoError(t, otherErr)

						pair.Private = other.Private
					}

					return pair, err
				}
			}

			material, err := generateDevelopmentMaterial(ctx, plan, DevelopmentMaterialPolicy{Curve: elliptic.P384(), Lifetime: time.Hour}, generate)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "PRIVATE")
			require.Zero(t, len(material.Secrets))
			require.Zero(t, len(material.Token))
		})
	}
}
