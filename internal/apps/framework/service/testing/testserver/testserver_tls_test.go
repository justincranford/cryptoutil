// Copyright (c) 2025 Justin Cranford
//
//

package testserver_test

import (
"crypto/x509"
"testing"

"github.com/stretchr/testify/require"

cryptoutilTestingTestserver "cryptoutil/internal/apps/framework/service/testing/testserver"
)

// nilPoolServer embeds mockServer but returns nil from TLSRootCAPool for error-path testing.
type nilPoolServer struct {
*mockServer
}

func (n *nilPoolServer) TLSRootCAPool() *x509.CertPool {
return nil
}

func TestNewTestTLSBundle_Success(t *testing.T) {
t.Parallel()

srv := newMockServer()

bundle := cryptoutilTestingTestserver.NewTestTLSBundle(t, srv)

require.NotNil(t, bundle)
}

func TestNewTestTLSBundle_NilServer(t *testing.T) {
t.Parallel()

ctb := &captureTB{T: t}
result := cryptoutilTestingTestserver.NewTestTLSBundle(ctb, nil)

require.True(t, ctb.hasFatal(), "Fatalf should be called with nil server")
require.Nil(t, result)
}

func TestNewTestTLSBundle_NilPool(t *testing.T) {
t.Parallel()

srv := &nilPoolServer{mockServer: newMockServer()}

ctb := &captureTB{T: t}
result := cryptoutilTestingTestserver.NewTestTLSBundle(ctb, srv)

require.True(t, ctb.hasFatal(), "Fatalf should be called when TLSRootCAPool() is nil")
require.Nil(t, result)
}

func TestTLSClientConfig_Success(t *testing.T) {
t.Parallel()

pool := x509.NewCertPool()
srv := newMockServer()
_ = pool

bundle := cryptoutilTestingTestserver.NewTestTLSBundle(t, srv)
require.NotNil(t, bundle)

cfg := cryptoutilTestingTestserver.TLSClientConfig(t, bundle)

require.NotNil(t, cfg)
require.Equal(t, uint16(0x0304), cfg.MinVersion, "MinVersion must be TLS 1.3")
require.NotNil(t, cfg.RootCAs)
require.False(t, cfg.InsecureSkipVerify, "InsecureSkipVerify must be false")
}

func TestTLSClientConfig_NilBundle(t *testing.T) {
t.Parallel()

ctb := &captureTB{T: t}
result := cryptoutilTestingTestserver.TLSClientConfig(ctb, nil)

require.True(t, ctb.hasFatal(), "Fatalf should be called with nil bundle")
require.Nil(t, result)
}
