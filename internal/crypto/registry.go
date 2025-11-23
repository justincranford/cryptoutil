package crypto

// HashSecret hashes a secret using PBKDF2-HMAC-SHA256 (FIPS-approved).
// FIPS mode is ALWAYS enabled - no configurable algorithm selection.
func HashSecret(secret string) (string, error) {
	return HashSecretPBKDF2(secret)
}
