// Copyright (c) 2025-2026 Justin Cranford.
package magic

// Framework Service Infrastructure Configuration.
// The framework package (internal/apps-framework/) provides the shared service
// infrastructure used by all services.
const (
	// FrameworkProductName is the top-level directory component name of the service framework.
	// Used by fitness linters that check import isolation — the framework directory
	// is at internal/apps-framework/ (outside internal/apps/), so cross-service
	// isolation checks skip it by matching the "apps-framework" prefix separately.
	FrameworkProductName = "framework"

	// FrameworkInternalDir is the relative path to the framework package from the project root.
	// All services are allowed to import from this directory.
	FrameworkInternalDir = "internal/apps-framework"
)
