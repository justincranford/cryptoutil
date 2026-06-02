// Copyright (c) 2025-2026 Justin Cranford.
//

//go:build e2e

// Package test_orch_e2e owns reusable E2E lifecycle orchestration for PS-ID stacks.
package test_orch_e2e

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// App variant port offsets relative to the PS-ID base public port.
	variantSQLiteOnePortOffset   = 0
	variantSQLiteTwoPortOffset   = 1
	variantPostgresOnePortOffset = 2
	variantPostgresTwoPortOffset = 3
)

// TLSAppVariantSpec describes one app variant in a PS-ID compose stack.
type TLSAppVariantSpec struct {
	Name           string
	PublicPort     int
	ServerCertCN   string
	ClientCertPath string
	ClientKeyPath  string
}

// TLSPSIDSpec contains all PS-ID specific inputs needed by TLS E2E orchestration tests.
type TLSPSIDSpec struct {
	PSID string

	ComposeFile         string
	ComposeOverrideFile string

	PublicCACertPath     string
	AppHealthEndpoint    string
	PKIInitServiceName   string
	OTelServiceName      string
	GrafanaServiceName   string
	OTelGRPCPort         int
	OTelHTTPPort         int
	OTelHealthPort       int
	OTelServerCertCN     string
	OTelClientCertPath   string
	OTelClientKeyPath    string
	GrafanaUIPort        int
	GrafanaOTLPGRPCPort  int
	GrafanaServerCertCN  string
	GrafanaInfraCertPath string
	GrafanaInfraKeyPath  string

	AppVariants []TLSAppVariantSpec
}

// OTelHealthURL returns the host URL for OTel collector health checks.
func (spec TLSPSIDSpec) OTelHealthURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/", spec.OTelHealthPort)
}

// StartupServices returns the compose services needed by TLS E2E tests.
func (spec TLSPSIDSpec) StartupServices() []string {
	services := make([]string, 0, 3+len(spec.AppVariants))
	services = append(services, spec.PKIInitServiceName, spec.OTelServiceName, spec.GrafanaServiceName)

	for _, variant := range spec.AppVariants {
		services = append(services, variant.Name)
	}

	return services
}

var psidPublicBasePorts = map[string]int{
	cryptoutilSharedMagic.OTLPServiceSMKMS:            cryptoutilSharedMagic.KMSServicePort,
	cryptoutilSharedMagic.OTLPServicePKICA:            cryptoutilSharedMagic.PKICAServicePort,
	cryptoutilSharedMagic.OTLPServiceIdentityAuthz:    int(cryptoutilSharedMagic.IdentityAuthzServicePort),
	cryptoutilSharedMagic.OTLPServiceIdentityIDP:      int(cryptoutilSharedMagic.IdentityIDPServicePort),
	cryptoutilSharedMagic.OTLPServiceIdentityRS:       int(cryptoutilSharedMagic.IdentityRSServicePort),
	cryptoutilSharedMagic.OTLPServiceIdentityRP:       int(cryptoutilSharedMagic.IdentityRPServicePort),
	cryptoutilSharedMagic.OTLPServiceIdentitySPA:      int(cryptoutilSharedMagic.IdentitySPAServicePort),
	cryptoutilSharedMagic.OTLPServiceSkeletonTemplate: cryptoutilSharedMagic.SkeletonTemplateServicePort,
}

// SupportedTLSPSIDs returns the ordered PS-ID list supported by TLS E2E orchestration.
func SupportedTLSPSIDs() []string {
	return append([]string(nil), cryptoutilSharedMagic.AllPSIDs...)
}

// NewTLSPSIDSpec builds a PS-ID specific TLS E2E orchestration specification.
func NewTLSPSIDSpec(psid string) (TLSPSIDSpec, error) {
	basePublicPort, ok := psidPublicBasePorts[psid]
	if !ok {
		return TLSPSIDSpec{}, fmt.Errorf("unsupported TLS E2E PS-ID %q (supported: %s)", psid, strings.Join(SupportedTLSPSIDs(), ", "))
	}

	projectRootPath, err := resolveProjectRootPath()
	if err != nil {
		return TLSPSIDSpec{}, err
	}

	deploymentDir := filepath.Join(projectRootPath, "deployments", psid)
	certRootDir := filepath.Join(deploymentDir, "certs", psid)

	sqliteOneClientCertPath, sqliteOneClientKeyPath := serviceUserCertAndKeyPath(certRootDir, psid, "sqlite-1-serviceuser-db")
	sqliteTwoClientCertPath, sqliteTwoClientKeyPath := serviceUserCertAndKeyPath(certRootDir, psid, "sqlite-2-serviceuser-db")
	postgresClientCertPath, postgresClientKeyPath := serviceUserCertAndKeyPath(certRootDir, psid, "postgres-serviceuser-db")
	otelClientCertPath, otelClientKeyPath := otelClientCertAndKeyPath(certRootDir, psid)
	grafanaInfraCertPath, grafanaInfraKeyPath := grafanaInfraCertAndKeyPath(certRootDir)

	appVariants := []TLSAppVariantSpec{
		{
			Name:           fmt.Sprintf("%s-app-sqlite-1", psid),
			PublicPort:     basePublicPort + variantSQLiteOnePortOffset,
			ServerCertCN:   fmt.Sprintf("public-https-server-entity-%s-sqlite-1", psid),
			ClientCertPath: sqliteOneClientCertPath,
			ClientKeyPath:  sqliteOneClientKeyPath,
		},
		{
			Name:           fmt.Sprintf("%s-app-sqlite-2", psid),
			PublicPort:     basePublicPort + variantSQLiteTwoPortOffset,
			ServerCertCN:   fmt.Sprintf("public-https-server-entity-%s-sqlite-2", psid),
			ClientCertPath: sqliteTwoClientCertPath,
			ClientKeyPath:  sqliteTwoClientKeyPath,
		},
		{
			Name:           fmt.Sprintf("%s-app-postgresql-1", psid),
			PublicPort:     basePublicPort + variantPostgresOnePortOffset,
			ServerCertCN:   fmt.Sprintf("public-https-server-entity-%s-postgres-1", psid),
			ClientCertPath: postgresClientCertPath,
			ClientKeyPath:  postgresClientKeyPath,
		},
		{
			Name:           fmt.Sprintf("%s-app-postgresql-2", psid),
			PublicPort:     basePublicPort + variantPostgresTwoPortOffset,
			ServerCertCN:   fmt.Sprintf("public-https-server-entity-%s-postgres-2", psid),
			ClientCertPath: postgresClientCertPath,
			ClientKeyPath:  postgresClientKeyPath,
		},
	}

	publicCACertPath := filepath.Join(deploymentDir, "certs", cryptoutilSharedMagic.PKIInitAdminCABundleFilename)

	return TLSPSIDSpec{
		PSID: psid,

		ComposeFile:         filepath.Join(deploymentDir, cryptoutilSharedMagic.COMPOSE_YML),
		ComposeOverrideFile: filepath.Join(deploymentDir, "compose-test-otel-expose.yml"),

		PublicCACertPath:     publicCACertPath,
		AppHealthEndpoint:    cryptoutilSharedMagic.KMSE2EHealthEndpoint,
		PKIInitServiceName:   cryptoutilSharedMagic.PSIDPKIInit,
		OTelServiceName:      cryptoutilSharedMagic.OtelTLSE2EContainer,
		GrafanaServiceName:   cryptoutilSharedMagic.GrafanaTLSE2EContainer,
		OTelGRPCPort:         cryptoutilSharedMagic.OtelTLSE2EGRPCPort,
		OTelHTTPPort:         cryptoutilSharedMagic.OtelTLSE2EHTTPPort,
		OTelHealthPort:       cryptoutilSharedMagic.OtelTLSE2EHealthPort,
		OTelServerCertCN:     cryptoutilSharedMagic.OtelTLSE2EOtelServerCertCN,
		OTelClientCertPath:   otelClientCertPath,
		OTelClientKeyPath:    otelClientKeyPath,
		GrafanaUIPort:        cryptoutilSharedMagic.GrafanaTLSE2EUIPort,
		GrafanaOTLPGRPCPort:  cryptoutilSharedMagic.GrafanaTLSE2EOTLPGRPCPort,
		GrafanaServerCertCN:  cryptoutilSharedMagic.GrafanaTLSE2EServerCertCN,
		GrafanaInfraCertPath: grafanaInfraCertPath,
		GrafanaInfraKeyPath:  grafanaInfraKeyPath,
		AppVariants:          appVariants,
	}, nil
}

func resolveProjectRootPath() (string, error) {
	_, sourceFilePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve project root path from runtime caller")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(sourceFilePath), "..", "..", "..", "..")), nil
}

func serviceUserCertAndKeyPath(certRootDir, psid, suffix string) (string, string) {
	entityName := fmt.Sprintf("public-https-client-entity-%s-%s", psid, suffix)
	entityDir := filepath.Join(certRootDir, entityName)

	return filepath.Join(entityDir, entityName+".crt"), filepath.Join(entityDir, entityName+".key")
}

func otelClientCertAndKeyPath(certRootDir, psid string) (string, string) {
	entityName := fmt.Sprintf("otel-collector-contrib-https-client-entity-%s-sqlite-1", psid)
	entityDir := filepath.Join(certRootDir, entityName)

	return filepath.Join(entityDir, entityName+".crt"), filepath.Join(entityDir, entityName+".key")
}

func grafanaInfraCertAndKeyPath(certRootDir string) (string, string) {
	entityName := "otel-collector-contrib-https-client-entity-infra"
	entityDir := filepath.Join(certRootDir, entityName)

	return filepath.Join(entityDir, entityName+".crt"), filepath.Join(entityDir, entityName+".key")
}
