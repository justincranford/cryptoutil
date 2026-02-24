// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
"testing"

testify "github.com/stretchr/testify/require"

cryptoutilIdentityJobs "cryptoutil/internal/apps/identity/jobs"
)

func TestNewServerManager(t *testing.T) {
t.Parallel()

tests := []struct {
name          string
authzServer   *AuthZServer
idpServer     *IDPServer
rsServer      *RSServer
cleanupJob    *cryptoutilIdentityJobs.CleanupJob
expectNonNil  bool
}{
{name: "all nil servers", authzServer: nil, idpServer: nil, rsServer: nil, cleanupJob: nil, expectNonNil: true},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

mgr := NewServerManager(tc.authzServer, tc.idpServer, tc.rsServer, tc.cleanupJob, nil)
if tc.expectNonNil {
testify.NotNil(t, mgr)
}
})
}
}

func TestServerManager_GetCleanupMetrics_NilJob(t *testing.T) {
t.Parallel()

mgr := NewServerManager(nil, nil, nil, nil, nil)
metrics := mgr.GetCleanupMetrics()
testify.Equal(t, cryptoutilIdentityJobs.CleanupJobMetrics{}, metrics)
}

func TestServerManager_IsCleanupHealthy_NilJob(t *testing.T) {
t.Parallel()

mgr := NewServerManager(nil, nil, nil, nil, nil)
testify.True(t, mgr.IsCleanupHealthy())
}
