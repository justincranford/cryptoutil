// Copyright (c) 2025-2026 Justin Cranford.
package e2e_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	postgresContinuityTestEnabled bool
	postgresContinuityContainer   *postgres.PostgresContainer
	postgresContinuityDSN         string
	postgresContinuitySetup       func(context.Context) (*postgres.PostgresContainer, string, error)
)

func TestMain(m *testing.M) {
	if !postgresContinuityTestEnabled {
		os.Exit(m.Run())
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 2*time.Minute)
	if postgresContinuitySetup == nil {
		cancelStartup()
		log.Printf("PostgreSQL continuity test setup is not registered")
		os.Exit(1)
	}

	var err error

	postgresContinuityContainer, postgresContinuityDSN, err = postgresContinuitySetup(startupCtx)
	if err != nil {
		cancelStartup()
		log.Printf("failed to set up PostgreSQL continuity test: %v", err)
		os.Exit(1)
	}

	cancelStartup()

	exitCode := m.Run()

	if err := postgresContinuityContainer.Terminate(context.Background()); err != nil {
		log.Printf("failed to terminate PostgreSQL continuity test container: %v", err)

		exitCode = 1
	}

	os.Exit(exitCode)
}
