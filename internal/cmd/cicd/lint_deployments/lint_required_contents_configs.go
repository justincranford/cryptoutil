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

	// Config directories (allowed but not strictly required at this time)
	// Future: add more specific validation once config patterns stabilize
	contents["cryptoutil/"] = OptionalFileStatus // suite: cryptoutil
	contents["identity/"] = OptionalFileStatus // product: identity
	contents["identity/authz"] = OptionalFileStatus // service: identity-authz
	contents["identity/idp"] = OptionalFileStatus // service: identity-idp
	contents["identity/rp"] = OptionalFileStatus // service: identity-rp
	contents["identity/rs"] = OptionalFileStatus // service: identity-rs
	contents["identity/spa"] = OptionalFileStatus // service: identity-spa
	contents["identity/policies/"] = OptionalFileStatus
	contents["identity/profiles/"] = OptionalFileStatus
	contents["jose/"] = OptionalFileStatus // product: jose
	contents["jose/ja/"] = OptionalFileStatus // service: jose-ja
	contents["pki/"] = OptionalFileStatus // product: pki
	contents["pki/ca/"] = OptionalFileStatus // service: pki-ca
	contents["sm/"] = OptionalFileStatus // product: sm
	contents["sm/im/"] = OptionalFileStatus // service: sm-im
	contents["sm/kms/"] = OptionalFileStatus // service: sm-kms

	// Future: add specific required files once config patterns are established
	// For now, configs/ validation is minimal to allow experimentation

	return contents
}
