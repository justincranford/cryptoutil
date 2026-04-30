// Copyright (c) 2025-2026 Justin Cranford.
package jobs

// boolPtr converts bool to *bool for struct literals requiring pointer fields.
func boolPtr(b bool) *bool {
	return &b
}
