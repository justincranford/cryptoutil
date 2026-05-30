// Copyright (c) 2025-2026 Justin Cranford.
//

package magic

// Legacy removed-service constants retained for test and tooling compatibility.
// Do not use these for new product/service registration.
var (
	OTLPServiceSMIM            = "sm-im"
	IMServiceID                = OTLPServiceSMIM
	IMProductName              = SMProductName
	IMServiceName              = "im"
	IMServicePort              = 8100
	IMAdminPort                = 9090
	IMPBKDF2Iterations         = 600000
	IME2EPostgreSQL1PublicPort = 8102
	IME2EPostgreSQL2PublicPort = 8103
	IME2ESQLiteContainer       = "sm-im-app-sqlite-1"
	IME2EPostgreSQL1Container  = "sm-im-app-postgresql-1"
	IME2EPostgreSQL2Container  = "sm-im-app-postgresql-2"
	IME2EHealthEndpoint        = "/service/api/v1/health"
	IMDisplayName              = "Instant Messenger"

	OTLPServiceJoseJA              = "sm-kms"
	JoseProductName                = "jose"
	JoseJAServiceID                = OTLPServiceJoseJA
	JoseJAServiceName              = "ja"
	JoseJAServicePort              = 8200
	JoseJAAdminPort                = 9090
	JoseJADisplayName              = "JWK Authority"
	JoseJAE2EGrafanaPort           = 3000
	JoseJAE2EOtelCollectorGRPCPort = 4317
	JoseJAE2EOtelCollectorHTTPPort = 4318
)
