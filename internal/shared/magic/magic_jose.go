// Copyright (c) 2025 Justin Cranford

// Package magic provides domain-specific constants and configuration values.
package magic

// JOSE-JA Service Configuration.
const (
	// OTLPServiceJoseJA is the OTLP service name for jose-ja telemetry.
	OTLPServiceJoseJA = "jose-ja"

	// JoseJAServicePort is the default public API port for jose-ja service.
	JoseJAServicePort = 9443

	// JoseJAAdminPort is the admin API port (same for all services).
	JoseJAAdminPort = 9090
)

// JOSE-JA Elastic Key Configuration.
const (
	// JoseJADefaultMaxMaterials is the default maximum number of material keys per elastic key.
	JoseJADefaultMaxMaterials = 10

	// JoseJAMinMaterials is the minimum number of material keys per elastic key.
	JoseJAMinMaterials = 1

	// JoseJAMaxMaterials is the maximum number of material keys per elastic key.
	JoseJAMaxMaterials = 100
)

// JOSE-JA Audit Configuration.
const (
	// JoseJAAuditDefaultEnabled is the default audit enabled state.
	JoseJAAuditDefaultEnabled = true

	// JoseJAAuditDefaultSamplingRate is the default audit sampling rate (100% = log all).
	JoseJAAuditDefaultSamplingRate = 100

	// JoseJAAuditFallbackSamplingRate is the fallback sampling rate when no config exists (1%).
	JoseJAAuditFallbackSamplingRate = 0.01

	// JoseJAAuditMinSamplingRate is the minimum audit sampling rate.
	JoseJAAuditMinSamplingRate = 0

	// JoseJAAuditMaxSamplingRate is the maximum audit sampling rate.
	JoseJAAuditMaxSamplingRate = 100
)

// JOSE-JA E2E Test Configuration.
const (
	// JoseJAE2ESQLitePublicPort is the public port for SQLite E2E tests.
	JoseJAE2ESQLitePublicPort = 9443

	// JoseJAE2EPostgreSQL1PublicPort is the public port for first PostgreSQL E2E tests.
	JoseJAE2EPostgreSQL1PublicPort = 9444

	// JoseJAE2EPostgreSQL2PublicPort is the public port for second PostgreSQL E2E tests.
	JoseJAE2EPostgreSQL2PublicPort = 9445

	// JoseJAE2EGrafanaPort is the Grafana port for E2E tests.
	JoseJAE2EGrafanaPort = 3000

	// JoseJAE2EOtelCollectorGRPCPort is the OpenTelemetry collector gRPC port for E2E tests.
	JoseJAE2EOtelCollectorGRPCPort = 4317

	// JoseJAE2EOtelCollectorHTTPPort is the OpenTelemetry collector HTTP port for E2E tests.
	JoseJAE2EOtelCollectorHTTPPort = 4318
)
