// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	"crypto/elliptic"
	"os"
	"path/filepath"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentGeneratedPublicationClosesProtectedPolicy(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err)

	read := func(name string) ([]byte, error) { return os.ReadFile(filepath.Join(root, filepath.FromSlash(name))) }
	raw, err := read("api/cryptosuite-registry/k8s/k8s-model.json")
	require.NoError(t, err)
	model, err := gen.DecodeModel(raw)
	require.NoError(t, err)
	raw, err = read("api/cryptosuite-registry/k8s/k8s-toolchain.lock.json")
	require.NoError(t, err)
	lock, err := gen.DecodeToolchainLock(raw)
	require.NoError(t, err)
	plan, err := gen.BuildDevelopmentRuntimePlan(model, lock, gen.DevelopmentSmokeRun, gen.DevelopmentSmokeSelector, lock.Images.Postgres.Reference)
	require.NoError(t, err)
	raw, err = read("configs/sm-kms/sm-kms-domain.yml")
	require.NoError(t, err)
	plan, err = gen.ConfigureDevelopmentSmoke(plan, raw)
	require.NoError(t, err)
	input, err := developmentSmokePreflight(plan, lock, read)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	material, err := gen.GenerateDevelopmentMaterial(ctx, plan, gen.DevelopmentMaterialPolicy{Curve: elliptic.P384(), Lifetime: 24 * time.Hour})
	require.NoError(t, err)
	manifest, err := developmentPublication(input.RuntimeValues)
	require.NoError(t, err)
	request, err := preparePKIVerification(input,
		execution.Request{Model: plan.Model, Inventory: input.Inventory, Input: plan.Lifecycle, Secrets: material.Secrets}, manifest)
	require.NoError(t, err)
	require.NotNil(t, request.VerifyPKI)
	require.Len(t, request.Secrets, len(plan.Lifecycle.Secrets))

	for _, secret := range request.Secrets {
		if secret.Metadata.Labels[cryptoutilSharedMagic.DeploymentMaterializerSecretLabelRole] != "" {
			require.Equal(t, "external-secret-import", secret.Metadata.Labels["cu.io/owner"])
		}
	}
}
