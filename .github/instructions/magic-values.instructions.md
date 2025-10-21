---
description: "Instructions for defining magic numbers and values in dedicated constants package"
applyTo: "**"
---
# Magic Values Instructions

- Define all magic numbers and magic values in Go files within the `cryptoutilMagic` package
- Create named constants for commonly used values to avoid magic number linter violations
- Group related constants logically (buffer sizes, timeouts, percentages, ports, etc.)
- Use descriptive constant names that clearly indicate their purpose and units
- Update .golangci.yml importas configuration to include cryptoutilMagic alias
- Remove magic numbers from mnd ignored-numbers list once they are properly defined as constants
- Follow the established pattern of centralizing constants to improve code maintainability and eliminate linter bypasses
