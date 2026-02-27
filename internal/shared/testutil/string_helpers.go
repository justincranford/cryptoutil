// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

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
