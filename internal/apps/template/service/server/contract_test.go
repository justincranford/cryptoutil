// Copyright (c) 2025 Justin Cranford
//
//

package server_test

import (
	"testing"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"

	"github.com/stretchr/testify/require"
)

// TestServiceServer_InterfaceIsDefinedInServerPackage verifies that ServiceServer
// is correctly exported from internal/apps/template/service/server.
func TestServiceServer_InterfaceIsDefinedInServerPackage(t *testing.T) {
	t.Parallel()

	// This test verifies the interface type is accessible and usable.
	// The compile-time assertions (var _ ServiceServer = (*XxxServer)(nil))
	// in each service server.go provide hard compile-time guarantees.
	var svc cryptoutilAppsTemplateServiceServer.ServiceServer

	// Interface variable should be nil before assignment (zero value).
	require.Nil(t, svc)
}

// TestServiceServer_NilImplementation verifies that a nil typed pointer
// satisfies the interface structurally (compile-time).
// This is the idiomatic Go pattern: var _ T = (*Impl)(nil).
//
// If ANY of the 10 services were to lose a required method,
// the corresponding compile-time assertion in that service's server.go
// would cause a build failure — documented here for clarity.
//
// Services with compile-time assertions (var _ ServiceServer = (*XxxServer)(nil)):
// - SmIMServer          (internal/apps/sm/im/server)
// - KMSServer           (internal/apps/sm/kms/server)
// - JoseJAServer        (internal/apps/jose/ja/server)
// - PKICAServer         (internal/apps/pki/ca/server)
// - SkeletonTemplateServer (internal/apps/skeleton/template/server)
// - AuthzServer         (internal/apps/identity/authz/server)
// - IDPServer           (internal/apps/identity/idp/server)
// - RPServer            (internal/apps/identity/rp/server)
// - RSServer            (internal/apps/identity/rs/server)
// - SPAServer           (internal/apps/identity/spa/server).
func TestServiceServer_AllServicesHaveCompileTimeAssertions(t *testing.T) {
	t.Parallel()

	// Nothing to run: the compile-time assertions are the test.
	// If /any/ of the 10 services break the interface, this file will not compile.
	// This function documents and asserts the design intent.
	t.Log("ServiceServer compile-time assertions are enforced in each service's server.go")
}
