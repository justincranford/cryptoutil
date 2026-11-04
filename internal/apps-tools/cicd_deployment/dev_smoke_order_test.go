// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/elliptic"
	sha256 "crypto/sha256"
	"encoding/hex"
	json "encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"

	"github.com/stretchr/testify/require"
)

type developmentOrderRunner struct {
	t         *testing.T
	lifecycle *completionVerificationRunner
	docker    string
	archive   []byte
	reference string
	fail      string
	corrupt   string
	cancel    context.CancelFunc
	calls     []string
}

func (r *developmentOrderRunner) Run(ctx context.Context, executable string, args []string, raw []byte) ([]byte, error) {
	label := strings.Join(args, " ")

	r.calls = append(r.calls, label)
	if r.fail != "" && strings.Contains(label, r.fail) {
		if r.cancel != nil {
			r.cancel()
		}

		return nil, errors.New("PRIVATE")
	}

	if executable == r.docker {
		if args[0] == developmentImage && args[1] == developmentInspect {
			if r.corrupt == testDevMalformedDigest {
				return []byte("unparseable-inspector-response"), nil
			}

			return json.Marshal([]string{r.reference})
		}

		if args[0] == developmentImage && args[1] == developmentImageSave {
			return r.archive, nil
		}

		require.Contains(r.t, label, "images import")

		return nil, nil
	}

	if executable == r.lifecycle.helm {
		if strings.Contains(label, testDevLifecycleRenderPath) {
			return r.lifecycle.Run(ctx, executable, args, raw)
		}

		return []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: synthetic-test-only\n"), nil
	}

	if strings.Contains(label, "--request-timeout=30s") {
		if strings.Contains(label, " exec ") {
			require.Equal(r.t, 2, r.lifecycle.jobCreates)
		}

		return nil, nil
	}

	return r.lifecycle.Run(ctx, executable, args, raw)
}

func TestDevelopmentDeployOrdersProtectedBarriersAndRejectsTransport(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{
		"", "image inspect", "image save", "images import", "create -f", "cryptoutil-infra", "cryptoutil-runtime", "rollout", " exec ",
		testDevMalformedDigest, testDevMalformedArchive, testDevUnreadyNodes, testDevBadLifecycle, testDevBadConfig, testDevBadPrepare, testDevBadSecretType, testDevAPIDenial, testDevOutputFailure,
	} {
		t.Run("transport="+failure, func(t *testing.T) {
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

			material, err := gen.GenerateDevelopmentMaterial(ctx, plan, gen.DevelopmentMaterialPolicy{Curve: elliptic.P384(), Lifetime: developmentMaterialLifetime})
			require.NoError(t, err)
			manifest, err := developmentPublication(input.RuntimeValues)
			require.NoError(t, err)
			request, err := preparePKIVerification(input, execution.Request{
				Model: plan.Model, Inventory: input.Inventory, Input: plan.Lifecycle, Secrets: material.Secrets,
			}, manifest)
			require.NoError(t, err)
			testRoot := t.TempDir()
			config := execution.KubectlConfig{
				RepositoryRoot: testRoot, Helm: filepath.Join(testRoot, "helm.exe"),
				Kubectl: filepath.Join(testRoot, "kubectl.exe"), Kubeconfig: filepath.Join(testRoot, "kubeconfig"),
				Context: gen.DevelopmentSmokeRun, PollInterval: time.Millisecond,
			}
			base := &completionVerificationRunner{publicationRunner: &publicationRunner{
				t: t, input: input, helm: config.Helm, documents: map[string]map[string]any{},
			}, resources: map[string]map[string]any{}}
			index := []byte(`{"schemaVersion":2,"manifests":[]}`)
			hash := sha256.Sum256(index)

			var archive bytes.Buffer

			writer := tar.NewWriter(&archive)
			require.NoError(t, writer.WriteHeader(&tar.Header{Name: "blobs/sha256/" + hex.EncodeToString(hash[:]), Typeflag: tar.TypeReg, Size: int64(len(index))}))
			_, err = writer.Write(index)
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			runner := &developmentOrderRunner{
				t: t, lifecycle: base, docker: filepath.Join(testRoot, "docker.exe"),
				archive: archive.Bytes(), reference: "synthetic/test@sha256:" + hex.EncodeToString(hash[:]), fail: failure,
			}

			stdout := io.Discard

			switch failure {
			case testDevMalformedDigest:
				runner.fail, runner.corrupt = "", failure
			case testDevMalformedArchive:
				runner.fail = ""
				runner.archive = []byte("test malformed archive bytes")
			case testDevUnreadyNodes:
				runner.fail, runner.cancel = "wait nodes", cancel
			case testDevBadLifecycle:
				runner.fail = ""
				request.Input.Version = 0
			case testDevBadConfig:
				runner.fail = ""
				config.Context = ""
			case testDevBadPrepare:
				runner.fail = testDevLifecycleRenderPath
			case testDevBadSecretType:
				runner.fail = ""
				request.Secrets[0].Type = "unsupported-type"
			case testDevAPIDenial:
				runner.fail = ""
			case testDevOutputFailure:
				runner.fail = ""
				stdout = executionPrivateWriter{}
			}

			apiChecks := 0

			err = deployDevelopmentSmokeWithTransports(ctx, developmentSmokeInput{root: root, image: testDevSyntheticImage, docker: runner.docker, timeout: time.Minute},
				input, plan.Cluster, request, config, runner, stdout,
				func(context.Context, gen.BootstrapPreflightInput, []gen.LifecycleSecret, execution.KubectlConfig) error {
					apiChecks++

					require.Equal(t, 2, base.jobCreates)
					require.Positive(t, base.secretReads)

					if failure == testDevAPIDenial {
						return errors.New("PRIVATE")
					}

					return nil
				}, testDevelopmentImageImport)
			if failure == "" {
				require.NoError(t, err)
				require.Equal(t, 1, apiChecks)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")

				if failure == testDevAPIDenial || failure == testDevOutputFailure {
					require.Equal(t, 1, apiChecks)
				} else {
					require.Zero(t, apiChecks)
				}
			}
		})
	}
}

// Small synthetic archives are deliberately confined to this test adapter.
// Production uses native streaming processes and exact captured node UIDs.
func testDevelopmentImageImport(ctx context.Context, input developmentSmokeInput, cluster gen.LocalK3sPlan, reference string, runner execution.Runner) error {
	archive, err := runner.Run(ctx, input.docker, []string{developmentImage, developmentImageSave, reference}, nil)
	if err != nil {
		return err
	}

	if err := verifyDevelopmentImageArchive(archive, reference); err != nil {
		return err
	}

	for _, node := range append([]gen.LocalK3sNode{cluster.Server}, cluster.Agents...) {
		if _, err := runner.Run(ctx, input.docker, []string{
			developmentExec, "-i", node.Container, developmentK3s,
			developmentContainerd, developmentNamespaceFlag, developmentContainerdNamespace, developmentContainerdImages, "import", "--platform", "linux/amd64", "--digests", developmentStdin,
		}, archive); err != nil {
			return err
		}
	}

	return nil
}
