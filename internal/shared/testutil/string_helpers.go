// Copyright (c) 2025-2026 Justin Cranford.
//
// SPDX-License-Identifier: AGPL-3.0-only
package testutil

import "strings"

// ContainsAny returns true if s contains any of the substrings.
func ContainsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}

	return false
}
