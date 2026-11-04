// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

// Build constructs the graph for an optional exact selector using only the
// canonical typed registry.
func Build(model Model, selector string) (Inventory, error) {
	return build(model, RegisteredScopes(), selector)
}

// Plan builds a canonical JSON deployment inventory from a strict versioned
// model document and an optional exact release selector.
func Plan(modelJSON []byte, selector string) ([]byte, error) {
	model, err := DecodeModel(modelJSON)
	if err != nil {
		return nil, err
	}

	inventory, err := Build(model, selector)
	if err != nil {
		return nil, err
	}

	return marshalCanonicalJSON(inventory)
}
