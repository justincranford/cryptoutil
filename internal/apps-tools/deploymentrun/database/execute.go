// Copyright (c) 2025-2026 Justin Cranford.

package database

import (
	"context"
	"errors"
	"fmt"
)

// Client separates SQL effects from ordered lifecycle orchestration.
type Client interface {
	Close() error
	RequireDatabase(context.Context, string, string) error
	EnsureDatabase(context.Context, string, string) error
	EnsurePublication(context.Context, string) error
	RequirePublication(context.Context, string) error
	Migrate(context.Context, string, string) error
	PrepareReplica(context.Context, string, string) error
	Schema(context.Context) (string, error)
	EnsureSubscription(context.Context, Target, Client) error
}

type Open func(context.Context, Connection, string) (Client, error)

func Execute(ctx context.Context, operation string, input Input, open Open) (result error) {
	if ctx == nil {
		return errors.New("database lifecycle context required")
	}

	if err := input.Validate(operation); err != nil {
		return fmt.Errorf("database lifecycle contract: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("database lifecycle context: %w", err)
	}

	if open == nil {
		return errors.New("database opener required")
	}

	clients := []Client{}

	defer func() {
		if err := contextFailure(ctx); err != nil {
			result = errors.Join(result, err)
		}

		for _, client := range clients {
			if err := client.Close(); err != nil {
				result = errors.Join(result, errors.New("database closure rejected"))
			}
		}
	}()

	acquire := func(connection Connection) (Client, error) {
		client, err := open(ctx, connection, input.Backend)
		if err != nil {
			return nil, errors.New("database connection rejected")
		}

		if client == nil {
			return nil, errors.New("database connection unavailable")
		}

		clients = append(clients, client)

		return client, nil
	}
	if operation == Initialize {
		leader, err := acquire(input.LeaderAdmin)
		if err != nil {
			return err
		}

		follower, err := acquire(input.FollowerAdmin)
		if err != nil {
			return err
		}

		for _, target := range input.Targets {
			if err := leader.EnsureDatabase(ctx, target.Database, target.Role); err != nil {
				return errors.New("leader database initialization rejected")
			}

			if err := follower.EnsureDatabase(ctx, target.Database, target.Role); err != nil {
				return errors.New("follower database initialization rejected")
			}

			source, err := acquire(target.Leader)
			if err != nil {
				return err
			}

			if err := source.RequireDatabase(ctx, target.Database, target.Role); err != nil {
				return errors.New("publication database identity rejected")
			}

			if err := source.EnsurePublication(ctx, target.Publication); err != nil {
				return errors.New("publication initialization rejected")
			}
		}

		return contextFailure(ctx)
	}

	for _, target := range input.Targets {
		leader, err := acquire(target.Leader)
		if err != nil {
			return err
		}

		if err := leader.RequireDatabase(ctx, target.Database, target.Role); err != nil {
			return errors.New("leader database identity rejected")
		}

		if operation == Migrate {
			if err := leader.Migrate(ctx, target.PSID, input.Backend); err != nil {
				return errors.New("service migration rejected")
			}

			continue
		}

		follower, err := acquire(target.Follower)
		if err != nil {
			return err
		}

		if err := follower.RequireDatabase(ctx, target.Database, target.Role); err != nil {
			return errors.New("follower database identity rejected")
		}
		// Leader migrations are an external ordered Job barrier; never silently
		// migrate the leader in the replication phase.
		if err := leader.RequirePublication(ctx, target.Publication); err != nil {
			return errors.New("source publication rejected")
		}

		sourceSchema, err := leader.Schema(ctx)
		if err != nil {
			return errors.New("leader migrated schema unavailable")
		}

		if err := follower.PrepareReplica(ctx, target.PSID, sourceSchema); err != nil {
			return errors.New("follower schema bootstrap rejected")
		}

		if err := follower.EnsureSubscription(ctx, target, leader); err != nil {
			return errors.New("guarded subscription reconciliation rejected")
		}

		followerSchema, err := follower.Schema(ctx)
		if err != nil || sourceSchema == "" || sourceSchema != followerSchema {
			return errors.New("leader/follower schema parity rejected")
		}
	}

	return contextFailure(ctx)
}

func contextFailure(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("database lifecycle context: %w", err)
	}

	return nil
}
