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
	contents["ca/"] = OptionalFileStatus
	contents["cipher/"] = OptionalFileStatus
	contents["cipher/im/"] = OptionalFileStatus
	contents["cryptoutil/"] = OptionalFileStatus
	contents["identity/"] = OptionalFileStatus
	contents["identity/policies/"] = OptionalFileStatus
	contents["identity/profiles/"] = OptionalFileStatus
	contents["jose/"] = OptionalFileStatus
	contents["jose/ja/"] = OptionalFileStatus
	contents["pki/"] = OptionalFileStatus
	contents["pki/ca/"] = OptionalFileStatus
	contents["sm/"] = OptionalFileStatus
	contents["sm/kms/"] = OptionalFileStatus

	// Future: add specific required files once config patterns are established
	// For now, configs/ validation is minimal to allow experimentation

	return contents
}
