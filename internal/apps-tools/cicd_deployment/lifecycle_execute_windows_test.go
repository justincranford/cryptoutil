// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	json "encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilDeploymentGen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilLifecycleRun "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const (
	testExecutionAppCommand = "/app/sm-kms"
)

// Only the final execution boundary is injected. The command dispatcher, actual
// file readers, model/registry/schema projections, admission budget and protected
// physical import all execute. All pins/material/commands below are TEST-ONLY.
func TestLifecycleExecutionCLIRealPhysicalAdmissionPipeline(t *testing.T) {
	t.Parallel()

	originalRoot := filepath.Join("..", "..", "..")
	root := t.TempDir()
	modelPath := filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-model.json")
	raw, err := os.ReadFile(filepath.Join(originalRoot, modelPath))
	require.NoError(t, err)
	model, err := cryptoutilDeploymentGen.DecodeModel(raw)
	require.NoError(t, err)

	selector := "s:sm-kms:mem"
	inventory, err := cryptoutilDeploymentGen.Build(model, selector)
	require.NoError(t, err)
	identities, err := cryptoutilDeploymentGen.BuildLifecycleInventory(inventory)
	require.NoError(t, err)

	resources := cryptoutilDeploymentGen.ResourceRequirements{
		Requests: cryptoutilDeploymentGen.ResourceSet{CPU: testPKIRequestCPU, Memory: testPKIRequestMemory},
		Limits:   cryptoutilDeploymentGen.ResourceSet{CPU: testPKILimitCPU, Memory: testPKILimitMemory},
	}
	input := cryptoutilDeploymentGen.LifecycleInput{Version: 1}

	for _, identity := range identities.Identities {
		switch identity.Kind {
		case testPKISecretKind:
			input.Secrets = append(input.Secrets, cryptoutilDeploymentGen.LifecycleSecretSource{
				Namespace: identity.Namespace, Name: identity.Name, Type: testPKIOpaque,
				Files: map[string]string{testPKITLSCertKey: testExecutionMaterialFile, testPKITLSPrivateKey: testExecutionMaterialFile, testPKICAKey: testExecutionMaterialFile, testPKIUsernameKey: testExecutionMaterialFile, testPKIPasswordKey: testExecutionMaterialFile},
			})
		case testPKIJobKind:
			input.Jobs = append(input.Jobs, cryptoutilDeploymentGen.LifecycleJobInput{
				Namespace: identity.Namespace, Name: strings.TrimSuffix(identity.Name, "-"+model.Revision),
				Image:   "registry.example.invalid/test-suite@sha256:" + strings.Repeat("a", 64),
				Command: []string{testPKISuiteCommand}, Args: []string{testPKIInitCommand, testPKIDomainArgument, testPKIOutputArgument},
				Resources: resources, ActiveDeadlineSeconds: 300, BackoffLimit: 1,
				SecretNames: []string{}, WriteSecretNames: []string{}, ConfigMapNames: []string{},
			})
		}
	}

	model.Revision, err = cryptoutilDeploymentGen.LifecycleRevision(input)
	require.NoError(t, err)
	inventory, err = cryptoutilDeploymentGen.Build(model, selector)
	require.NoError(t, err)

	chart := cryptoutilDeploymentGen.ChartValuesInput{
		Images: map[string]cryptoutilDeploymentGen.ContainerImage{
			testPKIServiceID: {
				Reference: "registry.example.invalid/test-app@sha256:" + strings.Repeat("b", 64),
				Command:   []string{testExecutionAppCommand, "serve"}, Drain: []string{testExecutionAppCommand, "drain"},
				Probes: cryptoutilDeploymentGen.ProbeCommands{Startup: []string{testExecutionAppCommand, "startupz"}, Liveness: []string{testExecutionAppCommand, "livez"}, Readiness: []string{testExecutionAppCommand, "readyz"}},
			},
		},
		PKIMaterializer: cryptoutilDeploymentGen.PKIMaterializerImage{
			Reference: "registry.example.invalid/test-init@sha256:" + strings.Repeat("c", 64), Command: []string{testPKISuiteCommand},
			Args: []string{"pki-materialize", "--instance=/etc/cryptoutil/instance/instance.yaml", "--input=/etc/cryptoutil/instance/materializer.json"}, Resources: resources,
		},
		Storage: cryptoutilDeploymentGen.ChartStorage{ClassName: "local-path", Size: "1Gi"},
		Runtime: cryptoutilDeploymentGen.RuntimeChartInputs{Workloads: map[string]cryptoutilDeploymentGen.RuntimeWorkloadInput{}},
	}
	for _, release := range inventory.Releases {
		for _, workload := range release.Workloads {
			chart.Runtime.Workloads[release.Key+"/"+workload.PSID+"/"+workload.Pool] = cryptoutilDeploymentGen.RuntimeWorkloadInput{
				MaterializerConfig:  executionMaterializerFixture(release, workload),
				FrameworkConfigFile: testPKIFrameworkFile, FrameworkConfig: "# test-only framework\n",
				DomainConfigFile: testPKIDomainFile, DomainConfig: "# test-only domain\n",
				Args: []string{"--backend=" + workload.Pool}, Resources: resources,
			}
		}
	}

	writeJSON := func(name string, value any) {
		raw, err := json.Marshal(value)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o700))
		require.NoError(t, os.WriteFile(name, raw, 0o600))
	}
	writeJSON(filepath.Join(root, modelPath), model)
	inputFile := filepath.Join(root, "operator.json")
	writeJSON(inputFile, input)

	runtimeFile := filepath.Join(root, "runtime.json")
	writeJSON(runtimeFile, chart)

	for _, name := range []string{
		filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-toolchain.lock.json"),
		filepath.Join("api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-chart-values-input.schema.json"),
		filepath.Join("api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-infra", "values.yaml"),
	} {
		content, err := os.ReadFile(filepath.Join(originalRoot, name))
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(root, name), content, 0o600))
	}

	secretRoot := filepath.Join(root, "protected")
	require.NoError(t, os.Mkdir(secretRoot, 0o700))

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	require.NoError(t, err)
	descriptor, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")")
	require.NoError(t, err)
	acl, _, err := descriptor.DACL()
	require.NoError(t, err)
	require.NoError(t, windows.SetNamedSecurityInfo(secretRoot, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil))
	require.NoError(t, os.WriteFile(filepath.Join(secretRoot, testExecutionMaterialFile), []byte(testExecutionPrivate), 0o600))

	_, err = loadLifecycleExecutionInput(root, selector, runtimeFile, os.ReadFile)
	require.NoError(t, err)

	executions := 0
	process := lifecycleExecutionProcess(func(_ context.Context, request cryptoutilLifecycleRun.Request, _ cryptoutilLifecycleRun.KubectlConfig, _ cryptoutilLifecycleRun.Policy) (cryptoutilLifecycleRun.Result, error) {
		executions++

		require.Equal(t, model.Revision, request.Model.Revision)
		require.Len(t, request.Secrets, len(input.Secrets))

		for _, secret := range request.Secrets {
			for _, data := range secret.Data {
				require.Equal(t, []byte(testExecutionPrivate), data)
			}
		}

		return cryptoutilLifecycleRun.Result{Secrets: len(request.Secrets), Jobs: len(request.Input.Jobs)}, nil
	})
	// This test owns only the protected importer/admission boundary. Explicitly
	// inject the later PKI verifier rather than accepting fixture generation.
	process.verifyPKI = func(_ cryptoutilDeploymentGen.BootstrapPreflightInput, request cryptoutilLifecycleRun.Request, _ string) (cryptoutilLifecycleRun.Request, error) {
		return request, nil
	}
	args, _ := executionCLIArgs(t)
	args = append(args, "-input", inputFile, "-runtime-input", runtimeFile, "-select", selector, "-repository-root", root, "-secret-root", secretRoot)

	var stdout, stderr bytes.Buffer

	command := append([]string{lifecycleExecuteCommand}, args...)
	code := mainWithLifecycle(command, &stdout, &stderr, func(args []string, stdout, stderr io.Writer) int {
		return lifecycleExecuteMainWithIO(args, stdout, stderr, process)
	})
	require.Zero(t, code, stderr.String())
	require.Equal(t, 1, executions)
	require.NotContains(t, stdout.String()+stderr.String(), testExecutionPrivate)
	// Same physical inputs cannot bypass a mismatched current contract revision.
	model.Revision = "r0"
	writeJSON(filepath.Join(root, modelPath), model)
	stdout.Reset()
	stderr.Reset()
	code = mainWithLifecycle(command, &stdout, &stderr, func(args []string, stdout, stderr io.Writer) int {
		return lifecycleExecuteMainWithIO(args, stdout, stderr, process)
	})
	require.Equal(t, 1, code)
	require.Equal(t, 1, executions)
	require.Contains(t, stderr.String(), "budget rejected")
	require.Empty(t, stdout.String())
}
