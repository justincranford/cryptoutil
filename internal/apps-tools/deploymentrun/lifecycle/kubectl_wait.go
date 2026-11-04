// Copyright (c) 2025-2026 Justin Cranford.
package lifecycle

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"fmt"
	"time"
)

func (c *KubectlClient) WaitJob(ctx context.Context, id Identity) error {
	if c.uids[identityKey(id)] == "" {
		return errors.New("job was not ensured")
	}

	return c.poll(ctx, id, true)
}

func (c *KubectlClient) WaitDatabase(ctx context.Context, id Identity) error {
	return c.poll(ctx, id, false)
}

func (c *KubectlClient) poll(ctx context.Context, id Identity, job bool) error {
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("lifecycle wait: %w", ctx.Err())
		}

		document, err := c.get(ctx, id)
		if errors.Is(err, errNotFound) {
			return ErrPrerequisite
		}

		if err != nil {
			return err
		}

		if !owns(document, id) {
			return ErrPrerequisite
		}

		uid := text(objectMap(document["metadata"])["uid"])
		if uid == "" || (c.uids[identityKey(id)] != "" && c.uids[identityKey(id)] != uid) {
			return ErrReplaced
		}

		c.uids[identityKey(id)] = uid

		status := objectMap(document[cryptoutilSharedMagic.StringStatus])
		if job {
			conditions, _ := status["conditions"].([]any)
			complete := false

			for _, condition := range conditions {
				value := objectMap(condition)
				if value[cryptoutilSharedMagic.StringStatus] == "True" {
					if value["type"] == "Failed" {
						return ErrJobFailed
					}

					complete = complete || value["type"] == "Complete"
				}
			}

			if complete {
				return nil
			}
		} else if databaseReady(document) {
			return nil
		}

		timer := time.NewTimer(c.config.PollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()

			return fmt.Errorf("lifecycle wait: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

func databaseReady(document map[string]any) bool {
	spec := objectMap(document["spec"])
	status := objectMap(document[cryptoutilSharedMagic.StringStatus])
	meta := objectMap(document["metadata"])
	replicas, _ := spec["replicas"].(float64)
	generation, _ := meta["generation"].(float64)
	observed, _ := status["observedGeneration"].(float64)

	return replicas > 0 && generation > 0 && observed >= generation &&
		status["readyReplicas"] == replicas && status["updatedReplicas"] == replicas &&
		text(status["currentRevision"]) != "" && status["currentRevision"] == status["updateRevision"]
}
