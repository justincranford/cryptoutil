// Copyright (c) 2025 Justin Cranford

// Package magic contains magic constants for OTel TLS E2E tests.
package magic

import "time"

// OTel TLS E2E Test Configuration (framework/tls/e2e package).
// These constants configure the TLS connectivity tests for OTel Collector
// using the sm-kms compose stack (includes pki-init + OTel Collector).
const (
	// OtelTLSE2EComposeFile is the sm-kms compose path relative to internal/apps/framework/tls/e2e/.
	// Levels: e2e→tls(1)→framework(2)→apps(3)→internal(4)→root(5), then deployments/sm-kms.
	OtelTLSE2EComposeFile = "../../../../../deployments/sm-kms/compose.yml"

	// OtelTLSE2EComposeOverrideFile exposes OTel OTLP ports to the host for TLS verification.
	// Loaded alongside OtelTLSE2EComposeFile so Go tests can dial gRPC :4317 and HTTP :4318.
	OtelTLSE2EComposeOverrideFile = "../../../../../deployments/sm-kms/compose-test-otel-expose.yml"

	// OtelTLSE2EHealthTimeout is the max wait for the OTel Collector to become ready.
	OtelTLSE2EHealthTimeout = 120 * time.Second

	// OtelTLSE2EGRPCPort is the host-side port used by the test compose override for OTel gRPC.
	// Must not conflict with other running stacks.
	OtelTLSE2EGRPCPort = 14317

	// OtelTLSE2EHTTPPort is the host-side port used by the test compose override for OTel HTTP.
	OtelTLSE2EHTTPPort = 14318

	// OtelTLSE2EHealthPort is the host-side port for the OTel health check endpoint.
	OtelTLSE2EHealthPort = 14133

	// OtelTLSE2ECACertPath is the Cat 1 public HTTPS server issuing CA truststore.
	// Used to verify the Cat 2 OTel Collector server cert in TLS handshakes.
	// Path is relative to internal/apps/framework/tls/e2e/.
	OtelTLSE2ECACertPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-server-issuing-ca/truststore/public-https-server-issuing-ca.crt"

	// OtelTLSE2EClientCertPath is the Cat 9 app client cert (sqlite-1) for mTLS to OTel.
	// Path is relative to internal/apps/framework/tls/e2e/.
	OtelTLSE2EClientCertPath = "../../../../../deployments/sm-kms/certs/sm-kms/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1.crt"

	// OtelTLSE2EClientKeyPath is the Cat 9 app client key (sqlite-1).
	// Path is relative to internal/apps/framework/tls/e2e/.
	OtelTLSE2EClientKeyPath = "../../../../../deployments/sm-kms/certs/sm-kms/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1.key"

	// OtelTLSE2EOtelServerCertCN is the expected Common Name of the Cat 2 OTel server cert.
	OtelTLSE2EOtelServerCertCN = "public-https-server-entity-otel-collector-contrib"

	// OtelTLSE2EContainer is the OTel Collector compose service name.
	OtelTLSE2EContainer = "opentelemetry-collector-contrib"
)
