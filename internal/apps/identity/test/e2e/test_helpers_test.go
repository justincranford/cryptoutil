// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e

// boolPtr converts bool to *bool for struct literals requiring pointer fields.
func boolPtr(b bool) *bool {
	return &b
}
