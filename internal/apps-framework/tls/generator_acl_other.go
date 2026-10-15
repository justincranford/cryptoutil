//go:build !windows

package tls

func securePrivateFileACL(string) error {
	return nil
}

func validatePrivateFileACL(string) error {
	return nil
}
