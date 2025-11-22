package magic

// Identity service scaling constants.
const (
	// IdentityScaling1x - Single instance scaling (demo, ci).
	IdentityScaling1x = 1
	// IdentityScaling2x - High availability scaling (development).
	IdentityScaling2x = 2
	// IdentityScaling3x - Production-like scaling (production).
	IdentityScaling3x = 3
)
