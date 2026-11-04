// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"context"
	json "encoding/json"
	"errors"
	"reflect"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
)

// VerifySecret checks an existing complete desired import without create/apply.
// It requires Prepare for this exact graph and the UID recorded during import.
func (c *KubectlClient) VerifySecret(ctx context.Context, secret gen.LifecycleSecret) error {
	source, exists := c.imports[secret.Metadata.Namespace+"/"+secret.Metadata.Name]
	if !exists || source.Type != secret.Type || secret.Metadata.Labels["cu.io/owner"] != importOwner {
		return ErrCollision
	}

	if _, err := validateSecrets(gen.LifecycleChartValues{Run: c.run, Secrets: []gen.LifecycleSecretSource{source}}, []gen.LifecycleSecret{secret}); err != nil {
		return ErrCollision
	}

	id := Identity{Namespace: secret.Metadata.Namespace, Name: secret.Metadata.Name, Kind: secretKind, Run: c.run}
	raw, _ := json.Marshal(secret)

	var desired map[string]any

	_ = json.Unmarshal(raw, &desired)

	return c.verifyExisting(ctx, id, desired)
}

// VerifyJob rechecks revision, complete desired spec and unchanged ensured UID.
// It is separate from the Job condition poll and never creates a resource.
func (c *KubectlClient) VerifyJob(ctx context.Context, id Identity) error {
	desired := c.resources[identityKey(id)]
	if desired == nil {
		return ErrPrerequisite
	}

	return c.verifyExisting(ctx, id, desired)
}

func (c *KubectlClient) verifyExisting(ctx context.Context, id Identity, desired map[string]any) error {
	if c.uids[identityKey(id)] == "" {
		return ErrPrerequisite
	}

	actual, err := c.get(ctx, id)
	if errors.Is(err, errNotFound) {
		return ErrPrerequisite
	}

	if err != nil {
		return err
	}

	if text(objectMap(actual["metadata"])["uid"]) != c.uids[identityKey(id)] {
		return ErrReplaced
	}

	if !owns(actual, id) || !containsDesired(actual, desired) || !safeExistingPod(actual, desired) {
		return ErrCollision
	}

	if id.Kind == secretKind && (!reflect.DeepEqual(actual["data"], desired["data"]) || actual["stringData"] != nil) {
		return ErrCollision
	}

	return nil
}
