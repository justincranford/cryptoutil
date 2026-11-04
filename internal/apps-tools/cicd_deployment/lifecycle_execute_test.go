// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	cryptoutilDeploymentGen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilLifecycleRun "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testExecutionPrivate       = "private material must not be printed"
	testExecutionInputFlag     = "-input"
	testExecutionContextFlag   = "-context"
	testExecutionAdmission     = "admit"
	testExecutionDecode        = "decode"
	testExecutionSelectedRun   = "selected"
	testExecutionRootFlag      = "-repository-root"
	testExecutionOperationFlag = "-operation-timeout"
	testExecutionStage         = "execute"
	testExecutionRuntimeFlag   = "-runtime-input"
	testExecutionPollFlag      = "-poll-interval"
	testExecutionTimeoutFlag   = "-timeout"
	testExecutionImportStage   = "import"
	testExecutionReadStage     = cryptoutilSharedMagic.ScopeRead
	testExecutionLoadStage     = cryptoutilSharedMagic.WorkflowNameLoad
	testExecutionContext       = "explicit-context"
	testExecutionMaterialFile  = "material"
	testPKISecretKind          = cryptoutilSharedMagic.KubernetesSecretKind
	testPKIOpaque              = cryptoutilSharedMagic.KubernetesSecretTypeOpaque
	testPKIJobKind             = "Job"
	testPKICAKey               = "ca.crt"
	testPKISuiteCommand        = cryptoutilSharedMagic.KubernetesSuiteImageBinary
	testPKIDomainArgument      = "--domain=cryptoutil"
	testPKIFrameworkFile       = "framework.yml"
	testPKIDomainFile          = "domain.yml"
	testPKIServiceID           = cryptoutilSharedMagic.IMJWTIssuer
	testPKIRequestCPU          = "25m"
	testPKILimitCPU            = "100m"
	testPKIPublicationFile     = "publication.json"
	testPKIRequestMemory       = "32Mi"
	testPKILimitMemory         = "64Mi"
	testPKIPasswordKey         = "password"
	testPKIOutputArgument      = "--output-dir=/tmp/pki"
	testPKIHelpFlag            = "-h"
	testPKITLSCertKey          = cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate
	testPKITLSPrivateKey       = cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey
	testPKIUsernameKey         = "username"
	testPKIInitCommand         = cryptoutilSharedMagic.PSIDPKIInit
	testDatabaseManifestStage  = "database-manifest"
)

const testExecutionPublicationFlag = "-publication-input"

func executionCLIArgs(t *testing.T) ([]string, string) {
	t.Helper()

	root := t.TempDir()

	return []string{
		testExecutionInputFlag, "operator.json", testExecutionRuntimeFlag, "runtime.json", testExecutionRootFlag, root,
		testExecutionPublicationFlag, testPKIPublicationFile,
		"-secret-root", filepath.Join(root, "protected"), "-kubectl", filepath.Join(root, "kubectl.exe"),
		"-helm", filepath.Join(root, "helm.exe"), "-kubeconfig", filepath.Join(root, "kubeconfig"),
		testExecutionContextFlag, testExecutionContext, "-admission-certified", testExecutionTimeoutFlag, "120s", testExecutionOperationFlag, cryptoutilSharedMagic.CICDTemplateHealthcheckTimeout, testExecutionPollFlag, "1s",
	}, root
}

func TestLifecycleExecutionCLIExplicitFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		code int
	}{
		{"execution help", []string{testPKIHelpFlag}, 0},
		{"unknown execution flag", []string{"-execution-unknown"}, 2},
		{"missing prior admission", []string{"-admission-certified=false"}, 2},
		{"relative target", []string{testExecutionRootFlag, "relative"}, 2},
		{"empty context", []string{testExecutionContextFlag, ""}, 2},
		{"missing runtime", []string{testExecutionRuntimeFlag, ""}, 2},
		{"missing certificate publication policy", []string{testExecutionPublicationFlag, ""}, 2},
		{"missing operator input", []string{testExecutionInputFlag, ""}, 2},
		{"zero deadline", []string{testExecutionTimeoutFlag, "0"}, 2},
		{"excessive deadline", []string{testExecutionTimeoutFlag, "25h"}, 2},
		{"operation exceeds overall", []string{testExecutionOperationFlag, "3m"}, 2},
		{"excessive operation", []string{testExecutionOperationFlag, "2h"}, 2},
		{"zero operation", []string{testExecutionOperationFlag, "0"}, 2},
		{"zero poll", []string{testExecutionPollFlag, "0"}, 2},
		{"excessive poll", []string{testExecutionPollFlag, "2m"}, 2},
		{"positional execution argument", []string{"trailing"}, 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			args, _ := executionCLIArgs(t)

			var stdout, stderr bytes.Buffer

			require.Equal(t, test.code, lifecycleExecuteMainWithIO(append(args, test.args...), &stdout, &stderr, lifecycleExecutionIO{}))
		})
	}

	var stdout, stderr bytes.Buffer
	require.Zero(t, Main([]string{lifecycleExecuteCommand, "-help"}, nil, &stdout, &stderr))
	require.Contains(t, stderr.String(), "admission-certified")
	require.Equal(t, 2, Main([]string{lifecycleExecuteCommand}, nil, &stdout, &stderr))
}

func TestLifecycleExecutionCLIOrdersBudgetsBeforeMaterialAndMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		failStage  string
		failOutput bool
		code       int
		expected   []string
	}{
		{"guarded execution", "", false, 0, []string{testExecutionLoadStage, testExecutionReadStage, testExecutionAdmission, testExecutionImportStage, testExecutionStage}},
		{"projection failure", testExecutionLoadStage, false, 1, []string{testExecutionLoadStage}},
		{"operator read failure", testExecutionReadStage, false, 1, []string{testExecutionLoadStage, testExecutionReadStage}},
		{"operator decode failure", testExecutionDecode, false, 1, []string{testExecutionLoadStage, testExecutionReadStage}},
		{"current quota failure", testExecutionAdmission, false, 1, []string{testExecutionLoadStage, testExecutionReadStage, testExecutionAdmission}},
		{"protected import failure", testExecutionImportStage, false, 1, []string{testExecutionLoadStage, testExecutionReadStage, testExecutionAdmission, testExecutionImportStage}},
		{"database manifest rejected before transport", testDatabaseManifestStage, false, 1, []string{testExecutionLoadStage, testExecutionReadStage, testExecutionAdmission, testExecutionImportStage}},
		{"mutation failure sanitized", testExecutionStage, false, 1, []string{testExecutionLoadStage, testExecutionReadStage, testExecutionAdmission, testExecutionImportStage, testExecutionStage}},
		{"output failure sanitized", "", true, 1, []string{testExecutionLoadStage, testExecutionReadStage, testExecutionAdmission, testExecutionImportStage, testExecutionStage}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			args, root := executionCLIArgs(t)
			calls := make([]string, 0)

			operator := cryptoutilDeploymentGen.LifecycleInput{Version: 1}
			if test.failStage == testDatabaseManifestStage {
				operator.Jobs = []cryptoutilDeploymentGen.LifecycleJobInput{{Name: "db-init", Namespace: "owned-db"}}
			}

			raw, err := json.Marshal(operator)
			require.NoError(t, err)

			input := cryptoutilDeploymentGen.BootstrapPreflightInput{
				Model:     cryptoutilDeploymentGen.Model{Run: testExecutionSelectedRun},
				Inventory: cryptoutilDeploymentGen.Inventory{Run: testExecutionSelectedRun},
			}
			private := []cryptoutilDeploymentGen.LifecycleSecret{{Data: map[string][]byte{"value": []byte(testExecutionPrivate)}}}
			fail := errors.New(testExecutionPrivate)
			process := lifecycleExecutionIO{
				verifyPKI: func(_ cryptoutilDeploymentGen.BootstrapPreflightInput, request cryptoutilLifecycleRun.Request, _ string) (cryptoutilLifecycleRun.Request, error) {
					return request, nil
				},
				load: func(gotRoot, selector, runtimePath string, _ func(string) ([]byte, error)) (cryptoutilDeploymentGen.BootstrapPreflightInput, error) {
					calls = append(calls, testExecutionLoadStage)

					require.Equal(t, root, gotRoot)
					require.Empty(t, selector)
					require.Equal(t, "runtime.json", runtimePath)

					if test.failStage == testExecutionLoadStage {
						return cryptoutilDeploymentGen.BootstrapPreflightInput{}, fail
					}

					return input, nil
				},
				read: func(name string) ([]byte, error) {
					calls = append(calls, testExecutionReadStage)

					require.Equal(t, "operator.json", name)

					if test.failStage == testExecutionReadStage {
						return nil, fail
					}

					if test.failStage == testExecutionDecode {
						return []byte("{"), nil
					}

					return raw, nil
				},
				admit: func(got cryptoutilDeploymentGen.BootstrapPreflightInput) (cryptoutilDeploymentGen.CapacityEstimate, error) {
					calls = append(calls, testExecutionAdmission)

					require.Equal(t, operator, *got.Lifecycle)

					if test.failStage == testExecutionAdmission {
						return cryptoutilDeploymentGen.CapacityEstimate{}, fail
					}

					return cryptoutilDeploymentGen.CapacityEstimate{}, nil
				},
				importSecrets: func(model cryptoutilDeploymentGen.Model, inventory cryptoutilDeploymentGen.Inventory, got cryptoutilDeploymentGen.LifecycleInput, directory string) ([]cryptoutilDeploymentGen.LifecycleSecret, error) {
					calls = append(calls, testExecutionImportStage)

					require.Equal(t, input.Model, model)
					require.Equal(t, input.Inventory, inventory)
					require.Equal(t, operator, got)
					require.Equal(t, filepath.Join(root, "protected"), directory)

					if test.failStage == testExecutionImportStage {
						return nil, fail
					}

					return private, nil
				},
				execute: func(ctx context.Context, request cryptoutilLifecycleRun.Request, config cryptoutilLifecycleRun.KubectlConfig, policy cryptoutilLifecycleRun.Policy) (cryptoutilLifecycleRun.Result, error) {
					calls = append(calls, testExecutionStage)
					_, bounded := ctx.Deadline()
					require.True(t, bounded)
					require.Equal(t, private, request.Secrets)
					require.Equal(t, operator, request.Input)
					require.Equal(t, root, config.RepositoryRoot)
					require.Equal(t, testExecutionContext, config.Context)
					require.Equal(t, time.Second, config.PollInterval)
					require.Equal(t, 2*time.Minute, policy.MaxDuration)
					require.Equal(t, 10*time.Second, policy.OperationTimeout)

					if test.failStage == testExecutionStage {
						return cryptoutilLifecycleRun.Result{}, fail
					}

					return cryptoutilLifecycleRun.Result{Secrets: 1, Jobs: 2}, nil
				},
			}

			var (
				stdout, stderr bytes.Buffer
				output         io.Writer = &stdout
			)
			if test.failOutput {
				output = executionPrivateWriter{}
			}

			require.Equal(t, test.code, lifecycleExecuteMainWithIO(args, output, &stderr, process), stderr.String())
			require.Equal(t, test.expected, calls)
			require.NotContains(t, stdout.String()+stderr.String(), testExecutionPrivate)

			if test.code == 0 {
				require.Contains(t, stdout.String(), "secrets=1 jobs=2")
				require.Contains(t, stdout.String(), "runtime acceptance remains outstanding")
			}
		})
	}
}

type executionPrivateWriter struct{}

func (executionPrivateWriter) Write([]byte) (int, error) { return 0, errors.New(testExecutionPrivate) }

func TestLifecycleExecutionSafeFailureCategories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		want string
	}{
		{context.DeadlineExceeded, "deadline"},
		{context.Canceled, "canceled"},
		{cryptoutilLifecycleRun.ErrCollision, "collision"},
		{cryptoutilLifecycleRun.ErrPrerequisite, "prerequisites"},
		{cryptoutilLifecycleRun.ErrJobFailed, "Job failed"},
		{cryptoutilLifecycleRun.ErrReplaced, "replaced"},
		{cryptoutilLifecycleRun.ErrTransport, "transport"},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			t.Parallel()

			got := lifecycleExecutionFailureCategory(fmt.Errorf("%w: %s", test.err, testExecutionPrivate))
			require.Contains(t, got, test.want)
			require.NotContains(t, got, testExecutionPrivate)
		})
	}

	require.Equal(t, 1, executionDiagnostic(executionPrivateWriter{}, 2, "safe diagnostic"))
}

func TestLifecycleExecutionRootedLoaderAndNativeConstructor(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	read := func(name string) ([]byte, error) {
		require.Equal(t, filepath.Join(root, "relative-model.json"), name)

		return []byte("{}"), nil
	}
	_, err := loadLifecycleExecutionInputWith(root, "selector", "runtime.json", read,
		func(selector, run, inputPath string, rootedRead func(string) ([]byte, error)) (cryptoutilDeploymentGen.BootstrapPreflightInput, error) {
			require.Equal(t, "selector", selector)
			require.Empty(t, run)
			require.Equal(t, "runtime.json", inputPath)

			_, err := rootedRead("relative-model.json")
			require.NoError(t, err)
			_, err = rootedRead(filepath.Join(root, "relative-model.json"))
			require.NoError(t, err)

			return cryptoutilDeploymentGen.BootstrapPreflightInput{}, nil
		})
	require.NoError(t, err)
	_, err = loadLifecycleExecutionInput(root, "", "runtime.json", func(string) ([]byte, error) {
		return nil, errors.New("input unavailable")
	})
	require.Error(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	_, err = executeLifecycleRequest(ctx, cryptoutilLifecycleRun.Request{}, cryptoutilLifecycleRun.KubectlConfig{}, cryptoutilLifecycleRun.Policy{})
	require.Error(t, err)
	_, err = executeLifecycleRequest(ctx, cryptoutilLifecycleRun.Request{}, cryptoutilLifecycleRun.KubectlConfig{
		Kubectl: filepath.Join(root, "kubectl.exe"), Helm: filepath.Join(root, "helm.exe"), RepositoryRoot: root,
		Kubeconfig: filepath.Join(root, "kubeconfig"), Context: testExecutionContext, PollInterval: time.Second,
	}, cryptoutilLifecycleRun.Policy{MaxDuration: time.Minute, OperationTimeout: time.Second})
	require.Error(t, err, "invalid request must fail before any executable runs")
}
