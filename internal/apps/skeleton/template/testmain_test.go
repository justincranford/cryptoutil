// Copyright (c) 2025 Justin Cranford
//

package template

import (
"context"
"fmt"
http "net/http"
"os"
"testing"
"time"

cryptoutilAppsSkeletonTemplateServer "cryptoutil/internal/apps/skeleton/template/server"
cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton/template/server/config"
cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

var (
testSkeletonTemplateService *cryptoutilAppsSkeletonTemplateServer.SkeletonTemplateServer
sharedHTTPClient            *http.Client
publicBaseURL               string
adminBaseURL                string
)

func TestMain(m *testing.M) {
// Create in-memory SQLite configuration for testing.
cfg := cryptoutilAppsSkeletonTemplateServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

ctx := context.Background()

// Create server.
var err error

testSkeletonTemplateService, err = cryptoutilAppsSkeletonTemplateServer.NewFromConfig(ctx, cfg)
if err != nil {
panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
}

// Start server in background.
errChan := make(chan error, 1)

go func() {
if startErr := testSkeletonTemplateService.Start(ctx); startErr != nil {
errChan <- startErr
}
}()

// Wait for server ports to be assigned.
const (
pollTimeout  = 5 * time.Second
pollInterval = 100 * time.Millisecond
)

pollErr := cryptoutilSharedUtilPoll.Until(ctx, pollTimeout, pollInterval, func(_ context.Context) (bool, error) {
select {
case startErr := <-errChan:
return false, fmt.Errorf("server failed to start: %w", startErr)
default:
}

return testSkeletonTemplateService.PublicPort() > 0 && testSkeletonTemplateService.AdminPort() > 0, nil
})
if pollErr != nil {
panic(fmt.Sprintf("TestMain: %v", pollErr))
}

// Mark server as ready.
testSkeletonTemplateService.SetReady(true)

// Store base URLs for tests.
publicBaseURL = testSkeletonTemplateService.PublicBaseURL()
adminBaseURL = testSkeletonTemplateService.AdminBaseURL()

// Create shared HTTP client for all tests (accepts self-signed certs).
sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

// Run all tests.
exitCode := m.Run()

// Cleanup: Shutdown server.
shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout*time.Second)
defer cancel()

_ = testSkeletonTemplateService.Shutdown(shutdownCtx)

os.Exit(exitCode)
}
