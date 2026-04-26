// Copyright (c) 2025 Justin Cranford

// Package magic contains magic constants for OTel TLS E2E tests.
package magic

import "time"

// OTel TLS E2E Test Configuration (framework/tls/e2e package).
// These constants configure the TLS connectivity tests for OTel Collector
// using the sm-kms compose stack (includes pki-init + OTel Collector).
const (
	// OtelTLSE2EComposeFile is the sm-kms compose path relative to internal/apps-framework/tls/e2e/.
	// Levels: e2e→tls(1)→framework(2)→apps(3)→internal(4)→root(5), then deployments/sm-kms.
	OtelTLSE2EComposeFile = "../../../../../deployments/sm-kms/compose.yml"

	// OtelTLSE2EComposeOverrideFile exposes OTel OTLP ports to the host for TLS verification.
	// Loaded alongside OtelTLSE2EComposeFile so Go tests can dial gRPC :4317 and HTTP :4318.
	// Uses ports 24317/24318/24133 (offset +10000 from Grafana 14317/14318) to avoid conflict
	// when Grafana and OTel containers both run in the full-pipeline stack.
	OtelTLSE2EComposeOverrideFile = "../../../../../deployments/sm-kms/compose-test-otel-expose.yml"

	// OtelTLSE2EHealthTimeout is the max wait for the OTel Collector to become ready.
	OtelTLSE2EHealthTimeout = 120 * time.Second

	// OtelTLSE2EGRPCPort is the host-side port used by the test compose override for OTel gRPC.
	// Port 24317 (not 14317) to avoid conflict with Grafana's OTLP gRPC binding 14317:4317.
	OtelTLSE2EGRPCPort = 24317

	// OtelTLSE2EHTTPPort is the host-side port used by the test compose override for OTel HTTP.
	// Port 24318 (not 14318) to avoid conflict with Grafana's OTLP HTTP binding 14318:4318.
	OtelTLSE2EHTTPPort = 24318

	// OtelTLSE2EHealthPort is the host-side port for the OTel health check endpoint.
	// Port 24133 (not 14133) to avoid conflict with Grafana bindings in the full-pipeline stack.
	OtelTLSE2EHealthPort = 24133

	// OtelTLSE2ECACertPath is the Cat 1 public HTTPS server issuing CA truststore.
	// Used to verify the Cat 2 OTel Collector server cert in TLS handshakes.
	// Path is relative to internal/apps-framework/tls/e2e/.
	OtelTLSE2ECACertPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-server-issuing-ca/truststore/public-https-server-issuing-ca.crt"

	// OtelTLSE2EClientCertPath is the Cat 9 app client cert (sqlite-1) for mTLS to OTel.
	// Path is relative to internal/apps-framework/tls/e2e/.
	OtelTLSE2EClientCertPath = "../../../../../deployments/sm-kms/certs/sm-kms/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1.crt"

	// OtelTLSE2EClientKeyPath is the Cat 9 app client key (sqlite-1).
	// Path is relative to internal/apps-framework/tls/e2e/.
	OtelTLSE2EClientKeyPath = "../../../../../deployments/sm-kms/certs/sm-kms/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1/otel-collector-contrib-https-client-entity-sm-kms-sqlite-1.key"

	// OtelTLSE2EOtelServerCertCN is the expected Common Name of the Cat 2 OTel server cert.
	OtelTLSE2EOtelServerCertCN = "public-https-server-entity-otel-collector-contrib"

	// OtelTLSE2EContainer is the OTel Collector compose service name.
	OtelTLSE2EContainer = "opentelemetry-collector-contrib"

	// GrafanaTLSE2EUIPort is the host-side HTTPS port for the Grafana web UI.
	// Exposed in shared-telemetry/compose.yml as "3000:3000".
	GrafanaTLSE2EUIPort = 3000

	// GrafanaTLSE2EOTLPGRPCPort is the host-side OTLP gRPC port for Grafana mTLS ingest.
	// Exposed in shared-telemetry/compose.yml as "14317:4317".
	GrafanaTLSE2EOTLPGRPCPort = 14317

	// GrafanaTLSE2EServerCertCN is the expected Common Name of the Cat 2 Grafana server cert.
	GrafanaTLSE2EServerCertCN = "public-https-server-entity-grafana-otel-lgtm"

	// GrafanaTLSE2ECACertPath is the Cat 1 public HTTPS server issuing CA truststore.
	// Used to verify the Cat 2 Grafana server cert in TLS handshakes.
	// Path is relative to internal/apps-framework/tls/e2e/.
	GrafanaTLSE2ECACertPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-server-issuing-ca/truststore/public-https-server-issuing-ca.crt"

	// GrafanaTLSE2EInfraCertPath is the Cat 9 infra client cert (OTel→Grafana) for mTLS.
	// Path is relative to internal/apps-framework/tls/e2e/.
	GrafanaTLSE2EInfraCertPath = "../../../../../deployments/sm-kms/certs/sm-kms/otel-collector-contrib-https-client-entity-infra/otel-collector-contrib-https-client-entity-infra.crt"

	// GrafanaTLSE2EInfraKeyPath is the Cat 9 infra client key (OTel→Grafana).
	// Path is relative to internal/apps-framework/tls/e2e/.
	GrafanaTLSE2EInfraKeyPath = "../../../../../deployments/sm-kms/certs/sm-kms/otel-collector-contrib-https-client-entity-infra/otel-collector-contrib-https-client-entity-infra.key"

	// GrafanaTLSE2EHealthTimeout is the max wait for Grafana to become ready.
	GrafanaTLSE2EHealthTimeout = 180 * time.Second

	// GrafanaTLSE2EContainer is the Grafana compose service name.
	GrafanaTLSE2EContainer = "grafana-otel-lgtm"

	// FullPipelineTLSE2ETimeout is the max wait for the full pipeline stack (all services).
	// Longer than OTel/Grafana timeouts because PostgreSQL startup adds ~30s.
	FullPipelineTLSE2ETimeout = 300 * time.Second

	// App public HTTPS ports — bound in deployments/sm-kms/compose.yml as {HOST}:8080.
	// Tests dial these from the host to verify Cat 3 server certs and Cat 4 mTLS enforcement.

	// AppSMKMSSQLite1PublicPort is the host port for sm-kms-app-sqlite-1 public HTTPS.
	AppSMKMSSQLite1PublicPort = 8000

	// AppSMKMSSQLite2PublicPort is the host port for sm-kms-app-sqlite-2 public HTTPS.
	AppSMKMSSQLite2PublicPort = 8001

	// AppSMKMSPostgres1PublicPort is the host port for sm-kms-app-postgresql-1 public HTTPS.
	AppSMKMSPostgres1PublicPort = 8002

	// AppSMKMSPostgres2PublicPort is the host port for sm-kms-app-postgresql-2 public HTTPS.
	AppSMKMSPostgres2PublicPort = 8003

	// Cat 3 server cert CNs — the expected Common Names of the public HTTPS server certs per variant.
	// These are the CNs the app presents during TLS handshake on the public port.

	// AppSMKMSSQLite1ServerCertCN is the Cat 3 server cert CN for sm-kms-app-sqlite-1.
	AppSMKMSSQLite1ServerCertCN = "public-https-server-entity-sm-kms-sqlite-1"

	// AppSMKMSSQLite2ServerCertCN is the Cat 3 server cert CN for sm-kms-app-sqlite-2.
	AppSMKMSSQLite2ServerCertCN = "public-https-server-entity-sm-kms-sqlite-2"

	// AppSMKMSPostgres1ServerCertCN is the Cat 3 server cert CN for sm-kms-app-postgresql-1.
	AppSMKMSPostgres1ServerCertCN = "public-https-server-entity-sm-kms-postgres-1"

	// AppSMKMSPostgres2ServerCertCN is the Cat 3 server cert CN for sm-kms-app-postgresql-2.
	AppSMKMSPostgres2ServerCertCN = "public-https-server-entity-sm-kms-postgres-2"

	// App public HTTPS client cert paths (Cat 5 service-user certs) for mTLS.
	// These certs are signed by the Cat 4 CA (public-https-client-issuing-ca-sm-kms-{variant})
	// and are used by test clients to satisfy the server's RequireAndVerifyClientCert policy.
	// Paths are relative to internal/apps-framework/tls/e2e/.

	// AppSMKMSSQLite1ClientCertPath is the Cat 5 service-user client cert for sqlite-1.
	AppSMKMSSQLite1ClientCertPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-client-entity-sm-kms-sqlite-1-serviceuser-db/public-https-client-entity-sm-kms-sqlite-1-serviceuser-db.crt"

	// AppSMKMSSQLite1ClientKeyPath is the Cat 5 service-user client key for sqlite-1.
	AppSMKMSSQLite1ClientKeyPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-client-entity-sm-kms-sqlite-1-serviceuser-db/public-https-client-entity-sm-kms-sqlite-1-serviceuser-db.key"

	// AppSMKMSSQLite2ClientCertPath is the Cat 5 service-user client cert for sqlite-2.
	AppSMKMSSQLite2ClientCertPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-client-entity-sm-kms-sqlite-2-serviceuser-db/public-https-client-entity-sm-kms-sqlite-2-serviceuser-db.crt"

	// AppSMKMSSQLite2ClientKeyPath is the Cat 5 service-user client key for sqlite-2.
	AppSMKMSSQLite2ClientKeyPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-client-entity-sm-kms-sqlite-2-serviceuser-db/public-https-client-entity-sm-kms-sqlite-2-serviceuser-db.key"

	// AppSMKMSPostgresClientCertPath is the Cat 5 service-user client cert for postgres variants.
	// Postgres-1 and postgres-2 share the same Cat 4 issuing CA (public-https-client-issuing-ca-sm-kms-postgres).
	AppSMKMSPostgresClientCertPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-client-entity-sm-kms-postgres-serviceuser-db/public-https-client-entity-sm-kms-postgres-serviceuser-db.crt"

	// AppSMKMSPostgresClientKeyPath is the Cat 5 service-user client key for postgres variants.
	AppSMKMSPostgresClientKeyPath = "../../../../../deployments/sm-kms/certs/sm-kms/public-https-client-entity-sm-kms-postgres-serviceuser-db/public-https-client-entity-sm-kms-postgres-serviceuser-db.key"

	// Compose service names for sm-kms app variants (full pipeline stack).

	// AppSMKMSSQLite1Container is the compose service name for sm-kms-app-sqlite-1.
	AppSMKMSSQLite1Container = "sm-kms-app-sqlite-1"

	// AppSMKMSSQLite2Container is the compose service name for sm-kms-app-sqlite-2.
	AppSMKMSSQLite2Container = "sm-kms-app-sqlite-2"

	// AppSMKMSPostgres1Container is the compose service name for sm-kms-app-postgresql-1.
	AppSMKMSPostgres1Container = "sm-kms-app-postgresql-1"

	// AppSMKMSPostgres2Container is the compose service name for sm-kms-app-postgresql-2.
	AppSMKMSPostgres2Container = "sm-kms-app-postgresql-2"
)
