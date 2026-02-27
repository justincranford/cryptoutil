// Copyright (c) 2025 Justin Cranford

// Package common provides shared types, constants, and utilities for lint_ports linters.
package common

import (
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ServicePortConfig defines the expected port configuration for a service.
type ServicePortConfig struct {
	Name          string   // Service name (e.g., "sm-im", "jose-ja").
	PublicPorts   []uint16 // Expected public ports (can have multiple for SQLite/PG variants).
	AdminPort     uint16   // Expected admin port (always 9090).
	LegacyPorts   []uint16 // Old ports that should be flagged as violations.
	MagicConstant string   // Magic constant name (e.g., "SMKMSServicePort").
}

// Violation represents a port configuration violation.
type Violation struct {
	File    string
	Line    int
	Content string
	Port    uint16
	Reason  string
}

// HealthViolation represents a health path configuration violation.
type HealthViolation struct {
	File    string
	Line    int
	Content string
	Reason  string
}

// StandardAdminPort is the universal admin port for all services.
const StandardAdminPort = cryptoutilSharedMagic.DefaultPrivatePortCryptoutil

// StandardHealthPath is the expected health check path for all services.
const StandardHealthPath = cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath

// LineSeparatorLength defines the length of line separators in output.
const LineSeparatorLength = 60

// ServicePorts defines the canonical port assignments for all cryptoutil services.
// This is the single source of truth for port assignments.
var ServicePorts = map[string]ServicePortConfig{
	cryptoutilSharedMagic.OTLPServiceSMKMS: {
		Name:          cryptoutilSharedMagic.OTLPServiceSMKMS,
		PublicPorts:   []uint16{cryptoutilSharedMagic.KMSServicePort, cryptoutilSharedMagic.KMSE2EPostgreSQL1PublicPort, cryptoutilSharedMagic.KMSE2EPostgreSQL2PublicPort}, // Base ports for SQLite/PostgreSQL variants
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{cryptoutilSharedMagic.DemoServerPort, 8081, 8082},
		MagicConstant: "KMSServicePort",
	},
	cryptoutilSharedMagic.OTLPServicePKICA: {
		Name:          cryptoutilSharedMagic.OTLPServicePKICA,
		PublicPorts:   []uint16{cryptoutilSharedMagic.PKICAServicePort},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8050, 8443},
		MagicConstant: "PKICAServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceIdentityAuthz: {
		Name:          cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		PublicPorts:   []uint16{cryptoutilSharedMagic.IdentityDefaultAuthZPort},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{cryptoutilSharedMagic.PKICAServicePort, 18000},
		MagicConstant: "IdentityAuthzServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceIdentityIDP: {
		Name:          cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		PublicPorts:   []uint16{cryptoutilSharedMagic.IdentityDefaultIDPPort, 8301}, // 8300 default, 8301 for E2E (avoids conflict with authz)
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8110, 8111, 8112, 18100},
		MagicConstant: "IdentityIdpServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceIdentityRS: {
		Name:          cryptoutilSharedMagic.OTLPServiceIdentityRS,
		PublicPorts:   []uint16{cryptoutilSharedMagic.IdentityDefaultRSPort},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8120, 8121, 8122, cryptoutilSharedMagic.IdentityE2EAuthzPublicPort},
		MagicConstant: "IdentityRsServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceIdentityRP: {
		Name:          cryptoutilSharedMagic.OTLPServiceIdentityRP,
		PublicPorts:   []uint16{cryptoutilSharedMagic.DefaultSPARPPort},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8130, 8131, 8132, cryptoutilSharedMagic.IdentityE2EIDPPublicPort},
		MagicConstant: "IdentityRpServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceIdentitySPA: {
		Name:          cryptoutilSharedMagic.OTLPServiceIdentitySPA,
		PublicPorts:   []uint16{8600},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8140, 8141, 8142, cryptoutilSharedMagic.IdentityE2ERSPublicPort},
		MagicConstant: "IdentitySpaServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceSMIM: {
		Name:          cryptoutilSharedMagic.OTLPServiceSMIM,
		PublicPorts:   []uint16{cryptoutilSharedMagic.IMServicePort, cryptoutilSharedMagic.IME2EPostgreSQL1PublicPort, cryptoutilSharedMagic.IME2EPostgreSQL2PublicPort},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8070, 8071, 8072, cryptoutilSharedMagic.DefaultPublicPortInternalMetrics, cryptoutilSharedMagic.PortOtelCollectorReceivedMetrics, 8890},
		MagicConstant: "IMServicePort",
	},
	cryptoutilSharedMagic.OTLPServiceJoseJA: {
		Name:          cryptoutilSharedMagic.OTLPServiceJoseJA,
		PublicPorts:   []uint16{cryptoutilSharedMagic.JoseJAServicePort},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8060, 9443, 8092},
		MagicConstant: "JoseJAServicePort",
	},
}

// AllLegacyPorts returns all legacy ports that should be flagged as violations.
func AllLegacyPorts() []uint16 {
	var ports []uint16

	seen := make(map[uint16]bool)

	for _, cfg := range ServicePorts {
		for _, port := range cfg.LegacyPorts {
			if !seen[port] {
				seen[port] = true
				ports = append(ports, port)
			}
		}
	}

	return ports
}

// AllValidPublicPorts returns all valid public ports.
func AllValidPublicPorts() []uint16 {
	var ports []uint16

	seen := make(map[uint16]bool)

	for _, cfg := range ServicePorts {
		for _, port := range cfg.PublicPorts {
			if !seen[port] {
				seen[port] = true
				ports = append(ports, port)
			}
		}
	}

	return ports
}

// OtelCollectorPorts are legitimate OpenTelemetry collector ports.
// These are NOT cryptoutil service ports and should be excluded from legacy port checks.
var OtelCollectorPorts = []uint16{
	cryptoutilSharedMagic.DefaultPublicPortInternalMetrics,
	cryptoutilSharedMagic.PortOtelCollectorReceivedMetrics,
}

// IsOtelCollectorPort checks if a port is a legitimate OTEL collector port.
func IsOtelCollectorPort(port uint16) bool {
	for _, p := range OtelCollectorPorts {
		if p == port {
			return true
		}
	}

	return false
}

// IsOtelRelatedFile checks if a file is related to OpenTelemetry configuration.
func IsOtelRelatedFile(filePath string) bool {
	lowerPath := strings.ToLower(filePath)

	return strings.Contains(lowerPath, "otel") ||
		strings.Contains(lowerPath, "opentelemetry") ||
		strings.Contains(lowerPath, "telemetry")
}

// IsOtelRelatedContent checks if a line of code contains OTEL-related terms.
// This catches cases like constant definitions with "Otel" in the name.
func IsOtelRelatedContent(line string) bool {
	lowerLine := strings.ToLower(line)

	return strings.Contains(lowerLine, "otel") ||
		strings.Contains(lowerLine, "opentelemetry") ||
		strings.Contains(lowerLine, "telemetry")
}

// IsComposeFile checks if a file is a Docker Compose file.
func IsComposeFile(filePath string) bool {
	base := filepath.Base(filePath)

	return base == "docker-compose.yml" ||
		base == "docker-compose.yaml" ||
		base == "compose.yml" ||
		base == "compose.yaml" ||
		strings.HasPrefix(base, "compose.") && strings.HasSuffix(base, ".yml") ||
		strings.HasPrefix(base, "compose.") && strings.HasSuffix(base, ".yaml") ||
		strings.HasPrefix(base, "docker-compose.") && strings.HasSuffix(base, ".yml") ||
		strings.HasPrefix(base, "docker-compose.") && strings.HasSuffix(base, ".yaml")
}
