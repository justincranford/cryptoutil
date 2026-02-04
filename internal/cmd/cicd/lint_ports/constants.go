// Copyright (c) 2025 Justin Cranford

// Package lint_ports provides port validation for cryptoutil services.
package lint_ports

// ServicePortConfig defines the expected port configuration for a service.
type ServicePortConfig struct {
	Name          string   // Service name (e.g., "cipher-im", "jose-ja").
	PublicPorts   []uint16 // Expected public ports (can have multiple for SQLite/PG variants).
	AdminPort     uint16   // Expected admin port (always 9090).
	LegacyPorts   []uint16 // Old ports that should be flagged as violations.
	MagicConstant string   // Magic constant name (e.g., "CipherServicePort").
}

// StandardAdminPort is the universal admin port for all services.
const StandardAdminPort uint16 = 9090

// ServicePorts defines the canonical port assignments for all cryptoutil services.
// This is the single source of truth for port assignments.
var ServicePorts = map[string]ServicePortConfig{
	"cipher-im": {
		Name:          "cipher-im",
		PublicPorts:   []uint16{8070, 8071, 8072},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8888, 8889, 8890},
		MagicConstant: "CipherServicePort",
	},
	"jose-ja": {
		Name:          "jose-ja",
		PublicPorts:   []uint16{8060},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{9443, 8092},
		MagicConstant: "JoseJAServicePort",
	},
	"pki-ca": {
		Name:          "pki-ca",
		PublicPorts:   []uint16{8050},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{8443},
		MagicConstant: "PKICAServicePort",
	},
	"sm-kms": {
		Name:          "sm-kms",
		PublicPorts:   []uint16{8080, 8081, 8082},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{},
		MagicConstant: "KMSServicePort",
	},
	"identity-authz": {
		Name:          "identity-authz",
		PublicPorts:   []uint16{18000},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{},
		MagicConstant: "IdentityAuthzServicePort",
	},
	"identity-idp": {
		Name:          "identity-idp",
		PublicPorts:   []uint16{18100},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{},
		MagicConstant: "IdentityIdpServicePort",
	},
	"identity-rs": {
		Name:          "identity-rs",
		PublicPorts:   []uint16{18200},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{},
		MagicConstant: "IdentityRsServicePort",
	},
	"identity-rp": {
		Name:          "identity-rp",
		PublicPorts:   []uint16{18300},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{},
		MagicConstant: "IdentityRpServicePort",
	},
	"identity-spa": {
		Name:          "identity-spa",
		PublicPorts:   []uint16{18400},
		AdminPort:     StandardAdminPort,
		LegacyPorts:   []uint16{},
		MagicConstant: "IdentitySpaServicePort",
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
	8888,
	8889,
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
