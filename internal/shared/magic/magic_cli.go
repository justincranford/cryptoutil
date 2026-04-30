// Copyright (c) 2025-2026 Justin Cranford.
//
//

package magic

const (
	// CLIHelpCommand is the help subcommand name.
	CLIHelpCommand = "help"
	// CLIHelpFlag is the long-form help flag.
	CLIHelpFlag = "--help"
	// CLIHelpShortFlag is the short-form help flag.
	CLIHelpShortFlag = "-h"
	// CLIURLFlag is the URL flag for health/client commands.
	CLIURLFlag = "--url"
	// CLICACertFlag is the CA certificate flag for health/client commands.
	CLICACertFlag = "--cacert"
	// CLICertFlag is the client certificate flag for mTLS client authentication.
	CLICertFlag = "--cert"
	// CLIKeyFlag is the client private key flag for mTLS client authentication.
	CLIKeyFlag = "--key"
	// CLIVersionCommand is the version subcommand name.
	CLIVersionCommand = "version"
	// CLIVersionFlag is the long-form version flag.
	CLIVersionFlag = "--version"
	// CLIVersionShortFlag is the short-form version flag.
	CLIVersionShortFlag = "-v"
	// CLIValidateSecretsCommand is the validate-secrets subcommand name.
	CLIValidateSecretsCommand = "validate-secrets"
	// DockerSecretsDir is the standard Docker secrets mount path.
	DockerSecretsDir = "/run/secrets"
	// DockerSecretMinLength is the minimum acceptable character length for high-entropy secret files.
	DockerSecretMinLength = 43
)
