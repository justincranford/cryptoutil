// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

// boolPtr converts bool to *bool for struct literals requiring pointer fields.
func boolPtr(b bool) *bool {
	return &b
}
