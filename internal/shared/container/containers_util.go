// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package container provides utilities for managing testcontainers.
package container

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/testcontainers/testcontainers-go"
)

// StartContainer starts a testcontainer with the given request and returns it with a cleanup function.
func StartContainer(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, containerRequest testcontainers.ContainerRequest) (testcontainers.Container, func(), error) {
	return startContainer(ctx, telemetryService, containerRequest, testcontainers.GenericContainer)
}

func startContainer(
	ctx context.Context,
	telemetryService *cryptoutilSharedTelemetry.TelemetryService,
	containerRequest testcontainers.ContainerRequest,
	genericContainerFn func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error),
) (testcontainers.Container, func(), error) {
	telemetryService.Slogger.Debug("starting container")

	startedContainer, err := genericContainerFn(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})
	if err != nil {
		telemetryService.Slogger.Error("failed to start container", cryptoutilSharedMagic.StringError, err)

		startErr := fmt.Errorf("failed to start container: %w", err)
		if !isNilTestContainer(startedContainer) {
			startErr = cleanupTestContainerAfterError(ctx, startedContainer, startErr, "failed to clean up container after startup failure")
		}

		return nil, nil, startErr
	}

	if isNilTestContainer(startedContainer) {
		return nil, nil, errors.New("container startup returned nil container")
	}

	terminateContainer := func() {
		telemetryService.Slogger.Debug("terminating container")

		err := TerminateTestContainer(ctx, startedContainer)
		if err == nil {
			telemetryService.Slogger.Debug("successfully terminated container")
		} else {
			telemetryService.Slogger.Error("failed to terminate container", cryptoutilSharedMagic.StringError, err)
		}
	}

	return startedContainer, terminateContainer, nil
}

func cleanupTestContainerAfterError(ctx context.Context, container testcontainers.Container, operationErr error, cleanupErrorPrefix string) error {
	if isNilTestContainer(container) {
		return operationErr
	}

	if cleanupErr := TerminateTestContainer(ctx, container); cleanupErr != nil {
		return errors.Join(operationErr, fmt.Errorf("%s: %w", cleanupErrorPrefix, cleanupErr))
	}

	return operationErr
}

// TerminateTestContainer terminates a test container with a bounded cleanup context
// that is not cancelled with the caller's context.
func TerminateTestContainer(ctx context.Context, container testcontainers.Container) error {
	if isNilTestContainer(container) {
		return nil
	}

	cleanupParentContext := ctx
	if cleanupParentContext == nil {
		cleanupParentContext = context.Background()
	}

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(cleanupParentContext), cryptoutilSharedMagic.DBTestContainerCleanupTimeout)
	defer cancel()

	if err := container.Terminate(cleanupCtx); err != nil {
		return fmt.Errorf("failed to terminate test container: %w", err)
	}

	return nil
}

func isNilTestContainer(container testcontainers.Container) bool {
	if container == nil {
		return true
	}

	containerValue := reflect.ValueOf(container)

	return containerValue.Kind() == reflect.Pointer && containerValue.IsNil()
}

// GetContainerHostAndMappedPort returns the host and mapped port for a container.
func GetContainerHostAndMappedPort(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, container testcontainers.Container, port string) (string, string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		telemetryService.Slogger.Error("failed to get container host", cryptoutilSharedMagic.StringError, err)

		return "", "", fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		telemetryService.Slogger.Error("failed to get container mapped port", cryptoutilSharedMagic.StringError, err)

		return "", "", fmt.Errorf("failed to get container mapped port: %w", err)
	}

	return host, mappedPort.Port(), nil
}
