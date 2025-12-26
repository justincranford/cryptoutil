// Copyright (c) 2025 Justin Cranford

package server_test

import (
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

var (
	testServerSettings *cryptoutilConfig.ServerSettings
	testPublicTLS      *cryptoutilTLSGenerator.TLSGeneratedSettings
	testPrivateTLS     *cryptoutilTLSGenerator.TLSGeneratedSettings
)

func TestMain(m *testing.M) {
	// Create shared ServerSettings fixture for tests (port 0 for dynamic allocation).
	testServerSettings = &cryptoutilConfig.ServerSettings{
		BindPublicProtocol:          "https",
		BindPublicAddress:           cryptoutilMagic.IPv4Loopback,
		BindPublicPort:              0,
		BindPrivateProtocol:         "https",
		BindPrivateAddress:          cryptoutilMagic.IPv4Loopback,
		BindPrivatePort:             0,
		PublicBrowserAPIContextPath: "/browser",
		PublicServiceAPIContextPath: "/service",
		PrivateAdminAPIContextPath:  "/admin",
		TLSPublicDNSNames:           []string{"localhost"},
		TLSPublicIPAddresses:        []string{cryptoutilMagic.IPv4Loopback},
		TLSPrivateDNSNames:          []string{"localhost"},
		TLSPrivateIPAddresses:       []string{cryptoutilMagic.IPv4Loopback},
		TLSPublicMode:               cryptoutilConfig.TLSModeAuto,
		TLSPrivateMode:              cryptoutilConfig.TLSModeAuto,
	}

	// Generate shared TLS fixtures for tests (auto-mode, localhost/IPs).
	var err error

	testPublicTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{cryptoutilMagic.IPv4Loopback}, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		panic("failed to generate public TLS fixtures: " + err.Error())
	}

	testPrivateTLS, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings([]string{"localhost"}, []string{cryptoutilMagic.IPv4Loopback}, cryptoutilMagic.TLSTestEndEntityCertValidity1Year)
	if err != nil {
		panic("failed to generate private TLS fixtures: " + err.Error())
	}

	os.Exit(m.Run())
}
