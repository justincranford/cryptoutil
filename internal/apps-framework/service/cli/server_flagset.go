// Copyright (c) 2025-2026 Justin Cranford.
//

package cli

// ServerFlagSetName returns the canonical pflag FlagSet name for a service's server subcommand.
func ServerFlagSetName(serviceID string) string {
	return serviceID + "-server"
}
