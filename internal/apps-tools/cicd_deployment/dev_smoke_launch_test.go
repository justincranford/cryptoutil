// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	sha256 "crypto/sha256"
	"encoding/hex"
	json "encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type developmentLaunchRunner struct {
	t            *testing.T
	input        developmentSmokeInput
	cluster      gen.LocalK3sPlan
	lock         gen.ToolchainLock
	fail         string
	admissions   int
	launches     int
	readyChecks  int
	cleanupCalls int
	kubeconfig   []byte
	cancel       context.CancelFunc
}

const (
	testLaunchMissingModel    = "launch-missing-model"
	testLaunchMissingLock     = "launch-missing-lock"
	testLaunchMissingDomain   = "launch-missing-domain"
	testLaunchMissingTopology = "launch-missing-topology"
	testLaunchReadiness       = "launch-readiness"
	testLaunchSnapshot        = "launch-snapshot"
	testLaunchImageStream     = "launch-image-stream"
	testLaunchImageReference  = "launch-image-reference"
	testLaunchTokenWrite      = "launch-token-write"
	testLaunchComposeWrite    = "launch-compose-write"
	testLaunchKubeconfigWrite = "launch-kubeconfig-write"
	testLaunchMaterialCancel  = "launch-material-cancel"
	testLaunchResumeSuccess   = "launch-resume-success"
	testLaunchResumeMissing   = "launch-resume-missing"
	testLaunchResumeForeign   = "launch-resume-foreign"
	testLaunchResumeApp       = "launch-resume-application"
	testLaunchResumePreflight = "launch-resume-preflight"
	testLaunchImageRead       = "launch-image-read"
)

func (r *developmentLaunchRunner) Run(ctx context.Context, binary string, args []string, raw []byte) ([]byte, error) {
	if binary == r.input.kubectl {
		require.True(r.t, r.input.resumeBootstrap)
		require.Contains(r.t, args, "namespaces,pods,jobs")

		var objects []map[string]any
		for _, name := range []string{developmentDefaultContext, developmentKubeSystem, developmentKubePublic, developmentKubeNodeLease} {
			objects = append(objects, map[string]any{developmentKind: developmentNamespaceKind, developmentMetadata: map[string]string{developmentName: name}})
		}

		if r.fail == testLaunchResumeApp {
			objects = append(objects, map[string]any{developmentKind: developmentNamespaceKind, developmentMetadata: map[string]string{developmentName: "previous-app"}})
		}

		return json.Marshal(map[string]any{developmentKind: developmentListKind, developmentItems: objects})
	}

	if binary == r.input.helm {
		if r.fail == testDevHelmFailure {
			return nil, errors.New("PRIVATE")
		}

		return []byte(r.lock.Kubernetes.Helm.Version), nil
	}

	require.Equal(r.t, r.input.docker, binary)

	switch args[0] {
	case "info":
		if args[len(args)-1] == "{{.ID}}" {
			return []byte("owned-test-engine"), nil
		}

		r.admissions++
		if r.fail == testDevFirstAdmission || (r.fail == testDevFinalAdmission && r.admissions == 2) {
			return nil, errors.New("PRIVATE")
		}

		return []byte("14 8326418432"), nil
	case developmentImage:
		if r.fail == testLaunchImageRead {
			return nil, errors.New(testDiagnosticPrivate)
		}

		if r.fail == developmentImage {
			return []byte("{}"), nil
		}

		reference := r.lock.Images.Postgres.Reference
		if r.fail == testLaunchImageReference {
			reference = "non-immutable-image-reference"
		}

		return json.Marshal(map[string]any{
			"digests": []string{reference},
			"os":      cryptoutilSharedMagic.MockRuntimeGoOS, "architecture": cryptoutilSharedMagic.MockRuntimeGoArch, developmentLabels: map[string]string{testPKIRunLabel: gen.DevelopmentSmokeRun, "cu.io/owner": "phase2-runtime"},
		})
	case "compose":
		require.Equal(r.t, 2, r.admissions)

		r.launches++
		if r.fail == testDevLaunchFailure {
			return nil, errors.New("PRIVATE")
		}

		return nil, nil
	case developmentExec:
		require.Equal(r.t, "owned-id-"+r.cluster.Server.Container, args[1])

		if args[2] == "cat" {
			if r.fail == testDevKubeconfigRead {
				return nil, errors.New("PRIVATE")
			}

			if r.fail == testDevKubeconfigIdentity {
				return []byte("{}"), nil
			}

			return r.kubeconfig, nil
		}

		r.readyChecks++
		if r.fail == testLaunchReadiness {
			r.cancel()

			return nil, errors.New("PRIVATE")
		}

		return nil, nil
	case "container", developmentNetwork:
		if args[1] == "ls" {
			return nil, nil
		}

		if args[len(args)-2] == developmentStateFormat {
			require.Equal(r.t, "owned-id-"+r.cluster.Server.Container, args[len(args)-1])

			return []byte(`{"Running":false}`), nil
		}

		r.cleanupCalls++

		owned := &developmentCleanupRunner{t: r.t, cluster: r.cluster}
		if r.fail == testLaunchSnapshot {
			owned.failure = testDevWrongRun
		}

		return owned.Run(ctx, binary, args, raw)
	case developmentVolume:
		if r.input.resumeBootstrap {
			var records []string

			for _, node := range append([]gen.LocalK3sNode{r.cluster.Server}, r.cluster.Agents...) {
				if args[1] == developmentInspect && args[len(args)-1] == node.Volume {
					labels := map[string]string{
						testPKIRunLabel: r.cluster.Run, developmentNodeIDLabel: node.Container,
						developmentComposeProject: r.cluster.Project, "com.docker.compose.config-hash": strings.Repeat("a", 64),
					}
					if r.fail == testLaunchResumeForeign {
						labels[testPKIRunLabel] = "other-owner-run"
					}

					return json.Marshal(gen.DevelopmentRetainedVolume{
						Name: node.Volume, Driver: developmentLocalVolume, Scope: developmentLocalVolume, CreatedAt: "2026-11-03T10:00:00Z",
						Mountpoint: path.Join("/var/lib/docker/volumes", node.Volume, "_data"), Labels: labels,
					})
				}

				raw, err := json.Marshal(map[string]string{"Name": node.Volume, "Labels": "cu.io/run=" + r.cluster.Run})
				require.NoError(r.t, err)

				records = append(records, string(raw))
			}

			require.Equal(r.t, "ls", args[1])

			return []byte(strings.Join(records, "\n")), nil
		}

		return nil, nil
	default:
		require.FailNow(r.t, "unexpected test-only transport command", args[0])

		return nil, nil
	}
}

func TestDevelopmentLaunchAdmissionProtectedGenerationAndOwnedCleanup(t *testing.T) {
	t.Parallel()

	for _, failure := range []string{
		testLaunchResumeSuccess, testLaunchResumeMissing, testLaunchResumeForeign, testLaunchResumeApp, testLaunchResumePreflight, testLaunchImageRead,
		"", testDevModelFailure, testDevLockFailure, testDevBinaryFailure, testDevHelmFailure, developmentImage, testDevDomainFailure, testDevInfrastructure,
		testDevFirstAdmission, testDevPrivateCollision, testDevTopologyFailure, testDevFinalAdmission, testDevLaunchFailure, testDevKubeconfigRead, testDevKubeconfigIdentity, testDevDeploymentFailure,
		testLaunchMissingModel, testLaunchMissingLock, testLaunchMissingDomain, testLaunchMissingTopology, testLaunchReadiness, testLaunchSnapshot, testLaunchImageStream, testLaunchImageReference,
		testLaunchTokenWrite, testLaunchComposeWrite, testLaunchKubeconfigWrite, testLaunchMaterialCancel,
	} {
		t.Run("launch="+failure, func(t *testing.T) {
			t.Parallel()

			sourceRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
			require.NoError(t, err)
			root := t.TempDir()

			for _, name := range []string{
				testDevModelPath, testDevLockPath,
				testDevDomainPath, testDevInfrastructurePath,
				testDevTopologyPath,
			} {
				data, err := os.ReadFile(filepath.Join(sourceRoot, filepath.FromSlash(name)))
				require.NoError(t, err)

				target := filepath.Join(root, filepath.FromSlash(name))
				require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o700))
				require.NoError(t, os.WriteFile(target, data, 0o600))
			}

			require.NoError(t, os.Mkdir(filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirTestOutput), 0o700))

			binary := []byte("public test-only binary pin; not runtime certification")
			input := developmentSmokeInput{
				root: root, image: testDevSyntheticImage, docker: filepath.Join(root, "docker.exe"),
				kubectl: filepath.Join(root, "kubectl.exe"), helm: filepath.Join(root, "helm.exe"), timeout: time.Minute,
			}
			input.resumeBootstrap = strings.HasPrefix(failure, "launch-resume-")
			input.preflightOnly = failure == testLaunchResumePreflight
			require.NoError(t, os.WriteFile(input.kubectl, binary, 0o600))

			lockPath := filepath.Join(root, "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-toolchain.lock.json")
			raw, err := os.ReadFile(lockPath)
			require.NoError(t, err)
			lock, err := gen.DecodeToolchainLock(raw)
			require.NoError(t, err)

			hash := sha256.Sum256(binary)
			lock.Kubernetes.Kubectl.WindowsAMD64SHA256 = hex.EncodeToString(hash[:])
			raw, err = json.Marshal(lock)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(lockPath, raw, 0o600))
			raw, err = os.ReadFile(filepath.Join(root, "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-model.json"))
			require.NoError(t, err)
			model, err := gen.DecodeModel(raw)
			require.NoError(t, err)
			plan, err := gen.BuildDevelopmentRuntimePlan(model, lock, gen.DevelopmentSmokeRun, gen.DevelopmentSmokeSelector, lock.Images.Postgres.Reference)
			require.NoError(t, err)
			runner := &developmentLaunchRunner{t: t, input: input, lock: lock, cluster: plan.Cluster, fail: failure}
			runner.kubeconfig = developmentKubeconfigIdentityFixture(t)

			if input.resumeBootstrap && failure != testLaunchResumeMissing {
				topology, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(testDevTopologyPath)))
				require.NoError(t, err)
				compose, err := gen.RenderLocalK3sCompose(topology, plan.Cluster)
				require.NoError(t, err)
				private, err := gen.CreateDevelopmentPrivateDirectory(filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirTestOutput))
				require.NoError(t, err)
				require.NoError(t, private.Write(cryptoutilSharedMagic.COMPOSE_YML, compose))
				require.NoError(t, private.Write("secrets/p2sm1103/k3s-join-token", []byte(strings.Repeat("a", 64))))
				require.NoError(t, private.Close())
			}

			badFiles := map[string]string{
				testDevModelFailure: testDevModelPath, testDevLockFailure: testDevLockPath,
				testDevDomainFailure: testDevDomainPath, testDevInfrastructure: testDevInfrastructurePath,
				testDevTopologyFailure: testDevTopologyPath,
			}
			if name, ok := badFiles[failure]; ok {
				require.NoError(t, os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte("{"), 0o600))
			}

			if failure == testDevBinaryFailure {
				require.NoError(t, os.WriteFile(input.kubectl, []byte("forged"), 0o600))
			}

			if failure == testDevPrivateCollision {
				require.NoError(t, os.Mkdir(filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirTestOutput, "phase2-runtime-p2sm1103"), 0o700))
			}

			missing := map[string]string{
				testLaunchMissingModel: testDevModelPath, testLaunchMissingLock: testDevLockPath,
				testLaunchMissingDomain: testDevDomainPath, testLaunchMissingTopology: testDevTopologyPath,
			}
			if name, ok := missing[failure]; ok {
				require.NoError(t, os.Remove(filepath.Join(root, filepath.FromSlash(name))))
			}

			deploys := 0
			operations := developmentSmokeOperations{
				writeProtected: func(directory *gen.DevelopmentPrivateDirectory, name string, data []byte) error {
					if input.resumeBootstrap {
						require.Equal(t, "kubeconfig", name, "recovery must never overwrite retained token or topology")
					}

					if (failure == testLaunchTokenWrite && strings.HasSuffix(name, cryptoutilSharedMagic.K3SBootstrapTokenSecretName)) ||
						(failure == testLaunchComposeWrite && name == cryptoutilSharedMagic.COMPOSE_YML) ||
						(failure == testLaunchKubeconfigWrite && name == "kubeconfig") {
						return errors.New("PRIVATE")
					}

					return writeDevelopmentSmokeProtected(directory, name, data)
				},
				verifyImages: func(_ context.Context, _ developmentSmokeInput, references []string) error {
					require.Equal(t, []string{lock.Images.Postgres.Reference, lock.Images.OTelCollector.Reference, lock.Images.LGTM.Reference}, references)
					require.Zero(t, runner.launches)

					if failure == testLaunchImageStream {
						return errors.New("PRIVATE")
					}

					if failure == testLaunchMaterialCancel {
						runner.cancel()
					}

					return nil
				},
				listen: func(network, _ string) (net.Listener, error) {
					return net.Listen(network, "127.0.0.1:0")
				},
				deploy: func(_ context.Context, _ developmentSmokeInput, preflight gen.BootstrapPreflightInput, _ gen.LocalK3sPlan,
					request execution.Request, config execution.KubectlConfig, _ execution.Runner, _ io.Writer,
				) error {
					deploys++

					require.Equal(t, 2, runner.admissions)
					require.Equal(t, 1, runner.launches)
					require.Equal(t, len(plan.Lifecycle.Secrets), len(request.Secrets))
					require.Len(t, preflight.RuntimeValues, 1)
					require.NotNil(t, request.VerifyPKI)
					require.FileExists(t, config.Kubeconfig)

					if failure == testDevDeploymentFailure {
						return errors.New("synthetic deployment failure")
					}

					return nil
				},
			}

			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			runner.cancel = cancel

			err = executeDevelopmentSmokeWithOperations(ctx, input, runner, io.Discard, operations)
			if input.preflightOnly {
				require.NoError(t, err)
				require.Zero(t, deploys)
				require.Zero(t, runner.launches)
				require.Zero(t, runner.cleanupCalls)
				require.NoFileExists(t, filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirTestOutput, "phase2-runtime-p2sm1103", "kubeconfig"))
			} else if failure == "" || failure == testLaunchResumeSuccess {
				require.NoError(t, err)
				require.Equal(t, 1, deploys)
				require.Positive(t, runner.cleanupCalls)
			} else {
				require.Error(t, err)
				require.NotContains(t, err.Error(), "PRIVATE")

				if !strings.HasPrefix(failure, "kubeconfig") && failure != testDevDeploymentFailure {
					require.Zero(t, deploys)
				}

				if runner.launches == 0 {
					require.Zero(t, runner.cleanupCalls)
				} else {
					require.Positive(t, runner.cleanupCalls)
				}
			}
		})
	}
}
