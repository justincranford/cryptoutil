package lint_deployments

// GetExpectedConfigsContents returns the expected file structure for configs/.
// This is intentionally less strict than deployments/ to allow for configuration flexibility.
//
// Format: map[relativePathFromConfigsRoot]fileStatus where fileStatus is:
//   - RequiredFileStatus - file/directory MUST exist
//   - OptionalFileStatus - file/directory MAY exist
//
// See: docs/ARCHITECTURE.md Section 12.4 for config organization patterns.
func GetExpectedConfigsContents() map[string]string {
	contents := make(map[string]string)

	// Config directories — flat configs/{PS-ID}/ layout (Decision 2=B).
	// Suite config directory.
	contents["cryptoutil/"] = OptionalFileStatus // suite: cryptoutil

	// Product-service config directories (flat, no product nesting).
	contents["identity-authz/"] = OptionalFileStatus                    // PS-ID: identity-authz
	contents["identity-authz/domain/policies/"] = OptionalFileStatus    // authz policies
	contents["identity-idp/"] = OptionalFileStatus                      // PS-ID: identity-idp
	contents["identity-rp/"] = OptionalFileStatus                       // PS-ID: identity-rp
	contents["identity-rs/"] = OptionalFileStatus                       // PS-ID: identity-rs
	contents["identity-spa/"] = OptionalFileStatus                      // PS-ID: identity-spa
	contents["jose-ja/"] = OptionalFileStatus                           // PS-ID: jose-ja
	contents["pki-ca/"] = OptionalFileStatus                            // PS-ID: pki-ca
	contents["pki-ca/profiles/"] = OptionalFileStatus                   // certificate profiles
	contents["skeleton-template/"] = OptionalFileStatus                 // PS-ID: skeleton-template
	contents["sm-im/"] = OptionalFileStatus                             // PS-ID: sm-im
	contents["sm-kms/"] = OptionalFileStatus                            // PS-ID: sm-kms

	return contents
}
