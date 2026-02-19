// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"fmt"
"strings"
)

// DuplicateGroup holds a literal value shared by multiple magic constants.
type DuplicateGroup struct {
// Value is the raw Go literal shared by all constants in the group.
Value string

// Constants are the definitions that share this value.
Constants []MagicConstant
}

// DuplicatesResult holds the outcome of duplicate-value detection.
type DuplicatesResult struct {
// Valid is false when at least one duplicate group was found.
Valid bool

// Duplicates lists every group of constants that share the same literal value.
Duplicates []DuplicateGroup

// Errors lists file-system or parsing errors encountered during validation.
Errors []string
}

// ValidateDuplicates scans the magic package at magicDir and returns every
// group of constants that share an identical literal value.  Finding duplicates
// indicates that a value is registered under more than one name, which can hide
// the "canonical" constant from callers and causes confusion.
//
// Only constants whose value is a basic literal are considered; constants that
// reference another identifier (e.g. DefaultProfile = EmptyString) are skipped
// because their resolved value requires full type-checking.
func ValidateDuplicates(magicDir string) (*DuplicatesResult, error) {
result := &DuplicatesResult{Valid: true}

inv, err := parseMagicPackage(magicDir)
if err != nil {
result.Errors = append(result.Errors, fmt.Sprintf("parse error: %v", err))
result.Valid = false

return result, nil
}

for value, consts := range inv.ByValue {
if len(consts) < 2 {
continue
}

group := DuplicateGroup{
Value:     value,
Constants: append([]MagicConstant(nil), consts...),
}

result.Duplicates = append(result.Duplicates, group)
result.Valid = false
}

// Sort for deterministic output: by value string.
sortDuplicateGroups(result.Duplicates)

return result, nil
}

// sortDuplicateGroups sorts duplicate groups and their member constants
// deterministically so that output is stable across runs.
func sortDuplicateGroups(groups []DuplicateGroup) {
for i := range groups {
consts := groups[i].Constants
sortMagicConstants(consts)
}

// Sort groups by value.
for i := 1; i < len(groups); i++ {
for j := i; j > 0 && groups[j].Value < groups[j-1].Value; j-- {
groups[j], groups[j-1] = groups[j-1], groups[j]
}
}
}

// sortMagicConstants sorts a slice of MagicConstant by file then name.
func sortMagicConstants(consts []MagicConstant) {
for i := 1; i < len(consts); i++ {
for j := i; j > 0; j-- {
a, b := consts[j-1], consts[j]
if a.File > b.File || (a.File == b.File && a.Name > b.Name) {
consts[j], consts[j-1] = consts[j-1], consts[j]
} else {
break
}
}
}
}

// FormatDuplicatesResult formats the duplicates result into a human-readable
// report suitable for CI/CD output.
func FormatDuplicatesResult(result *DuplicatesResult) string {
var sb strings.Builder

if len(result.Errors) > 0 {
for _, e := range result.Errors {
fmt.Fprintf(&sb, "ERROR: %s\n", e)
}
}

if len(result.Duplicates) == 0 && len(result.Errors) == 0 {
fmt.Fprint(&sb, "validate-duplicates: OK (no duplicate magic values found)\n")

return sb.String()
}

fmt.Fprintf(&sb, "validate-duplicates: FAIL (%d duplicate value group(s) found)\n\n", len(result.Duplicates))

for _, group := range result.Duplicates {
fmt.Fprintf(&sb, "  Duplicate value %s is defined %d times:\n", group.Value, len(group.Constants))

for _, mc := range group.Constants {
fmt.Fprintf(&sb, "    %s:%d  %s\n", mc.File, mc.Line, mc.Name)
}

fmt.Fprint(&sb, "\n")
}

return sb.String()
}
