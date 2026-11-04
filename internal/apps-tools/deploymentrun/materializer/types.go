// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// DefaultOutputDirectory is the init-container path for the atomically published volume root.
	DefaultOutputDirectory = cryptoutilSharedMagic.DeploymentMaterializerDefaultOutputDirectory

	// HealthClientDirectory is the fixed app-container health client directory.
	HealthClientDirectory = cryptoutilSharedMagic.DeploymentMaterializerHealthClientDirectory
	// OTLPClientDirectory is the fixed app-container OTLP client directory.
	OTLPClientDirectory = cryptoutilSharedMagic.DeploymentMaterializerOTLPClientDirectory

	// SecretLabelWorkloadName labels a Secret with the canonical PSID-pool workload.
	SecretLabelWorkloadName = cryptoutilSharedMagic.DeploymentMaterializerSecretLabelWorkloadName
	// SecretLabelPSID labels a Secret with its PS-ID.
	SecretLabelPSID = cryptoutilSharedMagic.DeploymentMaterializerSecretLabelPSID
	// SecretLabelPool labels a Secret with its pool identity.
	SecretLabelPool = cryptoutilSharedMagic.DeploymentMaterializerSecretLabelPool
	// SecretLabelRun labels a Secret with its run identity.
	SecretLabelRun = cryptoutilSharedMagic.DeploymentMaterializerSecretLabelRun
	// SecretLabelRelease labels a Secret with its release identity.
	SecretLabelRelease = cryptoutilSharedMagic.DeploymentMaterializerSecretLabelRelease
	// SecretLabelRole labels a Secret with its materializer role.
	SecretLabelRole = cryptoutilSharedMagic.DeploymentMaterializerSecretLabelRole

	// SecretRoleTLSBundle is the label value for a per-pod server certificate bundle.
	SecretRoleTLSBundle = cryptoutilSharedMagic.DeploymentMaterializerSecretRoleTLSBundle
	// SecretRoleHealthClient is the label value for a per-pod health client Secret.
	SecretRoleHealthClient = cryptoutilSharedMagic.DeploymentMaterializerSecretRoleHealth
	// SecretRoleOTLPClient is the label value for a per-pod OTLP client Secret.
	SecretRoleOTLPClient = cryptoutilSharedMagic.DeploymentMaterializerSecretRoleOTLP

	// SecretDataPublicCertificate is the TLS bundle key for the public server chain.
	SecretDataPublicCertificate = cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate
	// SecretDataPublicPrivateKey is the TLS bundle key for the public server private key.
	SecretDataPublicPrivateKey = cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey
	// SecretDataAdminCertificate is the TLS bundle key for the admin server chain.
	SecretDataAdminCertificate = cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate
	// SecretDataAdminPrivateKey is the TLS bundle key for the admin server private key.
	SecretDataAdminPrivateKey = cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey
	// SecretDataPostgreSQLCertificate is the TLS bundle key for the PostgreSQL client chain.
	SecretDataPostgreSQLCertificate = cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLCertificate
	// SecretDataPostgreSQLPrivateKey is the TLS bundle key for the PostgreSQL client private key.
	SecretDataPostgreSQLPrivateKey = cryptoutilSharedMagic.DeploymentMaterializerSecretDataPostgreSQLPrivateKey
	// SecretDataCertificate is the key for standalone health and OTLP client certificate chains.
	SecretDataCertificate = cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate
	// SecretDataPrivateKey is the key for standalone health and OTLP client private keys.
	SecretDataPrivateKey = cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey

	// PublicServerCertificateFile is the path relative to the output root for the public chain.
	PublicServerCertificateFile = cryptoutilSharedMagic.DeploymentMaterializerPublicServerCertificateFile
	// PublicServerPrivateKeyFile is the path relative to the output root for the public key.
	PublicServerPrivateKeyFile = cryptoutilSharedMagic.DeploymentMaterializerPublicServerPrivateKeyFile
	// AdminServerCertificateFile is the path relative to the output root for the admin chain.
	AdminServerCertificateFile = cryptoutilSharedMagic.DeploymentMaterializerAdminServerCertificateFile
	// AdminServerPrivateKeyFile is the path relative to the output root for the admin key.
	AdminServerPrivateKeyFile = cryptoutilSharedMagic.DeploymentMaterializerAdminServerPrivateKeyFile
	// HealthClientCertificateFile is the path relative to the output root for the health chain.
	HealthClientCertificateFile = cryptoutilSharedMagic.DeploymentMaterializerHealthClientCertificateFile
	// HealthClientPrivateKeyFile is the path relative to the output root for the health key.
	HealthClientPrivateKeyFile = cryptoutilSharedMagic.DeploymentMaterializerHealthClientPrivateKeyFile
	// OTLPClientCertificateFile is the path relative to the output root for the OTLP chain.
	OTLPClientCertificateFile = cryptoutilSharedMagic.DeploymentMaterializerOTLPClientCertificateFile
	// OTLPClientPrivateKeyFile is the path relative to the output root for the OTLP key.
	OTLPClientPrivateKeyFile = cryptoutilSharedMagic.DeploymentMaterializerOTLPClientPrivateKeyFile
	// PostgreSQLClientCertificateFile is the path relative to the output root for the PG chain.
	PostgreSQLClientCertificateFile = cryptoutilSharedMagic.DeploymentMaterializerPostgreSQLClientCertificateFile
	// PostgreSQLClientPrivateKeyFile is the path relative to the output root for the PG key.
	PostgreSQLClientPrivateKeyFile = cryptoutilSharedMagic.DeploymentMaterializerPostgreSQLClientPrivateKeyFile
)

const (
	defaultOutputDirectory = DefaultOutputDirectory

	labelWorkload = SecretLabelWorkloadName
	labelPSID     = SecretLabelPSID
	labelPool     = SecretLabelPool
	labelRun      = SecretLabelRun
	labelRelease  = SecretLabelRelease
	labelRole     = SecretLabelRole

	roleTLSBundle = SecretRoleTLSBundle
	roleHealth    = SecretRoleHealthClient
	roleOTLP      = SecretRoleOTLPClient
)

// RuntimeInputs are supplied by the controller from the pod Downward API and fixed runtime
// configuration; the namespace is never inferred from a pod name or service account token.
type RuntimeInputs struct {
	PodName          string
	Namespace        string
	APIEndpoint      string
	APIServerName    string
	ExpectedAdminDNS string
	// ExpectedPublicDNSNames must enumerate every required public certificate DNS SAN,
	// including derived service/pod FQDNs and any instance, pooled, or external SNI names.
	ExpectedPublicDNSNames []string
}

// TrustInputs contain explicit CA anchors from the trusted issuer mount. They are never
// obtained by reading a Kubernetes Secret through the init-container API role.
type TrustInputs struct {
	PublicServerPEM          []byte
	AdminServerPEM           []byte
	HealthClientPEM          []byte
	OTLPClientPEM            []byte
	PostgreSQLClientPEM      []byte
	PostgreSQLClientBindings map[string]PostgreSQLClientBinding
}

// PostgreSQLClientBinding binds one exact pod identity to its certificate CN and DB role.
type PostgreSQLClientBinding struct {
	CommonName   string
	DatabaseRole string
}

// HTTPDoer permits deterministic, listener-free Kubernetes API tests.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Dependencies are runtime seams. OutputDirectory must be the fixed emptyDir destination,
// never a value sourced from instance.yaml.
type Dependencies struct {
	HTTPDoer        HTTPDoer
	ReadFile        func(string) ([]byte, error)
	FileSystem      FileSystem
	Now             func() time.Time
	OutputDirectory string
}

// Materializer fetches only the current pod's three declared Secret resources and publishes
// verified role-specific certificate pairs.
type Materializer struct {
	httpDoer        HTTPDoer
	readFile        func(string) ([]byte, error)
	fileSystem      FileSystem
	now             func() time.Time
	outputDirectory string
}

// New constructs a materializer with operating-system dependencies unless test seams are supplied.
func New(dependencies Dependencies) *Materializer {
	if dependencies.ReadFile == nil {
		dependencies.ReadFile = osReadFile
	}

	if dependencies.FileSystem == nil {
		dependencies.FileSystem = operatingSystemFileSystem{}
	}

	if dependencies.Now == nil {
		dependencies.Now = func() time.Time { return time.Now().UTC() }
	}

	if dependencies.OutputDirectory == "" {
		dependencies.OutputDirectory = defaultOutputDirectory
	}

	return &Materializer{
		httpDoer:        dependencies.HTTPDoer,
		readFile:        dependencies.ReadFile,
		fileSystem:      dependencies.FileSystem,
		now:             dependencies.Now,
		outputDirectory: dependencies.OutputDirectory,
	}
}

type materializedFile struct {
	data []byte
}

// FileSystem provides testable directory and atomic-publication operations.
type FileSystem interface {
	Lstat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	MkdirTemp(string, string) (string, error)
	OpenFile(string, int, os.FileMode) (FileHandle, error)
	Chmod(string, os.FileMode) error
	Rename(string, string) error
	RemoveAll(string) error
}

// FileHandle exposes the write, sync, and close operations required before publication.
type FileHandle interface {
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

func contextError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context: %w", err)
	}

	return nil
}
