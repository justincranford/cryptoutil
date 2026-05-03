//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
package __SERVICE__

// NOTE: This template is documentation/scaffold only and is NOT currently enforced by lint-fitness.

// Lifecycle test pattern:
//   - server starts with dynamic ports
//   - livez/readyz transitions are validated
//   - graceful shutdown path returns cleanly
