// Copyright (c) 2025-2026 Justin Cranford.

package magic

const (
	// HelmCommandTemplate renders a chart without installing it.
	HelmCommandTemplate = "template"
	// KubernetesPodTemplateField identifies a workload's Pod template specification.
	KubernetesPodTemplateField = "template"
	// CICDCheckovPythonUnix is the pinned virtual environment's Unix interpreter filename.
	CICDCheckovPythonUnix = "python"
	// CICDCheckovPythonWindows is the pinned virtual environment's Windows interpreter filename.
	CICDCheckovPythonWindows = "python.exe"
	// KubernetesSuiteImageDeployment is the current-source image shared by local Helm selections.
	KubernetesSuiteImageDeployment = "k8s"
	// KubernetesSuiteImageBinary is the suite executable whose service route is selected by Helm.
	KubernetesSuiteImageBinary = "/app/cryptoutil"
	// CICDCmdDirCheckov is the cmd/ directory name for Checkov scanning tooling.
	CICDCmdDirCheckov = "cicd-checkov"
	// CICDCmdDirGolangci is the repository-safe golangci-lint command entry point.
	CICDCmdDirGolangci = "cicd-golangci"

	// CICDCmdDirDeployment is the cmd/ directory name for deployment tooling.
	CICDCmdDirDeployment       = "cicd-deployment"
	CICDTemplatesKubernetesDir = "kubernetes/"

	// AdminPrivateKeyFileName is the shared filename for the deployment admin private key.
	AdminPrivateKeyFileName = "admin.key"
	// DeploymentSecretsDirectoryPathSegment is the path segment for a deployment secrets directory.
	DeploymentSecretsDirectoryPathSegment = "/secrets/"
	// DeploymentPoolPG identifies the PostgreSQL deployment backend pool.
	DeploymentPoolPG = "pg"
	// DeploymentPoolMem identifies the memory deployment backend pool.
	DeploymentPoolMem = "mem"
	// DeploymentPoolDisk identifies the disk deployment backend pool.
	DeploymentPoolDisk = "disk"

	// K3SBootstrapAgentServerArgument selects the Compose-local K3s server for agents.
	K3SBootstrapAgentServerArgument = "--server=https://server-0:6443"
	// K3SBootstrapServerEncryptionArgument enables encryption at rest for Kubernetes Secrets.
	K3SBootstrapServerEncryptionArgument = "--secrets-encryption"
	// K3SBootstrapServerEncryptionProviderArgument pins the Secrets-at-rest provider.
	K3SBootstrapServerEncryptionProviderArgument = "--secrets-encryption-provider=aescbc"
	// K3SBootstrapServerURL is the stable Compose-local server endpoint used by agents.
	K3SBootstrapServerURL = "https://server-0:6443"
	// K3SBootstrapTokenArgument is forbidden because token values must never enter CLI args.
	K3SBootstrapTokenArgument = "--token"
	// K3SBootstrapTokenEnvironment is forbidden because the join token must never enter env.
	K3SBootstrapTokenEnvironment = "K3S_TOKEN"
	// K3SBootstrapTokenAssignmentPrefix is forbidden because token values must never enter CLI args.
	K3SBootstrapTokenAssignmentPrefix = "--token="
	// K3SBootstrapTokenFileArgument mounts the protected Docker Secret as a file.
	K3SBootstrapTokenFileArgument = "--token-file=/run/secrets/k3s-join-token"
	// K3SBootstrapTokenSecretFileFormat builds a relative, run-owned Docker Secret source.
	K3SBootstrapTokenSecretFileFormat = "./secrets/%s/k3s-join-token"
	// K3SBootstrapTokenSecretName is the non-secret Docker Secret reference.
	K3SBootstrapTokenSecretName = "k3s-join-token"
	// K3SBootstrapTokenSecretTarget is the read-only Docker Secret mount path.
	K3SBootstrapTokenSecretTarget = "/run/secrets/k3s-join-token"
	// K3SBootstrapURLEnvironment is forbidden because the agent URL belongs in CLI config.
	K3SBootstrapURLEnvironment = "K3S_URL"

	// PostgresDataDirectoryEnvironment redirects PostgreSQL into its retained PVC mount.
	PostgresDataDirectoryEnvironment = "PGDATA"
	// Postgres18ImageDefaultDataDirectory is outside the chart's configured persistent mount.
	Postgres18ImageDefaultDataDirectory = "/var/lib/postgresql/18/docker"
)
