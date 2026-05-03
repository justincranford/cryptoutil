//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
package __SERVICE__

// NOTE: This template is documentation/scaffold only and is NOT currently enforced by lint-fitness.

// Port conflict test pattern:
//   - reserve public/admin ports
//   - assert startup fails with address-in-use error
//   - release listeners and verify recovery path
